package httpjson

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/jsonschema"
	"github.com/teamkeel/keel/runtime/openapi"
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

		switch r.Method {
		case http.MethodGet:
			inputs = parseQueryParams(r.URL.Query())
		case http.MethodPost:
			var err error
			inputs, err = parsePostBody(r.Body)
			if err != nil {
				return NewErrorResponse(ctx, common.NewInputMalformedError("error parsing POST body"), nil)
			}
		default:
			return NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP POST or GET accepted"), nil)

		}

		action := proto.FindAction(p, actionName)
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

		response, headers, err := actions.Execute(scope, inputs)
		if err != nil {
			return NewErrorResponse(ctx, err, nil)
		}

		return common.NewJsonResponse(http.StatusOK, response, headers)
	}
}

func parseQueryParams(q url.Values) map[string]any {
	inputs := map[string]any{}
	for k := range q {
		inputs[k] = q.Get(k)
	}
	return inputs
}

func parsePostBody(b io.ReadCloser) (inputs any, err error) {
	body, err := io.ReadAll(b)
	if err != nil {
		return nil, err
	}

	// if no json body has been sent, just return an empty map for the inputs
	if string(body) == "" {
		return map[string]any{}, nil
	}

	err = json.Unmarshal(body, &inputs)
	return inputs, err
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
