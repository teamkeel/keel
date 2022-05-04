package validation

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/parser"
)

//Models are UpperCamel
func TestModelsAreUpperCamel(t *testing.T) {
	tests := map[string]struct {
		input    *parser.Schema
		expected error
	}{
		"simple":     {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "Book"}}}}, expected: nil},
		"long":       {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "BookAuthorLibrary"}}}}, expected: nil},
		"allLower":   {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "bookauthor"}}}}, expected: fmt.Errorf("you have a model name that is not UpperCamel bookauthor")},
		"allUpper":   {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "BOOKAUTHOR"}}}}, expected: fmt.Errorf("you have a model name that is not UpperCamel BOOKAUTHOR")},
		"underscore": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "book_author"}}}}, expected: fmt.Errorf("you have a model name that is not UpperCamel book_author")},
	}

	for name, tc := range tests {
		got := modelsUpperCamel(tc.input)
		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}
}

//Fields/operations/functions are lowerCamel
func TestFieldsOpsFuncsLowerCamel(t *testing.T) {
	tests := map[string]struct {
		input    *parser.Schema
		expected error
	}{
		"simple": {input: &parser.Schema{Declarations: []*parser.Declaration{{
			Model: &parser.Model{Name: "createBook"}}}}, expected: nil},
		"long": {input: &parser.Schema{Declarations: []*parser.Declaration{{
			Model: &parser.Model{Name: "createBooksForService"}}}}, expected: nil},
		"allLower": {input: &parser.Schema{Declarations: []*parser.Declaration{{
			Model: &parser.Model{Name: "bookauthor"}}}}, expected: nil},
		"allUpper": {input: &parser.Schema{Declarations: []*parser.Declaration{{
			Model: &parser.Model{Name: "CREATEBOOK"}}}}, expected: fmt.Errorf("you have a field name that is not lowerCamel CREATEBOOK")},
		"underscore": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{
			Name: "book_author"}}}}, expected: fmt.Errorf("you have a field name that is not lowerCamel book_author")},
	}

	for name, tc := range tests {
		got := fieldsOpsFuncsLowerCamel(tc.input)
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
		expected error
	}{
		"working": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input1},
		}}}}}, expected: nil},
		"long": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input2},
		}}}}}, expected: errors.New("you have duplicate field names name")},
	}

	for name, tc := range tests {
		got := fieldNamesMustBeUniqueInAModel(tc.input)
		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}

}

//Operations/functions must be globally unique
func TestFindOpsFuncsMustBeGloballyUnique(t *testing.T) {
	input1 := []GlobalOperations{
		{Name: "deleteBook", Model: "book"},
		{Name: "createBook", Model: "author"},
	}
	input2 := []GlobalOperations{
		{Name: "createBook", Model: "book"},
		{Name: "createBook", Model: "author"},
	}

	tests := map[string]struct {
		input    []GlobalOperations
		expected error
	}{
		"working": {input: input1, expected: nil},
		"invalid": {input: input2, expected: errors.New("you have duplicate operations [{createBook book} {createBook author}]")},
	}

	for name, tc := range tests {
		got := findDuplicatesOperations(tc.input)
		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}
}

//Inputs of ops must be model fields
func TestOpsFuncsMustBeGloballyUnique(t *testing.T) {
	err := operationsUniqueGlobally(&parser.Schema{Declarations: []*parser.Declaration{
		{
			Model: &parser.Model{
				Name: "book",
				Sections: []*parser.ModelSection{
					{
						Functions: []*parser.ModelAction{
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
						Functions: []*parser.ModelAction{
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

	assert.Equal(t, errors.New("you have duplicate operations [{createbook book} {createbook book}]"), err)
}

//Inputs of ops must be model fields
func TestInputsModelFields(t *testing.T) {
	tests := map[string]struct {
		input    *parser.Schema
		expected error
	}{
		"passing": {input: &parser.Schema{Declarations: []*parser.Declaration{
			{
				Model: &parser.Model{
					Sections: []*parser.ModelSection{

						{
							Fields: []*parser.ModelField{
								{
									Name: "id",
								},
							},
							Functions: []*parser.ModelAction{
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
							Functions: []*parser.ModelAction{
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
										{Name: "name"},
									},
								},
							},
						},
					}}}}}, expected: fmt.Errorf("you are using inputs that are not fields model:author, field:name")}}

	for name, tc := range tests {
		got := operationInputs(tc.input)

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
		expected error
	}{
		"working": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input1},
		}}}}}, expected: nil},
		"invalid": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input2},
		}}}}}, expected: errors.New("you have a reserved field name id")},
		"invalidUpperCase": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input3},
		}}}}}, expected: errors.New("you have a reserved field name ID")},
	}

	for name, tc := range tests {
		got := noReservedFieldNames(tc.input)
		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}
}

//No reserved model name (query)
func TestReservedModelNames(t *testing.T) {
	tests := map[string]struct {
		input    *parser.Schema
		expected error
	}{
		"working": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "book"}}}}, expected: nil},
		"invalid": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "query"}}}}, expected: errors.New("you have a reserved model name query")},
	}

	for name, tc := range tests {
		got := noReservedModelNames(tc.input)
		if !assert.Equal(t, tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}
}

//GET operation must take a unique field as an input (or a unique combinations of inputs)
func TestGetOperationMustTakeAUniqueFieldAsAnInput(t *testing.T) {
	tests := map[string]struct {
		input    *parser.Schema
		expected error
	}{
		"working": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "book", Sections: []*parser.ModelSection{
			{
				Fields: []*parser.ModelField{
					{Name: "id", Type: "int", Attributes: []*parser.Attribute{{Name: "primaryKey"}}},
					{Name: "name", Type: "string", Attributes: []*parser.Attribute{{Name: "unique"}}},
				},
			}, {
				Functions: []*parser.ModelAction{
					{
						Type: parser.ActionTypeGet,
						Name: "createBook",
						Arguments: []*parser.ActionArg{
							{Name: "id"},
						},
					},
				},
			},
		}}}}}, expected: nil},
		"invalid": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Name: "book", Sections: []*parser.ModelSection{
			{
				Fields: []*parser.ModelField{
					{Name: "id", Type: "int", Attributes: []*parser.Attribute{{Name: "primaryKey"}}},
					{Name: "name", Type: "string"},
				},
			}, {
				Functions: []*parser.ModelAction{
					{
						Type: parser.ActionTypeGet,
						Name: "createBook",
						Arguments: []*parser.ActionArg{
							{Name: "name"},
						},
					},
				},
			},
		}}}}}, expected: fmt.Errorf("operation createBook must take a unique field as an input")},
	}

	for name, tc := range tests {
		got := operationUniqueFieldInput(tc.input)
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
		expected error
	}{
		"working": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input1, Functions: []*parser.ModelAction{{Name: "createBook", Type: parser.ActionTypeGet, Arguments: []*parser.ActionArg{{Name: "userId"}}}}},
		}}}}}, expected: nil},
		"invalid": {input: &parser.Schema{Declarations: []*parser.Declaration{{Model: &parser.Model{Sections: []*parser.ModelSection{
			{Fields: input2, Functions: []*parser.ModelAction{{Name: "createBook", Type: parser.ActionTypeGet, Arguments: []*parser.ActionArg{{Name: "userId"}}}}},
		}}}}}, expected: errors.New("field userId has an unsupported type Invalid")},
	}

	for name, tc := range tests {
		got := supportedFieldTypes(tc.input)
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
