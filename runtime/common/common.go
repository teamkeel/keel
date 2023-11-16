package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

func NewJsonResponse(status int, body any, headers map[string][]string) Response {
	b, _ := json.Marshal(body)

	if headers == nil {
		headers = map[string][]string{}
	}

	// Content-Type must only be a single value
	// https://www.rfc-editor.org/rfc/rfc7230#section-3.2.2
	headers["Content-Type"] = []string{"application/json"}

	return Response{
		Status:  status,
		Body:    b,
		Headers: headers,
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

func NewNotFoundError() RuntimeError {
	return RuntimeError{
		Code:    ErrRecordNotFound,
		Message: "record not found",
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
		Message: "not authorized to access this action",
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
