package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
	toolsproto "github.com/teamkeel/keel/tools/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestService_GetTools(t *testing.T) {
	t.Parallel()
	testdataDir := "./testdata/composition"
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

			tools, err := toolsSvc.GetTools(t.Context())
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

func TestService_AddSpace(t *testing.T) {
	t.Parallel()
	testdataDir := "./testdata/spaces"

	testCases := []struct {
		name           string
		existingConfig string // path to file that has initial config
		spaceConfig    *SpaceConfig
		expected       []*toolsproto.Space
	}{
		{
			name:           "Add first space",
			existingConfig: "empty.json",
			spaceConfig: &SpaceConfig{
				Name:         "my space",
				Icon:         "icon",
				DisplayOrder: 2,
			},
			expected: []*toolsproto.Space{
				{
					Id:           "space-my-space",
					Name:         "my space",
					Icon:         "icon",
					DisplayOrder: 2,
				},
			},
		},
		{
			name:           "Add second space",
			existingConfig: "one_space.json",
			spaceConfig: &SpaceConfig{
				Name:         "Another Space",
				Icon:         "another-icon",
				DisplayOrder: 2,
			},
			expected: []*toolsproto.Space{
				{
					Id:           "space-space",
					Name:         "Space",
					Icon:         "icon",
					DisplayOrder: 1,
				},
				{
					Id:           "space-another-space",
					Name:         "Another Space",
					Icon:         "another-icon",
					DisplayOrder: 2,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			opts := []ServiceOpt{WithSchema(nil)}

			if tc.existingConfig != "" {
				existing, err := os.ReadFile(filepath.Join(testdataDir, tc.existingConfig))
				require.NoError(t, err)

				opts = append(opts, WithSpacesConfig(existing))
			}

			toolsSvc := NewService(opts...)

			// Add a new space
			_, err := toolsSvc.AddSpace(t.Context(), tc.spaceConfig)
			require.NoError(t, err, "Failed to add space")

			// Verify that the space was added
			updatedSpaces, err := toolsSvc.GetSpaces(t.Context())
			require.NoError(t, err)
			require.Len(t, updatedSpaces, len(tc.expected))

			// Validate the properties of the added space
			for i, expectedSpace := range tc.expected {
				actual := updatedSpaces[i]
				assert.Equal(t, expectedSpace.GetId(), actual.GetId(), " id mismatch - expected %q, got %q", i, expectedSpace.GetId(), actual.GetId())
				assert.Equal(t, expectedSpace.GetName(), actual.GetName(), "name mismatch - expected %q, got %q", expectedSpace.GetName(), actual.GetName())
				assert.Equal(t, expectedSpace.GetIcon(), actual.GetIcon(), "icon mismatch - expected %q, got %q", expectedSpace.GetIcon(), actual.GetIcon())
				assert.Equal(t, expectedSpace.GetDisplayOrder(), actual.GetDisplayOrder(), "display order mismatch - expected %v, got %v", expectedSpace.GetDisplayOrder(), actual.GetDisplayOrder())
			}
		})
	}
}
