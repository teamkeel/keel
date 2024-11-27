package runtime

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/apis/authapi"
	"github.com/teamkeel/keel/runtime/apis/graphql"
	"github.com/teamkeel/keel/runtime/apis/httpjson"
	"github.com/teamkeel/keel/runtime/apis/jsonrpc"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime")
var Version string

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(logLevel())
}

func GetVersion() string {
	return Version
}

func NewHttpHandler(currSchema *proto.Schema) http.Handler {
	var apiHandler common.HandlerFunc
	var authHandler func(http.ResponseWriter, *http.Request) common.Response
	if currSchema != nil {
		apiHandler = NewApiHandler(currSchema)
		authHandler = NewAuthHandler(currSchema)
	}

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "Runtime")
		defer span.End()

		span.SetAttributes(
			attribute.String("runtime_version", Version),
		)

		if apiHandler == nil || authHandler == nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("cannot serve requests when handlers are not set up"))
			return
		}

		r = r.WithContext(ctx)

		var response common.Response
		path := r.URL.Path
		switch {
		case strings.HasPrefix(path, "/auth"):
			response = authHandler(w, r)
		default:
			response = apiHandler(r)
		}

		// Add any custom headers to response, and join
		// into a single string where multi values exists
		for k, values := range response.Headers {
			for _, value := range values {
				w.Header().Add(k, value)
			}
		}

		span.SetAttributes(
			attribute.Int("response.status", response.Status),
		)

		w.WriteHeader(response.Status)
		_, _ = w.Write(response.Body)
	}

	return http.HandlerFunc(httpHandler)
}

// NewAuthHandler handles requests to the authentication endpoints
func NewAuthHandler(schema *proto.Schema) func(http.ResponseWriter, *http.Request) common.Response {
	handleProviders := authapi.ProvidersHandler(schema)
	handleToken := authapi.TokenEndpointHandler(schema)
	handleRevoke := authapi.RevokeHandler(schema)
	handleAuthorize := authapi.AuthorizeHandler(schema)
	handleCallback := authapi.CallbackHandler(schema)
	handleOpenApiRequest := authapi.OAuthOpenApiSchema()

	return func(w http.ResponseWriter, r *http.Request) common.Response {
		// Collect request headers and add to runtime context
		// These are exposed in custom functions and in expressions
		headers := map[string][]string{}
		for k := range r.Header {
			headers[k] = r.Header.Values(k)
		}
		r = r.WithContext(runtimectx.WithRequestHeaders(r.Context(), headers))

		switch {
		case r.URL.Path == "/auth/providers":
			return handleProviders(r)
		case r.URL.Path == "/auth/token":
			return handleToken(r)
		case r.URL.Path == "/auth/revoke":
			return handleRevoke(r)
		case strings.HasPrefix(r.URL.Path, "/auth/authorize"):
			return handleAuthorize(r)
		case strings.HasPrefix(r.URL.Path, "/auth/callback"):
			return handleCallback(r)
		case r.URL.Path == "/auth/openapi.json":
			return handleOpenApiRequest(r)
		default:
			return common.Response{
				Status: http.StatusNotFound,
			}
		}
	}
}

// NewApiHandler handles requests to the customer APIs
func NewApiHandler(s *proto.Schema) common.HandlerFunc {
	handlers := map[string]common.HandlerFunc{}

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
		ctx := r.Context()

		handler, ok := handlers[strings.ToLower(r.URL.Path)]
		if !ok {
			return common.Response{
				Status: http.StatusNotFound,
				Body:   []byte("Not found"),
			}
		}

		// Collect request headers and add to runtime context
		// These are exposed in custom functions and in expressions
		headers := map[string][]string{}
		for k := range r.Header {
			headers[k] = r.Header.Values(k)
		}
		ctx = runtimectx.WithRequestHeaders(ctx, headers)
		r = r.WithContext(ctx)

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
func (handler JobHandler) RunJob(ctx context.Context, jobName string, inputs map[string]any, trigger functions.TriggerType) error {
	ctx, span := tracer.Start(ctx, "Run job")
	defer span.End()

	job := handler.schema.FindJob(strcase.ToCamel(jobName))
	if job == nil {
		return fmt.Errorf("no job with the name '%s' exists", jobName)
	}

	scope := actions.NewJobScope(ctx, job, handler.schema)
	permissionState := common.NewPermissionState()

	if trigger == functions.ManualTrigger {
		// Check if authorisation can be concluded by role permissions
		canAuthorise, authorised, err := actions.TryResolveAuthorisationEarly(scope, inputs, job.Permissions)
		if err != nil {
			return err
		}

		if canAuthorise {
			if authorised {
				permissionState.Grant()
			} else {
				return common.NewPermissionError()
			}
		}
	}

	var err error
	if job.InputMessageName != "" {
		message := scope.Schema.FindMessage(job.InputMessageName)
		inputs, err = actions.TransformInputs(handler.schema, message, inputs, true)
		if err != nil {
			return err
		}
	}

	err = functions.CallJob(
		ctx,
		job,
		inputs,
		permissionState,
		trigger,
	)

	// Generate and send any events for this context.
	// This must run regardless of the job succeeding or failing.
	// Failure to generate events fail silently.
	eventsErr := events.SendEvents(ctx, scope.Schema)
	if eventsErr != nil {
		span.RecordError(eventsErr)
		span.SetStatus(codes.Error, eventsErr.Error())
	}

	return err
}

type SubscriberHandler struct {
	schema *proto.Schema
}

func NewSubscriberHandler(currSchema *proto.Schema) SubscriberHandler {
	return SubscriberHandler{
		schema: currSchema,
	}
}

// RunSubscriber will run the subscriber function in the runtime with the event payload.
func (handler SubscriberHandler) RunSubscriber(ctx context.Context, subscriberName string, event *events.Event) error {
	ctx, span := tracer.Start(ctx, "Run subscriber")
	defer span.End()

	subscriber := proto.FindSubscriber(handler.schema.Subscribers, subscriberName)
	if subscriber == nil {
		return fmt.Errorf("no subscriber with the name '%s' exists", subscriberName)
	}

	err := functions.CallSubscriber(
		ctx,
		subscriber,
		event,
	)

	// Generate and send any events for this context.
	// This must run regardless of the function succeeding or failing.
	// Failure to generate events fail silently.
	eventsErr := events.SendEvents(ctx, handler.schema)
	if eventsErr != nil {
		span.RecordError(eventsErr)
		span.SetStatus(codes.Error, eventsErr.Error())
	}

	return err
}

func withRequestResponseLogging(handler common.HandlerFunc) common.HandlerFunc {
	return func(request *http.Request) common.Response {
		log.WithFields(log.Fields{
			"url":     request.URL,
			"uri":     request.RequestURI,
			"headers": request.Header,
			"method":  request.Method,
			"host":    request.Host,
		}).Info("Runtime request")

		response := handler(request)

		entry := log.WithFields(log.Fields{
			"headers": response.Headers,
			"status":  response.Status,
			"url":     request.URL,
		})
		if response.Status >= 300 {
			entry.WithField("body", string(response.Body))
		}
		entry.Info("Runtime response")

		return response
	}
}

func logLevel() log.Level {
	switch os.Getenv("KEEL_LOG_LEVEL") {
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
