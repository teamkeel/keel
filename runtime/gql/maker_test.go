package gql_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/runtime/gql"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
)

func TestMaker(t *testing.T) {
	testFiles, err := ioutil.ReadDir("./testdata")
	require.NoError(t, err)

	type testCase struct {
		schema  string
		graphql string
	}

	testCases := map[string]testCase{}

	for _, f := range testFiles {
		parts := strings.Split(f.Name(), ".")
		name, ext := parts[0], parts[1]

		tc := testCases[name]

		b, err := ioutil.ReadFile(filepath.Join("./testdata", f.Name()))
		require.NoError(t, err)

		switch ext {
		case "keel":
			tc.schema = string(b)
		case "graphql":
			tc.graphql = string(b)
		}

		testCases[name] = tc
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			builder := schema.Builder{}
			proto, err := builder.MakeFromInputs(&reader.Inputs{
				SchemaFiles: []reader.SchemaFile{
					{
						Contents: tc.schema,
					},
				},
			})
			require.NoError(t, err)

			gqlSchemas, err := gql.MakeSchemas(proto)
			require.NoError(t, err)

			actual := gql.ToSchemaLanguage(*gqlSchemas["Test"])
			expected := tc.graphql

			assert.Equal(t, expected, actual)

			if expected != actual {
				// Print the actual result for easier debugging
				fmt.Println("Actual GraphQL schema:")
				fmt.Println(actual)
			}
		})
	}
}
