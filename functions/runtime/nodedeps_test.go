package runtime_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/functions/runtime"
)

var TestCases = []string{
	"non_existent_package_json",
}

func TestAllTestCases(t *testing.T) {
	testCases, err := ioutil.ReadDir("testdata")

	require.NoError(t, err)

	for _, testCase := range testCases {
		if !lo.Contains(TestCases, testCase.Name()) {
			continue
		}

		t.Run(testCase.Name(), func(t *testing.T) {

			workingDir := filepath.Join("testdata", testCase.Name())
			outDir := filepath.Join("testdata", testCase.Name(), "tmp")

			runtime, err := runtime.NewRuntime(workingDir, outDir)

			require.NoError(t, err)

			err = runtime.BootstrapPackageJson()

			require.NoError(t, err)

			packageJsonPath := filepath.Join(workingDir, "package.json")

			if _, err := os.Stat(packageJsonPath); errors.Is(err, os.ErrNotExist) {
				assert.Fail(t, "package.json not created")
			}
		})
	}
}
