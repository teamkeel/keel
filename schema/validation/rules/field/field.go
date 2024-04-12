package field

import (
	"fmt"
	"sort"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func UniqueFieldNamesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		fieldNames := map[string]*parser.FieldNode{}
		for _, field := range query.ModelFields(model) {
			if existingField, ok := fieldNames[field.Name.Value]; ok {
				if field.BuiltIn {
					errs.Append(errorhandling.ErrorReservedFieldName,
						map[string]string{
							"Name": field.Name.Value,
							"Line": fmt.Sprint(field.Name.Pos.Line),
						},
						existingField.Name,
					)
				} else {
					errs.Append(errorhandling.ErrorFieldNamesUniqueInModel,
						map[string]string{
							"Name": field.Name.Value,
							"Line": fmt.Sprint(field.Name.Pos.Line),
						},
						field.Name,
					)
				}
			}

			fieldNames[field.Name.Value] = field
		}
	}

	return
}

func ValidFieldTypesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {

			if parser.IsBuiltInFieldType(field.Type.Value) {
				continue
			}

			if query.IsUserDefinedType(asts, field.Type.Value) {
				continue
			}

			validTypes := query.UserDefinedTypes(asts)
			for t := range parser.BuiltInTypes {
				validTypes = append(validTypes, t)
			}

			// todo feed hint suggestions into validation error somehow.
			sort.Strings(validTypes)

			hint := errorhandling.NewCorrectionHint(validTypes, field.Type.Value)

			suggestions := formatting.HumanizeList(hint.Results, formatting.DelimiterOr)

			errs.Append(errorhandling.ErrorUnsupportedFieldType,
				map[string]string{
					"Name":        field.Name.Value,
					"Type":        field.Type.Value,
					"Suggestions": suggestions,
				},
				field.Name,
			)
		}
	}

	return
}

// FieldNamesMaxLengthRule will validate that field names are smaller than the maximum allowed by postgres (63 bytes).
func FieldNamesMaxLengthRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	const MAX_BYTES = 63

	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {
			identifier := casing.ToSnake(field.Name.Value)

			if len(identifier) > MAX_BYTES {
				errs.Append(errorhandling.ErrorFieldNamesMaxLength,
					map[string]string{
						"Name": field.Name.Value,
					},
					field.Name,
				)
			}
		}
	}

	return
}
