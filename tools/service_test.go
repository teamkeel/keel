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

func TestService_RemoveSpace(t *testing.T) {
	t.Parallel()
	testdataDir := "./testdata/spaces"

	testCases := []struct {
		name           string
		existingConfig string
		spaceIDToRemove string
		expectedCount  int
		expectedIDs    []string
	}{
		{
			name:            "Remove only space",
			existingConfig:  "one_space.json",
			spaceIDToRemove: "space-space",
			expectedCount:   0,
			expectedIDs:     []string{},
		},
		{
			name:            "Remove first of two spaces",
			existingConfig:  "two_spaces.json",
			spaceIDToRemove: "space-first",
			expectedCount:   1,
			expectedIDs:     []string{"space-second"},
		},
		{
			name:            "Remove second of two spaces",
			existingConfig:  "two_spaces.json",
			spaceIDToRemove: "space-second",
			expectedCount:   1,
			expectedIDs:     []string{"space-first"},
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

			// Remove the space
			err := toolsSvc.RemoveSpace(t.Context(), tc.spaceIDToRemove)
			require.NoError(t, err, "Failed to remove space")

			// Verify that the space was removed
			remainingSpaces, err := toolsSvc.GetSpaces(t.Context())
			require.NoError(t, err)
			require.Len(t, remainingSpaces, tc.expectedCount)

			// Validate the remaining spaces
			for i, expectedID := range tc.expectedIDs {
				assert.Equal(t, expectedID, remainingSpaces[i].GetId(), "space id mismatch at index %d", i)
			}
		})
	}
}

func TestService_AddSpaceAction(t *testing.T) {
	t.Parallel()
	testdataDir := "./testdata/spaces"

	testCases := []struct {
		name           string
		existingConfig string
		spaceID        string
		payload        *toolsproto.CreateSpaceActionPayload
		expectedActionCount int
	}{
		{
			name:           "Add action to space without group",
			existingConfig: "one_space.json",
			spaceID:        "space-space",
			payload: &toolsproto.CreateSpaceActionPayload{
				SpaceId: "space-space",
				GroupId: nil,
				Link: &toolsproto.ToolLink{
					ToolId: "my-tool",
					Title:  &toolsproto.StringTemplate{Template: "My Action"},
				},
			},
			expectedActionCount: 1,
		},
		{
			name:           "Add action to space with group",
			existingConfig: "space_with_items.json",
			spaceID:        "space-test",
			payload: &toolsproto.CreateSpaceActionPayload{
				SpaceId: "space-test",
				GroupId: func() *string { s := "group-test-group"; return &s }(),
				Link: &toolsproto.ToolLink{
					ToolId: "another-tool",
					Title:  &toolsproto.StringTemplate{Template: "Another Action"},
				},
			},
			expectedActionCount: 1, // 1 action in the group
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

			// Add the action
			space, err := toolsSvc.AddSpaceAction(t.Context(), tc.payload)
			require.NoError(t, err, "Failed to add space action")
			require.NotNil(t, space)

			// Verify the action was added
			if tc.payload.GetGroupId() != "" {
				// Find the group and check its actions
				var targetGroup *toolsproto.SpaceGroup
				for _, group := range space.GetGroups() {
					if group.GetId() == tc.payload.GetGroupId() {
						targetGroup = group
						break
					}
				}
				require.NotNil(t, targetGroup, "Group not found")
				require.Len(t, targetGroup.GetActions(), tc.expectedActionCount, "action count mismatch in group")
			} else {
				// Check top-level actions
				require.GreaterOrEqual(t, len(space.GetActions()), tc.expectedActionCount, "action count mismatch")
			}
		})
	}
}

func TestService_RemoveSpaceItem(t *testing.T) {
	t.Parallel()
	testdataDir := "./testdata/spaces"

	testCases := []struct {
		name           string
		existingConfig string
		spaceID        string
		itemID         string
		expectError    bool
	}{
		{
			name:           "Remove action from space",
			existingConfig: "space_with_items.json",
			spaceID:        "space-test",
			itemID:         "action-test-tool",
			expectError:    false,
		},
		{
			name:           "Remove group from space",
			existingConfig: "space_with_items.json",
			spaceID:        "space-test",
			itemID:         "group-test-group",
			expectError:    false,
		},
		{
			name:           "Remove metric from space",
			existingConfig: "space_with_items.json",
			spaceID:        "space-test",
			itemID:         "metric-test-metric",
			expectError:    false,
		},
		{
			name:           "Remove link from space",
			existingConfig: "space_with_items.json",
			spaceID:        "space-test",
			itemID:         "link-test-link",
			expectError:    false,
		},
		{
			name:           "Remove non-existent item",
			existingConfig: "space_with_items.json",
			spaceID:        "space-test",
			itemID:         "non-existent-id",
			expectError:    false, // RemoveSpaceItem doesn't error on missing items
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

			// Get initial item counts
			initialSpaces, err := toolsSvc.GetSpaces(t.Context())
			require.NoError(t, err)

			var initialSpace *toolsproto.Space
			for _, s := range initialSpaces {
				if s.GetId() == tc.spaceID {
					initialSpace = s
					break
				}
			}
			require.NotNil(t, initialSpace)

			// Remove the item
			updatedSpace, err := toolsSvc.RemoveSpaceItem(t.Context(), tc.spaceID, tc.itemID)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err, "Failed to remove space item")
			require.NotNil(t, updatedSpace)

			// Verify the item was removed by checking if item count decreased
			// or stayed the same (for non-existent items)
			totalInitialItems := len(initialSpace.GetActions()) +
				len(initialSpace.GetGroups()) +
				len(initialSpace.GetMetrics()) +
				len(initialSpace.GetLinks())

			totalUpdatedItems := len(updatedSpace.GetActions()) +
				len(updatedSpace.GetGroups()) +
				len(updatedSpace.GetMetrics()) +
				len(updatedSpace.GetLinks())

			if tc.itemID != "non-existent-id" {
				assert.Less(t, totalUpdatedItems, totalInitialItems, "item count should decrease")
			}
		})
	}
}

