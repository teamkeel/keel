package testing

import (
	"context"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/apis/httpjson"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const (
	AuthPath       = "auth"
	ActionApiPath  = "testingactionsapi"
	JobPath        = "testingjobs"
	SubscriberPath = "testingsubscribers"
)

type TestOutput struct {
	Output  string
	Success bool
}

type RunnerOpts struct {
	Dir           string
	Pattern       string
	DbConnInfo    *db.ConnectionInfo
	Secrets       map[string]string
	TestGroupName string
	// Generates a Keel client (keelClient.ts) for this test
	GenerateClient bool
}

var tracer = otel.Tracer("github.com/teamkeel/keel/testing")

func Run(ctx context.Context, opts *RunnerOpts) error {
	builder := &schema.Builder{}

	schema, err := builder.MakeFromDirectory(opts.Dir)
	if err != nil {
		return err
	}

	envVars := builder.Config.GetEnvVars()

	testApi := &proto.Api{
		// TODO: make random so doesn't clash
		Name: ActionApiPath,
	}
	for _, m := range schema.Models {
		apiModel := &proto.ApiModel{
			ModelName:    m.Name,
			ModelActions: []*proto.ApiModelAction{},
		}

		testApi.ApiModels = append(testApi.ApiModels, apiModel)
		for _, a := range m.Actions {
			apiModel.ModelActions = append(apiModel.ModelActions, &proto.ApiModelAction{ActionName: a.Name})
		}
	}

	schema.Apis = append(schema.Apis, testApi)

	spanName := opts.TestGroupName
	if spanName == "" {
		spanName = "testing.Run"
	}
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	dbName := "keel_test"
	database, err := testhelpers.SetupDatabaseForTestCase(ctx, opts.DbConnInfo, schema, dbName, true)
	if err != nil {
		return err
	}
	defer database.Close()

	dbConnString := opts.DbConnInfo.WithDatabase(dbName).String()

	files, err := node.Generate(
		ctx,
		schema,
		node.WithDevelopmentServer(true),
	)
	if err != nil {
		return err
	}

	if opts.GenerateClient {
		clientFiles, err := node.GenerateClient(
			ctx,
			schema,
			false,
			ActionApiPath,
		)
		if err != nil {
			return err
		}

		files = append(files, clientFiles...)
	}

	err = files.Write(opts.Dir)
	if err != nil {
		return err
	}

	var functionsServer *node.DevelopmentServer
	var functionsTransport functions.Transport

	if node.HasFunctions(schema) {
		functionEnvVars := map[string]string{
			"KEEL_DB_CONN_TYPE":        "pg",
			"KEEL_DB_CONN":             dbConnString,
			"KEEL_TRACING_ENABLED":     os.Getenv("TRACING_ENABLED"),
			"OTEL_RESOURCE_ATTRIBUTES": "service.name=functions",
		}

		for key, value := range envVars {
			functionEnvVars[key] = value
		}

		functionsServer, err = node.StartDevelopmentServer(ctx, opts.Dir, &node.ServerOpts{
			EnvVars: functionEnvVars,
			Output:  os.Stdout,
			Debug:   os.Getenv("DEBUG_FUNCTIONS") == "true",
		})

		if err != nil {
			if functionsServer != nil && functionsServer.Output() != "" {
				return errors.New(functionsServer.Output())
			}
			return err
		}

		defer func() {
			_ = functionsServer.Kill()
		}()

		functionsTransport = functions.NewHttpTransport(functionsServer.URL)
	}

	runtimePort, err := util.GetFreePort()
	if err != nil {
		return err
	}

	for key, value := range envVars {
		os.Setenv(key, value)
	}

	// Server to handle receiving HTTP requests from the ActionExecutor, JobExecutor and SubscriberExecutor.
	runtimeServer := http.Server{
		Addr: fmt.Sprintf(":%s", runtimePort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), strings.Trim(r.URL.Path, "/"))
			defer span.End()

			ctx = runtimectx.WithEnv(ctx, runtimectx.KeelEnvTest)
			ctx = db.WithDatabase(ctx, database)
			ctx = runtimectx.WithSecrets(ctx, opts.Secrets)
			ctx = runtimectx.WithOAuthConfig(ctx, &builder.Config.Auth)

			span.SetAttributes(attribute.String("request.url", r.URL.String()))

			// Use the embedded private key for the tests
			pk, err := testhelpers.GetEmbeddedPrivateKey()
			if err != nil {
				panic(err)
			}

			if pk == nil {
				panic("No private key")
			}

			ctx = runtimectx.WithPrivateKey(ctx, pk)

			if functionsTransport != nil {
				ctx = functions.WithFunctionsTransport(ctx, functionsTransport)
			}

			// Synchronous event handling
			ctx, err = events.WithEventHandler(ctx, func(ctx context.Context, subscriber string, event *events.Event, traceparent string) error {
				return runtime.NewSubscriberHandler(schema).RunSubscriber(ctx, subscriber, event)
			})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
			}

			pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

			switch pathParts[0] {
			case AuthPath:
				r = r.WithContext(ctx)
				runtime.NewHttpHandler(schema).ServeHTTP(w, r)
			case ActionApiPath:
				r = r.WithContext(ctx)
				runtime.NewHttpHandler(schema).ServeHTTP(w, r)
			case JobPath:
				if len(pathParts) != 3 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				err := HandleJobExecutorRequest(ctx, schema, pathParts[2], r)
				if err != nil {
					response := httpjson.NewErrorResponse(ctx, err, nil)
					w.WriteHeader(response.Status)
					_, _ = w.Write(response.Body)
				}
			case SubscriberPath:
				if len(pathParts) != 3 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				err := HandleSubscriberExecutorRequest(ctx, schema, pathParts[2], r)
				if err != nil {
					response := httpjson.NewErrorResponse(ctx, err, nil)
					w.WriteHeader(response.Status)
					_, _ = w.Write(response.Body)
				}
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	}

	go func() {
		_ = runtimeServer.ListenAndServe()
	}()

	defer func() {
		_ = runtimeServer.Shutdown(ctx)
	}()

	cmd := exec.Command("./node_modules/.bin/tsc", "--noEmit", "--pretty")
	cmd.Dir = opts.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	if opts.Pattern == "" {
		opts.Pattern = "(.*)"
	}

	pk, _ := testhelpers.GetEmbeddedPrivateKey()

	pkBytes := x509.MarshalPKCS1PrivateKey(pk)
	pkPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: pkBytes,
		},
	)

	pkBase64 := base64.StdEncoding.EncodeToString(pkPem)

	cmd = exec.Command("./node_modules/.bin/vitest", "run", "--color", "--reporter", "verbose", "--config", "./.build/vitest.config.mjs", "--testNamePattern", opts.Pattern)
	cmd.Dir = opts.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), []string{
		fmt.Sprintf("KEEL_TESTING_ACTIONS_API_URL=http://localhost:%s/%s/json", runtimePort, ActionApiPath),
		fmt.Sprintf("KEEL_TESTING_JOBS_URL=http://localhost:%s/%s/json", runtimePort, JobPath),
		fmt.Sprintf("KEEL_TESTING_SUBSCRIBERS_URL=http://localhost:%s/%s/json", runtimePort, SubscriberPath),
		fmt.Sprintf("KEEL_TESTING_CLIENT_API_URL=http://localhost:%s/%s", runtimePort, ActionApiPath),
		fmt.Sprintf("KEEL_TESTING_AUTH_API_URL=http://localhost:%s/auth", runtimePort),
		"KEEL_DB_CONN_TYPE=pg",
		fmt.Sprintf("KEEL_DB_CONN=%s", dbConnString),
		// Disables experimental fetch warning that pollutes console experience when running tests
		"NODE_NO_WARNINGS=1",
		fmt.Sprintf("KEEL_DEFAULT_PK=%s", pkBase64),
	}...)

	return cmd.Run()
}

