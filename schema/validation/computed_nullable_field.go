package validation

import (
	"fmt"

	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ComputedNullableFieldRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var model *parser.ModelNode
	var field *parser.FieldNode
	var attribute *parser.AttributeNode

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			model = m
		},
		LeaveModel: func(*parser.ModelNode) {
			model = nil
		},
		EnterField: func(f *parser.FieldNode) {
			field = f
		},
		LeaveField: func(n *parser.FieldNode) {
			field = nil
		},
		EnterAttribute: func(attr *parser.AttributeNode) {
			if field == nil || attr.Name.Value != parser.AttributeComputed {
				return
			}

			if field.Optional {
				return
			}

			attribute = attr
		},
		LeaveAttribute: func(*parser.AttributeNode) {
			attribute = nil
		},
		EnterExpression: func(expression *parser.Expression) {
			if attribute == nil {
				return
			}

			operands, err := resolve.IdentOperands(expression)
			if err != nil {
				return
			}

			for _, operand := range operands {

				currModel := model
				currField := field
				for i, ident := range operand.Fragments {
					if i == 0 {
						continue
					}
					currField = query.Field(currModel, ident)
					if currField == nil {
						return
					}

					isToManyLookup := query.Model(asts, currField.Type.Value) != nil && currField.Repeated
					if isToManyLookup {
						// nullable fields are not a concern in 1:M lookups because
						// the aggregate functions will always coalesce to a default value
						continue
					}
					if currField.Optional {
						errs.AppendError(
							errorhandling.NewValidationErrorWithDetails(
								errorhandling.AttributeExpressionError,
								errorhandling.ErrorDetails{
									Message: fmt.Sprintf("this @computed field is required and cannot perform a lookup to the nullable field '%s'", ident),
									Hint:    "make this field or all target fields nullable",
								},
								operand,
							),
						)
						break
					}
					currModel = query.Model(asts, currField.Type.Value)
				}
				// for i := 1; i < len(fragments)-1; i++ {
				// 	currentFragment := fragments[i]
				// 	field := proto.FindField(v.schema.Models, model.Name, currentFragment)
				// 	if field.Type.Type == proto.Type_TYPE_MODEL && field.Type.Repeated {
				// 		return true, nil
				// 	}
				// 	model = v.schema.FindModel(field.Type.ModelName.Value)
				// }

			}
		},
	}
}
