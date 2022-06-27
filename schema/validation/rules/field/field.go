package field

import (
	"fmt"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var (
	reservedFieldNames = []string{"id", "createdAt", "updatedAt"}
	BuiltInFieldTypes  = map[string]bool{
		"Text":             true,
		"Date":             true,
		"Timestamp":        true,
		"Image":            true,
		"Boolean":          true,
		"Number":           true,
		parser.FieldTypeID: true,
	}
)

func ReservedNameRule(asts []*parser.AST) []error {
	var errors []error

	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {

			if field.BuiltIn {
				continue
			}

			for _, reserved := range reservedFieldNames {
				if strings.EqualFold(reserved, field.Name.Value) {
					errors = append(
						errors,
						errorhandling.NewValidationError(errorhandling.ErrorReservedFieldName,
							errorhandling.TemplateLiterals{
								Literals: map[string]string{
									"Name":       field.Name.Value,
									"Suggestion": fmt.Sprintf("%ser", field.Name.Value),
								},
							},
							field.Name,
						),
					)

				}
			}
		}
	}

	return errors
}

func FieldNamingRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {
			if field.BuiltIn {
				continue
			}
			if strcase.ToLowerCamel(field.Name.Value) != field.Name.Value {
				errors = append(
					errors,
					errorhandling.NewValidationError(errorhandling.ErrorFieldNameLowerCamel,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"Name":      field.Name.Value,
								"Suggested": strcase.ToLowerCamel(strings.ToLower(field.Name.Value)),
							},
						},
						field.Name,
					),
				)
			}
		}
	}

	return errors
}

func UniqueFieldNamesRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		fieldNames := map[string]bool{}
		for _, field := range query.ModelFields(model) {
			// Ignore built in fields as usage of these field names is handled
			// by reservedFieldNamesRule
			if field.BuiltIn {
				continue
			}
			if _, ok := fieldNames[field.Name.Value]; ok {
				errors = append(
					errors,
					errorhandling.NewValidationError(errorhandling.ErrorFieldNamesUniqueInModel,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"Name": field.Name.Value,
								"Line": fmt.Sprint(field.Name.Pos.Line),
							},
						},
						field.Name,
					),
				)
			}

			fieldNames[field.Name.Value] = true
		}
	}

	return errors
}

func ValidFieldTypesRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {

			if _, ok := BuiltInFieldTypes[field.Type]; ok {
				continue
			}

			if query.IsUserDefinedType(asts, field.Type) {
				continue
			}

			validTypes := query.UserDefinedTypes(asts)
			for t := range BuiltInFieldTypes {
				validTypes = append(validTypes, t)
			}

			// todo feed hint suggestions into validation error somehow.
			sort.Strings(validTypes)

			hint := errorhandling.NewCorrectionHint(validTypes, field.Type)

			suggestions := formatting.HumanizeList(hint.Results, formatting.DelimiterOr)

			errors = append(
				errors,
				errorhandling.NewValidationError(errorhandling.ErrorUnsupportedFieldType,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Name":        field.Name.Value,
							"Type":        field.Type,
							"Suggestions": suggestions,
						},
					},
					field.Name,
				),
			)
		}
	}

	return errors
}
