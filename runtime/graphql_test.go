package runtime_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/graphql-go/graphql/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
)

func TestGraphQL(t *testing.T) {
	testFiles, err := ioutil.ReadDir("./testdata/graphql")
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

		b, err := ioutil.ReadFile(filepath.Join("./testdata/graphql", f.Name()))
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
			protoSchema, err := builder.MakeFromInputs(&reader.Inputs{
				SchemaFiles: []reader.SchemaFile{
					{
						Contents: tc.schema,
					},
				},
			})
			require.NoError(t, err)

			handler := runtime.NewHandler(protoSchema)

			body, err := json.Marshal(map[string]string{
				"query": testutil.IntrospectionQuery,
			})
			require.NoError(t, err)

			response, err := handler(&runtime.Request{
				Path: "/Test",
				Body: body,
			})
			require.NoError(t, err)
			assert.Equal(t, 200, response.Status)

			actual := runtime.ToGraphQLSchemaLanguage(response)
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
