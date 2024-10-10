package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/db"
	keelevents "github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func apiHandler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	defer func() {
		if tracerProvider != nil {
			tracerProvider.ForceFlush(ctx)
		}
	}()

	ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", request.RequestContext.HTTP.Method, request.RequestContext.HTTP.Path))
	defer span.End()

	log.WithFields(logrus.Fields{
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

	ctx = runtimectx.WithOAuthConfig(ctx, &keelConfig.Auth)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	ctx = runtimectx.WithSecrets(ctx, secrets)
	ctx = runtimectx.WithStorage(ctx, NewS3BucketStore(ctx))
	ctx = db.WithDatabase(ctx, dbConn)
	ctx = functions.WithFunctionsTransport(ctx, NewLambdaInvokeTransport(os.Getenv("KEEL_FUNCTIONS_ARN")))

	ctx, err := keelevents.WithEventHandler(ctx, sqsEventHandler)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	// mailClient := mail.NewSMTPClientFromEnv()
	// if mailClient != nil {
	// 	ctx = runtimectx.WithMailClient(ctx, mailClient)
	// } else {
	// 	ctx = runtimectx.WithMailClient(ctx, mail.NoOpClient())
	// }

	handler := runtime.NewHttpHandler(schema)

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

	return events.LambdaFunctionURLResponse{
		StatusCode: crw.StatusCode,
		Body:       crw.Body.String(),
		Headers:    responseHeaders,
	}, nil
}
