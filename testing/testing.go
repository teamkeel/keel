package testing

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/cmd/database"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
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

const (
	StatusPass      = "pass"
	StatusFail      = "fail"
	StatusException = "exception"
)

type EventStatus string

const (
	EventStatusPending  EventStatus = "pending"
	EventStatusComplete EventStatus = "complete"
)

type Event struct {
	EventStatus EventStatus

	Result *TestResult
	Meta   *TestCase
}

type TestCase struct {
	TestName string `json:"name"`
	FilePath string `json:"filePath"`
}

type TestResult struct {
	TestCase

	Status   string          `json:"status"`
	Expected json.RawMessage `json:"expected,omitempty"`
	Actual   json.RawMessage `json:"actual,omitempty"`
	Err      json.RawMessage `json:"err,omitempty"`
}

type ActionRequest struct {
	ActionName string         `json:"actionName"`
	Payload    map[string]any `json:"payload"`
}

type ResetRequest struct {
	FilePath string `json:"filePath"`
	TestCase string `json:"testCase"`
}

//go:embed tsconfig.json
var sampleTsConfig string

func Run(dir string, pattern string) (chan []*Event, error) {
	builder := &schema.Builder{}
	shortDir := filepath.Base(dir)
	dbName := testhelpers.DbNameForTestName(shortDir)

	var db *gorm.DB
	var testProcess *exec.Cmd

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

	output, err := typecheck(dir)

	if err != nil {
		fmt.Print(output)
		return nil, err
	}

	expectedResults := 0

	// Server for node test process to talk to
	srv := http.Server{
		Addr: fmt.Sprintf(":%s", reportingPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)

			switch r.URL.Path {
			case "/action":
				body := &ActionRequest{}
				json.Unmarshal(b, body)

				identityIdHeader := r.Header.Get("identityId")
				identityId, _ := ksuid.Parse(identityIdHeader)

				client := &functions.HttpFunctionsClient{
					Port: customFunctionRuntimePort,
					Host: "0.0.0.0",
				}

				for _, model := range schema.Models {
					for _, operation := range model.Operations {
						if operation.Name == body.ActionName {
							ctx := r.Context()
							ctx = runtimectx.WithDatabase(ctx, db)
							ctx = functions.WithFunctionsClient(ctx, client)

							if identityId != ksuid.Nil {
								ctx = runtimectx.WithIdentity(ctx, &identityId)
							}

							argParser := &IntegrationTestArgParser{}
							scope, err := actions.NewScope(ctx, operation, schema)

							if err != nil {
								panic(err)
							}
							switch operation.Type {
							case proto.OperationType_OPERATION_TYPE_GET:
								args, err := argParser.ParseGet(operation, body.Payload)
								if err != nil {
									panic(err)
								}

								result, err := scope.Get(args)

								r := map[string]any{
									"object": nil,
									"errors": serializeError(err),
								}

								if result != nil {
									r["object"] = result.Object
								}

								writeResponse(r, w)
							case proto.OperationType_OPERATION_TYPE_CREATE:
								args, err := argParser.ParseCreate(operation, body.Payload)
								if err != nil {
									panic(err)
								}

								result, err := scope.Create(args)

								r := map[string]any{
									"object": nil,
									"errors": serializeError(err),
								}

								if result != nil {
									r["object"] = result.Object
								}

								writeResponse(r, w)
							case proto.OperationType_OPERATION_TYPE_UPDATE:
								args, err := argParser.ParseUpdate(operation, body.Payload)
								if err != nil {
									panic(err)
								}

								result, err := scope.Update(args)

								r := map[string]any{
									"object": nil,
									"errors": serializeError(err),
								}

								if result != nil {
									r["object"] = result.Object
								}

								writeResponse(r, w)
							case proto.OperationType_OPERATION_TYPE_LIST:
								args, err := argParser.ParseList(operation, body.Payload)
								if err != nil {
									panic(err)
								}

								result, err := scope.List(args)

								r := map[string]any{
									"errors": serializeError(err),
								}

								if result != nil {
									r["collection"] = result.Collection
									r["hasNextPage"] = result.HasNextPage
								}

								writeResponse(r, w)
							case proto.OperationType_OPERATION_TYPE_DELETE:
								args, err := argParser.ParseDelete(operation, body.Payload)
								if err != nil {
									panic(err)
								}

								result, err := scope.Delete(args)

								r := map[string]any{
									"errors": serializeError(err),
								}

								if result != nil {
									r["success"] = result.Success
								}

								writeResponse(r, w)
							case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
								authArgs := actions.AuthenticateArgs{
									CreateIfNotExists: body.Payload["createIfNotExists"].(bool),
									Email:             body.Payload["email"].(string),
									Password:          body.Payload["password"].(string),
								}

								token, identityCreated, err := actions.Authenticate(ctx, schema, &authArgs)

								var identityId *ksuid.KSUID
								if err == nil {
									identityId, err = actions.ParseBearerToken(token)
								}

								var r map[string]any
								if identityId == nil {
									r = map[string]any{
										"identityId":      nil,
										"identityCreated": false,
										"errors":          serializeError(err),
									}
								} else {
									r = map[string]any{
										"identityId":      identityId.String(),
										"identityCreated": identityCreated,
										"errors":          serializeError(err),
									}
								}

								writeResponse(r, w)
							default:
								w.WriteHeader(400)
								panic(fmt.Sprintf("%s not yet implemented", operation.Type))
							}

							return
						}
					}
				}
			case "/report":
				result := []*TestResult{}
				json.Unmarshal(b, &result)

				ch <- []*Event{{EventStatus: EventStatusComplete, Result: result[0]}}
				w.Write([]byte("ok"))
				expectedResults--
				if expectedResults == 0 {
					// Now that all tests have been reported on we can kill the node process running the tests
					// TODO: we shouldn't really need to do this but it seems the process hangs on for 10s or
					// so after all tests have run. Possibly to do with an open database connection?
					testProcess.Process.Kill()
				}
			case "/collect":
				cases := []*TestCase{}
				json.Unmarshal(b, &cases)
				expectedResults = len(cases)
				for _, tc := range cases {
					ch <- []*Event{{EventStatus: EventStatusPending, Meta: tc}}
				}
				w.Write([]byte("ok"))
			case "/reset":
				resetRequestBody := &ResetRequest{}
				err = json.Unmarshal(b, &resetRequestBody)
				if err != nil {
					panic(err)
				}
				if db == nil {
					db, err = testhelpers.SetupDatabaseForTestCase(schema, dbName)
				} else {
					err = testhelpers.TruncateTables(db)
				}
				if err != nil {
					panic(err)
				}
				w.Write([]byte("ok"))
			}
		}),
	}
	go srv.ListenAndServe()

	fs := os.DirFS(dir)

	// go func() {
	// 	for {
	// 		val := <-ch
	// 		fmt.Println(val)
	// 	}
	// }()

	testFiles, err := doublestar.Glob(fs, "**/*.test.ts")

	if err != nil {
		return nil, err
	}

	go func() {
		for _, file := range testFiles {
			if strings.Contains(file, "node_modules") {
				continue
			}

			err := WrapTestFileWithShim(reportingPort, filepath.Join(dir, file), pattern)

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
			cmd.Env = append(cmd.Env, "DB_CONN_TYPE=pg")
			cmd.Env = append(cmd.Env, fmt.Sprintf("FORCE_COLOR=%d", 1))

			cmd.Dir = dir
			// cmd.Stdout = os.Stdout
			// cmd.Stderr = os.Stderr

			testProcess = cmd

			err = cmd.Run()

			if err != nil && err.Error() != "signal: killed" {
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
	cmd.Env = append(cmd.Env, "DB_CONN_TYPE=pg")
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

func writeResponse(data interface{}, w http.ResponseWriter) {
	b, err := json.Marshal(data)

	if err != nil {
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

func typecheck(dir string) (output string, err error) {
	// todo: we need to generate a tsconfig to be able to run tsc for typechecking
	// however, when we come to use the testing package in real projects, there may already
	// be a tsconfig file that we need to respect

	f, err := os.Create(filepath.Join(dir, "tsconfig.json"))

	if err != nil {
		return "", err
	}

	f.WriteString(sampleTsConfig)

	defer f.Close()
	cmd := exec.Command("npx", "tsc", "--noEmit", "--skipLibCheck", "--incremental", "--project", filepath.Base(f.Name()))
	cmd.Dir = dir

	b, e := cmd.CombinedOutput()

	if e != nil {
		err = e
	}

	return string(b), err
}

// Converts the input args (JSON) sent from the JavaScript process
// into a format that the actions code understands.
// Dates in JSON will come in as strings in ISO8601 format whereas
// the actions code expects time.Time
// In the future, this method can be extended to handle other conversions
// or refactored completely into a deserializer type pattern, shared with
// the graphql code
func toNativeMap(args map[string]interface{}, action *proto.Operation) (map[string]any, error) {
	out := map[string]any{}

	for _, input := range action.Inputs {
		match, ok := args[input.Name]

		if ok {
			inputType := input.Type.Type

			switch inputType {
			case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP, proto.Type_TYPE_DATE:
				str, ok := match.(string)

				if !ok {
					return nil, fmt.Errorf("%s arg with value %v is not a string", input.Name, match)
				}

				time, err := time.Parse("2006-01-02T15:04:05-0700", str)
				if err != nil {
					return nil, fmt.Errorf("%s is not ISO8601 formatted date: %s", input.Name, str)
				}

				out[input.Name] = time
			default:
				out[input.Name] = match
			}
		}
	}

	return out, nil
}
