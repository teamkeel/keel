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
