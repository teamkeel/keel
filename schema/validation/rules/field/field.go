package field

import (
	"fmt"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var (
	reservedFieldNames = []string{"id", "createdAt", "updatedAt"}
)

func ReservedNameRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {

	for _, model := range parser.Models(asts) {
		for _, field := range parser.ModelFields(model) {

			if field.BuiltIn {
				continue
			}

			for _, reserved := range reservedFieldNames {
				if strings.EqualFold(reserved, field.Name.Value) {
					errs.Append(errorhandling.ErrorReservedFieldName,
						map[string]string{
							"Name":       field.Name.Value,
							"Suggestion": fmt.Sprintf("%ser", field.Name.Value),
						},
						field.Name,
					)
				}
			}
		}
	}

	return
}

func FieldNamingRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range parser.Models(asts) {
		for _, field := range parser.ModelFields(model) {
			if field.BuiltIn {
				continue
			}
			if strcase.ToLowerCamel(field.Name.Value) != field.Name.Value {
				errs.Append(errorhandling.ErrorFieldNameLowerCamel,
					map[string]string{
						"Name":      field.Name.Value,
						"Suggested": strcase.ToLowerCamel(strings.ToLower(field.Name.Value)),
					},
					field.Name,
				)
			}
		}
	}

	return
}

func UniqueFieldNamesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range parser.Models(asts) {
		fieldNames := map[string]bool{}
		for _, field := range parser.ModelFields(model) {
			// Ignore built in fields as usage of these field names is handled
			// by reservedFieldNamesRule
			if field.BuiltIn {
				continue
			}
			if _, ok := fieldNames[field.Name.Value]; ok {
				errs.Append(errorhandling.ErrorFieldNamesUniqueInModel,
					map[string]string{
						"Name": field.Name.Value,
						"Line": fmt.Sprint(field.Name.Pos.Line),
					},
					field.Name,
				)
			}

			fieldNames[field.Name.Value] = true
		}
	}

	return
}

func ValidFieldTypesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range parser.Models(asts) {
		for _, field := range parser.ModelFields(model) {

			if parser.IsBuiltInFieldType(field.Type) {
				continue
			}

			if parser.IsUserDefinedType(asts, field.Type) {
				continue
			}

			validTypes := parser.UserDefinedTypes(asts)
			for t := range parser.BuiltInTypes {
				validTypes = append(validTypes, t)
			}

			// todo feed hint suggestions into validation error somehow.
			sort.Strings(validTypes)

			hint := errorhandling.NewCorrectionHint(validTypes, field.Type)

			suggestions := formatting.HumanizeList(hint.Results, formatting.DelimiterOr)

			errs.Append(errorhandling.ErrorUnsupportedFieldType,
				map[string]string{
					"Name":        field.Name.Value,
					"Type":        field.Type,
					"Suggestions": suggestions,
				},
				field.Name,
			)
		}
	}

	return
}
