package integration_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	gotest "testing"

	"github.com/joho/godotenv"
	"github.com/nsf/jsondiff"
	"github.com/teamkeel/keel/nodedeps"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *gotest.T) {

	testToTarget := readTestToTargetFromEnv()

	entries, err := ioutil.ReadDir("./testdata")
	require.NoError(t, err)

	for _, e := range entries {
		entryName := e.Name()
		_ = entryName
		if testToTarget != nil && e.Name() != testToTarget.Directory {
			continue
		}

		fmt.Printf("XXXX targeting %s because %+v\n", entryName, testToTarget)

		t.Run(e.Name(), func(t *gotest.T) {
			workingDir, err := testhelpers.WithTmpDir(filepath.Join("./testdata", e.Name()))

			require.NoError(t, err)

			fmt.Println("TEST DIRECTORY:")
			fmt.Println(workingDir)

			packageJson, err := nodedeps.NewPackageJson(filepath.Join(workingDir, "package.json"))

			require.NoError(t, err)

			// todo: to save time during test suite run, do the package.json creation
			// plus injection only once per suite.
			err = packageJson.Inject(map[string]string{
				"@teamkeel/testing": "*",
				"@teamkeel/sdk":     "*",
				"@teamkeel/runtime": "*",
				"ts-node":           "*",
				// https://typestrong.org/ts-node/docs/swc/
				"@swc/core":           "*",
				"regenerator-runtime": "*",
			}, true)

			require.NoError(t, err)

			ch, err := testing.Run(t, workingDir)
			require.NoError(t, err)

			events := []*testing.Event{}
			for newEvents := range ch {
				events = append(events, newEvents...)
			}

			actual, err := json.Marshal(events)
			require.NoError(t, err)

			expected, err := ioutil.ReadFile(filepath.Join("./testdata", e.Name(), "expected.json"))
			require.NoError(t, err)

			opts := jsondiff.DefaultConsoleOptions()

			diff, explanation := jsondiff.Compare(expected, actual, &opts)
			if diff == jsondiff.FullMatch {
				return
			}

			assert.Fail(t, "actual test output did not match expected", explanation)
		})
	}
}

// TestToTarget models a test that should be targeted (i.e. isolated).
type TestToTarget struct {
	Directory string
	File      string
	CaseName  string
}

// readTestToTargetFromEnv provides either nil, or a well-formed TestToTarget object
// depending on the value of the KEEL_TARGET_TEST environment variable.
func readTestToTargetFromEnv() *TestToTarget {
	godotenv.Load()
	conf := os.Getenv("KEEL_TARGET_TEST")
	if conf == "" {
		return nil
	}
	segments := strings.Split(conf, "/")
	if len(segments) != 3 {
		panic(fmt.Sprintf("your KEEL_TARGET_TEST env var value: %s must have 3 slash-delimited segments", conf))
	}
	return &TestToTarget{
		Directory: segments[0],
		File:      segments[1],
		CaseName:  segments[2],
	}
}
