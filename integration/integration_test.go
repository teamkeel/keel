package integration_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	gotest "testing"

	cp "github.com/otiai10/copy"

	"github.com/alexflint/go-restructure/regex"
	"github.com/nsf/jsondiff"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/cmd"
	"github.com/teamkeel/keel/nodedeps"
	"github.com/teamkeel/keel/testing"
)

var pattern = flag.String("pattern", "", "Pattern to match individual test case names")

type Assertion struct {
	TestName string `json:"testName"`
	Status   string `json:"status"`
	Actual   any    `json:"actual,omitempty"`
	Expected any    `json:"expected,omitempty"`
}

func TestIntegration(t *gotest.T) {
	entries, err := os.ReadDir("./testdata")
	require.NoError(t, err)

	allResults := []*testing.TestResult{}

	// Make a temp dir for all tests to run in (each tests files will be copied to this dir)
	tmpDir, err := os.MkdirTemp("", t.Name())
	require.NoError(t, err)
	t.Cleanup(func() {
		// Remove it when done
		os.RemoveAll(tmpDir)
	})

	packageJson, err := nodedeps.NewPackageJson(filepath.Join(tmpDir, "package.json"))
	require.NoError(t, err)

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

	// Whatever files/dirs are present now can stay between tests
	// e.g. node_modules, package.json
	genericEntries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	for _, e := range entries {
		t.Run(e.Name(), func(t *gotest.T) {
			testDir := filepath.Join("./testdata", e.Name())

			// Copy test files to temp dir
			require.NoError(t, cp.Copy(testDir, tmpDir))

			// At the end of this tests remove all the test files
			t.Cleanup(func() {
				entries, err := os.ReadDir(tmpDir)
				require.NoError(t, err)
			outer:
				for _, entry := range entries {
					for _, g := range genericEntries {
						if g.Name() == entry.Name() {
							continue outer
						}
					}
					os.RemoveAll(filepath.Join(tmpDir, entry.Name()))
				}
			})

			ch, err := testing.Run(tmpDir, *pattern)
			require.NoError(t, err)

			results := []*testing.TestResult{}
			for newEvents := range ch {
				resultEvents := lo.Filter(newEvents, func(e *testing.Event, _ int) bool {
					return e.EventStatus == testing.EventStatusComplete
				})

				for _, e := range resultEvents {
					results = append(results, e.Result)
				}
			}

			allResults = append(allResults, results...)

			actual := []*Assertion{}

			for _, r := range results {
				assertion := &Assertion{
					TestName: r.TestName,
					Status:   r.Status,
				}

				if r.Expected != nil {
					assertion.Expected = r.Expected
				}
				if r.Actual != nil {
					assertion.Actual = r.Actual
				}

				actual = append(actual, assertion)
			}

			a, err := json.Marshal(actual)
			require.NoError(t, err)
			expected, err := os.ReadFile(filepath.Join("./testdata", e.Name(), "expected.json"))
			require.NoError(t, err)

			if pattern != nil && *pattern != "" {
				// subset of tests

				allExpected := []*Assertion{}

				err := json.Unmarshal(expected, &allExpected)

				require.NoError(t, err)

				filteredExpected := []*Assertion{}

				for _, e := range allExpected {
					match, _ := regex.MatchString(fmt.Sprintf("^%s$", *pattern), e.TestName)

					if match {
						filteredExpected = append(filteredExpected, e)
					}
				}

				b, err := json.Marshal(filteredExpected)

				require.NoError(t, err)

				CompareJson(t, b, a)
			} else {
				CompareJson(t, expected, a)
			}
		})
	}

	cmd.PrintSummary(allResults)
}

func CompareJson(t *gotest.T, expected []byte, actual []byte) {
	opts := jsondiff.DefaultConsoleOptions()

	diff, explanation := jsondiff.Compare(expected, actual, &opts)
	if diff == jsondiff.FullMatch {
		return
	}

	assert.Fail(t, "actual test output did not match expected", explanation)
}
