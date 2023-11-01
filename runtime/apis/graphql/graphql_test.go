package graphql_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/graphql/testutil"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/apis/graphql"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
)

func TestGraphQL(t *testing.T) {
	testFiles, err := os.ReadDir("./testdata/graphql")
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

		b, err := os.ReadFile(filepath.Join("./testdata/graphql", f.Name()))
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
				SchemaFiles: []*reader.SchemaFile{
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

			response := handler(&http.Request{
				URL: &url.URL{
					Path: "/test/graphql",
				},
				Method: http.MethodPost,
				Body:   io.NopCloser(bytes.NewReader(body)),
			})

			require.NoError(t, err)
			assert.Equal(t, 200, response.Status, string(response.Body))

			actual := graphql.ToGraphQLSchemaLanguage(response)
			expected := tc.graphql

			assert.Equal(t, expected, actual)

			if expected != actual {
				// Print the actual result for easier debugging
				fmt.Println("Actual GraphQL schema for", name, ":")
				fmt.Println(actual)
			}
		})
	}
}
