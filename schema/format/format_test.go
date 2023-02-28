package format_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/format"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
)

func TestFormat(t *testing.T) {
	testCases, err := os.ReadDir("testdata")
	require.NoError(t, err)

	for _, testCase := range testCases {
		t.Run(strings.TrimSuffix(testCase.Name(), ".txt"), func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join("testdata", testCase.Name()))
			require.NoError(t, err)

			parts := strings.Split(string(b), "===")
			require.Equal(t, 2, len(parts), "fixture file should contain two sections seperated by \"===\"")

			ast, err := parser.Parse(&reader.SchemaFile{
				Contents: parts[0],
			})
			require.NoError(t, err)

			formatted := format.Format(ast)

			_, err = parser.Parse(&reader.SchemaFile{
				Contents: formatted,
			})
			assert.NoError(t, err, "formatter produced an invalid schema")

			expected := strings.TrimSpace(parts[1]) + "\n"
			if !assert.Equal(t, expected, formatted) {
				// Print actual output as the output of the assert is not
				// very readable

				fmt.Println("Expected:")
				fmt.Println(expected)
				fmt.Println()
				fmt.Println("Actual:")
				fmt.Println(formatted)

			}

		})
	}
}
