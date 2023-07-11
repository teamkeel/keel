package schema_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
	"google.golang.org/protobuf/encoding/protojson"
)

type Error struct {
	Code string `json:"code"`
}

type Errors struct {
	Errors []Error `json:"Errors"`
}

func TestSchema(t *testing.T) {
	testdataDir := "./testdata"
	testCases, err := os.ReadDir(testdataDir)
	require.NoError(t, err)

	for _, testCase := range testCases {
		if !testCase.IsDir() {
			continue
		}

		if testCase.Name() != "validation_duplicate_inputs" {
			continue
		}

		testCaseDir := testdataDir + "/" + testCase.Name()

		t.Run(testCase.Name(), func(t *testing.T) {

			files, err := os.ReadDir(testCaseDir)
			require.NoError(t, err)

			s2m := schema.Builder{}

			filesByName := map[string][]byte{}
			for _, f := range files {
				if f.IsDir() {
					continue
				}
				b, err := os.ReadFile(testCaseDir + "/" + f.Name())
				require.NoError(t, err)
				filesByName[f.Name()] = b
			}

			protoSchema, err := s2m.MakeFromDirectory(testCaseDir)

			var expectedJSON []byte
			var actualJSON []byte

			var actualProtoJSONPretty string

			// This is used when expected error json differs from actual error json,
			// and provides something you can copy and paste into your errors.json file,
			// once you've got it looking right.
			var prettyJSONErr string

			if expectedProto, ok := filesByName["proto.json"]; ok {
				require.NoError(t, err)
				expectedJSON = expectedProto
				actualJSON, err = protojson.Marshal(protoSchema)
				require.NoError(t, err)
				actualProtoJSONPretty = protojson.Format(protoSchema)
				_ = actualProtoJSONPretty

			} else if expectedErrors, ok := filesByName["errors.json"]; ok {
				require.NotNil(t, err, "expected there to be validation errors")
				expectedJSON = expectedErrors
				capturedErr := err
				actualJSON, err = json.Marshal(capturedErr)
				require.NoError(t, err)
				q, err := json.MarshalIndent(capturedErr, "", "  ")
				prettyJSONErr = string(q)
				require.NoError(t, err)
			} else {
				// if no proto.json file or errors.json file is provided then we assume this
				// is a test case that is just expected to parse and validate with no errors
				require.NoError(t, err)
				return
			}

			opts := jsondiff.DefaultConsoleOptions()

			diff, explanation := jsondiff.Compare(expectedJSON, actualJSON, &opts)

			switch diff {
			case jsondiff.FullMatch:
				// success
			case jsondiff.SupersetMatch, jsondiff.NoMatch:
				// These printfs produce output you can copy and paste to rectify
				// errors during development.
				fmt.Printf("Pretty json error (%s): \n%s\n", testCase.Name(), prettyJSONErr)
				fmt.Printf("Pretty actual proto json (%s): \n%s\n", testCase.Name(), actualProtoJSONPretty)
				assert.Fail(t, "actual result does not match expected", explanation)
			case jsondiff.FirstArgIsInvalidJson:
				assert.Fail(t, "expected JSON is invalid")
			case jsondiff.SecondArgIsInvalidJson:
				// highly unlikely (almost impossible) to happen
				assert.Fail(t, "actual JSON (proto or errors) is invalid")
			case jsondiff.BothArgsAreInvalidJson:
				// also highly unlikely (almost impossible) to happen
				assert.Fail(t, "both expected and actual JSON are invalid")
			}
		})
	}
}
