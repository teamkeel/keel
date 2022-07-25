package nodedeps_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/functions/runtime/nodedeps"
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
			packageJsonPath := filepath.Join(workingDir, "package.json")

			packageJson, err := nodedeps.NewPackageJson(packageJsonPath)

			require.NoError(t, err)

			err = packageJson.Bootstrap()

			require.NoError(t, err)

			b, err := os.ReadFile(packageJsonPath)

			require.NoError(t, err)

			packageJsonContents := map[string]interface{}{}

			err = json.Unmarshal(b, &packageJsonContents)

			require.NoError(t, err)

			devDeps, ok := packageJsonContents["devDependencies"].(map[string]interface{})

			if !ok {
				assert.Fail(t, "devDeps not in expected format")
			}

			assert.ObjectsAreEqual(devDeps, nodedeps.DEV_DEPENDENCIES)
		})
	}
}
