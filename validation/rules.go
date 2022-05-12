package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
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
	validatorFuncs := []func([]Input) []error{
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
		modelsGloballyUnique,
	}
	var errors []*ValidationError
	for _, vf := range validatorFuncs {
		err := vf(v.inputs)
		for _, e := range err {
			if verrs, ok := e.(*ValidationError); ok {
				errors = append(errors, verrs)
			}
		}
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// Models are UpperCamel
func modelsUpperCamel(inputs []Input) []error {
	var errors []error
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
				errors = append(errors, validationError(fmt.Sprintf("you have a model name that is not UpperCamel %s", decl.Model.Name),
					fmt.Sprintf("%s is not UpperCamel", decl.Model.Name),
					strcase.ToCamel(strings.ToLower(decl.Model.Name)),
					decl.Model.Pos))
			}
		}
	}
	return errors
}

//Fields/operations are lowerCamel
func fieldsOpsFuncsLowerCamel(inputs []Input) []error {
	var errors []error
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
						errors = append(errors, validationError(fmt.Sprintf("you have a field name that is not lowerCamel %s", field.Name),
							fmt.Sprintf("%s isn't lower camel", field.Name),
							strcase.ToLowerCamel(strings.ToLower(field.Name)),
							field.Pos))

					}
				}
				for _, function := range model.Operations {
					if strcase.ToLowerCamel(function.Name) != function.Name {
						errors = append(errors, validationError(fmt.Sprintf("you have a function name that is not lowerCamel %s", function.Name),
							fmt.Sprintf("%s isn't lower camel", function.Name),
							strcase.ToLowerCamel(strings.ToLower(function.Name)),
							function.Pos))

					}
				}
			}
		}
	}

	return errors
}

//Field names must be unique in a model
func fieldNamesMustBeUniqueInAModel(inputs []Input) []error {
	var errors []error
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
						errors = append(errors, validationError(
							fmt.Sprintf("you have duplicate field names %s", name.Name),
							fmt.Sprintf("%s is duplicated", name.Name),
							fmt.Sprintf(`Remove '%s' on line %v`, name.Name, name.Pos.Line),
							name.Pos))

					}
					fieldNames[name.Name] = true
				}
			}
		}
	}
	return errors
}

type GlobalOperations struct {
	Name  string
	Model string
	Pos   lexer.Position
}

func uniqueOperationsGlobally(inputs []Input) []GlobalOperations {
	var globalOperations []GlobalOperations
	for _, input := range inputs {
		schema := input.ParsedSchema
		for _, declaration := range schema.Declarations {
			if declaration.Model == nil {
				continue
			}
			for _, sec := range declaration.Model.Sections {
				for _, functionNames := range sec.Operations {
					globalOperations = append(globalOperations, GlobalOperations{
						Name: functionNames.Name, Model: declaration.Model.Name, Pos: functionNames.Pos,
					})
				}
			}
		}
	}
	return globalOperations
}

//Operations must be globally unique
func operationsUniqueGlobally(inputs []Input) []error {
	var errors []error
	var operationNames []string

	globalOperations := uniqueOperationsGlobally(inputs)

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

	for _, nameError := range duplicationOperations {
		errors = append(errors, validationError(
			fmt.Sprintf("you have duplicate operations Model:%s Name:%s", nameError.Model, nameError.Name),
			fmt.Sprintf("%s is duplicated", nameError.Name),
			fmt.Sprintf(`Remove '%s' on line %v`, nameError.Name, nameError.Pos.Line),
			nameError.Pos))

	}

	return errors
}

type operationInputFields struct {
	Fields []*parser.ActionArg
	Pos    lexer.Position
}

//Inputs of ops must be model fields
func operationInputs(inputs []Input) []error {
	var errors []error

	functionFields := make(map[string]*operationInputFields, 0)

	for _, input := range inputs {
		schema := input.ParsedSchema

		for _, declaration := range schema.Declarations {
			if declaration.Model == nil {
				continue
			}
			for _, section := range declaration.Model.Sections {
				for _, function := range section.Operations {
					if len(function.Arguments) == 0 {
						continue
					}
					functionFields[function.Name] = &operationInputFields{
						Fields: function.Arguments,
						Pos:    function.Pos,
					}
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
						for _, functionFieldName := range functionField.Fields {
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
			for _, field := range v.Fields {
				message := fmt.Sprintf("model:%s, field:%v", k, field.Name)
				errors = append(errors, validationError(fmt.Sprintf("you are using inputs that are not fields %s", message),
					fmt.Sprintf("Replace %s", field.Name),
					fmt.Sprintf("Check inputs for %s", k),
					field.Pos,
				))
			}

		}
	}

	return errors
}

//No reserved field names (id, createdAt, updatedAt)
func noReservedFieldNames(inputs []Input) []error {
	var errors []error
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
							errors = append(errors, validationError(fmt.Sprintf("you have a reserved field name %s", field.Name),
								fmt.Sprintf("cannot use %s", field.Name),
								fmt.Sprintf("You cannot use %s as field name, it is reserved, try %ser", field.Name, field.Name),
								field.Pos))
						}
					}
				}
			}
		}
	}

	return errors
}

