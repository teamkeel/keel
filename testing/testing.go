package testing

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/cmd/database"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/util"
	"gorm.io/gorm"
)

const (
	EventTypeTestRun = "TestRun"
)

var TestIgnorePatterns []string = []string{"node_modules"}

const dbConnUri = "postgres://postgres:postgres@0.0.0.0:8001/%s"

type Event struct {
	Status   string          `json:"status"`
	TestName string          `json:"testName"`
	Expected json.RawMessage `json:"expected,omitempty"`
	Actual   json.RawMessage `json:"actual,omitempty"`
	Err      json.RawMessage `json:"err,omitempty"`
}

type ActionRequest struct {
	ActionName string         `json:"actionName"`
	Payload    map[string]any `json:"payload"`
}

func Run(t *testing.T, dir string) (<-chan []*Event, error) {
	builder := &schema.Builder{}
	shortDir := filepath.Base(dir)
	dbName := testhelpers.DbNameForTestName(shortDir)

	var db *gorm.DB

	schema, err := builder.MakeFromDirectory(dir)
	if err != nil {
		return nil, err
	}

	ch := make(chan []*Event)

	reportingPort, err := util.GetFreePort()

	if err != nil {
		return nil, err
	}

	customFunctionsRuntime, err := functions.NewRuntime(schema, dir)

	if err != nil {
		return nil, err
	}

	err = customFunctionsRuntime.Bootstrap()

	if err != nil {
		return nil, err
	}

	var ops []*proto.Operation = []*proto.Operation{}

	for _, model := range schema.Models {
		ops = append(ops, model.Operations...)
	}

	hasCustomFunctions := lo.SomeBy(ops, func(o *proto.Operation) bool {
		return o.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM
	})

	var customFunctionRuntimeProcess *os.Process
	var customFunctionRuntimePort string

	if hasCustomFunctions {
		customFunctionRuntimePort, err = util.GetFreePort()

		if err != nil {
			panic(err)
		}

		customFunctionRuntimeProcess, err = RunServer(dir, customFunctionRuntimePort, reportingPort, fmt.Sprintf(dbConnUri, dbName))

		if err != nil {
			panic(err)
		}
	}

	injector := NewInjector(dir, schema)

	err = injector.Inject()

	if err != nil {
		panic(err)
	}

	// Server for node test process to talk to
	srv := http.Server{
		Addr: fmt.Sprintf(":%s", reportingPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := ioutil.ReadAll(r.Body)

			switch r.URL.Path {
			case "/action":
				body := &ActionRequest{}
				json.Unmarshal(b, body)

				for _, model := range schema.Models {
					for _, action := range model.Operations {
						if action.Name == body.ActionName {
							if action.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO {
								ctx := r.Context()
								ctx = runtimectx.WithDatabase(ctx, db)

								switch action.Type {
								case proto.OperationType_OPERATION_TYPE_GET:
									res, err := actions.Get(ctx, action, schema, body.Payload)

									r := map[string]any{
										"object": res,
										"errors": serializeError(err),
									}

									WriteResponse(r, w)
								case proto.OperationType_OPERATION_TYPE_CREATE:
									res, err := actions.Create(ctx, action, schema, body.Payload)

									r := map[string]any{
										"object": res,
										"errors": serializeError(err),
									}

									WriteResponse(r, w)
								default:
									w.WriteHeader(400)
									panic(fmt.Sprintf("%s not yet implemented", action.Type))
								}

								return
							}

							if action.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
								// call node process
								ctx := r.Context()

								client := &functions.HttpFunctionsClient{
									Port: customFunctionRuntimePort,
									Host: "0.0.0.0",
								}

								ctx = runtime.WithFunctionsClient(ctx, client)
								res, err := client.Request(ctx, body.ActionName, body.Payload)

								if err != nil {
									// transport error with http request only
									r := map[string]any{
										"object": nil,
										"errors": []map[string]string{
											{
												"message": err.Error(),
											},
										},
									}

									WriteResponse(r, w)

									return
								}

								// for custom functions, we just want to return whatever response
								// shape is returned from the node handler
								WriteResponse(res, w)
							}
						}
					}
				}
			default:
				events := []*Event{}
				json.Unmarshal(b, &events)
				ch <- events
				w.Write([]byte("ok"))
			}
		}),
	}
	go srv.ListenAndServe()

	fs := os.DirFS(dir)

	// todo: test.ts files only
	testFiles, err := doublestar.Glob(fs, "**/*.test.ts")

	if err != nil {
		return nil, err
	}

	go func() {
		for _, file := range testFiles {
			if strings.Contains(file, "node_modules") {
				continue
			}

			db = testhelpers.SetupDatabaseForTestCase(t, schema, dbName)

			err := WrapTestFileWithShim(reportingPort, filepath.Join(dir, file))

			if err != nil {
				panic(err)
			}

			// We need to pass the skipIgnore flag to ts-node as by default
			// ts-node does not transpile stuff in node_modules
			// Given we are publishing a pure typescript module in the form of
			// @teamkeel/testing, we need ts-node to also process these files
			// ref: https://github.com/TypeStrong/ts-node#skipping-node_modules
			cmd := exec.Command("./node_modules/.bin/ts-node", "--skipIgnore", "--swc", file)
			cmd.Env = os.Environ()

			// The HOST_PORT is the port of the "reporting server" - the reporting server's job
			// is to receive messages from the node process about passing / failing tests
			cmd.Env = append(cmd.Env, fmt.Sprintf("HOST_PORT=%s", reportingPort))

			// We need to pass across the connection string to the database
			// so that slonik (query builder lib) can create a database pool which will be used
			// by the generated Query API code
			cmd.Env = append(cmd.Env, fmt.Sprintf("DB_CONN=%s", fmt.Sprintf(dbConnUri, dbName)))
			cmd.Env = append(cmd.Env, fmt.Sprintf("FORCE_COLOR=%d", 1))

			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err = cmd.Run()

			if err != nil {
				panic(err)
			}
		}

		// Cleanup
		srv.Close()
		if customFunctionRuntimeProcess != nil {
			customFunctionRuntimeProcess.Kill()
		}
		database.Stop()
		close(ch)
	}()

	return ch, nil
}

func RunServer(workingDir string, port string, parentPort string, dbConnectionString string) (*os.Process, error) {
	serverDistPath := filepath.Join(workingDir, "node_modules", "@teamkeel", "client", "dist", "handler.js")

	if _, err := os.Stat(serverDistPath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	cmd := exec.Command("node", filepath.Join("node_modules", "@teamkeel", "client", "dist", "handler.js"))

	cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%s", port))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DB_CONN=%s", dbConnectionString))
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOST_PORT=%s", parentPort))
	cmd.Env = append(cmd.Env, fmt.Sprintf("FORCE_COLOR=%d", 1))

	cmd.Dir = workingDir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()

	if err != nil {
		return nil, err
	}

	return cmd.Process, nil
}

func WriteResponse(data interface{}, w http.ResponseWriter) {
	b, err := json.Marshal(data)

	if err != nil {
		fmt.Print("banana")

		fmt.Print(err)
		panic(err)
	}

	w.Write(b)
}

func serializeError(err error) []map[string]string {
	if err == nil {
		return []map[string]string{}
	}

	return []map[string]string{
		{
			"message": err.Error(),
		},
	}
}
