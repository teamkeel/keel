package gql

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

			m := newMaker(proto)
			gqlSchemas, err := m.make()
			require.NoError(t, err)

			result := graphql.Do(graphql.Params{
				Schema:         *gqlSchemas["Test"],
				Context:        context.Background(),
				RequestString:  testutil.IntrospectionQuery,
				VariableValues: map[string]any{},
			})

			assert.Len(t, result.Errors, 0)

			b, err := json.MarshalIndent(result.Data, "", " ")
			require.NoError(t, err)

			var introspectionResult IntrospectionQueryResult
			json.Unmarshal(b, &introspectionResult)

			assert.Equal(t, tc.graphql, toSchemaString(&introspectionResult))
		})
	}
}

type TypeRef struct {
	Name   string   `json:"name"`
	Kind   string   `json:"kind"`
	OfType *TypeRef `json:"ofType"`
}

func (t TypeRef) String() string {
	if t.Kind == "NON_NULL" {
		return t.OfType.String() + "!"
	}
	if t.Kind == "LIST" {
		return "[" + t.OfType.String() + "]"
	}
	return t.Name
}

type Field struct {
	Args []struct {
		DefaultValue interface{} `json:"defaultValue"`
		Name         string      `json:"name"`
		Type         TypeRef     `json:"type"`
	} `json:"args"`
	Name string  `json:"name"`
	Type TypeRef `json:"type"`
}

// Represents the result of executing github.com/graphql-go/graphql/testutil.IntrospectionQuery
type IntrospectionQueryResult struct {
	Schema struct {
		MutationType struct {
			Name string `json:"name"`
		} `json:"mutationType"`
		QueryType struct {
			Name string `json:"name"`
		} `json:"queryType"`
		Types []struct {
			EnumValues []struct {
				Name string
			} `json:"enumValues"`
			Fields        []Field     `json:"fields"`
			InputFields   []Field     `json:"inputFields"`
			Interfaces    interface{} `json:"interfaces"`
			Kind          string      `json:"kind"`
			Name          string      `json:"name"`
			PossibleTypes interface{} `json:"possibleTypes"`
		} `json:"types"`
	} `json:"__schema"`
}

// toSchemaString converts the result of an introspection query
// into a GraphQL schema string
// Note: this implementation is not complete and only covers cases
// that are relevant to us, for example directives are not handled
func toSchemaString(r *IntrospectionQueryResult) string {
	result := []string{}

	sort.Slice(r.Schema.Types, func(i, j int) bool {
		if r.Schema.Types[i].Name == "Query" {
			return true
		}
		if r.Schema.Types[j].Name == "Query" {
			return false
		}
		return r.Schema.Types[i].Name < r.Schema.Types[j].Name
	})

	for _, t := range r.Schema.Types {
		if t.Kind == "SCALAR" {
			continue
		}
		if strings.HasPrefix(t.Name, "__") {
			continue
		}

		keyword, ok := map[string]string{
			"OBJECT":       "type",
			"INPUT_OBJECT": "input",
			"ENUM":         "enum",
		}[t.Kind]
		if !ok {
			continue
		}

		b := strings.Builder{}
		b.WriteString(keyword)
		b.WriteString(" ")
		b.WriteString(t.Name)
		b.WriteString(" {\n")

		if t.Kind == "ENUM" {
			values := t.EnumValues
			sort.Slice(values, func(i, j int) bool {
				return values[i].Name < values[j].Name
			})

			for _, v := range values {
				b.WriteString("  ")
				b.WriteString(v.Name)
				b.WriteString("\n")
			}
		} else {
			fields := t.Fields
			if t.Kind == "INPUT_OBJECT" {
				fields = t.InputFields
			}

			sort.Slice(fields, func(i, j int) bool {
				return fields[i].Name < fields[j].Name
			})

			for _, field := range fields {
				b.WriteString("  ")
				b.WriteString(field.Name)

				sort.Slice(field.Args, func(i, j int) bool {
					return field.Args[i].Name < field.Args[j].Name
				})

				if len(field.Args) > 0 {
					b.WriteString("(")
					for i, arg := range field.Args {
						if i > 0 {
							b.WriteString(", ")
						}
						b.WriteString(arg.Name)
						b.WriteString(": ")
						b.WriteString(arg.Type.String())
					}
					b.WriteString(")")
				}

				b.WriteString(": ")
				b.WriteString(field.Type.String())
				b.WriteString("\n")
			}
		}

		b.WriteString("}")

		result = append(result, b.String())
	}

	return strings.Join(result, "\n\n") + "\n"
}
