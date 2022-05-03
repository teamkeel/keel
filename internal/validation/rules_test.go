package validation

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/parser"
)

//Models are UpperCamel
func TestModelsAreUpperCamel(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"simple":     {input: "Book", expected: true},
		"long":       {input: "BookAuthorLibrary", expected: true},
		"allLower":   {input: "bookauthor", expected: false},
		"allUpper":   {input: "BOOKAUTHOR", expected: false},
		"underscore": {input: "book_author", expected: false},
	}

	for name, tc := range tests {
		got := ModelsUpperCamel(tc.input)
		if !reflect.DeepEqual(tc.expected, got) {
			t.Fatalf("%s: expected: %v, got: %v", name, tc.expected, got)
		}
	}
}

//Fields/operations/functions are lowerCamel
func TestFieldsOpsFuncsLowerCamel(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected bool
	}{
		"simple":     {input: "createBook", expected: true},
		"long":       {input: "createBooksForService", expected: true},
		"allLower":   {input: "bookauthor", expected: true},
		"allUpper":   {input: "CREATEBOOK", expected: false},
		"underscore": {input: "book_author", expected: false},
	}

	for name, tc := range tests {
		got := FieldsOpsFuncsLowerCamel(tc.input)
		if !reflect.DeepEqual(tc.expected, got) {
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
		input    []*parser.ModelField
		expected error
	}{
		"working": {input: input1, expected: nil},
		"long":    {input: input2, expected: errors.New("you have duplicate field names [name]")},
	}

	for name, tc := range tests {
		got := FieldNamesMustBeUniqueInAModel(tc.input)
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
	err := OperationsUniqueGlobally([]*parser.Declaration{
		{
			Model: &parser.Model{
				Name: "book",
				Sections: []*parser.ModelSection{
					{
						Functions: []*parser.ModelFunction{
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
						Functions: []*parser.ModelFunction{
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
	})

	assert.Equal(t, errors.New("you have duplicate operations [{createbook book} {createbook book}]"), err)

}
func TestInputsModelFields(t *testing.T) {

	//Inputs of ops must be model fields

}

//No reserved field names (id, createdAt, updatedAt)
func TestNoReservedFieldNames(t *testing.T) {

}

//No reserved model name (query)
func TestReservedModelNames(t *testing.T) {

}

//GET operation must take a unique field as an input (or a unique combinations of inputs)
func TestGetOperationMustTakeAUniqueFieldAsAnInput(t *testing.T) {

}

//Supported field types
func TestSupportedFieldTypes(t *testing.T) {
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
