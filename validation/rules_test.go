package validation

import (
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/parser"
)

//Models are UpperCamel
func TestModelsAreUpperCamel(t *testing.T) {
	tests := map[string]struct {
		input    *parser.Schema
		expected []error
	}{
		"simple": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "Book"}}}},
			expected: nil},
		"long": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "BookAuthorLibrary"}}}},
			expected: nil},
		"allLower": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "bookauthor"}}}},
			expected: []error{&ValidationError{
				Code: "E001",
				ErrorDetails: ErrorDetails{
					Message:      "You have a model name that is not UpperCamel bookauthor",
					ShortMessage: "bookauthor is not UpperCamel",
					Hint:         "Bookauthor",
				},
			},
			}},
		"allUpper": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "BOOKAUTHOR"}}}},
			expected: []error{&ValidationError{
				Code: "E001",
				ErrorDetails: ErrorDetails{
					Message:      "You have a model name that is not UpperCamel BOOKAUTHOR",
					ShortMessage: "BOOKAUTHOR is not UpperCamel",
					Hint:         "Bookauthor",
				},
			}}},
		"underscore": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "book_author"}}}},
			expected: []error{&ValidationError{
				Code: "E001",
				ErrorDetails: ErrorDetails{
					Message:      "You have a model name that is not UpperCamel book_author",
					ShortMessage: "book_author is not UpperCamel",
					Hint:         "BookAuthor",
				},
			}}},
	}

	for name, tc := range tests {
		got := modelsUpperCamel(asInputs(tc.input))
		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}
}

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

//Field names must be unique in a model
func TestFieldNamesMustBeUniqueInAModel(t *testing.T) {
	input1 := []*parser.ModelField{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "createdAt", Type: "time.Time"},
	}
	input2 := []*parser.ModelField{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "name", Type: "time.Time"},
	}

	tests := map[string]struct {
		input    *parser.Schema
		expected []error
	}{
		"working": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input1},
		}}}}}, expected: nil},
		"long": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input2},
		}}}}}, expected: []error{&ValidationError{
			Code: "E003",
			ErrorDetails: ErrorDetails{
				Message:      "You have duplicate field names name",
				ShortMessage: "name is duplicated",
				Hint:         "Remove 'name' on line 0",
			},
		}}},
	}

	for name, tc := range tests {
		got := fieldNamesMustBeUniqueInAModel(asInputs(tc.input))
		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}

}

//Operations/functions must be globally unique
func TestFindOpsFuncsMustBeGloballyUnique(t *testing.T) {
	input := asInputs(&parser.Schema{Declarations: []*parser.Declaration{
		{
			Model: &parser.Model{
				Name: "book",
				Sections: []*parser.ModelSection{
					{
						Operations: []*parser.ModelAction{
							{
								Name: "createbook",
							},
							{
								Name: "dave",
							},
						},
					},
				},
			},
		},
		{
			Model: &parser.Model{
				Name: "book",
				Sections: []*parser.ModelSection{
					{
						Operations: []*parser.ModelAction{
							{
								Name: "createbook",
							},
							{
								Name: "dave1",
							},
						},
					},
				},
			},
		},
	}})

	expected := []GlobalOperations{
		{Name: "createbook", Model: "book"},
		{Name: "dave", Model: "book"},
		{Name: "createbook", Model: "book"},
		{Name: "dave1", Model: "book"},
	}

	got := uniqueOperationsGlobally(input)
	if !assert.Equal(t, expected, got) {
		t.Fatalf("%s: expected: %v, got: %v", "name", expected, got)
	}
}

func TestHintCorrection(t *testing.T) {
	apiAttribute := parser.Attribute{Name: "graphq", Pos: lexer.Position{Line: 23, Column: 1}}

	err := supportedAttributeTypes(asInputs(&parser.Schema{
		Declarations: []*parser.Declaration{
			{
				API: &parser.API{
					Name: "Web",
					Sections: []*parser.APISection{
						{
							Attribute: &apiAttribute,
						},
					},
				},
			},
		},
	}))

	expected := []error{
		&ValidationError{
			Code: "E011",
			ErrorDetails: ErrorDetails{
				Message:      "api 'Web' has an unrecognised attribute @graphq",
				ShortMessage: "Unrecognised attribute @graphq",
				Hint:         "Did you mean @graphql?",
			},
			Pos: LexerPos{Line: 23, Column: 1},
		}}

	assert.Equal(t, err, expected)
}

