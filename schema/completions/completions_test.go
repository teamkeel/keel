package completions_test

import (
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/completions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
)

func TestCompletions(t *testing.T) {
	testCases, err := ioutil.ReadDir("testdata")
	require.NoError(t, err)

	toRun := []fs.FileInfo{}
	for _, testCase := range testCases {

		if strings.HasSuffix(testCase.Name(), ".only") {
			toRun = append(toRun, testCase)
		}
	}

	if len(toRun) > 0 {
		testCases = toRun
	}

	for _, testCase := range testCases {
		t.Run(strings.TrimSuffix(testCase.Name(), ".txt"), func(t *testing.T) {
			b, err := ioutil.ReadFile(filepath.Join("testdata", testCase.Name()))
			require.NoError(t, err)

			parts := strings.Split(string(b), "===")
			require.Equal(t, 3, len(parts), "fixture file should contain 3 sections seperated by \"===\"")

			ast, _ := parser.Parse(&reader.SchemaFile{
				Contents: string(parts[0]),
			})

			lineColPart := parts[1]

			lines := strings.Split(lineColPart, "\n")

			line := 0
			column := 0

			for _, l := range lines {
				if strings.HasPrefix(l, "line:") {
					lineStr := strings.Replace(l, "line:", "", 1)

					lineParse, err := strconv.Atoi(strings.TrimSpace(lineStr))
					require.NoError(t, err)

					line = lineParse
				}

				if strings.HasPrefix(l, "column:") {
					columnStr := strings.Replace(l, "column:", "", 1)

					columnParse, err := strconv.Atoi(strings.TrimSpace(columnStr))
					require.NoError(t, err)

					column = columnParse
				}
			}

			res := completions.ProvideCompletions(ast, node.Position{
				Line:   line,
				Column: column,
			})

			deleteEmpty := func(strs []string) (ret []string) {
				for _, str := range strs {
					if len(strings.TrimSpace(str)) > 0 {
						ret = append(ret, str)
					}
				}

				return ret
			}

			labelsOnly := func(completions []*completions.CompletionItem) (strs []string) {
				for _, c := range completions {
					strs = append(strs, c.Label)
				}

				return strs
			}

			expected := deleteEmpty(strings.Split(parts[2], "\n"))
			actual := labelsOnly(res)

			assert.ElementsMatch(t, expected, actual)
		})
	}
}
