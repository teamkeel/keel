package common

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/iancoleman/strcase"
)

type Response struct {
	Status  int
	Body    []byte
	Headers map[string][]string
}

func NewJsonResponse(status int, body any, headers map[string][]string) Response {
	b, _ := json.Marshal(body)
	return Response{
		Status:  status,
		Body:    b,
		Headers: headers,
	}
}

type ApiHandlerFunc func(r *http.Request) Response

const (
	ErrInternal         = "ERR_INTERNAL"
	ErrInvalidInput     = "ERR_INVALID_INPUT"
	ErrPermissionDenied = "ERR_PERMISSION_DENIED"
	ErrRecordNotFound   = "ERR_RECORD_NOT_FOUND"
)

type RuntimeError struct {
	Code    string
	Message string
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
	field := strcase.ToLowerCamel(column)

	return RuntimeError{
		Code:    ErrInvalidInput,
		Message: fmt.Sprintf("field '%s' cannot be null", field),
	}
}

func NewUniquenessError(column string) RuntimeError {
	// Parses from the database casing back to the schema casing.
	// Important since these error messages are delivered to the user.
	field := strcase.ToLowerCamel(column)

	return RuntimeError{
		Code:    ErrInvalidInput,
		Message: fmt.Sprintf("field '%s' can only contain unique values", field),
	}
}

func NewForeignKeyConstraintError(column string) RuntimeError {
	// Parses from the database casing back to the schema casing.
	// Important since these error messages are delivered to the user.
	field := strcase.ToLowerCamel(column)

	return RuntimeError{
		Code:    ErrInvalidInput,
		Message: fmt.Sprintf("the relationship lookup for field '%s' does not exist", field),
	}
}
