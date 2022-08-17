package testing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/util"
)

const (
	EventTypeTestRun = "TestRun"
)

var TestIgnorePatterns []string = []string{"node_modules"}

type Event struct {
	Status   string          `json:"status"`
	TestName string          `json:"testName"`
	Expected json.RawMessage `json:"expected,omitempty"`
	Actual   json.RawMessage `json:"actual,omitempty"`
}

func Run(dir string) (<-chan []*Event, error) {
	builder := &schema.Builder{}

	schema, err := builder.MakeFromDirectory(dir)
	if err != nil {
		return nil, err
	}

	ch := make(chan []*Event)

	freePort, err := util.GetFreePort()

	if err != nil {
		return nil, err
	}

	// TODO: Generate custom function lib
	// TODO: Generate testing lib
	// TODO: Start database using cmd/database.Start(true)
	// TODO: Run migrations
	// TODO: Start custom functions node process (if any custom functions defined)

	// Server for node test process to talk to
	srv := http.Server{
		Addr: fmt.Sprintf(":%s", freePort),
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

			err := WrapTestFileWithShim(freePort, filepath.Join(dir, file))

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
			cmd.Env = append(cmd.Env, fmt.Sprintf("HOST_PORT=%s", freePort))
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err = cmd.Run()

			if err != nil {
				panic(err)
			}
		}

		srv.Close()
		close(ch)
	}()

	return ch, nil
}
