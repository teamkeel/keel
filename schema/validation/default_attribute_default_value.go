package validation

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func DefaultAttributeExpressionRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var field *parser.FieldNode

	return Visitor{
		EnterField: func(f *parser.FieldNode) {
			field = f
		},
		LeaveField: func(_ *parser.FieldNode) {
			field = nil
		},
		EnterAttribute: func(attribute *parser.AttributeNode) {
			if attribute == nil || attribute.Name.Value != parser.AttributeDefault {
				return
			}

			typesWithZeroValue := []string{"Text", "Number", "Boolean", "ID", "Timestamp"}
			if len(attribute.Arguments) == 0 && !lo.Contains(typesWithZeroValue, field.Type.Value) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "This default attribute requires an expression",
						Hint:    "Try @default(MyDefaultValue) instead",
					},
					attribute,
				))
			}
		},
	}
}
