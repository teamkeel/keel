package functions_test

import (
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/schema"
)

func TestCodeGeneration(t *testing.T) {
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
		t.Run(strings.TrimSuffix(testCase.Name(), ".txt"), func(t *testing.T) {
			b, err := ioutil.ReadFile(filepath.Join("testdata", testCase.Name()))
			require.NoError(t, err)

			parts := strings.Split(string(b), "===")

			require.Equal(t, 2, len(parts), "fixture file should contain 2 sections separated by ===")

			require.NoError(t, err)

			scm := schema.Builder{}

			proto, err := scm.MakeFromString(parts[0])

			require.NoError(t, err)

			generator := functions.NewCodeGenerator(proto)

			if strings.HasPrefix(testCase.Name(), "model_") {

				result := generator.GenerateBaseTypes() + generator.GenerateModels()

				assert.Equal(t, strings.Trim(parts[1], "\n"), strings.Trim(result, "\n"))
			} else if strings.HasPrefix(testCase.Name(), "enum_") {
				result := generator.GenerateEnums()

				assert.Equal(t, strings.Trim(parts[1], "\n"), strings.Trim(result, "\n"))
			} else if strings.HasPrefix(testCase.Name(), "api") {
				result := generator.GenerateAPIs()

				assert.Equal(t, strings.Trim(parts[1], "\n"), strings.Trim(result, "\n"))
			} else {
				t.Fatal("Test case names must follow convention XXX_name where XXX is one of model, enum, api")
			}
		})
	}

}
