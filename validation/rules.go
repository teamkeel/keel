package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/parser"
)

var (
	ReservedNames  = []string{"id", "createdAt", "updatedAt"}
	ReservedModels = []string{"query"}
)

type Validator struct {
	schema *parser.Schema
}

func NewValidator(schema *parser.Schema) *Validator {
	return &Validator{
		schema: schema,
	}
}

func (v *Validator) RunAllValidators() error {
	validatorFuncs := []func(*parser.Schema) error{
		modelsUpperCamel,
		fieldsOpsFuncsLowerCamel,
		fieldNamesMustBeUniqueInAModel,
		operationsUniqueGlobally,
		operationInputs,
		noReservedFieldNames,
		noReservedModelNames,
		operationUniqueFieldInput,
		supportedFieldTypes,
	}
	for _, vf := range validatorFuncs {
		err := vf(v.schema)
		if err != nil {
			return err
		}
	}
	return nil
}

//Models are UpperCamel
func modelsUpperCamel(schema *parser.Schema) error {
	for _, input := range schema.Declarations {
		reg := regexp.MustCompile("([A-Z][a-z0-9]+)+")

		if reg.FindString(input.Model.Name) != input.Model.Name {
			return fmt.Errorf("you have a model name that is not UpperCamel %s", input.Model.Name)
		}
	}

	return nil
}

//Fields/operations/functions are lowerCamel
func fieldsOpsFuncsLowerCamel(schema *parser.Schema) error {
	for _, input := range schema.Declarations {
		if strcase.ToLowerCamel(input.Model.Name) != input.Model.Name {
			return fmt.Errorf("you have a field name that is not lowerCamel %s", input.Model.Name)
		}

	}

	return nil
}

//Field names must be unique in a model
func fieldNamesMustBeUniqueInAModel(schema *parser.Schema) error {
	for _, model := range schema.Declarations {
		for _, sections := range model.Model.Sections {
			fieldNames := map[string]bool{}
			for _, name := range sections.Fields {
				if _, ok := fieldNames[name.Name]; ok {
					return fmt.Errorf("you have duplicate field names %s", name.Name)
				}
				fieldNames[name.Name] = true
			}
		}
	}

	return nil
}

type GlobalOperations struct {
	Name  string
	Model string
}

//Operations/functions must be globally unique
func operationsUniqueGlobally(schema *parser.Schema) error {
	var globalOperations []GlobalOperations

	for _, declaration := range schema.Declarations {
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
func operationInputs(schema *parser.Schema) error {
	functionFields := make(map[string][]*parser.FunctionArg, 0)
	for _, declaration := range schema.Declarations {
		for _, section := range declaration.Model.Sections {
			for _, function := range section.Functions {
				functionFields[function.Name] = function.Arguments
			}
		}

	}

	for _, input := range schema.Declarations {
		for _, modelName := range input.Model.Sections {
			for _, fields := range modelName.Fields {
				for functionName, functionField := range functionFields {
					for _, functionFieldName := range functionField {
						if functionFieldName.Name == fields.Name {
							delete(functionFields, functionName)
						}
					}
				}
			}
		}
	}

	if len(functionFields) > 0 {
		for k, v := range functionFields {
			var fields []string
			for _, field := range v {
				fields = append(fields, field.Name)
			}
			message := fmt.Sprintf("model:%s, field:%v", k, strings.Join(fields, ","))
			return fmt.Errorf("you are using inputs that are not fields %s", message)
		}
	}

	return nil
}

//No reserved field names (id, createdAt, updatedAt)
func noReservedFieldNames(schema *parser.Schema) error {
	for _, name := range ReservedNames {
		for _, dec := range schema.Declarations {
			for _, section := range dec.Model.Sections {
				for _, field := range section.Fields {
					if strings.EqualFold(name, field.Name) {
						return fmt.Errorf("you have a reserved field name %s", field.Name)
					}
				}
			}
		}
	}

	return nil
}

//No reserved model name (query)
func noReservedModelNames(schema *parser.Schema) error {
	for _, name := range ReservedModels {
		for _, dec := range schema.Declarations {
			if strings.EqualFold(name, dec.Model.Name) {
				return fmt.Errorf("you have a reserved model name %s", dec.Model.Name)
			}
		}
	}

	return nil
}

//GET operation must take a unique field as an input (or a unique combinations of inputs)
func operationUniqueFieldInput(schema *parser.Schema) error {
	// getAuthor(id)
	// A get operation can only accept a single field. Needs to be unique OR primary key

	for _, dec := range schema.Declarations {
		if dec.Model == nil {
			continue
		}

		for _, section := range dec.Model.Sections {
			if len(section.Functions) == 0 {
				continue
			}

			for _, function := range section.Functions {
				if !function.Get {
					continue
				}

				if len(function.Arguments) != 1 {
					return fmt.Errorf("get operation must take a unique field as an input")
				}

				arg := function.Arguments[0]
				isValid := false

				for _, section2 := range dec.Model.Sections {
					if len(section2.Fields) == 0 {
						continue
					}
					for _, field := range section2.Fields {
						if field.Name != arg.Name {
							continue
						}
						for _, attr := range field.Attributes {
							if attr.Name == "unique" {
								isValid = true
							}
							if attr.Name == "primaryKey" {
								isValid = true
							}
						}
					}
				}

				if !isValid {
					return fmt.Errorf("operation %s must take a unique field as an input", function.Name)
				}

			}
		}

	}

	return nil
}

//Supported field types
func supportedFieldTypes(schema *parser.Schema) error {
	var fieldTypes = map[string]bool{"Text": true, "Date": true, "Timestamp": true, "Image": true, "Boolean": true, "Enum": true, "Identity": true}

	for _, dec := range schema.Declarations {
		if dec.Model == nil {
			continue
		}

		for _, section := range dec.Model.Sections {
			for _, field := range section.Fields {
				if _, ok := fieldTypes[field.Type]; !ok {
					return fmt.Errorf("field %s has an unsupported type %s", field.Name, field.Type)
				}
			}
		}
	}

	return nil
}

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
