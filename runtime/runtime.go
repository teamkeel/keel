package runtime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/apis/authapi"
	"github.com/teamkeel/keel/runtime/apis/flowsapi"
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
	var flowsHandler common.HandlerFunc
	var authHandler func(http.ResponseWriter, *http.Request) common.Response
	var router *httprouter.Router
	if currSchema != nil {
		apiHandler = NewApiHandler(currSchema)
		flowsHandler = NewFlowsHandler(currSchema)
		authHandler = NewAuthHandler(currSchema)
		router = NewRouter(currSchema)
	}

	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "Runtime")
		defer span.End()

		span.SetAttributes(
			attribute.String("runtime_version", Version),
		)

		if apiHandler == nil || authHandler == nil || flowsHandler == nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("cannot serve requests when handlers are not set up"))
			return
		}

		r = r.WithContext(ctx)

		var response common.Response
		path := r.URL.Path
		switch {
		case strings.HasPrefix(path, "/flows"):
			response = flowsHandler(r)
		case strings.HasPrefix(path, "/auth"):
			response = authHandler(w, r)
		default:
			response = apiHandler(r)
		}

		// TODO: this is a bit of a hack - what we want to do is only use the routes router if no API or auth
		// route matched, but we can't just check for a 404 because that happens for "record not found".
		// So we check for both a 404 status and a non-JSON body.
		// Probably the right thing to do is refactor this whole thing to just use a single httprouter but that
		// is a bigger change.
		if response.Status == http.StatusNotFound && (len(response.Body) == 0 || response.Body[0] != '{') {
			router.ServeHTTP(w, r)
			return
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

// NewAuthHandler handles requests to the authentication endpoints.
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

// NewApiHandler handles requests to the customer APIs.
func NewApiHandler(s *proto.Schema) common.HandlerFunc {
	handlers := map[string]common.HandlerFunc{}

	for _, api := range s.GetApis() {
		root := "/" + strings.ToLower(api.GetName())

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

// NewFlowsHandler handles requests to the customer flows.
func NewFlowsHandler(s *proto.Schema) common.HandlerFunc {
	defaultFlowHandler := flowsapi.FlowHandler(s)

	explicitHandlers := map[string]common.HandlerFunc{
		"/flows/json":              flowsapi.ListFlowsHandler(s),
		"/flows/json/openapi.json": flowsapi.OpenAPISchemaHandler(s),
		// TODO: "/flows/json/myRuns"
	}

	return withRequestResponseLogging(func(r *http.Request) common.Response {
		ctx := r.Context()

		handler, ok := explicitHandlers[strings.ToLower(r.URL.Path)]
		if !ok {
			handler = defaultFlowHandler
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
func (handler JobHandler) RunJob(ctx context.Context, jobName string, input map[string]any, trigger functions.TriggerType) error {
	ctx, span := tracer.Start(ctx, "Run job")
	defer span.End()

	job := handler.schema.FindJob(strcase.ToCamel(jobName))
	if job == nil {
		return fmt.Errorf("no job with the name '%s' exists", jobName)
	}

	scope := actions.NewJobScope(ctx, job, handler.schema)
	permissionState := common.NewPermissionState()

	if trigger == functions.ManualTrigger {
		// Check if authorisation can be achieved early.
		canAuthoriseEarly, authorised, err := actions.TryResolveAuthorisationEarly(scope, input, job.GetPermissions())
		if err != nil {
			return err
		}

		if canAuthoriseEarly {
			if authorised {
				permissionState.Grant()
			} else {
				return common.NewPermissionError()
			}
		}
	}

	var err error
	if job.GetInputMessageName() != "" {
		message := scope.Schema.FindMessage(job.GetInputMessageName())
		input, err = actions.TransformInputs(handler.schema, message, input, true)
		if err != nil {
			return err
		}
	}

	err = functions.CallJob(
		ctx,
		job,
		input,
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

	subscriber := proto.FindSubscriber(handler.schema.GetSubscribers(), subscriberName)
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

func NewRouter(s *proto.Schema) *httprouter.Router {
	router := httprouter.New()

	for _, route := range s.GetRoutes() {
		var method string

		switch route.GetMethod() {
		case proto.HttpMethod_HTTP_METHOD_GET:
			method = http.MethodGet
		case proto.HttpMethod_HTTP_METHOD_POST:
			method = http.MethodPost
		case proto.HttpMethod_HTTP_METHOD_PUT:
			method = http.MethodPut
		case proto.HttpMethod_HTTP_METHOD_DELETE:
			method = http.MethodDelete
		}

		router.Handle(method, route.GetPattern(), func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("cannot read body"))
				return
			}

			paramsMap := map[string]string{}
			for _, p := range params {
				paramsMap[p.Key] = p.Value
			}

			headers := map[string][]string{}
			for k := range r.Header {
				headers[k] = r.Header.Values(k)
			}
			ctx := runtimectx.WithRequestHeaders(r.Context(), headers)

			resp, _, err := functions.CallRoute(ctx, route.GetHandler(), &functions.RouteRequest{
				Body:   string(body),
				Method: r.Method,
				Path:   r.URL.Path,
				Params: paramsMap,
				Query:  r.URL.RawQuery,
			})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("error calling handler"))
				return
			}

			for k, v := range resp.Headers {
				w.Header().Add(k, v)
			}

			if resp.StatusCode != 0 {
				w.WriteHeader(resp.StatusCode)
			} else {
				w.WriteHeader(http.StatusOK)
			}
			_, _ = w.Write([]byte(resp.Body))
		})
	}

	return router
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