//Inputs of ops must be model fields
func TestInputsModelFields(t *testing.T) {
	tests := map[string]struct {
		input    *parser.Schema
		expected []error
	}{
		"passingOperation": {input: &parser.Schema{Declarations: []*parser.Declaration{
			{
				Model: &parser.Model{
					Sections: []*parser.ModelSection{

						{
							Fields: []*parser.ModelField{
								{
									Name: "id",
								},
							},
							Operations: []*parser.ModelAction{
								{
									Name: "createBook",
									Arguments: []*parser.ActionArg{
										{Name: "id"},
									},
								},
							},
						},
						{
							Fields: []*parser.ModelField{
								{
									Name: "id",
								},
							},
							Operations: []*parser.ModelAction{
								{
									Name: "author",
									Arguments: []*parser.ActionArg{
										{Name: "id"},
									},
								},
							},
						},
					}}}}}, expected: nil},
		"passingFunction": {input: &parser.Schema{Declarations: []*parser.Declaration{
			{
				Model: &parser.Model{
					Sections: []*parser.ModelSection{

						{
							Fields: []*parser.ModelField{
								{
									Name: "id",
								},
							},
							Operations: []*parser.ModelAction{
								{
									Name: "createBook",
									Arguments: []*parser.ActionArg{
										{Name: "id"},
									},
								},
							},
						},
						{
							Fields: []*parser.ModelField{
								{
									Name: "id",
								},
							},
							Functions: []*parser.ModelAction{
								{
									Name: "author",
									Arguments: []*parser.ActionArg{
										{Name: "id"},
									},
								},
							},
						},
					}}}}}, expected: nil},
		"failing": {input: &parser.Schema{Declarations: []*parser.Declaration{
			{
				Model: &parser.Model{
					Sections: []*parser.ModelSection{
						{
							Fields: []*parser.ModelField{
								{
									Name: "id",
								},
							},
							Operations: []*parser.ModelAction{
								{
									Name: "createBook",
									Arguments: []*parser.ActionArg{
										{Name: "id"},
									},
								},
							},
						},
						{
							Fields: []*parser.ModelField{
								{
									Name: "id",
								},
							},
							Operations: []*parser.ModelAction{
								{
									Name: "author",
									Arguments: []*parser.ActionArg{
										{Name: "name"},
									},
								},
							},
						},
					}}}}}, expected: []error{
			&ValidationError{
				Code: "E005",
				ErrorDetails: ErrorDetails{
					Message:      "You are using inputs that are not fields model:author, field:name",
					ShortMessage: "Replace name",
					Hint:         "Check inputs for author",
				},
			},
		}}}

	for name, tc := range tests {
		got := operationFunctionInputs(asInputs(tc.input))

		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}

}

