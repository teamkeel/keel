package authapi

import (
	"context"
	"net/url"

	"github.com/teamkeel/keel/runtime/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// https://openid.net/specs/openid-connect-standard-1_0-21_orig.html#AccessTokenErrorResponse
// https://datatracker.ietf.org/doc/html/rfc7009#section-2.2

type ErrorResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// Errors which are returned in the body as JSON.
func jsonErrResponse(ctx context.Context, status int, errorType string, errorDescription string, err error) common.Response {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Error, errorType)

	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
	}

	span.SetAttributes(
		attribute.String("auth.error", errorType),
		attribute.String("auth.error_description", errorDescription),
	)

	return common.NewJsonResponse(status, &ErrorResponse{
		Error:            errorType,
		ErrorDescription: errorDescription,
	}, nil)
}

// Errors which are captured in the redirect query.
func redirectErrResponse(ctx context.Context, redirectUrl *url.URL, errorType string, errorDescription string, err error) common.Response {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Error, errorType)

	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
	}

	span.SetAttributes(
		attribute.String("auth.error", errorType),
		attribute.String("auth.error_description", errorDescription),
	)

	values := url.Values{}
	values.Add("error", errorType)
	values.Add("error_description", errorDescription)
	redirectUrl.RawQuery = values.Encode()

	return common.NewRedirectResponse(redirectUrl)
}
