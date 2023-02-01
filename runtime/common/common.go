package common

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Body   []byte
	Status int
}

func NewJsonResponse(status int, body any) Response {
	b, _ := json.Marshal(body)
	return Response{
		Status: status,
		Body:   b,
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
