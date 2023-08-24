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

	keelconfig "github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
)

const (
	ActionApiPath  = "testingactionsapi"
	JobPath        = "testingjobs"
	SubscriberPath = "testingsubscribers"
)

type TestOutput struct {
	Output  string
	Success bool
}

type RunnerOpts struct {
	Dir             string
	Pattern         string
	DbConnInfo      *db.ConnectionInfo
	FunctionsOutput io.Writer
	EnvVars         map[string]string
	Secrets         map[string]string
	TestGroupName   string
}

var tracer = otel.Tracer("github.com/teamkeel/keel/testing")

func Run(opts *RunnerOpts) (*TestOutput, error) {
	builder := &schema.Builder{}

	schema, err := builder.MakeFromDirectory(opts.Dir)
	if err != nil {
		return nil, err
	}

	testApi := &proto.Api{
		// TODO: make random so doesn't clash
		Name: ActionApiPath,
	}
	for _, m := range schema.Models {
		testApi.ApiModels = append(testApi.ApiModels, &proto.ApiModel{
			ModelName: m.Name,
		})
	}

	schema.Apis = append(schema.Apis, testApi)

	context := context.Background()

	dbName := "keel_test"
	database, err := testhelpers.SetupDatabaseForTestCase(context, opts.DbConnInfo, schema, dbName)
	if err != nil {
		return nil, err
	}
	defer database.Close()

	dbConnString := opts.DbConnInfo.WithDatabase(dbName).String()

	files, err := node.Generate(
		context,
		schema,
		node.WithDevelopmentServer(true),
	)

	if err != nil {
		return nil, err
	}

	err = files.Write(opts.Dir)
	if err != nil {
		return nil, err
	}

	var functionsServer *node.DevelopmentServer
	var functionsTransport functions.Transport

	if node.HasFunctions(schema) {
		keelEnvVars := map[string]string{
			"KEEL_DB_CONN_TYPE":        "pg",
			"KEEL_DB_CONN":             dbConnString,
			"KEEL_TRACING_ENABLED":     "true",
			"OTEL_RESOURCE_ATTRIBUTES": "service.name=functions",
		}

		for key, value := range keelEnvVars {
			opts.EnvVars[key] = value
		}

		functionsServer, err = node.RunDevelopmentServer(opts.Dir, &node.ServerOpts{
			EnvVars: opts.EnvVars,
			Output:  opts.FunctionsOutput,
			Debug:   true, // todo: configurable
		})

		if err != nil {
			if functionsServer != nil && functionsServer.Output() != "" {
				return nil, errors.New(functionsServer.Output())
			}
			return nil, err
		}

		defer func() {
			_ = functionsServer.Kill()
		}()

		functionsTransport = functions.NewHttpTransport(functionsServer.URL)
	}

	runtimePort, err := util.GetFreePort()
	if err != nil {
		return nil, err
	}

	config, err := keelconfig.Load(opts.Dir)
	if err != nil {
		return nil, err
	}

	envVars := config.GetEnvVars("test")
	for key, value := range envVars {
		os.Setenv(key, value)
	}

	// Server to handle receiving HTTP requests from the ActionExecutor, JobExecutor and SubscriberExecutor.
	runtimeServer := http.Server{
		Addr: fmt.Sprintf(":%s", runtimePort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = runtimectx.WithEnv(ctx, runtimectx.KeelEnvTest)
			ctx = runtimectx.WithDatabase(ctx, database)
			ctx = runtimectx.WithSecrets(ctx, opts.Secrets)

			exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())
			if err != nil {
				panic(err)
			}

			provider := traceSdk.NewTracerProvider(
				traceSdk.WithBatcher(exporter),
				traceSdk.WithResource(
					resource.NewSchemaless(attribute.String("service.name", "runtime")),
				),
			)
			otel.SetTracerProvider(provider)
			otel.SetTextMapPropagator(propagation.TraceContext{})

			ctx, span := tracer.Start(ctx, opts.TestGroupName)

			span.SetAttributes(attribute.String("request.url", r.URL.String()))

			defer span.End()

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

			pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(pathParts) != 3 {
				w.WriteHeader(http.StatusNotFound)
				_, err = w.Write([]byte(fmt.Sprintf("invalid url received on testing server '%s'", r.URL.Path)))
				if err != nil {
					panic(err)
				}
			}

			switch pathParts[0] {
			case ActionApiPath:
				r = r.WithContext(ctx)
				runtime.NewHttpHandler(schema).ServeHTTP(w, r)
			case JobPath:
				jobName := pathParts[2]
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				identity, err := actions.HandleAuthorizationHeader(ctx, schema, r.Header)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, err = w.Write([]byte(err.Error()))
					if err != nil {
						panic(err)
					}
				}

				if identity != nil {
					ctx = runtimectx.WithIdentity(ctx, identity)
				}

				var inputs map[string]any
				// if no json body has been sent, just return an empty map for the inputs
				if string(body) == "" {
					inputs = nil
				} else {
					err = json.Unmarshal(body, &inputs)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

				trigger := functions.TriggerType(r.Header.Get("X-Trigger-Type"))

				err = runtime.NewJobHandler(schema).RunJob(ctx, jobName, inputs, trigger)

				if err != nil {
					response := common.NewJsonErrorResponse(err)

					w.WriteHeader(response.Status)
					_, err = w.Write(response.Body)
					if err != nil {
						panic(err)
					}
				} else {
					w.WriteHeader(http.StatusOK)
				}
			case SubscriberPath:
				subscriberName := pathParts[2]
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				var event *events.Event
				err = json.Unmarshal(body, &event)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				err = runtime.NewSubscriberHandler(schema).RunSubscriber(ctx, subscriberName, event)

				if err != nil {
					response := common.NewJsonErrorResponse(err)

					w.WriteHeader(response.Status)
					_, err = w.Write(response.Body)
					if err != nil {
						panic(err)
					}
				} else {
					w.WriteHeader(http.StatusOK)
				}
			default:
				w.WriteHeader(http.StatusNotFound)
				_, err = w.Write([]byte(fmt.Sprintf("invalid url received on testing server '%s'", r.URL.Path)))
				if err != nil {
					panic(err)
				}
			}
		}),
	}

	go func() {
		_ = runtimeServer.ListenAndServe()
	}()

	defer func() {
		_ = runtimeServer.Shutdown(context)
	}()

	cmd := exec.Command("npx", "tsc", "--noEmit", "--pretty")
	cmd.Dir = opts.Dir

	b, err := cmd.CombinedOutput()
	exitError := &exec.ExitError{}
	if err != nil && !errors.As(err, &exitError) {
		return nil, err
	}
	if err != nil {
		return &TestOutput{Output: string(b), Success: false}, nil
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

	cmd = exec.Command("npx", "vitest", "run", "--color", "--reporter", "verbose", "--config", "./.build/vitest.config.mjs", "--testNamePattern", opts.Pattern)
	cmd.Dir = opts.Dir
	cmd.Env = append(os.Environ(), []string{
		fmt.Sprintf("KEEL_TESTING_ACTIONS_API_URL=http://localhost:%s/%s/json", runtimePort, ActionApiPath),
		fmt.Sprintf("KEEL_TESTING_JOBS_URL=http://localhost:%s/%s/json", runtimePort, JobPath),
		fmt.Sprintf("KEEL_TESTING_SUBSCRIBERS_URL=http://localhost:%s/%s/json", runtimePort, SubscriberPath),
		"KEEL_DB_CONN_TYPE=pg",
		fmt.Sprintf("KEEL_DB_CONN=%s", dbConnString),
		// Disables experimental fetch warning that pollutes console experience when running tests
		"NODE_NO_WARNINGS=1",
		fmt.Sprintf("KEEL_DEFAULT_PK=%s", pkBase64),
	}...)

	b, err = cmd.CombinedOutput()
	if err != nil && !errors.As(err, &exitError) {
		return nil, err
	}

	return &TestOutput{Output: string(b), Success: err == nil}, nil
}
