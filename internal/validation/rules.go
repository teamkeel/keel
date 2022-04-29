package validation

import (
	"fmt"
	"regexp"

	"github.com/teamkeel/keel/parser"
)

type GlobalOperations struct {
	Name  string
	Model string
}

//Models are UpperCamel
func ModelsUpperCamel(input string) bool {
	r, err := regexp.Compile("([A-Z][a-z0-9]+)+")
	if err != nil {
		return false
	}

	return r.FindString(input) == input
}

//Fields/operations/functions are lowerCamel
func FieldsOpsFuncsLowerCamel(input string) bool {
	r, err := regexp.Compile("[a-z]+[A-Z0-9][a-z0-9]+[A-Za-z0-9]*")
	if err != nil {
		return false
	}

	if r.FindString(input) == input {
		return true
	}

	lower, err := regexp.Compile("[a-z]+")
	if err != nil {
		return false
	}

	return lower.FindString(input) == input

}

//Field names must be unique in a model
func FieldNamesMustBeUniqueInAModel(input []*parser.ModelField) error {
	var fieldNames []string

	for _, name := range input {
		fieldNames = append(fieldNames, name.Name)
	}
	duplicates := findDuplicates(fieldNames)

	if len(duplicates) > 0 {
		return fmt.Errorf("you have duplicate field names %v", duplicates)
	}

	return nil
}

//Operations/functions must be globally unique
func OperationsUniqueGlobally(input []*parser.Declaration) error {
	var globalOperations []GlobalOperations

	for _, declaration := range input {
		for _, sec := range declaration.Model.Sections {
			for _, functionNames := range sec.Functions {
				globalOperations = append(globalOperations, GlobalOperations{
					Name: functionNames.Name, Model: declaration.Model.Name,
				})
			}
		}
	}
	return findDuplicatesOperations(globalOperations)
}

func findDuplicatesOperations(globalOperations []GlobalOperations) error {
	var operationNames []string

	for _, name := range globalOperations {
		operationNames = append(operationNames, name.Name)
	}
	duplicates := findDuplicates(operationNames)

	if len(duplicates) == 0 {
		return nil
	}

	var duplicationOperations []GlobalOperations

	for _, function := range globalOperations {
		for _, duplicate := range duplicates {
			if function.Name == duplicate {
				duplicationOperations = append(duplicationOperations, function)
			}
		}
	}

	return fmt.Errorf("you have duplicate operations %v", duplicationOperations)
}

//Inputs of ops must be model fields

//No reserved field names (id, createdAt, updatedAt)

//No reserved model name (query)

//GET operation must take a unique field as an input (or a unique combinations of inputs)

//Supported field types

func findDuplicates(s []string) []string {
	inResult := make(map[string]bool)
	var result []string

	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
		} else {
			result = append(result, str)
		}
	}
	return result
}
