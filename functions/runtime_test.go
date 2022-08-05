package functions_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
)

type PostResponse struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func TestAllCases(t *testing.T) {
	// todo: reinstate
	t.Skip()

	testCases, err := ioutil.ReadDir("runtime_testdata")
	require.NoError(t, err)

	for _, testCase := range testCases {
		if !testCase.IsDir() {
			continue
		}

		t.Run(testCase.Name(), func(t *testing.T) {
			// The base working directory - in this case, the test case directory
			testDir := filepath.Join("runtime_testdata", testCase.Name())
			builder := schema.Builder{}

			schema, err := builder.MakeFromDirectory(testDir)
			require.NoError(t, err)

			testhelpers.WithTmpDir(testDir, func(workingDir string) {
				runtime, err := functions.NewRuntime(schema, workingDir)
				require.NoError(t, err)

				// Checks if the correct dependencies are listed in the target app's package.json
				err = runtime.ReconcilePackageJson()
				require.NoError(t, err)

				// Generates client code files (typescript)
				// output path will be {app}/node_modules/@teamkeel/client/src/index.ts
				err = runtime.GenerateClient()

				require.NoError(t, err)

				// Generates runtime handler code (typescript)
				// output path will be {app}/node_modules/@teamkeel/client/src/handler.ts
				err = runtime.GenerateHandler()

				require.NoError(t, err)

				// Generates a package.json file in the ephemeral @teamkeel/client package
				// required for resolution from other @teamkeel npm modules
				err = runtime.GenerateClientPackageJson()

				require.NoError(t, err)

				// Check that the whole project, including generated code, typechecks
				typecheckResult, output := typecheck(workingDir)

				assert.True(t, typecheckResult, output)

				// Bundle all of the generated typescript code in @teamkeel/client
				// necessary to run the node server
				_, errs := runtime.Bundle(true)

				require.Len(t, errs, 0)

				port := 3002

				// Runs the node. js server
				// the entry point will be {app}/node_modules/@teamkeel/client/dist/handler.js
				process, err := RunServer(workingDir, port)

				require.NoError(t, err)

				// Loop until we receive a 200 status from the node server
				// If there is never a 200, then the test will timeout after prescribed timeout period, and fail
				for {
					time.Sleep(time.Second / 2)

					expected := map[string]string{
						"title": "something",
					}

					j, err := json.Marshal(expected)

					if err != nil {
						panic(err)
					}

					res, err := http.Post(fmt.Sprintf("http://localhost:%d/createPost", port), "application/json", bytes.NewBuffer(j))

					if err != nil {
						panic(err)
					}

					defer res.Body.Close()

					b, err := io.ReadAll(res.Body)

					if err != nil {
						panic(err)
					}

					if res.StatusCode == 200 {
						body := PostResponse{}
						err := json.Unmarshal(b, &body)

						if err != nil {
							assert.Fail(t, "Could not unmarshal JSON response from node server")
						}

						actual := body

						assert.Equal(t, expected["title"], actual.Title)

						// Kill the node server after assertion is successful
						process.Kill()

						break
					}
				}
			})
		})
	}
}

// Runs tsc against a tsconfig.json located in a particular directory
// returns bool, stdout string
func typecheck(workingDir string) (bool, string) {
	command := exec.Command("node_modules/.bin/tsc", "-p", "tsconfig.json", "--noEmit")
	command.Dir = workingDir
	outputBytes, err := command.CombinedOutput()

	if err != nil {
		return false, string(outputBytes)
	}

	return true, string(outputBytes)
}

func RunServer(workingDir string, port int) (*os.Process, error) {
	serverDistPath := filepath.Join(workingDir, "node_modules", "@teamkeel", "client", "dist", "handler.js")

	if _, err := os.Stat(serverDistPath); errors.Is(err, os.ErrNotExist) {
		fmt.Print(err)
		return nil, err
	}

	cmd := exec.Command("node", serverDistPath)
	cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", port))
	cmd.Dir = workingDir

	var buf bytes.Buffer
	w := io.MultiWriter(os.Stdout, &buf)

	cmd.Stdout = w
	cmd.Stderr = w

	err := cmd.Start()

	if err != nil {
		return nil, err
	}

	return cmd.Process, nil
}
