package functions_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/testhelpers"
)

func TestNodeDeps(t *testing.T) {
	testCases, err := ioutil.ReadDir("nodedeps_testdata")

	require.NoError(t, err)

	for _, testCase := range testCases {
		t.Run(testCase.Name(), func(t *testing.T) {
			testDir := filepath.Join("nodedeps_testdata", testCase.Name())

			testhelpers.WithTmpDir(testDir, func(tmpDir string) {
				packageJsonPath := filepath.Join(tmpDir, "package.json")

				packageJson, err := functions.NewPackageJson(packageJsonPath)

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

				assert.ObjectsAreEqual(devDeps, functions.DEV_DEPENDENCIES)
			})
		})
	}
}
