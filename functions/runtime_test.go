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
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func TestAllCases(t *testing.T) {
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

			workingDir, err := testhelpers.WithTmpDir(testDir)

			require.NoError(t, err)

			t.Cleanup(func() {
				os.RemoveAll(workingDir)
			})

			runtime, err := functions.NewRuntime(schema, workingDir)
			require.NoError(t, err)

			err = runtime.Bootstrap()

			require.NoError(t, err)

			// Check that the whole project, including generated code, typechecks
			typecheckResult, output := typecheck(workingDir)

			assert.True(t, typecheckResult, output)

			port := 3002

			dbConnString := fmt.Sprintf(
				"postgresql://%s:%s@%s:%s/%s",
				"postgres",
				"postgres",
				"localhost",
				"8001",
				"keel",
			)
			// Runs the node. js server
			// the entry point will be {app}/node_modules/@teamkeel/client/dist/handler.js
			process, err := RunServer(workingDir, port, dbConnString)

			require.NoError(t, err)

			// Loop until we receive a 200 status from the node server
			// If there is never a 200, then the test will timeout after prescribed timeout period, and fail
			for {
				time.Sleep(time.Second / 2)

				expected := map[string]string{
					"id":    "123",
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
					assert.Equal(t, expected["id"], actual.ID)

					// Kill the node server after assertion is successful
					process.Kill()

					break
				}
			}
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

func RunServer(workingDir string, port int, dbConnString string) (*os.Process, error) {
	serverDistPath := filepath.Join(workingDir, "node_modules", "@teamkeel", "client", "dist", "handler.js")

	if _, err := os.Stat(serverDistPath); errors.Is(err, os.ErrNotExist) {
		fmt.Print(err)
		return nil, err
	}

	cmd := exec.Command("node", filepath.Join("node_modules", "@teamkeel", "client", "dist", "handler.js"))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", port))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DB_CONN=%s", dbConnString))
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
