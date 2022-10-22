package testing

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

const (
	StatusPass      = "pass"
	StatusFail      = "fail"
	StatusException = "exception"
)

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

type ResetRequest struct {
	FilePath string `json:"filePath"`
	TestCase string `json:"testCase"`
}

//go:embed tsconfig.json
var sampleTsConfig string

type RunType = string

const (
	RunTypeIntegration = "integration"
	RunTypeTestCmd     = "testCmd"
)

func Run(dir string, pattern string, runType RunType) (<-chan []*Event, error) {
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

	output, err := typecheck(dir, runType)

	if err != nil {
		fmt.Print(output)
		return nil, err
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

				identityIdHeader := r.Header.Get("identityId")
				identityId, _ := ksuid.Parse(identityIdHeader)

				for _, model := range schema.Models {
					for _, action := range model.Operations {
						if action.Name == body.ActionName {

							if action.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO {
								ctx := r.Context()
								ctx = runtimectx.WithDatabase(ctx, db)
								if identityId != ksuid.Nil {
									ctx = runtimectx.WithIdentity(ctx, &identityId)
								}

								scope, err := actions.NewScope(ctx, action, schema)

								if err != nil {
									panic(err)
								}
								switch action.Type {
								case proto.OperationType_OPERATION_TYPE_GET:

									var builder actions.GetAction
									body.Payload, err = toNativeMap(body.Payload, action)

									result, err := builder.
										Initialise(scope).
										ApplyImplicitFilters(body.Payload).
										ApplyExplicitFilters(body.Payload).
										IsAuthorised(body.Payload).
										Execute(body.Payload)

									r := map[string]any{
										"object": nil,
										"errors": serializeError(err),
									}

									if result != nil {
										r["object"] = result.Value.Object
									}

									writeResponse(r, w)
								case proto.OperationType_OPERATION_TYPE_CREATE:
									var builder actions.CreateAction
									body.Payload, err = toNativeMap(body.Payload, action)

									result, err := builder.
										Initialise(scope).
										CaptureImplicitWriteInputValues(body.Payload). // todo: err?
										CaptureSetValues(body.Payload).
										IsAuthorised(body.Payload).
										Execute(body.Payload)

									r := map[string]any{
										"object": nil,
										"errors": serializeError(err),
									}

									if result != nil {
										r["object"] = result.Value.Object
									}

									writeResponse(r, w)
								case proto.OperationType_OPERATION_TYPE_UPDATE:
									var builder actions.UpdateAction

									if err != nil {
										panic(err)
									}
									body.Payload, err = toNativeMap(body.Payload, action)

									// toArgsMap covers if the key isnt present
									// if this is the case, an empty map will be returned
									values, err := toArgsMap(body.Payload, "values", true)
									if err != nil {
										panic(err)
									}

									wheres, err := toArgsMap(body.Payload, "where", false)
									if err != nil {
										panic(err)
									}

									result, err := builder.
										Initialise(scope).
										// first capture any implicit inputs
										CaptureImplicitWriteInputValues(values).
										// then capture explicitly used inputs
										CaptureSetValues(values).
										// then apply unique filters
										ApplyImplicitFilters(wheres).
										ApplyExplicitFilters(wheres).
										IsAuthorised(body.Payload).
										Execute(body.Payload)

									r := map[string]any{
										"object": nil,
										"errors": serializeError(err),
									}

									if result != nil {
										r["object"] = result.Value.Object
									}

									writeResponse(r, w)
								case proto.OperationType_OPERATION_TYPE_LIST:
									var builder actions.ListAction

									// todo: body.Payload is currently just the wheres
									// needs to follow this structure:
									// where: {},
									// pageInfo: {}

									args := map[string]any{
										"where": body.Payload,
									}

									body.Payload, err = toNativeMap(body.Payload, action)

									result, err := builder.
										Initialise(scope).
										ApplyImplicitFilters(body.Payload).
										ApplyExplicitFilters(body.Payload).
										IsAuthorised(args).
										Execute(args)

									r := map[string]any{
										"errors": serializeError(err),
									}

									if result != nil {
										r["collection"] = result.Value.Collection
										r["hasNextPage"] = result.Value.HasNextPage
									}

									writeResponse(r, w)
								case proto.OperationType_OPERATION_TYPE_DELETE:
									var builder actions.DeleteAction
									body.Payload, err = toNativeMap(body.Payload, action)

									result, err := builder.
										Initialise(scope).
										ApplyImplicitFilters(body.Payload).
										ApplyExplicitFilters(body.Payload).
										IsAuthorised(body.Payload).
										Execute(body.Payload)

									r := map[string]any{
										"errors": serializeError(err),
									}

									if result != nil {
										r["success"] = result.Value.Success
									}

									writeResponse(r, w)
								case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
									authArgs := actions.AuthenticateArgs{
										CreateIfNotExists: body.Payload["createIfNotExists"].(bool),
										Email:             body.Payload["email"].(string),
										Password:          body.Payload["password"].(string),
									}

									identityId, identityCreated, err := actions.Authenticate(ctx, schema, &authArgs)

									// todo: this doesn't nicely match what the user might expect to be returned from authenticate()
									// do we refactor this to return a token?
									var r map[string]any
									if identityId == nil {
										r = map[string]any{
											"identityId":      nil,
											"identityCreated": identityCreated,
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
								res, err := client.Request(ctx, body.ActionName, action.Type, body.Payload)

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

									writeResponse(r, w)

									return
								}

								// for custom functions, we just want to return whatever response
								// shape is returned from the node handler
								writeResponse(res, w)
							}
						}
					}
				}
			case "/report":
				events := []*Event{}
				json.Unmarshal(b, &events)
				ch <- events
				w.Write([]byte("ok"))

			case "/reset":
				resetRequestBody := &ResetRequest{}
				err = json.Unmarshal(b, &resetRequestBody)
				if err != nil {
					panic(err)
				}
				db, err = testhelpers.SetupDatabaseForTestCase(schema, dbName)
				if err != nil {
					panic(err)
				}
				w.Write([]byte("ok"))
			}
		}),
	}
	go srv.ListenAndServe()

	fs := os.DirFS(dir)

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

func typecheck(dir string, runType RunType) (output string, err error) {
	// todo: we need to generate a tsconfig to be able to run tsc for typechecking
	// however, when we come to use the testing package in real projects, there may already
	// be a tsconfig file that we need to respect

	// switch runType {
	// case RunTypeIntegration:
	// case RunTypeTestCmd:
	// default:
	// 	panic("unrecognised run type")
	// }
	f, err := os.Create(filepath.Join(dir, "tsconfig.json"))

	if err != nil {
		return "", err
	}

	f.WriteString(sampleTsConfig)

	defer f.Close()
	cmd := exec.Command("npx", "tsc", "--noEmit", "--skipLibCheck", "--project", filepath.Base(f.Name()))
	cmd.Dir = dir

	b, e := cmd.CombinedOutput()

	if e != nil {
		err = e
	}

	return string(b), err
}

func toArgsMap(input map[string]any, key string, defaultToEmpty bool) (map[string]any, error) {
	subKey, ok := input[key]

	if !ok {
		if defaultToEmpty {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("%s missing", key)
	}

	subMap, ok := subKey.(map[string]any)

	if !ok {
		return nil, fmt.Errorf("%s does not match expected format", key)
	}

	return subMap, nil
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
