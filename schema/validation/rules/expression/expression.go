package expression

import (
	"strings"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/relationships"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

type Rules func(asts []*parser.AST, expression *expressions.Expression, context interface{}) []error

func ValidateExpression(asts []*parser.AST, expression *expressions.Expression, customRules []Rules, context interface{}) (errors []error) {
	baseRules := []Rules{
		OperandResolutionRule,
		MismatchedTypesRule,
		CtxResolutionRule,
	}

	baseRules = append(baseRules, customRules...)

	for _, rule := range baseRules {
		errs := rule(asts, expression, context)

		for _, err := range errs {
			if verrs, ok := err.(*errorhandling.ValidationError); ok {
				errors = append(errors, verrs)
			}
		}
	}

	return errors
}

// Validates that all conditions in an expression use assignment
func OperatorAssignmentRule(asts []*parser.AST, expression *expressions.Expression, context interface{}) (errors []error) {
	conditions := expression.Conditions()

	for _, condition := range conditions {
		if condition.Type() != expressions.AssignmentCondition {
			errors = append(errors,
				errorhandling.NewValidationError(
					errorhandling.ErrorForbiddenExpressionOperation,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Operator":   condition.Operator.Symbol,
							"Suggestion": "==",
							"Area":       "//todo: need to provide context",
						},
					},
					condition.Operator,
				),
			)
		}
	}

	return errors
}

// Validates that no value conditions are used
// e.g just true or false as a condition with  would not be permitted
func PreventValueConditionRule(asts []*parser.AST, expression *expressions.Expression, context interface{}) (errors []error) {
	conditions := expression.Conditions()

	for _, cond := range conditions {
		if cond.Type() == expressions.ValueCondition {
			errors = append(errors,
				errorhandling.NewValidationError(
					errorhandling.ErrorForbiddenValueCondition,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Value": cond.ToString(),
							"Area":  "// todo",
						},
					},
					cond,
				),
			)
		}
	}

	return errors
}

// Validates that all conditions in an expression use logical operators
func OperatorLogicalRule(asts []*parser.AST, expression *expressions.Expression, context interface{}) (errors []error) {
	conditions := expression.Conditions()

	for _, condition := range conditions {
		if condition.Type() != expressions.LogicalCondition {
			errors = append(errors,
				errorhandling.NewValidationError(
					errorhandling.ErrorForbiddenExpressionOperation,
					errorhandling.TemplateLiterals{},
					condition.Operator,
				),
			)
		}
	}

	return errors
}

// Validates that all operands resolve correctly
// This handles operands of all types including operands such as model.associationA.associationB
// as well as simple value types such as string, number, bool etc
func OperandResolutionRule(asts []*parser.AST, expression *expressions.Expression, context interface{}) (errors []error) {
	conditions := expression.Conditions()

	for _, condition := range conditions {
		resolvedLHS, resolvedRHS, _ := resolveConditionOperands(asts, condition)

		errors = append(errors, buildOperandResolutionErrors(asts, resolvedLHS, context)...)
		errors = append(errors, buildOperandResolutionErrors(asts, resolvedRHS, context)...)
	}

	return errors
}

func buildOperandResolutionErrors(asts []*parser.AST, resolution *expressions.OperandResolution, context interface{}) (errors []error) {
	contextModel := ""
	if model, ok := context.(*parser.ModelNode); ok {
		contextModel = strings.ToLower(model.Name.Value)
	}

	if len(resolution.UnresolvedFragments()) > 0 {
		unresolved := resolution.UnresolvedFragments()[0]

		if unresolved.Parent == nil {
			literals := map[string]string{
				"Type":  "relationship",
				"Root":  unresolved.Value,
				"Model": contextModel,
			}

			errors = append(errors, errorhandling.NewValidationError(
				errorhandling.ErrorUnresolvedRootModel,
				errorhandling.TemplateLiterals{
					Literals: literals,
				},
				unresolved,
			))
		} else {
			if unresolved.Parent.Resolvable {
				parentModel := query.Model(asts, unresolved.Parent.Model)

				fieldsOnParent := query.ModelFieldNames(parentModel)
				correctionHint := errorhandling.NewCorrectionHint(fieldsOnParent, unresolved.Value)

				literals := map[string]string{
					"Type":       "relationship",
					"Fragment":   unresolved.Value,
					"Parent":     unresolved.Parent.Model,
					"Suggestion": correctionHint.ToString(),
				}

				errors = append(errors,
					errorhandling.NewValidationError(
						errorhandling.ErrorUnresolvableExpression,
						errorhandling.TemplateLiterals{
							Literals: literals,
						},
						unresolved,
					))
			} else {
				literals := map[string]string{
					"Type":       "relationship",
					"Fragment":   unresolved.Value,
					"Parent":     unresolved.Parent.Value,
					"Suggestion": contextModel,
				}

				errors = append(errors,
					errorhandling.NewValidationError(
						errorhandling.ErrorUnresolvableExpression,
						errorhandling.TemplateLiterals{
							Literals: literals,
						},
						unresolved,
					),
				)
			}
		}
	}

	return errors
}

// Validates ctx access
func CtxResolutionRule(asts []*parser.AST, expression *expressions.Expression, context interface{}) []error {
	return nil
}

// Validates that all lhs and rhs operands of each condition in an expression match
func MismatchedTypesRule(asts []*parser.AST, expression *expressions.Expression, context interface{}) (errors []error) {
	conditions := expression.Conditions()

	for _, condition := range conditions {
		resolvedLHS, resolvedRHS, _ := resolveConditionOperands(asts, condition)

		// ident resolution logic
		if !resolvedLHS.TypesMatch(resolvedRHS) {
			errors = append(errors,
				errorhandling.NewValidationError(
					errorhandling.ErrorExpressionTypeMismatch,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"LHS":     resolvedLHS.LastFragment().Value,
							"LHSType": resolvedLHS.LastFragment().Type,
							"RHS":     resolvedRHS.LastFragment().Value,
							"RHSType": resolvedRHS.LastFragment().Type,
						},
					},
					condition,
				),
			)
		}
	}

	return errors
}

// Validates inverse traversal in a relationship based expression
// e.g walking backwards (model => association => model) is not permitted
func PreventInverseTraversalRule(asts []*parser.AST, expression *expressions.Expression, context interface{}) []error {
	return nil
}

func resolveConditionOperands(asts []*parser.AST, cond *expressions.Condition) (*expressions.OperandResolution, *expressions.OperandResolution, []error) {
	lhs := cond.LHS
	rhs := cond.RHS

	resolvedLHS, lhsErrors := resolveOperand(asts, lhs)
	resolvedRHS, rhsErrors := resolveOperand(asts, rhs)

	return &resolvedLHS, &resolvedRHS, append(lhsErrors, rhsErrors...)
}

func resolveOperand(asts []*parser.AST, o *expressions.Operand) (expressions.OperandResolution, []error) {
	if ok, v := o.IsValueType(); ok {
		return expressions.OperandResolution{
			Parts: []expressions.OperandPart{
				{
					Value:      v,
					Type:       o.Type(),
					Resolvable: true,
					Node:       o.Node,
				},
			},
		}, nil
	} else if ok, _ := o.IsCtx(); ok {
		// resolve ctx
		panic("context resolution not yet implemented")
	} else {
		relationshipResolution, errs := relationships.TryResolveIdent(asts, o)

		return *relationshipResolution, errs
	}

}
