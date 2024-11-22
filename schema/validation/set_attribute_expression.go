package validation

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/attributes"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var (
	fieldsNotMutable = []string{
		parser.FieldNameCreatedAt,
		parser.FieldNameUpdatedAt,
	}
)

func SetAttributeExpressionRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var model *parser.ModelNode
	var action *parser.ActionNode

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			model = m
		},
		LeaveModel: func(_ *parser.ModelNode) {
			model = nil
		},
		EnterAction: func(a *parser.ActionNode) {
			action = a
		},
		LeaveAction: func(_ *parser.ActionNode) {
			action = nil
		},
		EnterAttribute: func(attribute *parser.AttributeNode) {
			if attribute == nil || attribute.Name.Value != parser.AttributeSet {
				return
			}

			expr := attribute.Arguments[0].Expression

			if expr.LHS == nil || expr.LHS.Operator == nil || expr.LHS.Term == nil {
				errs.AppendError(makeSetExpressionError(
					errorhandling.AttributeExpressionError,
					"The @set attribute cannot be a value condition and must express an assignment",
					fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
					expr,
				))
				return
			}

			if !lo.Contains(parser.AssignmentOperators, expr.LHS.Operator.Symbol) {
				errs.AppendError(makeSetExpressionError(
					errorhandling.AttributeExpressionError,
					fmt.Sprintf("Operator '%s' not permitted on @set", expr.LHS.Operator.Symbol),
					fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
					expr.LHS.Operator,
				))
				return
			}

			operand, assignmentExpression, err := attribute.Arguments[0].Expression.ToAssignmentExpression()
			switch err {
			case parser.ErrNotAssignmentOperand:
				errs.AppendError(makeSetExpressionError(
					errorhandling.AttributeExpressionError,
					"The @set attribute can only be used to set model fields",
					fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
					expr,
				))
				return
			case parser.ErrNotAssignmentOperator:
				errs.AppendError(makeSetExpressionError(
					errorhandling.AttributeExpressionError,
					"The @set attribute expression must be an assignment expression",
					fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
					expr,
				))
				return
			case parser.ErrNotAssignmentExpression:
				errs.AppendError(makeSetExpressionError(
					errorhandling.AttributeExpressionError,
					"The @set attribute expression's requires a valid right-hand side expression",
					fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
					expr,
				))
				return
			}

			if operand.Ident == nil {
				errs.AppendError(makeSetExpressionError(
					errorhandling.AttributeExpressionError,
					"The @set attribute can only be used to set model fields",
					fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
					operand,
				))
				return
			}

			// if len(conditions) > 1 {
			// 	errs.AppendError(makeSetExpressionError(
			// 		"A @set attribute can only consist of a single assignment expression",
			// 		fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
			// 		attribute.Arguments[0].Expression,
			// 	))
			// 	return
			// }

			// expressionContext := expressions.ExpressionContext{
			// 	Model:     model,
			// 	Attribute: attribute,
			// 	Action:    action,
			// }

			// if conditions[0].Type() == parser.ValueCondition {
			// 	errs.AppendError(makeSetExpressionError(
			// 		"The @set attribute cannot be a value condition and must express an assignment",
			// 		fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
			// 		conditions[0],
			// 	))
			// 	return
			// }

			// TODO: replace with expression parser
			// // We resolve whether the actual fragments are valid idents in other validations,
			// // but we need to exit early here if they dont resolve.
			// resolver := expressions.NewConditionResolver(conditions[0], asts, &expressionContext)
			// _, _, err := resolver.Resolve()
			// if err != nil {
			// 	return
			// }

			// Drop the 'id' at the end if it exists
			fragments := []*parser.IdentFragment{}
			for _, fragment := range operand.Ident.Fragments {
				if fragment.Fragment != "id" {
					fragments = append(fragments, fragment)
				}
			}

			incompatableInputs := []*parser.ActionInputNode{}
			currentModel := model
			var currentField *parser.FieldNode

			// Iterate through each fragment in the LHS operand, and ensure:
			// - is starts at the root model
			// - it is a field which is part of a model being created or updated (including nested creates)
			for i, fragment := range fragments {
				if i == 0 && fragment.Fragment != strcase.ToLowerCamel(model.Name.Value) {
					errs.AppendError(makeSetExpressionError(
						errorhandling.AttributeExpressionError,
						fmt.Sprintf("The identifier '%s' does not exist or is not available to be set", operand.Ident.ToString()),
						fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
						operand.Ident,
					))
					return
				}

				if i > 0 {
					// get the next field in the relationship fragments
					currentField = query.ModelField(currentModel, fragment.Fragment)

					if currentField == nil {
						errs.AppendError(makeSetExpressionError(
							errorhandling.AttributeExpressionError,
							fmt.Sprintf("The field '%s' does not exist", fragment.Fragment),
							fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
							operand.Ident,
						))
						return
					}

					// currentModel will be null if this is not a model field
					currentModel = query.Model(asts, currentField.Type.Value)
				}

				// The purpose of this part is to check that the nested field being set
				// is part of the nested create inputs. You can set any field within the models
				// being created. You cannot set fields on models which already reside in the database.
				if i < len(fragments)-1 {
					withinWriteScope := false

					if i < 2 {
						withinWriteScope = true
					}

					for _, input := range action.With {
						if input.Label != nil {
							continue
						}

						if lo.Contains(incompatableInputs, input) {
							continue
						}

						if i > len(input.Type.Fragments)-1 {
							withinWriteScope = i == len(fragments)-1
							continue
						}

						if i == 0 {
							withinWriteScope = true
							continue
						}

						if operand.Ident.Fragments[i].Fragment == input.Type.Fragments[i-1].Fragment {
							if input.Type.Fragments[i].Fragment != "id" {
								withinWriteScope = true
							}
						} else {
							incompatableInputs = append(incompatableInputs, input)
						}
					}

					if !withinWriteScope {
						errs.AppendError(makeSetExpressionError(
							errorhandling.AttributeExpressionError,
							"Cannot set a field which is beyond scope of the data being created or updated",
							"Use a field which is part of a model being created or updated within this action's inputs",
							operand,
						))
						return
					}
				}

				// The purpose of this part is to check that the nested model/id being set
				// is not being provided in the nested create inputs, because that means it
				// is being created and not associated.
				if i == len(fragments)-1 && currentModel != nil {
					// We know this is setting (associating to an existing model) at this point
					setFrags := lo.Map(fragments, func(f *parser.IdentFragment, _ int) string {
						return f.Fragment
					})

					setFragsString := strings.Join(setFrags[1:], ".")

					for _, input := range action.With {
						inputFrags := lo.Map(input.Type.Fragments, func(f *parser.IdentFragment, _ int) string {
							return f.Fragment
						})

						inputFragsString := strings.Join(inputFrags, ".")

						cut, has := strings.CutPrefix(inputFragsString, setFragsString)
						if has {
							if cut == ".id" || len(cut) == 0 {
								errs.AppendError(makeSetExpressionError(
									errorhandling.AttributeExpressionError,
									fmt.Sprintf("Cannot associate to the %s model here as it is already provided as an action input.", currentModel.Name.Value),
									"",
									operand,
								))
								return
							}
						}
					}
				}

				if i == len(fragments)-1 && currentModel == nil {
					if lo.Contains(fieldsNotMutable, currentField.Name.Value) {
						errs.AppendError(makeSetExpressionError(
							errorhandling.AttributeExpressionError,
							fmt.Sprintf("Cannot set the field '%s' as it is a built-in field and can only be mutated internally", currentField.Name.Value),
							"Target another field on the model or remove the @set attribute entirely",
							fragment,
						))
						return
					}
				}
			}

			p, err := attributes.NewSetExpressionParser(asts, operand.Ident, action)
			if err != nil {
				panic(err)
			}
			issues, err := p.Validate(assignmentExpression.String())

			if len(issues) > 0 {
				for _, issue := range issues {
					errs.AppendError(makeSetExpressionError(
						errorhandling.AttributeExpressionError,
						issue,
						fmt.Sprintf("For example, assign a value to a field on this model with @set(%s.isActive = true)", strcase.ToLowerCamel(model.Name.Value)),
						assignmentExpression.AstNode(),
					))
				}
				return
			}

		},
	}
}

func makeSetExpressionError(t errorhandling.ErrorType, message string, hint string, node node.ParserNode) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		t,
		errorhandling.ErrorDetails{
			Message: message,
			Hint:    hint,
		},
		node,
	)
}
