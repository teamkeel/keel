package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/graphql/gqlerrors"
	"github.com/teamkeel/keel/casing"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Response struct {
	Status  int
	Body    []byte
	Headers map[string][]string
}

type ResponseMetadata struct {
	Headers http.Header
	Status  int
}

func NewJsonResponse(status int, body any, meta *ResponseMetadata) Response {
	b, _ := json.Marshal(body)

	r := Response{
		Status:  status,
		Body:    b,
		Headers: map[string][]string{},
	}

	if meta != nil {
		r.Headers = meta.Headers

		if meta.Status != 0 {
			r.Status = meta.Status
		}
	}

	// Content-Type must only be a single value
	// https://www.rfc-editor.org/rfc/rfc7230#section-3.2.2
	r.Headers["Content-Type"] = []string{"application/json"}

	return r
}

func NewRedirectResponse(url *url.URL) Response {
	// This response code means that the URI of requested resource has been changed temporarily.
	// Further changes in the URI might be made in the future.
	code := http.StatusFound

	// The body content for a 302 usually contains a short hypertext note with a hyperlink to the different URI(s).
	// https://www.rfc-editor.org/rfc/rfc9110.html#name-302-found
	return Response{
		Status: code,
		Body:   []byte("<a href=\"" + url.String() + "\">" + http.StatusText(code) + "</a>.\n"),
		Headers: map[string][]string{
			"Location":     {url.String()},
			"Content-Type": {"text/html; charset=utf-8"},
		},
	}
}

func InternalServerErrorResponse(ctx context.Context, err error) Response {
	span := trace.SpanFromContext(ctx)

	span.RecordError(err, trace.WithStackTrace(true))
	span.SetStatus(codes.Error, err.Error())

	return NewJsonResponse(http.StatusInternalServerError, nil, nil)
}

type HandlerFunc func(r *http.Request) Response

const (
	// An unexpected internal error happened.
	ErrInternal = "ERR_INTERNAL"
	// The input arguments provided are not valid.
	ErrInvalidInput = "ERR_INVALID_INPUT"
	// The input provided is malformed and cannot be parsed.
	ErrInputMalformed = "ERR_INPUT_MALFORMED"
	// Authentication failed when trying to identify the identity.
	ErrAuthenticationFailed = "ERR_AUTHENTICATION_FAILED"
	// Permission denied when trying to access some resource.
	ErrPermissionDenied = "ERR_PERMISSION_DENIED"
	// Record cannot be found with the provided parameters.
	ErrRecordNotFound = "ERR_RECORD_NOT_FOUND"
	// The path or action does not exist.
	ErrMethodNotFound = "ERR_ACTION_NOT_FOUND"
	// The HTTP method is not allowed for this request.
	ErrHttpMethodNotAllowed = "ERR_HTTP_METHOD_NOT_ALLOWED"
	// An unexpected error happened from user code.
	ErrUnknown = "ERR_UNKNOWN"
)

type PermissionStatus string

const (
	PermissionGranted PermissionStatus = "granted"
	PermissionUnknown PermissionStatus = "unknown"
)

type PermissionState struct {
	Status PermissionStatus `json:"status"`
}

func NewPermissionState() *PermissionState {
	return &PermissionState{
		Status: PermissionUnknown,
	}
}

func (ps *PermissionState) Grant() {
	ps.Status = PermissionGranted
}

type RuntimeError struct {
	Code    string
	Message string
}

var _ gqlerrors.ExtendedError = RuntimeError{}

func (r RuntimeError) Extensions() map[string]any {
	return map[string]any{
		"code":    r.Code,
		"message": r.Message,
	}
}

func (r RuntimeError) Error() string {
	return r.Message
}

func NewValidationError(message string) RuntimeError {
	return RuntimeError{
		Code:    ErrInvalidInput,
		Message: message,
	}
}

func NewNotFoundError(message string) RuntimeError {
	if message == "" {
		message = "record not found"
	}
	return RuntimeError{
		Code:    ErrRecordNotFound,
		Message: message,
	}
}

func NewMethodNotFoundError() RuntimeError {
	return RuntimeError{
		Code:    ErrMethodNotFound,
		Message: "method not found",
	}
}

func NewHttpMethodNotAllowedError(message string) RuntimeError {
	return RuntimeError{
		Code:    ErrHttpMethodNotAllowed,
		Message: message,
	}
}

func NewInputMalformedError(message string) RuntimeError {
	return RuntimeError{
		Code:    ErrInputMalformed,
		Message: message,
	}
}

func NewNotNullError(column string) RuntimeError {
	// Parses from the database casing back to the schema casing.
	// Important since these error messages are delivered to the user.
	field := casing.ToLowerCamel(column)

	return RuntimeError{
		Code:    ErrInvalidInput,
		Message: fmt.Sprintf("field '%s' cannot be null", field),
	}
}

func NewUniquenessError(columns []string) RuntimeError {
	columns = lo.Map(columns, func(c string, _ int) string {
		// Parses from the database casing back to the schema casing.
		// Important since these error messages are delivered to the user.
		return casing.ToLowerCamel(c)
	})

	var message string
	if len(columns) == 1 {
		message = fmt.Sprintf("the value for the unique field '%s' must be unique", columns[0])
	} else {
		message = fmt.Sprintf("the values for the unique composite fields (%s) must be unique", strings.Join(columns, ", "))
	}

	return RuntimeError{
		Code:    ErrInvalidInput,
		Message: message,
	}
}

func NewForeignKeyConstraintError(column string) RuntimeError {
	// Parses from the database casing back to the schema casing.
	// Important since these error messages are delivered to the user.
	field := casing.ToLowerCamel(column)

	return RuntimeError{
		Code:    ErrInvalidInput,
		Message: fmt.Sprintf("the record referenced in field '%s' does not exist", field),
	}
}

func NewPermissionError() RuntimeError {
	return RuntimeError{
		Code:    ErrPermissionDenied,
		Message: "not authorized to access",
	}
}

func NewAuthenticationFailedErr() RuntimeError {
	return RuntimeError{
		Code:    ErrAuthenticationFailed,
		Message: "authentication failed",
	}
}

func NewAuthenticationFailedMessageErr(message string) RuntimeError {
	return RuntimeError{
		Code:    ErrAuthenticationFailed,
		Message: message,
	}
}

// ParseQueryParams will parse the parmeters in the request query string.
func ParseQueryParams(r *http.Request) map[string]any {
	q := r.URL.Query()
	inputs := map[string]any{}
	for k := range q {
		if len(q[k]) > 1 {
			inputs[k] = q[k]
		} else {
			inputs[k] = q.Get(k)
		}
	}
	return inputs
}

// ParseRequestData will parse the request based on the Content-Type header.
// Defaults to parsing as a JSON request body.
func ParseRequestData(r *http.Request) (any, error) {
	switch {
	case HasContentType(r.Header, "application/x-www-form-urlencoded"):
		return parseFormUrlEncoded(r)
	case HasContentType(r.Header, "application/json"):
		return parseJsonBody(r)
	default:
		return parseJsonBody(r)
	}
}

func parseFormUrlEncoded(r *http.Request) (any, error) {
	data := map[string]any{}
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	for k, v := range r.Form {
		data[k] = v[0]
	}

	return data, nil
}

func parseJsonBody(r *http.Request) (data any, err error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if string(body) == "" {
		return map[string]any{}, nil
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func HasContentType(headers http.Header, mimetype string) bool {
	contentType := headers.Get("Content-Type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}

		if t == mimetype {
			return true
		}
	}
	return false
}