// HandleJobExecutorRequest handles requests the job module in the testing package.
func HandleJobExecutorRequest(ctx context.Context, schema *proto.Schema, jobName string, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	identity, err := actions.HandleAuthorizationHeader(ctx, schema, r.Header)
	if err != nil {
		return err
	}

	if identity != nil {
		ctx = auth.WithIdentity(ctx, identity)
	}

	var inputs map[string]any
	// if no json body has been sent, just return an empty map for the inputs
	if string(body) == "" {
		inputs = nil
	} else {
		err = json.Unmarshal(body, &inputs)
		if err != nil {
			return err
		}
	}

	trigger := functions.TriggerType(r.Header.Get("X-Trigger-Type"))

	err = runtime.NewJobHandler(schema).RunJob(ctx, jobName, inputs, trigger)

	if err != nil {
		return err
	}

	return nil
}

// HandleSubscriberExecutorRequest handles requests the subscriber module in the testing package.
func HandleSubscriberExecutorRequest(ctx context.Context, schema *proto.Schema, subscriberName string, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var event *events.Event
	err = json.Unmarshal(body, &event)
	if err != nil {
		return err
	}

	err = runtime.NewSubscriberHandler(schema).RunSubscriber(ctx, subscriberName, event)

	if err != nil {
		return err
	}

	return nil
}
