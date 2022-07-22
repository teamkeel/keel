package runtime_test

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
	"github.com/teamkeel/keel/functions/runtime"
)

// Tests:
// - Run tsc against generated "whole thing"
// - Run code with node.js
// - Throw some requests at the server and assert that the right response is returned

type PostResponse struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type Response struct {
	Post PostResponse `json:"post"`
}

func TestAllCases(t *testing.T) {
	testCases, err := ioutil.ReadDir("testdata")
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
			workingDir := filepath.Join("testdata", testCase.Name())

			// Directory where the generated typescript code will be placed
			outDir := filepath.Join("testdata", testCase.Name(), runtime.DEV_DIRECTORY)

			// We need to install deps specified in the test case's package.json
			// as these are not installed by default.
			npmInstall := exec.Command("npm", "install")
			npmInstall.Dir = workingDir

			// .Run() waits for the npm install command to complete
			npmInstall.Run()

			runtime, err := runtime.NewRuntime(workingDir, outDir)
			require.NoError(t, err)

			_, err = runtime.Generate()
			require.NoError(t, err)

			typecheckResult, output := typecheck(workingDir)

			assert.True(t, typecheckResult, output)

			errs := runtime.Bundle(true)

			require.Len(t, errs, 0)

			port := 3001
			err = runtime.RunServer(port, func(p *os.Process) {
				for {

					time.Sleep(time.Second / 2)

					values := map[string]string{
						"title": "something",
					}
					j, err := json.Marshal(values)

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
							t.Fail()
						}
						assert.Equal(t, "something", body.Post.Title)

						p.Kill()
						break
					}
				}
			})

			assert.NoError(t, err)
		})
	}
}

func typecheck(workingDir string) (bool, string) {
	command := exec.Command("tsc", "-p", "tsconfig.json", "--noEmit")
	command.Dir = workingDir
	outputBytes, err := command.Output()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode() == 0, string(outputBytes)
		}
	}
	return true, string(outputBytes)
}
