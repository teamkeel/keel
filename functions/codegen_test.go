package functions_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/schema"
)

func TestCodeGeneration(t *testing.T) {
	testCases, err := ioutil.ReadDir("codegen_testdata")
	require.NoError(t, err)

	var permittedTestCaseTypes = []string{
		"model",
		"enum",
		"inputs",
		"api",
		"handler",
		"custom_function",
		"func_wrapper",
	}

	for _, testCase := range testCases {
		t.Run(strings.TrimSuffix(testCase.Name(), ".txt"), func(t *testing.T) {
			b, err := ioutil.ReadFile(filepath.Join("codegen_testdata", testCase.Name()))
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

				assert.Equal(t, strings.TrimSpace(parts[1]), strings.TrimSpace(result))
			} else if strings.HasPrefix(testCase.Name(), "enum_") {
				result := generator.GenerateEnums()

				assert.Equal(t, strings.TrimSpace(parts[1]), strings.TrimSpace(result))
			} else if strings.HasPrefix(testCase.Name(), "inputs_") {
				result := generator.GenerateInputs()

				assert.Equal(t, strings.TrimSpace(parts[1]), strings.TrimSpace(result))
			} else if strings.HasPrefix(testCase.Name(), "api_") {
				result := generator.GenerateAPIs()

				assert.Equal(t, strings.TrimSpace(parts[1]), strings.TrimSpace(result))
			} else if strings.HasPrefix(testCase.Name(), "handler_") {
				result := generator.GenerateEntryPoint()

				assert.Equal(t, strings.TrimSpace(parts[1]), strings.TrimSpace(result))
			} else if strings.HasPrefix(testCase.Name(), "custom_function_") {
				result := generator.GenerateFunction(proto.Models[0].Operations[0].Name)

				assert.Equal(t, strings.TrimSpace(parts[1]), strings.TrimSpace(result))
			} else if strings.HasPrefix(testCase.Name(), "func_wrapper_") {
				result := generator.GenerateWrappers()

				assert.Equal(t, strings.TrimSpace(parts[1]), strings.TrimSpace(result))
			} else {
				t.Fatalf("Test case names must follow convention XXX_name where XXX is one of %s", formatting.HumanizeList(permittedTestCaseTypes, formatting.DelimiterOr))
			}
		})
	}

}
