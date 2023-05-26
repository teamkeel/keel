package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/karlseguin/typed"
	"github.com/teamkeel/keel/casing"
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

type PermissionStatus string

const (
	PermissionGranted PermissionStatus = "granted"
	PermissionUnknown PermissionStatus = "unknown"
)

type PermissionState struct {
	Status PermissionStatus `json:"status"`
	Reason GrantReason      `json:"reason"`
}

type GrantReason string

const (
	GrantReasonRole       GrantReason = "role"
	GrantReasonExpression GrantReason = "expression"
)

func NewPermissionState() *PermissionState {
	return &PermissionState{
		Status: PermissionUnknown,
	}
}

func (ps *PermissionState) Grant(reason GrantReason) {
	ps.Status = PermissionGranted
	ps.Reason = reason
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

func NewInputValidationError(message string) RuntimeError {
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

func NewNotNullError(column string) RuntimeError {
	// Parses from the database casing back to the schema casing.
	// Important since these error messages are delivered to the user.
	field := casing.ToLowerCamel(column)

	return RuntimeError{
		Code:    ErrInvalidInput,
		Message: fmt.Sprintf("field '%s' cannot be null", field),
	}
}

func NewUniquenessError(column string) RuntimeError {
	// Parses from the database casing back to the schema casing.
	// Important since these error messages are delivered to the user.
	field := casing.ToLowerCamel(column)

	return RuntimeError{
		Code:    ErrInvalidInput,
		Message: fmt.Sprintf("field '%s' can only contain unique values", field),
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

// Extracts value from a nullable input type, or errors if incorrectly formatted.
// - If isNull is true, value must not be provided.
// - If isNull is false, value must be provided
// - If isNull is omitted, then value must be provided.
func ValueFromNullableInput(input any) (any, error) {
	asMap, ok := input.(map[string]any)
	if !ok {
		return nil, errors.New("input must be a nullable type")
	}

	typedInput := typed.New(asMap)
	value, hasValue := typedInput.InterfaceIf("value")
	isNull, hasIsNull := typedInput.BoolIf("isNull")

	if hasValue && hasIsNull && isNull {
		return nil, errors.New("nullable input cannot have a value if isNull is true")
	}

	if hasIsNull && !isNull && !hasValue {
		return nil, errors.New("nullable input must have a value if isNull is false")
	}

	if !hasValue && !isNull {
		return nil, errors.New("nullable input must have a value or isNull set")
	}

	if hasIsNull && isNull {
		return nil, nil
	} else {
		return value, nil
	}
}
