package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var (
	UniqueLookupActionTypes = []string{
		parser.ActionTypeGet,
		parser.ActionTypeUpdate,
		parser.ActionTypeDelete,
	}
)

// UniqueLookup checks that the filters will guarantee that one or zero record returned
// for get, update and delete actions
func UniqueLookup(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var model *parser.ModelNode
	var action *parser.ActionNode
	var hasUniqueLookup bool
	var fieldsInCompositeUnique map[*parser.ModelNode][]*parser.FieldNode

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			model = m

		},
		LeaveModel: func(_ *parser.ModelNode) {
			model = nil
		},
		EnterAction: func(a *parser.ActionNode) {
			// Only specific action types required unique lookups
			if !lo.Contains(UniqueLookupActionTypes, a.Type.Value) {
				return
			}

			// Functions do not require a unique lookup
			for _, attr := range a.Attributes {
				if attr.Name.Value == parser.AttributeFunction {
					return
				}
			}

			action = a
			hasUniqueLookup = false
			fieldsInCompositeUnique = map[*parser.ModelNode][]*parser.FieldNode{}
		},
		LeaveAction: func(a *parser.ActionNode) {
			// Action not relevant for unique lookups
			if action == nil {
				return
			}

			if !hasUniqueLookup {
				// Determine if any composite lookups are satisfied
				for m, fields := range fieldsInCompositeUnique {
					for _, attribute := range query.ModelAttributes(m) {
						if attribute.Name.Value != parser.AttributeUnique {
							continue
						}

						uniqueFields := query.CompositeUniqueFields(m, attribute)
						diff, _ := lo.Difference(uniqueFields, fields)
						if len(diff) == 0 {
							hasUniqueLookup = true
						}
					}
				}
			}

			if !hasUniqueLookup {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.ActionInputError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("The action '%s' can only %s a single record and therefore must be filtered by unique fields", action.Name.Value, action.Type.Value),
						Hint:    "Did you mean to filter by 'id' or some other unique fields in the action's inputs or @where attributes?",
					},
					action.Name,
				))
			}

			action = nil
		},
		EnterActionInput: func(input *parser.ActionInputNode) {
			// Action does not require unique lookups
			if action == nil {
				return
			}

			// A unique lookup has already been found
			if hasUniqueLookup {
				return
			}

			// We are only concerned with filters inputs (and not 'with' inputs)
			if !lo.Contains(action.Inputs, input) {
				return
			}

			// Ignore if it's a named input
			// for example `get getMyThing(name: Text)`
			if query.ResolveInputField(asts, input, model) == nil {
				return
			}

			var fieldsInComposite map[*parser.ModelNode][]*parser.FieldNode
			hasUniqueLookup, fieldsInComposite = fragmentsUnique(asts, model, input.Type.Fragments)

			for k, v := range fieldsInComposite {
				fieldsInCompositeUnique[k] = append(fieldsInCompositeUnique[k], v...)
			}

		},
		EnterAttribute: func(attr *parser.AttributeNode) {
			// Action does not require unique lookups
			if action == nil {
				return
			}

			// A unique lookup has already been found
			if hasUniqueLookup {
				return
			}

			// Is not a @where attribute
			if attr.Name.Value != parser.AttributeWhere {
				return
			}

			// Does not have an expression
			if len(attr.Arguments) == 0 || attr.Arguments[0].Expression == nil {
				return
			}

			hasUniqueLookup = expressionHasUniqueLookup(asts, attr.Arguments[0].Expression, fieldsInCompositeUnique)
		},
	}
}

// expressionHasUniqueLookup will work through the logical expression syntax to determine if a unique lookup is possible
func expressionHasUniqueLookup(asts []*parser.AST, expression *parser.Expression, fieldsInCompositeUnique map[*parser.ModelNode][]*parser.FieldNode) bool {
	hasUniqueLookup := false
	for _, or := range expression.Or {
		for _, and := range or.And {
			if and.Expression != nil {
				hasUniqueLookup = expressionHasUniqueLookup(asts, and.Expression, fieldsInCompositeUnique)
			}

			if and.Condition != nil {
				if and.Condition.Type() != parser.LogicalCondition {
					continue
				}

				operator := and.Condition.Operator.Symbol

				// Only the equal operator can guarantee unique lookups
				if operator != parser.OperatorEquals {
					continue
				}

				operands := []*parser.Operand{and.Condition.LHS}

				// If it's an equal operator we can check both sides
				if operator == parser.OperatorEquals {
					operands = append(operands, and.Condition.RHS)
				}

				for _, op := range operands {
					if op.Null {
						hasUniqueLookup = false
						break
					}

					if op.Ident == nil {
						continue
					}

					modelName := op.Ident.Fragments[0].Fragment
					model := query.Model(asts, casing.ToCamel(modelName))

					if model == nil {
						// For example, ctx, or an explicit input
						continue
					}

					// If there is only a single fragment in the expression,
					// and we know it's the model, therefore this is a unique lookup
					if len(op.Ident.Fragments) == 1 {
						return true
					}

					var fieldsInComposite map[*parser.ModelNode][]*parser.FieldNode
					hasUniqueLookup, fieldsInComposite = fragmentsUnique(asts, model, op.Ident.Fragments[1:])

					if len(expression.Or) == 1 {
						for k, v := range fieldsInComposite {
							fieldsInCompositeUnique[k] = append(fieldsInCompositeUnique[k], v...)
						}
					}
				}
			}

			// Once we find a unique lookup between ANDs,
			// then we know the expression is a unique lookup
			if hasUniqueLookup {
				break
			}

		}

		// There is no point checking further conditions in this expression
		// because all ORed conditions need to be unique lookup
		if !hasUniqueLookup {
			return false
		}

	}

	return hasUniqueLookup
}

func fragmentsUnique(asts []*parser.AST, model *parser.ModelNode, fragments []*parser.IdentFragment) (bool, map[*parser.ModelNode][]*parser.FieldNode) {
	fieldsInCompositeUnique := map[*parser.ModelNode][]*parser.FieldNode{}

	hasUniqueLookup := true
	for i, fragment := range fragments {
		field := query.ModelField(model, fragment.Fragment)
		if field == nil {
			// Input field does not exist on the model
			return false, nil
		}

		isComposite := query.FieldIsInCompositeUnique(model, field)
		if isComposite {
			fieldsInCompositeUnique[model] = append(fieldsInCompositeUnique[model], field)
		}

		if !query.FieldIsUnique(field) &&
			!query.IsBelongsToModelField(asts, model, field) &&
			!query.IsHasManyModelField(asts, field) {

			if !isComposite {
				return false, nil
			}
			hasUniqueLookup = false
		}

		if i == len(fragments)-1 {
			return hasUniqueLookup, fieldsInCompositeUnique
		}

		if i < len(fragments)-1 {
			model = query.Model(asts, field.Type.Value)
			if model == nil {
				// Model does not exist in the schema
				return false, nil
			}
		}
	}

	return false, nil
}
