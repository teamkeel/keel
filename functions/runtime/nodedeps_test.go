package runtime_test

import (
	"encoding/json"
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
	"existing_package_json",
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

			r, err := runtime.NewRuntime(workingDir, outDir)

			require.NoError(t, err)

			_, err = r.Generate()

			require.NoError(t, err)

			err = r.BootstrapPackageJson()

			require.NoError(t, err)

			packageJsonPath := filepath.Join(workingDir, "package.json")

			if _, err := os.Stat(packageJsonPath); errors.Is(err, os.ErrNotExist) {
				assert.Fail(t, "package.json not created")
			}

			b, err := os.ReadFile(packageJsonPath)

			require.NoError(t, err)

			packageJsonContents := map[string]interface{}{}

			err = json.Unmarshal(b, &packageJsonContents)

			require.NoError(t, err)

			devDeps, ok := packageJsonContents["devDependencies"].(map[string]interface{})

			if !ok {
				assert.Fail(t, "devDeps not in expected format")
			}

			assert.ObjectsAreEqual(devDeps, runtime.DEV_DEPENDENCIES)
		})
	}
}
