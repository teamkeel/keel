package runtime

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/compression"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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

func (h *Handler) APIHandler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	defer func() {
		if h.tracerProvider != nil {
			h.tracerProvider.ForceFlush(ctx)
		}
	}()

	ctx, span := h.tracer.Start(ctx, fmt.Sprintf("%s %s", request.RequestContext.HTTP.Method, request.RequestContext.HTTP.Path))
	defer span.End()

	h.log.WithFields(logrus.Fields{
		"method": request.RequestContext.HTTP.Method,
		"path":   request.RequestContext.HTTP.Path,
	}).Info("API request")

	span.SetAttributes(
		attribute.String("type", "request"),
		attribute.String("http.method", request.RequestContext.HTTP.Method),
		attribute.String("http.path", request.RequestContext.HTTP.Path),
		attribute.String("http.useragent", request.RequestContext.HTTP.UserAgent),
		attribute.String("http.ipAddress", request.RequestContext.HTTP.SourceIP),
		attribute.String("aws.requestID", request.RequestContext.RequestID),
	)

	if request.RawPath == "/_health" {
		statusResponse, _ := json.Marshal(map[string]string{"status": "ok"})
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusOK,
			Body:       string(statusResponse),
		}, nil
	}

	ctx, err := h.buildContext(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	handler := runtime.NewHttpHandler(h.schema)

	headers := http.Header{}
	for k, v := range request.Headers {
		headers.Set(k, v)
	}

	body := request.Body
	if request.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
			return events.LambdaFunctionURLResponse{
				StatusCode: http.StatusInternalServerError,
			}, nil
		}
		body = string(decoded)
	}

	runtimeRequest := &http.Request{
		Method: request.RequestContext.HTTP.Method,
		URL: &url.URL{
			Path:     request.RequestContext.HTTP.Path,
			RawQuery: request.RawQueryString,
		},
		Header: headers,
		Body:   io.NopCloser(strings.NewReader(body)),
	}

	runtimeRequest = runtimeRequest.WithContext(ctx)

	crw := &ResponseWriter{
		HeaderMap: make(http.Header),
	}

	handler.ServeHTTP(crw, runtimeRequest)

	responseHeaders := map[string]string{}
	for k, v := range crw.HeaderMap {
		responseHeaders[k] = v[0]
	}

	// Apply gzip compression if appropriate
	responseBody := crw.Body.Bytes()
	isBase64Encoded := false

	if compression.ShouldCompress(responseBody, headers) {
		compressed, err := compression.Compress(responseBody)
		if err != nil {
			span.RecordError(err)
			h.log.WithError(err).Error("failed to compress response")
		} else {
			responseBody = compressed
			isBase64Encoded = true
			compression.SetCompressionHeaders(crw.HeaderMap)

			// Update response headers with compression headers
			responseHeaders = map[string]string{}
			for k, v := range crw.HeaderMap {
				responseHeaders[k] = v[0]
			}
		}
	}

	responseBodyStr := string(responseBody)
	if isBase64Encoded {
		responseBodyStr = base64.StdEncoding.EncodeToString(responseBody)
	}

	return events.LambdaFunctionURLResponse{
		StatusCode:      crw.StatusCode,
		Body:            responseBodyStr,
		Headers:         responseHeaders,
		IsBase64Encoded: isBase64Encoded,
	}, nil
}
