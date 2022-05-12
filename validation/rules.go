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

// A Validator knows how to validate a parsed Keel schema.
//
// Conceptually we are validating a single schema.
// But the Validator supports it being "delivered" as a collection
// of *parser.Schema objects - to match up with a user's schema likely
// being written across N files.
//
// We use a []Input to model the inputs - so that the original file names are
// available for error reporting. (TODO although that is not implemented yet).
type Validator struct {
	inputs []Input
}

func NewValidator(inputs []Input) *Validator {
	return &Validator{
		inputs: inputs,
	}
}

func (v *Validator) RunAllValidators() error {
	validatorFuncs := []func([]Input) error{
		modelsUpperCamel,
		fieldsOpsFuncsLowerCamel,
		fieldNamesMustBeUniqueInAModel,
		operationsUniqueGlobally,
		operationInputs,
		noReservedFieldNames,
		noReservedModelNames,
		operationUniqueFieldInput,
		supportedFieldTypes,
		supportedAttributeTypes,
	}
	for _, vf := range validatorFuncs {
		err := vf(v.inputs)
		if err != nil {
			return err
		}
	}
	return nil
}

//Models are UpperCamel
func modelsUpperCamel(inputs []Input) error {
	for _, input := range inputs {
		schema := input.ParsedSchema
		for _, decl := range schema.Declarations {
			if decl.Model == nil {
				continue
			}
			// todo - these MustCompile regex would be better at module scope, to
			// make the MustCompile panic a load-time thing rather than a runtime thing.
			reg := regexp.MustCompile("([A-Z][a-z0-9]+)+")

			if reg.FindString(decl.Model.Name) != decl.Model.Name {
				return fmt.Errorf("you have a model name that is not UpperCamel %s", decl.Model.Name)
			}
		}
	}
	return nil
}

//Fields/operations/functions are lowerCamel
func fieldsOpsFuncsLowerCamel(inputs []Input) error {
	for _, input := range inputs {
		schema := input.ParsedSchema

		for _, decl := range schema.Declarations {
			if decl.Model == nil {
				continue
			}
			for _, model := range decl.Model.Sections {
				for _, field := range model.Fields {
					if field.BuiltIn {
						continue
					}
					if strcase.ToLowerCamel(field.Name) != field.Name {
						return fmt.Errorf("you have a field name that is not lowerCamel %s", field.Name)
					}
				}
				for _, function := range model.Functions {
					if strcase.ToLowerCamel(function.Name) != function.Name {
						return fmt.Errorf("you have a function name that is not lowerCamel %s", function.Name)
					}
				}
			}
		}
	}
	return nil
}

//Field names must be unique in a model
func fieldNamesMustBeUniqueInAModel(inputs []Input) error {
	for _, input := range inputs {
		schema := input.ParsedSchema
		for _, model := range schema.Declarations {
			if model.Model == nil {
				continue
			}
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
	}
	return nil
}

type GlobalOperations struct {
	Name  string
	Model string
}

//Operations/functions must be globally unique
func operationsUniqueGlobally(inputs []Input) error {
	var globalOperations []GlobalOperations
	for _, input := range inputs {
		schema := input.ParsedSchema
		for _, declaration := range schema.Declarations {
			if declaration.Model == nil {
				continue
			}
			for _, sec := range declaration.Model.Sections {
				for _, functionNames := range sec.Functions {
					globalOperations = append(globalOperations, GlobalOperations{
						Name: functionNames.Name, Model: declaration.Model.Name,
					})
				}
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
func operationInputs(inputs []Input) error {
	functionFields := make(map[string][]*parser.ActionArg, 0)

	for _, input := range inputs {
		schema := input.ParsedSchema

		for _, declaration := range schema.Declarations {
			if declaration.Model == nil {
				continue
			}
			for _, section := range declaration.Model.Sections {
				for _, function := range section.Functions {
					functionFields[function.Name] = function.Arguments
				}
			}

		}

		for _, input := range schema.Declarations {
			if input.Model == nil {
				continue
			}
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
func noReservedFieldNames(inputs []Input) error {
	for _, input := range inputs {
		schema := input.ParsedSchema
		for _, name := range ReservedNames {
			for _, dec := range schema.Declarations {
				if dec.Model == nil {
					continue
				}
				for _, section := range dec.Model.Sections {
					for _, field := range section.Fields {
						if field.BuiltIn {
							continue
						}
						if strings.EqualFold(name, field.Name) {
							return fmt.Errorf("you have a reserved field name %s", field.Name)
						}
					}
				}
			}
		}
	}
	return nil
}

//No reserved model name (query)
func noReservedModelNames(inputs []Input) error {
	for _, input := range inputs {
		schema := input.ParsedSchema
		for _, name := range ReservedModels {
			for _, dec := range schema.Declarations {
				if dec.Model == nil {
					continue
				}
				if strings.EqualFold(name, dec.Model.Name) {
					return fmt.Errorf("you have a reserved model name %s", dec.Model.Name)
				}
			}
		}
	}
	return nil
}

//GET operation must take a unique field as an input (or a unique combinations of inputs)
func operationUniqueFieldInput(inputs []Input) error {
	for _, input := range inputs {
		schema := input.ParsedSchema
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
					if function.Type != parser.ActionTypeGet {
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
	}

	return nil
}

//Supported field types
func supportedFieldTypes(inputs []Input) error {
	var fieldTypes = map[string]bool{"Text": true, "Date": true, "Timestamp": true, "Image": true, "Boolean": true, "Enum": true, "Identity": true, parser.FieldTypeID: true}

	for _, input := range inputs {
		schema := input.ParsedSchema
		for _, dec := range schema.Declarations {
			if dec.Model != nil {
				fieldTypes[dec.Model.Name] = true
			}
		}

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
	}

	return nil
}

func supportedAttributeTypes(inputs []Input) error {
	var supportedAttributes = map[string][]string{
		"model":     {"permission"},
		"api":       {"graphql"},
		"field":     {"unique"},
		"operation": {"set", "where", "get"},
	}

	for _, input := range inputs {
		schema := input.ParsedSchema

		for _, dec := range schema.Declarations {

			// Validate attributes defined within model sections

			if dec.Model != nil {
				for _, section := range dec.Model.Sections {
					if section.Attribute != nil {
						if !contains(supportedAttributes["model"], section.Attribute.Name) {
							return fmt.Errorf("model '%s' has an unrecognised attribute @%s", dec.Model.Name, section.Attribute.Name)
						}
					}

					if section.Operations != nil {
						for _, op := range section.Operations {
							for _, operationAttr := range op.Attributes {
								if !contains(supportedAttributes["operation"], operationAttr.Name) {
									return fmt.Errorf("operation '%s' has an unrecognised attribute @%s", op.Name, operationAttr.Name)
								}
							}
						}
					}

					for _, field := range section.Fields {
						for _, fieldAttr := range field.Attributes {
							if !contains(supportedAttributes["field"], fieldAttr.Name) {
								return fmt.Errorf("field '%s' has an unrecognised attribute @%s", field.Name, fieldAttr.Name)
							}
						}
					}
				}
			}

			// Validate attributes defined within api sections
			if dec.API != nil {
				for _, section := range dec.API.Sections {
					if section.Attribute != nil {
						if !contains(supportedAttributes["api"], section.Attribute.Name) {
							return fmt.Errorf("api '%s' has an unrecognised attribute @%s", dec.API.Name, section.Attribute.Name)
						}
					}
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

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))

	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]

	return ok
}
