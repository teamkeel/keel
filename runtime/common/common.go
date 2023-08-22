package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/graphql/gqlerrors"
	"github.com/teamkeel/keel/casing"
)

type Response struct {
	Status  int
	Body    []byte
	Headers map[string][]string
}

type HttpJsonErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

func NewJsonResponse(status int, body any, headers map[string][]string) Response {
	b, _ := json.Marshal(body)
	return Response{
		Status:  status,
		Body:    b,
		Headers: headers,
	}
}

func NewJsonErrorResponse(err error) Response {
	code := "ERR_INTERNAL"
	message := "error executing request"
	httpCode := http.StatusInternalServerError

	var runtimeErr RuntimeError
	if errors.As(err, &runtimeErr) {
		code = runtimeErr.Code
		message = runtimeErr.Message

		switch code {
		case ErrInvalidInput:
			httpCode = http.StatusBadRequest
		case ErrRecordNotFound:
			httpCode = http.StatusNotFound
		case ErrPermissionDenied:
			httpCode = http.StatusForbidden
		case ErrAuthenticationFailed:
			httpCode = http.StatusUnauthorized
		}
	}

	return NewJsonResponse(httpCode, HttpJsonErrorResponse{
		Code:    code,
		Message: message,
	}, nil)
}

type ApiHandlerFunc func(r *http.Request) Response

const (
	ErrInternal             = "ERR_INTERNAL"
	ErrInvalidInput         = "ERR_INVALID_INPUT"
	ErrPermissionDenied     = "ERR_PERMISSION_DENIED"
	ErrRecordNotFound       = "ERR_RECORD_NOT_FOUND"
	ErrAuthenticationFailed = "ERR_AUTHENTICATION_FAILED"
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

func NewNotFoundError() RuntimeError {
	return RuntimeError{
		Code:    ErrRecordNotFound,
		Message: "record not found",
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
		Message: fmt.Sprintf("the relationship lookup for field '%s' does not exist", field),
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
		Message: "not authenticated",
	}
}
