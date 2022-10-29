package integration_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	gotest "testing"

	"github.com/alexflint/go-restructure/regex"
	"github.com/nsf/jsondiff"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/cmd"
	"github.com/teamkeel/keel/nodedeps"
	"github.com/teamkeel/keel/testhelpers"
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

	for _, e := range entries {
		t.Run(e.Name(), func(t *gotest.T) {
			workingDir, err := testhelpers.WithTmpDir(filepath.Join("./testdata", e.Name()))
			fmt.Println(workingDir)

			require.NoError(t, err)

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

			ch, err := testing.Run(workingDir, *pattern)
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