func TestService_UpdateSpaceItem(t *testing.T) {
	t.Parallel()
	testdataDir := "./testdata/spaces"

	testCases := []struct {
		name           string
		existingConfig string
		payload        any
		itemID         string
		validateFunc   func(*testing.T, *toolsproto.Space)
	}{
		{
			name:           "Update space action",
			existingConfig: "space_with_items.json",
			payload: &toolsproto.UpdateSpaceActionPayload{
				Id: "action-test-tool",
				Link: &toolsproto.ToolLink{
					ToolId: "updated-tool",
					Title:  &toolsproto.StringTemplate{Template: "Updated Action"},
				},
			},
			itemID: "action-test-tool",
			validateFunc: func(t *testing.T, space *toolsproto.Space) {
				require.NotNil(t, space)
				require.Len(t, space.GetActions(), 1)
				assert.Equal(t, "updated-tool", space.GetActions()[0].GetLink().GetToolId())
				assert.Equal(t, "Updated Action", space.GetActions()[0].GetLink().GetTitle().GetTemplate())
			},
		},
		{
			name:           "Update space metric",
			existingConfig: "space_with_items.json",
			payload: &toolsproto.UpdateSpaceMetricPayload{
				Id:     "metric-test-metric",
				Label:  &toolsproto.StringTemplate{Template: "Updated Metric"},
				ToolId: "updated-tool",
				FacetLocation: &toolsproto.JsonPath{Path: "$.newPath"},
				DisplayOrder: 5,
			},
			itemID: "metric-test-metric",
			validateFunc: func(t *testing.T, space *toolsproto.Space) {
				require.NotNil(t, space)
				require.Len(t, space.GetMetrics(), 1)
				assert.Equal(t, "Updated Metric", space.GetMetrics()[0].GetLabel().GetTemplate())
				assert.Equal(t, "updated-tool", space.GetMetrics()[0].GetToolId())
				assert.Equal(t, "$.newPath", space.GetMetrics()[0].GetFacetLocation().GetPath())
				assert.Equal(t, int32(5), space.GetMetrics()[0].GetDisplayOrder())
			},
		},
		{
			name:           "Update space group",
			existingConfig: "space_with_items.json",
			payload: &toolsproto.UpdateSpaceGroupPayload{
				Id:          "group-test-group",
				Name:        &toolsproto.StringTemplate{Template: "Updated Group"},
				Description: &toolsproto.StringTemplate{Template: "Updated description"},
				DisplayOrder: 10,
			},
			itemID: "group-test-group",
			validateFunc: func(t *testing.T, space *toolsproto.Space) {
				require.NotNil(t, space)
				require.Len(t, space.GetGroups(), 1)
				assert.Equal(t, "Updated Group", space.GetGroups()[0].GetName().GetTemplate())
				assert.Equal(t, "Updated description", space.GetGroups()[0].GetDescription().GetTemplate())
				assert.Equal(t, int32(10), space.GetGroups()[0].GetDisplayOrder())
			},
		},
		{
			name:           "Update space link",
			existingConfig: "space_with_items.json",
			payload: &toolsproto.UpdateSpaceLinkPayload{
				Id: "link-test-link",
				Link: &toolsproto.ExternalLink{
					Label: &toolsproto.StringTemplate{Template: "Updated Link"},
					Href:  &toolsproto.StringTemplate{Template: "https://updated.example.com"},
				},
			},
			itemID: "link-test-link",
			validateFunc: func(t *testing.T, space *toolsproto.Space) {
				require.NotNil(t, space)
				require.Len(t, space.GetLinks(), 1)
				assert.Equal(t, "Updated Link", space.GetLinks()[0].GetLink().GetLabel().GetTemplate())
				assert.Equal(t, "https://updated.example.com", space.GetLinks()[0].GetLink().GetHref().GetTemplate())
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

			// Update the item
			updatedSpace, err := toolsSvc.UpdateSpaceItem(t.Context(), tc.payload)
			require.NoError(t, err, "Failed to update space item")

			// Validate using the custom function
			tc.validateFunc(t, updatedSpace)
		})
	}
}
