package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/apis/graphql"
	"github.com/teamkeel/keel/runtime/apis/httpjson"
	"github.com/teamkeel/keel/runtime/apis/jsonrpc"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime")
var Version string

const (
	authorizationHeaderName string = "Authorization"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(logLevel())
}

func GetVersion() string {
	return Version
}

func NewHttpHandler(currSchema *proto.Schema) http.Handler {
	var handler common.ApiHandlerFunc
	if currSchema != nil {
		handler = NewHandler(currSchema)
	}

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "Runtime")
		defer span.End()

		span.SetAttributes(
			attribute.String("runtime_version", Version),
		)

		if handler == nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Cannot serve requests when schema contains errors"))
			return
		}

		identity, err := HandleAuthorizationHeader(ctx, currSchema, r.Header, w)
		if err != nil {
			return
		}
		if identity != nil {
			ctx = runtimectx.WithIdentity(ctx, identity)
		}

		// Collect request headers and add to runtime context
		// These are exposed in custom functions and in expressions
		headers := map[string][]string{}
		for k := range r.Header {
			headers[k] = r.Header.Values(k)
		}
		ctx = runtimectx.WithRequestHeaders(ctx, headers)
		r = r.WithContext(ctx)

		response := handler(r)

		// Add any custom headers to response, and join
		// into a single string where multi values exists
		for k, values := range response.Headers {
			for _, value := range values {
				w.Header().Add(k, value)
			}
		}

		w.Header().Add("Content-Type", "application/json")

		span.SetAttributes(
			attribute.Int("response.status", response.Status),
		)

		w.WriteHeader(response.Status)
		_, _ = w.Write(response.Body)
	}

	return http.HandlerFunc(httpHandler)
}

func NewHandler(s *proto.Schema) common.ApiHandlerFunc {
	handlers := map[string]common.ApiHandlerFunc{}

	for _, api := range s.Apis {
		root := "/" + strings.ToLower(api.Name)

		handlers[root+"/graphql"] = graphql.NewHandler(s, api)
		handlers[root+"/rpc"] = jsonrpc.NewHandler(s, api)

		httpJson := httpjson.NewHandler(s, api)
		for _, name := range proto.GetActionNamesForApi(s, api) {
			handlers[root+"/json/"+strings.ToLower(name)] = httpJson
		}
		handlers[root+"/json/openapi.json"] = httpJson
	}

	return withRequestResponseLogging(func(r *http.Request) common.Response {
		handler, ok := handlers[strings.ToLower(r.URL.Path)]
		if !ok {
			return common.Response{
				Status: 404,
				Body:   []byte("Not found"),
			}
		}

		return handler(r)
	})
}

type JobHandler struct {
	schema *proto.Schema
}

func NewJobHandler(currSchema *proto.Schema) JobHandler {
	return JobHandler{
		schema: currSchema,
	}
}

// RunJob will run the job function in the runtime.
func (handler JobHandler) RunJob(ctx context.Context, jobName string, inputs map[string]any) error {
	job := proto.FindJob(handler.schema.Jobs, strcase.ToCamel(jobName))
	if job == nil {
		return fmt.Errorf("no job with the name '%s' exists", jobName)
	}

	scope := actions.NewJobScope(ctx, job, handler.schema)

	// Check if authorisation can be achieved early.
	permissions := proto.PermissionsForJob(handler.schema, job)
	canAuthoriseEarly, authorised, err := actions.TryResolveAuthorisationEarly(scope, permissions)
	if err != nil {
		return err
	}

	permissionState := common.NewPermissionState()
	if canAuthoriseEarly {
		if authorised {
			permissionState.Grant()
		} else {
			return common.NewPermissionError()
		}
	}

	return functions.CallJob(
		ctx,
		handler.schema,
		job,
		inputs,
		permissionState,
	)
}

func withRequestResponseLogging(handler common.ApiHandlerFunc) common.ApiHandlerFunc {
	return func(request *http.Request) common.Response {
		log.WithFields(log.Fields{
			"url":     request.URL,
			"uri":     request.RequestURI,
			"headers": request.Header,
			"method":  request.Method,
			"host":    request.Host,
		})

		response := handler(request)

		entry := log.WithFields(log.Fields{
			"headers": response.Headers,
			"status":  response.Status,
		})
		if response.Status >= 300 {
			entry.WithField("body", string(response.Body))
		}
		entry.Info("response")

		return response
	}
}

func logLevel() log.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "trace":
		return log.TraceLevel
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.ErrorLevel
	}
}

func HandleAuthorizationHeader(ctx context.Context, schema *proto.Schema, headers http.Header, w http.ResponseWriter) (*runtimectx.Identity, error) {
	ctx, span := tracer.Start(ctx, "Authorization")
	defer span.End()

	header := headers.Get(authorizationHeaderName)
	if header == "" {
		return nil, nil
	}

	headerSplit := strings.Split(header, "Bearer ")
	if len(headerSplit) != 2 {
		span.SetStatus(codes.Error, "no 'Bearer' prefix in the authentication header")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("no 'Bearer' prefix in the authentication header"))
		return nil, errors.New("invalid authorization header")
	}

	subject, issuer, err := actions.ValidateBearerToken(ctx, headerSplit[1])
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
	}

	switch {
	case errors.Is(err, actions.ErrInvalidToken) || errors.Is(err, actions.ErrTokenExpired):
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return nil, err
	case err != nil:
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("error validating bearer token"))
		return nil, err
	}

	// Check that identity actually does exist as it could
	// have been deleted after the bearer token was generated.
	var identity *runtimectx.Identity
	if issuer == "keel" || issuer == "" {
		identity, err = actions.FindIdentityById(ctx, schema, subject)
	} else {
		identity, err = actions.FindIdentityByExternalId(ctx, schema, subject, issuer)
		if identity == nil {
			identity, err = actions.CreateExternalIdentity(ctx, schema, subject, issuer)
		}
	}

	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error validating identity"))
		return nil, err
	}

	if identity == nil {
		span.SetStatus(codes.Error, "identity not found")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(actions.ErrIdentityNotFound.Error()))
		return nil, err
	}

	span.SetAttributes(attribute.String("identity.id", identity.Id))

	return identity, nil
}
