package codegenerator_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	codegenerator "github.com/teamkeel/keel/node/codegen"
	"github.com/teamkeel/keel/schema"
)

type TestCase struct {
	Name                       string
	Schema                     string
	TypeScriptDefinitionOutput string
	JavaScriptOutput           string
}

func TestSdk(t *testing.T) {
	cases := []TestCase{
		{
			Name: "model-generation-simple",
			Schema: `
			model Person {
				fields {
					name Text
					age Number
				}
			}
			`,
			JavaScriptOutput: `
			const doSomething = () => 'hello';
			const variableName = '';
			`,
			TypeScriptDefinitionOutput: "",
		},
	}

	for _, tc := range cases {
		builder := schema.Builder{}

		sch, err := builder.MakeFromString(tc.Schema)

		require.NoError(t, err)

		tmpDir, err := os.MkdirTemp("", tc.Name)

		require.NoError(t, err)

		cg := codegenerator.NewGenerator(sch, tmpDir)

		generatedFiles, err := cg.GenerateSDK()

		require.NoError(t, err)

		for _, f := range generatedFiles {
			actual := normaliseString(f.Contents)
			expected := ""

			switch f.Type {
			case codegenerator.SourceCodeTypeJavaScript:
				expected = normaliseString(tc.JavaScriptOutput)
			case codegenerator.SourceCodeTypeDefinition:
				expected = normaliseString(tc.TypeScriptDefinitionOutput)
			}

			assert.Equal(t, expected, actual)
		}
	}
}

func normaliseString(str string) string {
	str = strings.Replace(str, "\t", "", -1)
	str = strings.Replace(str, "\n", "", -1)

	return str
}
