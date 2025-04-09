package validation

import (
	"fmt"
	"strings"

	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func SequenceAttributeRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var field *parser.FieldNode
	hasSequence := false
	otherAttrs := []*parser.AttributeNode{}

	return Visitor{
		EnterField: func(f *parser.FieldNode) {
			field = f
		},
		LeaveField: func(n *parser.FieldNode) {
			if hasSequence && len(otherAttrs) > 0 {
				for _, attr := range otherAttrs {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.AttributeNotAllowedError,
							errorhandling.ErrorDetails{
								Message: fmt.Sprintf("@%s cannot be used together with @sequence", attr.Name.Value),
							},
							attr.Name,
						),
					)
				}
			}
			otherAttrs = []*parser.AttributeNode{}
			hasSequence = false
			field = nil
		},
		EnterAttribute: func(attr *parser.AttributeNode) {
			if field == nil {
				return
			}
			if attr.Name.Value != parser.AttributeSequence {
				otherAttrs = append(otherAttrs, attr)
				return
			}

			hasSequence = true

			if field.Type.Value != parser.FieldTypeText {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeNotAllowedError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("@sequence cannot be used on field of type %s", field.Type.Value),
						},
						attr.Name,
					),
				)
				return
			}

			if field.Repeated {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeNotAllowedError,
						errorhandling.ErrorDetails{
							Message: "@sequence cannot be used on repeated fields",
						},
						attr.Name,
					),
				)
			}

			if len(attr.Arguments) == 0 {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: "missing prefix argument e.g. @sequence(\"MYPREFIX\")",
						},
						attr,
					),
				)
			} else {
				expr := attr.Arguments[0].Expression
				v, isNull, err := resolve.ToValue[string](expr)
				if err != nil || isNull {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.AttributeArgumentError,
							errorhandling.ErrorDetails{
								Message: "first argument to @sequence must be a string",
							},
							expr,
						),
					)
				}
				if strings.Contains(v, " ") {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.AttributeArgumentError,
							errorhandling.ErrorDetails{
								Message: "prefix cannot contain spaces",
							},
							expr,
						),
					)
				}
			}

			if len(attr.Arguments) > 2 {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: "too many arguments to @sequence. Usage is @sequence(\"MYPREFIX\") or @sequence(\"MYPREFIX\", 1000)",
						},
						attr.Arguments[2],
					),
				)
			}

			if len(attr.Arguments) == 2 {
				expr := attr.Arguments[1].Expression
				v, isNull, err := resolve.ToValue[int64](expr)
				if err != nil || isNull {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.AttributeArgumentError,
							errorhandling.ErrorDetails{
								Message: "second argument to @sequence must be a number",
							},
							expr,
						),
					)
				} else if v < 0 {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.AttributeArgumentError,
							errorhandling.ErrorDetails{
								Message: "starting sequence value cannot be negative",
							},
							expr,
						),
					)
				}
			}
		},
	}
}
