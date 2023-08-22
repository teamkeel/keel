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
	"github.com/teamkeel/keel/runtime/apis/graphql"
	"github.com/teamkeel/keel/runtime/apis/httpjson"
	"github.com/teamkeel/keel/runtime/apis/jsonrpc"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

		w.Header().Add("Content-Type", "application/json")

		if handler == nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Cannot serve requests when schema contains errors"))
			return
		}

		ctx = runtimectx.WithIssuersFromEnv(ctx)

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
func (handler JobHandler) RunJob(ctx context.Context, jobName string, inputs map[string]any, trigger functions.TriggerType) error {
	ctx, span := tracer.Start(ctx, "Run job")
	defer span.End()

	job := proto.FindJob(handler.schema.Jobs, strcase.ToCamel(jobName))
	if job == nil {
		return fmt.Errorf("no job with the name '%s' exists", jobName)
	}

	span.SetAttributes(
		attribute.String("job.name", job.Name),
	)

	scope := actions.NewJobScope(ctx, job, handler.schema)

	permissionState := common.NewPermissionState()

	if trigger == functions.ManualTrigger {
		// Check if authorisation can be achieved early.
		canAuthoriseEarly, authorised, err := actions.TryResolveAuthorisationEarly(scope, job.Permissions)
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

	err := functions.CallJob(
		ctx,
		job,
		inputs,
		permissionState,
		trigger,
	)

	// Generate and send any events for this context.
	eventsErr := events.GenerateEvents(ctx)
	if eventsErr != nil {
		span.RecordError(eventsErr)
	}

	return err
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