//No reserved field names (id, createdAt, updatedAt)
func TestNoReservedFieldNames(t *testing.T) {
	input1 := []*parser.ModelField{
		{Name: "userId", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "userCreatedAt", Type: "time.Time"},
	}
	input2 := []*parser.ModelField{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "createdAt", Type: "time.Time"},
	}
	input3 := []*parser.ModelField{
		{Name: "ID", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "createdAt", Type: "time.Time"},
	}

	tests := map[string]struct {
		input    *parser.Schema
		expected []error
	}{
		"working": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input1},
		}}}}}, expected: nil},
		"invalid": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input2},
		}}}}}, expected: []error{
			&ValidationError{
				Code: "E006",
				ErrorDetails: ErrorDetails{
					Message:      "You have a reserved field name id",
					ShortMessage: "cannot use id",
					Hint:         "You cannot use id as field name, it is reserved, try ider",
				},
			},
			&ValidationError{
				Code: "E006",
				ErrorDetails: ErrorDetails{
					Message:      "You have a reserved field name createdAt",
					ShortMessage: "cannot use createdAt",
					Hint:         "You cannot use createdAt as field name, it is reserved, try createdAter",
				},
			}}},
		"invalidUpperCase": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input3},
		}}}}}, expected: []error{&ValidationError{
			Code: "E006",
			ErrorDetails: ErrorDetails{
				Message:      "You have a reserved field name ID",
				ShortMessage: "cannot use ID",
				Hint:         "You cannot use ID as field name, it is reserved, try IDer",
			},
		},
			&ValidationError{
				Code: "E006",
				ErrorDetails: ErrorDetails{
					Message:      "You have a reserved field name createdAt",
					ShortMessage: "cannot use createdAt",
					Hint:         "You cannot use createdAt as field name, it is reserved, try createdAter",
				},
			}}},
	}

	for name, tc := range tests {
		got := noReservedFieldNames(asInputs(tc.input))
		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}
}

//No reserved model name (query)
func TestReservedModelNames(t *testing.T) {
	tests := map[string]struct {
		input    *parser.Schema
		expected []error
	}{
		"working": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "book"}}}},
			expected: nil},
		"invalid": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "query"}}}},
			expected: []error{&ValidationError{
				Code: "E007",
				ErrorDetails: ErrorDetails{
					Message:      "You have a reserved model name query",
					ShortMessage: "query is reserved",
					Hint:         "You cannot use query as a model name, it is reserved, try queryer",
				},
			}}},
	}

	for name, tc := range tests {
		got := noReservedModelNames(asInputs(tc.input))
		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}
}

//Supported field types
func TestSupportedFieldTypes(t *testing.T) {

	input1 := []*parser.ModelField{
		{Name: "userId", Type: "Text"},
	}
	input2 := []*parser.ModelField{
		{Name: "userId", Type: "Invalid"},
	}
	tests := map[string]struct {
		input    *parser.Schema
		expected []error
	}{
		"working": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input1, Operations: []*parser.ModelAction{{Name: "createBook", Type: parser.ActionTypeGet, Arguments: []*parser.ActionArg{{Name: "userId"}}}}},
		}}}}}, expected: nil},
		"invalid": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input2, Operations: []*parser.ModelAction{{Name: "createBook", Type: parser.ActionTypeGet, Arguments: []*parser.ActionArg{{Name: "userId"}}}}},
		}}}}}, expected: []error{&ValidationError{
			Code: "E009",
			ErrorDetails: ErrorDetails{
				Message:      "field userId has an unsupported type Invalid",
				ShortMessage: "Invalid isn't supported",
				Hint:         "Did you mean one of Boolean, Date, Enum, ID, Identity, Image, Text, Timestamp?",
			},
		}}},
	}

	for name, tc := range tests {
		got := supportedFieldTypes(asInputs(tc.input))
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

//Models must be globally unique
func TestModelsBeGloballyUnique(t *testing.T) {
	err := modelsGloballyUnique(asInputs(&parser.Schema{Declarations: []*parser.Declaration{
		{
			Model: &parser.Model{
				Name: "Book",
				Sections: []*parser.ModelSection{
					{
						Operations: []*parser.ModelAction{
							{
								Name: "createbook",
							},
							{
								Name: "dave",
							},
						},
					},
				},
			},
		},
		{
			Model: &parser.Model{
				Name: "Book",
				Sections: []*parser.ModelSection{
					{
						Operations: []*parser.ModelAction{
							{
								Name: "createbook",
							},
							{
								Name: "dave1",
							},
						},
					},
				},
			},
		},
	}}))

	expected := []error{
		&ValidationError{
			Code: "E010",
			ErrorDetails: ErrorDetails{
				Message:      "You have duplicate Models Model:Book",
				ShortMessage: "Book is duplicated",
				Hint:         "Remove Book",
			},
		},
	}

	assert.Equal(t, expected, err)
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
