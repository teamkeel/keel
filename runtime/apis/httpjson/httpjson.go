package httpjson

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/jsonschema"
	"github.com/teamkeel/keel/runtime/locale"
	"github.com/teamkeel/keel/runtime/openapi"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/apis/httpjson")

func NewHandler(p *proto.Schema, api *proto.Api) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "HttpJson")
		defer span.End()

		// Special case for exposing an OpenAPI response
		if strings.HasSuffix(r.URL.Path, "/openapi.json") {
			sch := openapi.Generate(ctx, p, api)
			return common.NewJsonResponse(http.StatusOK, sch, nil)
		}

		pathParts := strings.Split(r.URL.Path, "/")
		actionName := pathParts[len(pathParts)-1]
		var inputs any

		span.SetAttributes(
			attribute.String("api.protocol", "HTTP JSON"),
		)

		identity, err := actions.HandleAuthorizationHeader(ctx, p, r.Header)
		if err != nil {
			return NewErrorResponse(ctx, err, nil)
		}
		if identity != nil {
			ctx = auth.WithIdentity(ctx, identity)
		}

		// handle any Time-Zone headers
		location, err := locale.HandleTimezoneHeader(ctx, r.Header)
		if err != nil {
			return NewErrorResponse(ctx, common.NewInputMalformedError(err.Error()), nil)
		}
		ctx = locale.WithTimeLocation(ctx, location)

		switch r.Method {
		case http.MethodGet:
			inputs = common.ParseQueryParams(r)
		case http.MethodPost:
			var err error
			inputs, err = common.ParseRequestData(r)
			if err != nil {
				return NewErrorResponse(ctx, common.NewInputMalformedError("error parsing POST body"), nil)
			}
		default:
			return NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP POST or GET accepted"), nil)
		}

		action := p.FindAction(actionName)
		if action == nil {
			return NewErrorResponse(ctx, common.NewMethodNotFoundError(), nil)
		}

		validation, err := jsonschema.ValidateRequest(ctx, p, action, inputs)
		if err != nil {
			// Validation cannot complete due to an invalid JSON schema
			return NewErrorResponse(ctx, err, nil)
		}

		if !validation.Valid() {
			messages := []string{}
			errs := []map[string]string{}
			attrs := []attribute.KeyValue{}
			for _, e := range validation.Errors() {
				messages = append(messages, fmt.Sprintf("%s: %s", e.Field(), e.Description()))
				attrs = append(attrs, attribute.String(e.Field(), e.Description()))
				errs = append(errs, map[string]string{
					"field": e.Field(),
					"error": e.Description(),
				})
			}

			span.AddEvent("errors", trace.WithAttributes(attrs...))
			span.SetStatus(codes.Error, strings.Join(messages, ", "))

			err = common.NewValidationError("one or more errors found validating request object")
			return NewErrorResponse(ctx, err, map[string]any{
				"errors": errs,
			})
		}

		scope := actions.NewScope(ctx, action, p)

		response, meta, err := actions.Execute(scope, inputs)
		if err != nil {
			return NewErrorResponse(ctx, err, nil)
		}

		return common.NewJsonResponse(http.StatusOK, response, meta)
	}
}

type HttpJsonErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

func NewErrorResponse(ctx context.Context, err error, data map[string]any) common.Response {
	span := trace.SpanFromContext(ctx)

	code := common.ErrInternal
	message := "error executing request"
	httpCode := http.StatusInternalServerError

	if runtimectx.GetEnv(ctx) == "test" {
		message = fmt.Sprintf("%s (%s)", message, err.Error())
	}

	var runtimeError common.RuntimeError
	if errors.As(err, &runtimeError) {
		code = runtimeError.Code
		message = runtimeError.Message

		switch code {
		case common.ErrInternal:
			httpCode = http.StatusInternalServerError
		case common.ErrInvalidInput:
			httpCode = http.StatusBadRequest
		case common.ErrRecordNotFound:
			httpCode = http.StatusNotFound
		case common.ErrPermissionDenied:
			httpCode = http.StatusForbidden
		case common.ErrAuthenticationFailed:
			httpCode = http.StatusUnauthorized
		case common.ErrMethodNotFound:
			httpCode = http.StatusNotFound
		case common.ErrHttpMethodNotAllowed:
			httpCode = http.StatusMethodNotAllowed
		case common.ErrInputMalformed:
			httpCode = http.StatusBadRequest
		}

		span.SetAttributes(
			attribute.String("error.code", runtimeError.Code),
			attribute.String("error.message", runtimeError.Message),
		)
	}

	span.RecordError(err, trace.WithStackTrace(true))
	span.SetStatus(codes.Error, err.Error())

	return common.NewJsonResponse(httpCode, HttpJsonErrorResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}, nil)
}
