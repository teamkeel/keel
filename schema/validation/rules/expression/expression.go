package expression

import (
	"fmt"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/schema/validation/operand"
)

type RuleContext struct {
	Model     *parser.ModelNode
	Attribute *parser.AttributeNode
}

type Rules func(asts []*parser.AST, expression *expressions.Expression, context RuleContext) []error

func ValidateExpression(asts []*parser.AST, expression *expressions.Expression, customRules []Rules, context RuleContext) (errors []error) {
	baseRules := []Rules{
		OperandResolutionRule,
		// MismatchedTypesRule,
	}

	baseRules = append(baseRules, customRules...)

	for _, rule := range baseRules {
		errs := rule(asts, expression, context)
		errors = append(errors, errs...)
	}

	return errors
}

// Validates that all conditions in an expression use assignment
func OperatorAssignmentRule(asts []*parser.AST, expression *expressions.Expression, context RuleContext) (errors []error) {
	conditions := expression.Conditions()

	for _, condition := range conditions {
		// If there is no operator, then it means there is no rhs
		if condition.Operator.Symbol == "" {
			continue
		}

		if condition.Type() != expressions.AssignmentCondition {
			errors = append(errors,
				errorhandling.NewValidationError(
					errorhandling.ErrorForbiddenExpressionOperation,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Operator":   condition.Operator.Symbol,
							"Suggestion": "=",
							"Area":       fmt.Sprintf("@%s", context.Attribute.Name.Value),
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
func PreventValueConditionRule(asts []*parser.AST, expression *expressions.Expression, context RuleContext) (errors []error) {
	conditions := expression.Conditions()

	for _, cond := range conditions {
		if cond.Type() == expressions.ValueCondition {
			errors = append(errors,
				errorhandling.NewValidationError(
					errorhandling.ErrorForbiddenValueCondition,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Value": cond.ToString(),
							"Area":  fmt.Sprintf("@%s", context.Attribute.Name.Value),
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
func OperatorLogicalRule(asts []*parser.AST, expression *expressions.Expression, context RuleContext) (errors []error) {
	conditions := expression.Conditions()

	for _, condition := range conditions {
		// If there is no operator, then it means there is no rhs
		if condition.Operator.Symbol == "" {
			continue
		}

		if condition.Type() != expressions.LogicalCondition {
			errors = append(errors,
				errorhandling.NewValidationError(
					errorhandling.ErrorForbiddenExpressionOperation,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Operator":   condition.Operator.Symbol,
							"Area":       fmt.Sprintf("@%s", context.Attribute.Name.Value),
							"Suggestion": "==",
						},
					},
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
func OperandResolutionRule(asts []*parser.AST, expression *expressions.Expression, context RuleContext) (errors []error) {
	conditions := expression.Conditions()

	for _, condition := range conditions {
		_, _, errs := resolveConditionOperands(asts, condition, context)

		errors = append(errors, errs...)
	}

	return errors
}

// func TypeCheck(a *relationships.ExpressionScopeEntity, b *relationships.ExpressionScopeEntity, operator string) *errorhandling.ValidationError {
// 	allowedOperators := []string{}

// 	if a.Model != nil {
// 		resolvedA = a.Model.Name.Value
// 		allowedOperators = []string{"=="}
// 	}

// 	if a.TypeIdentifier() != b.TypeIdentifier() {
// 		// error
// 	}

// 	return nil
// }

// Validates that all lhs and rhs operands of each condition in an expression match
// func MismatchedTypesRule(asts []*parser.AST, expression *expressions.Expression, context RuleContext) (errors []*errorhandling.ValidationError) {
// 	conditions := expression.Conditions()

// 	for _, condition := range conditions {
// 		resolvedLHS, resolvedRHS, _ := resolveConditionOperands(asts, condition, context)

// 		// if there is no rhs (value only conditions with only a lhs)
// 		// then we do not care about validating this rule for this condition
// 		if resolvedRHS == nil {
// 			continue
// 		}

// 		// check the type of the last fragment in both lhs and rhs operands match
// 		if !resolvedLHS.TypesMatch(resolvedRHS) {
// 			errors = append(errors,
// 				errorhandling.NewValidationError(
// 					errorhandling.ErrorExpressionTypeMismatch,
// 					errorhandling.TemplateLiterals{
// 						Literals: map[string]string{
// 							"LHS":     resolvedLHS.LastFragment().Value,
// 							"LHSType": resolvedLHS.LastFragment().Type,
// 							"RHS":     resolvedRHS.LastFragment().Value,
// 							"RHSType": resolvedRHS.LastFragment().Type,
// 						},
// 					},
// 					condition,
// 				),
// 			)
// 		}
// 	}

// 	return errors
// }

func resolveConditionOperands(asts []*parser.AST, cond *expressions.Condition, context RuleContext) (resolvedLhs *operand.ExpressionScopeEntity, resolvedRhs *operand.ExpressionScopeEntity, errors []error) {
	lhs := cond.LHS
	rhs := cond.RHS

	resolvedLhs, lhsErr := operand.ResolveOperand(
		asts,
		lhs,
		operand.DefaultExpressionScope(asts).Merge(
			&operand.ExpressionScope{
				Entities: []*operand.ExpressionScopeEntity{
					{
						Model: context.Model,
					},
				},
			},
		),
	)

	if lhsErr != nil {
		errors = append(errors, lhsErr)
	}

	if rhs != nil {
		resolvedRhs, rhsErr := operand.ResolveOperand(
			asts,
			rhs,
			operand.DefaultExpressionScope(asts).Merge(
				&operand.ExpressionScope{
					Entities: []*operand.ExpressionScopeEntity{
						{
							Model: context.Model,
						},
					},
				},
			),
		)

		if rhsErr != nil {
			errors = append(errors, rhsErr)
		}

		return resolvedLhs, resolvedRhs, errors
	}

	return resolvedLhs, nil, errors
}
