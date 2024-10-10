package main

import (
	"bytes"
	"net/http"
)

// ResponseWriter is a minimal implementation of http.ResponseWriter that
// simply stores the response for later inspection.
type ResponseWriter struct {
	StatusCode int
	Body       bytes.Buffer
	HeaderMap  http.Header
}

// Header needed to implement http.ResponseWriter.
func (c *ResponseWriter) Header() http.Header {
	return c.HeaderMap
}

// Write needed to implement http.ResponseWriter.
func (c *ResponseWriter) Write(b []byte) (int, error) {
	return c.Body.Write(b)
}

// WriteHeader needed to implement http.ResponseWriter.
func (c *ResponseWriter) WriteHeader(statusCode int) {
	c.StatusCode = statusCode
}
