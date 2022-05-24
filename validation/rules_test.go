package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/parser"
)

//Fields/operations/functions are lowerCamel
func TestFieldsOpsFuncsLowerCamel(t *testing.T) {
	tests := map[string]struct {
		input    *parser.Schema
		expected []error
	}{
		"simpleFieldName": {input: &parser.Schema{Declarations: []*parser.Declaration{{
			Model: &parser.Model{Name: "createBook", Sections: []*parser.ModelSection{
				{Fields: []*parser.ModelField{
					{Name: "title", Type: "string"},
				},
				},
			}}}}}, expected: nil},
		"simpelFunction": {input: &parser.Schema{Declarations: []*parser.Declaration{{
			Model: &parser.Model{Name: "createBook", Sections: []*parser.ModelSection{{
				Operations: []*parser.ModelAction{
					{Name: "createBook"},
				}},
			}}}}}, expected: nil},
		"allLower": {input: &parser.Schema{Declarations: []*parser.Declaration{{
			Model: &parser.Model{Name: "createbook", Sections: []*parser.ModelSection{{
				Operations: []*parser.ModelAction{
					{Name: "createbook"},
				}},
			}}}}}, expected: nil},
		"allUpperFunction": {input: &parser.Schema{Declarations: []*parser.Declaration{{
			Model: &parser.Model{Name: "createBook", Sections: []*parser.ModelSection{{
				Operations: []*parser.ModelAction{
					{Name: "CREATEBOOK"},
				}},
			}}}}}, expected: []error{&ValidationError{
			Code: "E002",
			ErrorDetails: ErrorDetails{
				Message:      "You have a function name that is not lowerCamel CREATEBOOK",
				ShortMessage: "CREATEBOOK isn't lower camel",
				Hint:         "createbook",
			},
		}}},
		"underscore": {input: &parser.Schema{Declarations: []*parser.Declaration{{
			Model: &parser.Model{Name: "createbook", Sections: []*parser.ModelSection{{
				Fields: []*parser.ModelField{
					{Name: "title", Type: "string"},
				},
				Operations: []*parser.ModelAction{
					{Name: "book_author"},
				}},
			}}}}}, expected: []error{&ValidationError{
			Code: "E002",
			ErrorDetails: ErrorDetails{
				Message:      "You have a function name that is not lowerCamel book_author",
				ShortMessage: "book_author isn't lower camel",
				Hint:         "bookAuthor",
			},
		}}},
	}

	for name, tc := range tests {
		got := fieldsOpsFuncsLowerCamel(asInputs(tc.input))
		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}

}

// test findDuplicates
func TestFindDuplicates(t *testing.T) {
	input1 := []string{"a", "b", "b"}
	input2 := []string{"a", "b", "c"}

	tests := map[string]struct {
		input    []string
		expected []string
	}{
		"working": {input: input1, expected: []string{"b"}},
		"nodups":  {input: input2, expected: []string(nil)},
	}

	for name, tc := range tests {
		got := findDuplicates(tc.input)
		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}
}

// asInputs wraps a single parser.Schema into an []Inputs - as required by most of the
// functions under test.
func asInputs(oneSchema *parser.Schema) []Input {
	oneInput := Input{
		FileName:     "unused",
		ParsedSchema: oneSchema,
	}
	return []Input{oneInput}
}

func TestCheckAttributeExpressions(t *testing.T) {
	input := []*parser.Attribute{
		{
			Name: "test",
			Arguments: []*parser.AttributeArgument{
				{
					Name: "test",
					Expression: &expressions.Expression{
						Or: []*expressions.OrExpression{
							{
								And: []*expressions.ConditionWrap{
									{
										Condition: &expressions.Condition{
											Operator: "=",
											LHS: &expressions.Value{
												Ident: []string{"profile", "identity"},
											},
											RHS: &expressions.Value{
												Ident: []string{"foo", "name"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	got := checkAttributeExpressions(input, "Profile", &parser.ModelField{Name: "identity", Attributes: []*parser.Attribute{
		{
			Name: "unique",
		},
	}})
	assert.Equal(t, true, got)

}
