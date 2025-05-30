package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/graphql"
	"github.com/teamkeel/keel/runtime/common"
)

type Request struct {
	Context context.Context
	Path    string
	Body    []byte
}

type Response struct {
	Body   []byte
	Status int
}

type introsepctionTypeRef struct {
	Name   string                `json:"name"`
	Kind   string                `json:"kind"`
	OfType *introsepctionTypeRef `json:"ofType"`
}

func (t introsepctionTypeRef) String() string {
	if t.Kind == "NON_NULL" {
		return t.OfType.String() + "!"
	}
	if t.Kind == "LIST" {
		return "[" + t.OfType.String() + "]"
	}
	return t.Name
}

type introspectionField struct {
	Args []struct {
		DefaultValue interface{}          `json:"defaultValue"`
		Name         string               `json:"name"`
		Type         introsepctionTypeRef `json:"type"`
	} `json:"args"`
	Name string               `json:"name"`
	Type introsepctionTypeRef `json:"type"`
}

// Represents the result of executing github.com/teamkeel/graphql/testutil.IntrospectionQuery
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
			Fields        []introspectionField `json:"fields"`
			InputFields   []introspectionField `json:"inputFields"`
			Interfaces    interface{}          `json:"interfaces"`
			Kind          string               `json:"kind"`
			Name          string               `json:"name"`
			PossibleTypes interface{}          `json:"possibleTypes"`
		} `json:"types"`
	} `json:"__schema"`
}

var (
	BuiltInScalars = []string{
		"Int",
		"Float",
		"String",
		"Boolean",
		"ID",
	}
)

// ToGraphQLSchemaLanguage converts the result of an introspection query
// into a GraphQL schema string
// Note: this implementation is not complete and only covers cases
// that are relevant to us, for example directives are not handled
func ToGraphQLSchemaLanguage(response common.Response) string {
	// First we have the marshal the response bytes back into
	// a graphql.Result
	var result graphql.Result
	_ = json.Unmarshal(response.Body, &result)

	// Then we pull out the data contained in the result and convert
	// back into JSON
	b, _ := json.Marshal(result.Data)

	// Finally we marshal that JSON into the IntrospectionQueryResult
	// type... urgh
	var r IntrospectionQueryResult
	_ = json.Unmarshal(b, &r)

	definitions := []string{}

	sort.Slice(r.Schema.Types, func(a, b int) bool {
		aType := r.Schema.Types[a]
		bType := r.Schema.Types[b]

		// Make sure Query and Mutation come at the top of the generated
		// schema with Query first and Mutation second
		typeNameOrder := []string{"Mutation", "Query"}
		aIndex := lo.IndexOf(typeNameOrder, aType.Name)
		bIndex := lo.IndexOf(typeNameOrder, bType.Name)
		if aIndex != -1 || bIndex != -1 {
			return aIndex > bIndex
		}

		// Then order by input types, and enums
		kindOrder := []string{"ENUM", "OBJECT", "INPUT_OBJECT"}
		aIndex = lo.IndexOf(kindOrder, aType.Kind)
		bIndex = lo.IndexOf(kindOrder, bType.Kind)
		if aIndex != bIndex {
			return aIndex > bIndex
		}

		// Order same kind by name
		return aType.Name < bType.Name
	})

	for _, t := range r.Schema.Types {
		if t.Kind == "SCALAR" {
			if lo.Contains(BuiltInScalars, t.Name) {
				continue
			}

			definitions = append(definitions, fmt.Sprintf("scalar %s", t.Name))

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

		definitions = append(definitions, b.String())
	}

	return strings.Join(definitions, "\n\n") + "\n"
}
