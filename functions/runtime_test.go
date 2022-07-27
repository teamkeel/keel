package functions_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/functions"
)

type PostResponse struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type Response struct {
	Post PostResponse `json:"post"`
}

func TestAllCases(t *testing.T) {
	testCases, err := ioutil.ReadDir("runtime_testdata")
	require.NoError(t, err)

	toRun := []fs.FileInfo{}

	for _, testCase := range testCases {
		if strings.HasSuffix(testCase.Name(), ".only") {
			toRun = append(toRun, testCase)
		}
	}

	if len(toRun) > 0 {
		testCases = toRun
	}

	for _, testCase := range testCases {
		if !testCase.IsDir() {
			continue
		}

		t.Run(testCase.Name(), func(t *testing.T) {
			// The base working directory - in this case, the test case directory
			workingDir := filepath.Join("runtime_testdata", testCase.Name())

			runtime, err := functions.NewRuntime(workingDir)
			require.NoError(t, err)

			err = runtime.ReconcilePackageJson()
			require.NoError(t, err)

			err = runtime.GenerateClient()

			require.NoError(t, err)

			err = runtime.GenerateHandler()

			require.NoError(t, err)

			typecheckResult, output := typecheck(workingDir)

			assert.True(t, typecheckResult, output)

			_, errs := runtime.Bundle(true)

			require.Len(t, errs, 0)

			port := 3002

			err = runtime.RunServer(port, func(p *os.Process) {
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
						fmt.Print("status is 200")

						body := Response{}
						err := json.Unmarshal(b, &body)

						if err != nil {
							assert.Fail(t, "Could not unmarshal JSON response from node server")
						}

						actual := body.Post

						assert.Equal(t, expected["title"], actual.Title)

						// Kill the node server after assertion is successful
						p.Kill()

						break
					}
				}
			})

			assert.NoError(t, err)
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
