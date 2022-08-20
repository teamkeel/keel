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
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/util"
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

func Run(t *testing.T, dir string) (<-chan []*Event, error) {
	builder := &schema.Builder{}
	shortDir := filepath.Base(dir)
	dbName := testhelpers.DbNameForTestName(shortDir)

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

	if hasCustomFunctions {
		customFunctionRuntimePort, err := util.GetFreePort()

		if err != nil {
			panic(err)
		}

		customFunctionRuntimeProcess, err = RunServer(dir, customFunctionRuntimePort, fmt.Sprintf(dbConnUri, dbName))

		if err != nil {
			panic(err)
		}
	}

	injector := NewInjector(dir, schema)

	err = injector.Inject()

	if err != nil {
		panic(err)
	}

	// TODO: Start custom functions node process (if any custom functions defined)

	// Server for node test process to talk to
	srv := http.Server{
		Addr: fmt.Sprintf(":%s", reportingPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := ioutil.ReadAll(r.Body)

			switch r.URL.Path {
			case "/action":
				var req map[string]any
				json.Unmarshal(b, &req)

				actionName := ""
				for _, model := range schema.Models {
					for _, action := range model.Operations {
						if action.Name == actionName {
							if action.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO {
								switch action.Type {
								case proto.OperationType_OPERATION_TYPE_GET:
									ctx := r.Context()
									ctx = runtimectx.WithDatabase(ctx, nil)
									actions.Get(ctx, action, schema, req)
								}
							}
							if action.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
								// call node process
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

			fmt.Println(dir)

			_ = testhelpers.SetupDatabaseForTestCase(t, schema, dbName)

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

			cmd.Env = append(cmd.Env, "TS_NODE_TRANSPILE_ONLY=true")
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

func RunServer(workingDir string, port string, dbConnectionString string) (*os.Process, error) {
	serverDistPath := filepath.Join(workingDir, "node_modules", "@teamkeel", "client", "dist", "handler.js")

	if _, err := os.Stat(serverDistPath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	cmd := exec.Command("node", filepath.Join("node_modules", "@teamkeel", "client", "dist", "handler.js"))

	cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%s", port))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DB_CONN=%s", dbConnectionString))
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
