package schema_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestProto(t *testing.T) {
	t.Parallel()
	testdataDir := "./testdata/proto"
	testCases, err := os.ReadDir(testdataDir)
	require.NoError(t, err)

	for _, tc := range testCases {
		testCase := tc
		if !testCase.IsDir() {
			t.Errorf("proto test data directory should only contain directories - file found: %s", testCase.Name())
			continue
		}

		testCaseDir := filepath.Join(testdataDir, testCase.Name())

		t.Run(testCase.Name(), func(t *testing.T) {
			t.Parallel()
			expected, err := os.ReadFile(filepath.Join(testCaseDir, "proto.json"))
			require.NoError(t, err)

			builder := schema.Builder{}
			protoSchema, err := builder.MakeFromDirectory(testCaseDir)
			require.NoError(t, err)

			actual, err := protojson.Marshal(protoSchema)
			require.NoError(t, err)

			opts := jsondiff.DefaultConsoleOptions()

			diff, explanation := jsondiff.Compare(expected, actual, &opts)
			if diff == jsondiff.FullMatch {
				return
			}

			fmt.Println(testCase.Name())
			fmt.Println(string(actual))
			fmt.Println()

			assert.Fail(t, "actual proto JSON does not match expected", explanation)
		})
	}
}

var expectErrorCommentRegex = regexp.MustCompile(`^\s*\/\/\s{0,1}expect-error:`)

func TestValidation(t *testing.T) {
	t.Parallel()
	dir := "./testdata/errors"
	testCases, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, tc := range testCases {
		// if tc.Name() != "unique_composite_lookup.keel" {
		// 	continue
		// }

		testCase := tc
		if testCase.IsDir() {
			t.Errorf("errors test data directory should only contain keel schema files - directory found: %s", testCase.Name())
			continue
		}

		testCaseDir := filepath.Join(dir, testCase.Name())
		t.Run(testCase.Name(), func(t *testing.T) {
			t.Parallel()
			b, err := os.ReadFile(testCaseDir)
			require.NoError(t, err)

			builder := &schema.Builder{}
			_, err = builder.MakeFromString(string(b), config.Empty)

			verrs := &errorhandling.ValidationErrors{}
			if !errors.As(err, &verrs) {
				t.Errorf("no validation errors returned in %s: %v", testCase.Name(), err)
			}

			expectedErrors := []*errorhandling.ValidationError{}
			lines := strings.Split(string(b), "\n")
			for i, line := range lines {
				if !expectErrorCommentRegex.MatchString(line) {
					continue
				}

				line := expectErrorCommentRegex.ReplaceAllString(line, "")
				parts := strings.SplitN(line, ":", 4)

				column, err := strconv.Atoi(parts[0])
				require.NoError(t, err, "unable to parse start column from //expect-error comment")

				endColumn, err := strconv.Atoi(parts[1])
				require.NoError(t, err, "unable to parse end column from //expect-eror comment")

				code := parts[2]
				message := parts[3]

				// A line can have multiple expected errors - so we find the next line that is not an "expect-error" comment
				errorLine := i + 2
				for j, l := range lines[i+1:] {
					if !expectErrorCommentRegex.MatchString(l) {
						errorLine += j
						break
					}
				}

				expectedErrors = append(expectedErrors, &errorhandling.ValidationError{
					ErrorDetails: &errorhandling.ErrorDetails{
						Message: message,
					},
					Code: code,
					Pos: errorhandling.LexerPos{
						Line:   errorLine,
						Column: column,
					},
					EndPos: errorhandling.LexerPos{
						Line:   errorLine,
						Column: endColumn,
					},
				})
			}

			missing, unexpected := lo.Difference(lo.Map(expectedErrors, errorToString), lo.Map(verrs.Errors, errorToString))
			for _, v := range missing {
				t.Errorf("  Expected in %s:   %s", testCase.Name(), v)
			}
			for _, v := range unexpected {
				t.Errorf("  Unexpected in %s: %s", testCase.Name(), v)
			}
		})
	}
}

func errorToString(err *errorhandling.ValidationError, _ int) string {
	return fmt.Sprintf("%d:%d:%d:%s:%s", err.Pos.Line, err.Pos.Column, err.EndPos.Column, err.Code, err.Message)
}
