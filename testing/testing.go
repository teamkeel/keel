package testing

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	keelconfig "github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/util"
)

const (
	ActionApiPath = "testingactionsapi"
	JobPath       = "testingjobs"
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
}

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
			"KEEL_DB_CONN_TYPE": "pg",
			"KEEL_DB_CONN":      dbConnString,
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

	// Server to handle receiving HTTP requests from the ActionExecutor and JobExecutor.
	runtimeServer := http.Server{
		Addr: fmt.Sprintf(":%s", runtimePort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = runtimectx.WithDatabase(ctx, database)
			ctx = runtimectx.WithSecrets(ctx, opts.Secrets)
			issuersFromEnv, _ := runtimectx.ExternalIssuersFromEnv()
			ctx = runtimectx.WithExternalIssuers(ctx, issuersFromEnv)
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

				identity, err := runtime.HandleAuthorizationHeader(ctx, schema, r.Header, w)
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

				err = runtime.NewJobHandler(schema).RunJob(ctx, jobName, inputs)

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

	cmd = exec.Command("npx", "vitest", "run", "--color", "--reporter", "verbose", "--config", "./.build/vitest.config.mjs", "--testNamePattern", opts.Pattern)
	cmd.Dir = opts.Dir
	cmd.Env = append(os.Environ(), []string{
		fmt.Sprintf("KEEL_TESTING_ACTIONS_API_URL=http://localhost:%s/testingactionsapi/json", runtimePort),
		fmt.Sprintf("KEEL_TESTING_JOBS_URL=http://localhost:%s/testingjobs/json", runtimePort),
		"KEEL_DB_CONN_TYPE=pg",
		fmt.Sprintf("KEEL_DB_CONN=%s", dbConnString),
		// Disables experimental fetch warning that pollutes console experience when running tests
		"NODE_NO_WARNINGS=1",
	}...)

	b, err = cmd.CombinedOutput()
	if err != nil && !errors.As(err, &exitError) {
		return nil, err
	}

	return &TestOutput{Output: string(b), Success: err == nil}, nil
}