//No reserved model name (query)
func noReservedModelNames(inputs []Input) []error {
	var errors []error
	for _, input := range inputs {
		schema := input.ParsedSchema
		for _, name := range ReservedModels {
			for _, dec := range schema.Declarations {
				if dec.Model == nil {
					continue
				}
				if strings.EqualFold(name, dec.Model.Name) {
					errors = append(errors, validationError(fmt.Sprintf("you have a reserved model name %s", dec.Model.Name),
						fmt.Sprintf("%s is reserved", dec.Model.Name),
						fmt.Sprintf("You cannot use %s as a model name, it is reserved, try %ser", dec.Model.Name, dec.Model.Name),
						dec.Model.Pos))

				}
			}
		}
	}

	return errors
}

//GET operation must take a unique field as an input (or a unique combinations of inputs)
func operationUniqueFieldInput(inputs []Input) []error {
	var errors []error

	for _, input := range inputs {
		schema := input.ParsedSchema

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
						errors = append(errors, validationError(
							fmt.Sprintf("operation %s must take a unique field as an input", function.Name),
							fmt.Sprintf("%s requires a unique field", function.Name),
							"Are you using a unique field?",
							function.Pos))
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
						errors = append(errors, validationError(
							fmt.Sprintf("operation %s must take a unique field as an input", function.Name),
							fmt.Sprintf("%s requires a unique field", function.Name),
							"Are you sure you are using a unique field?",
							function.Pos))
					}

				}
			}

		}
	}

	return errors
}

//Supported field types
func supportedFieldTypes(inputs []Input) []error {
	var errors []error

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
						errors = append(errors, validationError(
							fmt.Sprintf("field %s has an unsupported type %s", field.Name, field.Type),
							fmt.Sprintf("%s isn't supported", field.Type),
							"Have you tried Text?",
							field.Pos))
					}
				}
			}
		}
	}

	return errors
}

//Models are globally unique
func modelsGloballyUnique(inputs []Input) []error {
	var errors []error
	var modelNames []string

	globalOperations := uniqueModelsGlobally(inputs)

	for _, name := range globalOperations {
		modelNames = append(modelNames, name.Model)
	}
	duplicates := findDuplicates(modelNames)

	if len(duplicates) == 0 {
		return nil
	}

	var duplicateModels []GlobalOperations
	for _, model := range globalOperations {
		for _, duplicate := range duplicates {
			if model.Model == duplicate {
				duplicateModels = append(duplicateModels, model)
			}
		}
	}

	for _, nameError := range duplicateModels {
		errors = append(errors, validationError(
			fmt.Sprintf("you have duplicate Models Model:%s Pos:%s", nameError.Model, nameError.Pos),
			fmt.Sprintf("%s is duplicated", nameError.Model),
			fmt.Sprintf("Remove %s", nameError.Model),
			nameError.Pos))
	}

	return errors
}

func uniqueModelsGlobally(inputs []Input) []GlobalOperations {
	var globalOperations []GlobalOperations
	for _, input := range inputs {
		schema := input.ParsedSchema
		for _, declaration := range schema.Declarations {
			if declaration.Model == nil {
				continue
			}

			globalOperations = append(globalOperations, GlobalOperations{
				Model: declaration.Model.Name,
			})
		}
	}
	return globalOperations
}

func supportedAttributeTypes(inputs []Input) []error {
	var errors []error

	for _, input := range inputs {
		schema := input.ParsedSchema

		for _, dec := range schema.Declarations {
			if dec.Model != nil {
				for _, section := range dec.Model.Sections {
					if section.Attribute != nil {
						errors = append(errors, checkAttributes([]*parser.Attribute{section.Attribute}, "model", dec.Model.Name)...)
					}

					if section.Operations != nil {
						for _, op := range section.Operations {
							errors = append(errors, checkAttributes(op.Attributes, "operation", op.Name)...)
						}
					}

					if section.Functions != nil {
						for _, function := range section.Functions {
							errors = append(errors, checkAttributes(function.Attributes, "function", function.Name)...)
						}
					}

					if section.Fields != nil {
						for _, field := range section.Fields {
							errors = append(errors, checkAttributes(field.Attributes, "field", field.Name)...)
						}
					}
				}
			}

			// Validate attributes defined within api sections
			if dec.API != nil {
				for _, section := range dec.API.Sections {
					if section.Attribute != nil {
						errors = append(errors, checkAttributes([]*parser.Attribute{section.Attribute}, "api", dec.API.Name)...)
					}
				}
			}
		}
	}

	return errors
}

func checkAttributes(attributes []*parser.Attribute, definedOn string, parentName string) []error {
	var supportedAttributes = map[string][]string{
		"model":     {"permission"},
		"api":       {"graphql"},
		"field":     {"unique"},
		"operation": {"set", "where", "get", "permission"},
		"function":  {"permission"},
	}

	var builtIns = map[string][]string{
		"model":     {},
		"api":       {},
		"operation": {},
		"function":  {},
		"field":     {"primaryKey"},
	}

	errors := make([]error, 0)

	for _, attr := range attributes {
		if contains(builtIns[definedOn], attr.Name) {
			continue
		}

		if !contains(supportedAttributes[definedOn], attr.Name) {
			errors = append(errors, validationError(fmt.Sprintf("%s '%s' has an unrecognised attribute @%s", definedOn, parentName, attr.Name), fmt.Sprintf("Unrecognised attribute %s", attr.Name), "Did you mean XX?", attr.Pos))
		}
	}

	return errors
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
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}
