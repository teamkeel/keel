package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/rpc/rpc"
	"github.com/teamkeel/keel/schema"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestGenerateTools(t *testing.T) {
	t.Parallel()
	testdataDir := "./testdata/generator"
	testCases, err := os.ReadDir(testdataDir)

	require.NoError(t, err)

	for _, tc := range testCases {
		testCase := tc

		if !testCase.IsDir() {
			t.Errorf("test data directory should only contain directories - file found: %s", testCase.Name())
			continue
		}

		testCaseDir := filepath.Join(testdataDir, testCase.Name())

		t.Run(testCase.Name(), func(t *testing.T) {
			t.Parallel()
			expected, err := os.ReadFile(filepath.Join(testCaseDir, "tools.json"))
			require.NoError(t, err)

			builder := schema.Builder{}
			schema, err := builder.MakeFromDirectory(testCaseDir)
			require.NoError(t, err)

			gen, err := NewGenerator(schema, builder.Config)
			require.NoError(t, err)

			err = gen.Generate(t.Context())
			require.NoError(t, err)

			tools := gen.GetTools()

			response := &rpc.ListToolsResponse{
				ToolConfigs: tools,
			}

			actual, err := protojson.Marshal(response)
			require.NoError(t, err)

			opts := jsondiff.DefaultConsoleOptions()

			diff, explanation := jsondiff.Compare(expected, actual, &opts)
			if diff == jsondiff.FullMatch {
				return
			}

			fmt.Printf("%s - %s\n\n", testCase.Name(), string(actual))

			assert.Fail(t, "actual tools JSON does not match expected", explanation)
		})
	}
}
