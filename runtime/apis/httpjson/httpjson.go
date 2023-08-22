package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/jsonschema"
	"github.com/teamkeel/keel/runtime/openapi"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/apis/httpjson")

func NewHandler(p *proto.Schema, api *proto.Api) common.ApiHandlerFunc {
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
			return common.NewJsonResponse(http.StatusUnauthorized, common.NewAuthenticationFailedErr(), nil)
		}
		if identity != nil {
			ctx = runtimectx.WithIdentity(ctx, identity)
		}

		switch r.Method {
		case http.MethodGet:
			inputs = parseQueryParams(r.URL.Query())
		case http.MethodPost:
			var err error
			inputs, err = parsePostBody(r.Body)
			if err != nil {
				return common.NewJsonResponse(http.StatusBadRequest, common.HttpJsonErrorResponse{
					Code:    "ERR_INTERNAL",
					Message: "error parsing POST body",
				}, nil)
			}
		default:
			return common.NewJsonResponse(http.StatusMethodNotAllowed, common.HttpJsonErrorResponse{
				Code:    "ERR_HTTP_METHOD_NOT_ALLOWED",
				Message: "only HTTP POST or GET accepted",
			}, nil)
		}

		op := proto.FindOperation(p, actionName)
		if op == nil {
			span.SetStatus(codes.Error, "action not found")
			return common.NewJsonResponse(http.StatusNotFound, common.HttpJsonErrorResponse{
				Code:    "ERR_NOT_FOUND",
				Message: "method not found",
			}, nil)
		}

		validation, err := jsonschema.ValidateRequest(ctx, p, op, inputs)
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
			// I think this can only happen if we generate an invalid JSON Schema for the
			// request type
			return common.NewJsonResponse(http.StatusBadRequest, common.HttpJsonErrorResponse{
				Code:    "ERR_INTERNAL",
				Message: "error validating request body",
			}, nil)
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

			return common.NewJsonResponse(http.StatusBadRequest, common.HttpJsonErrorResponse{
				Code:    "ERR_INVALID_INPUT",
				Message: "one or more errors found validating request object",
				Data: map[string]any{
					"errors": errs,
				},
			}, nil)
		}

		scope := actions.NewScope(ctx, op, p)

		keelEnv := runtimectx.GetEnv(ctx)

		response, headers, err := actions.Execute(scope, inputs)
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())

			code := "ERR_INTERNAL"

			message := "error executing request"

			if keelEnv == runtimectx.KeelEnvTest {
				message = fmt.Sprintf("error executing request - %s", err.Error())
			}

			httpCode := http.StatusInternalServerError

			var runtimeErr common.RuntimeError
			if errors.As(err, &runtimeErr) {
				code = runtimeErr.Code
				message = runtimeErr.Message

				span.SetAttributes(
					attribute.String("error.code", runtimeErr.Code),
					attribute.String("error.message", runtimeErr.Message),
				)

				switch code {
				case common.ErrInvalidInput:
					httpCode = http.StatusBadRequest
				case common.ErrRecordNotFound:
					httpCode = http.StatusNotFound
				case common.ErrPermissionDenied:
					httpCode = http.StatusForbidden
				}
			}

			return common.NewJsonResponse(httpCode, common.HttpJsonErrorResponse{
				Code:    code,
				Message: message,
			}, nil)
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
