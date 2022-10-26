package common

import "context"

type Request struct {
	Context context.Context
	Path    string
	Body    []byte
}

type Response struct {
	Body   []byte
	Status int
}
