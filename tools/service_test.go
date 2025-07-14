package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestService_GetTools(t *testing.T) {
	t.Parallel()
	testdataDir := "./testdata/configuration"
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

			expectedDir := filepath.Join(testCaseDir, "expected")

			// make proto schema from test folder
			builder := schema.Builder{}
			schema, err := builder.MakeFromDirectory(testCaseDir)
			require.NoError(t, err)

			// create tools service with test case schema and test case tools config (testcaseDir/tools)
			toolsSvc := NewService(WithSchema(schema), WithFileStorage(testCaseDir))

			tools, err := toolsSvc.GetTools(context.Background())
			require.NoError(t, err)

			// check all expected configs
			expectedTools, err := os.ReadDir(expectedDir)
			require.NoError(t, err)

			for _, expectedTool := range expectedTools {
				expected, err := os.ReadFile(filepath.Join(expectedDir, expectedTool.Name()))
				require.NoError(t, err)

				toolID := strings.TrimSuffix(expectedTool.Name(), ".json")

				actualTool := tools.FindByID(toolID)
				require.NotNil(t, actualTool)

				actual, err := protojson.Marshal(actualTool)
				require.NoError(t, err)

				opts := jsondiff.DefaultConsoleOptions()

				diff, explanation := jsondiff.Compare(expected, actual, &opts)
				if diff != jsondiff.FullMatch {
					assert.Fail(t, fmt.Sprintf("Tool %s does not match expected", toolID), explanation)
					fmt.Printf("%s:%s - %s\n\n", testCase.Name(), toolID, string(actual))
				}
			}
		})
	}
}
