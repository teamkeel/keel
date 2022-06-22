package expression

import (
	"fmt"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/schema/validation/operand"
	"github.com/teamkeel/keel/util/collection"
	"golang.org/x/exp/slices"
)

type RuleContext struct {
	Model     *parser.ModelNode
	Attribute *parser.AttributeNode
}

type Rule func(asts []*parser.AST, expression *expressions.Expression, context RuleContext) []error

func ValidateExpression(asts []*parser.AST, expression *expressions.Expression, rules []Rule, context RuleContext) (errors []error) {
	for _, rule := range rules {
		errs := rule(asts, expression, context)
		errors = append(errors, errs...)
	}

	return errors
}

// Validates that all operands resolve correctly
// This handles operands of all types including operands such as model.associationA.associationB
// as well as simple value types such as string, number, bool etc
func OperandResolutionRule(asts []*parser.AST, condition *expressions.Condition, context RuleContext) (errors []error) {
	_, _, errs := resolveConditionOperands(asts, condition, context)
	errors = append(errors, errs...)

	return errors
}

// Validates that all conditions in an expression use assignment
func OperatorAssignmentRule(asts []*parser.AST, expression *expressions.Expression, context RuleContext) (errors []error) {
	conditions := expression.Conditions()

	for _, condition := range conditions {
		// If there is no operator, then it means there is no rhs
		if condition.Operator == nil {
			continue
		}

		if condition.Type() != expressions.AssignmentCondition {
			correction := errorhandling.NewCorrectionHint([]string{"="}, condition.Operator.Symbol)

			errors = append(errors,
				errorhandling.NewValidationError(
					errorhandling.ErrorForbiddenExpressionOperation,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Operator":   condition.Operator.Symbol,
							"Suggestion": correction.ToString(),
							"Attribute":  fmt.Sprintf("@%s", context.Attribute.Name.Value),
						},
					},
					condition.Operator,
				),
			)

			continue
		}

		errors = append(errors, runSideEffectOperandRules(asts, condition, context)...)
	}

	return errors
}

// Validates that all conditions in an expression use logical operators
func OperatorLogicalRule(asts []*parser.AST, expression *expressions.Expression, context RuleContext) (errors []error) {
	conditions := expression.Conditions()

	for _, condition := range conditions {
		// If there is no operator, then it means there is no rhs
		if condition.Operator == nil {
			continue
		}
		correction := errorhandling.NewCorrectionHint([]string{"=="}, condition.Operator.Symbol)

		if condition.Type() != expressions.LogicalCondition {
			errors = append(errors,
				errorhandling.NewValidationError(
					errorhandling.ErrorForbiddenExpressionOperation,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Operator":   condition.Operator.Symbol,
							"Attribute":  fmt.Sprintf("@%s", context.Attribute.Name.Value),
							"Suggestion": correction.ToString(),
						},
					},
					condition.Operator,
				),
			)

			continue
		}

		errors = append(errors, runSideEffectOperandRules(asts, condition, context)...)
	}

	return errors
}

// Validates that no value conditions are used
func PreventValueConditionRule(asts []*parser.AST, expression *expressions.Expression, context RuleContext) (errors []error) {
	conditions := expression.Conditions()

	for _, condition := range conditions {
		if condition.Type() == expressions.ValueCondition {
			errors = append(errors,
				errorhandling.NewValidationError(
					errorhandling.ErrorForbiddenValueCondition,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Value":      condition.ToString(),
							"Attribute":  fmt.Sprintf("@%s", context.Attribute.Name.Value),
							"Suggestion": fmt.Sprintf("%s = xxx", condition.ToString()),
						},
					},
					condition,
				),
			)
		}
	}

	return errors
}

func InvalidOperatorForOperandsRule(asts []*parser.AST, condition *expressions.Condition, context RuleContext) (errors []error) {
	resolvedLHS, resolvedRHS, _ := resolveConditionOperands(asts, condition, context)

	// If there is no operator, then we are not interested in validating this rule
	if condition.Operator == nil {
		return nil
	}

	allowedOperatorsLHS := resolvedLHS.AllowedOperators()
	allowedOperatorsRHS := resolvedRHS.AllowedOperators()

	if slices.Equal(allowedOperatorsLHS, allowedOperatorsRHS) {
		if !collection.Contains(allowedOperatorsLHS, condition.Operator.Symbol) {
			corrections := errorhandling.NewCorrectionHint(allowedOperatorsLHS, condition.Operator.Symbol)

			errors = append(errors, errorhandling.NewValidationError(
				errorhandling.ErrorForbiddenOperator,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"LHS":        resolvedLHS.Value(),
						"RHS":        resolvedRHS.Value(),
						"Operator":   condition.Operator.Symbol,
						"Suggestion": corrections.ToString(),
					},
				},
				condition.Operator,
			))
		}
	} else {
		// todo: lhs: single vs rhs: array
	}

	return errors
}

// Validates that all lhs and rhs operands of each condition in an expression match
func OperandTypesMatchRule(asts []*parser.AST, condition *expressions.Condition, context RuleContext) (errors []error) {

	resolvedLHS, resolvedRHS, _ := resolveConditionOperands(asts, condition, context)

	// if there is no lhs or rhs
	// then we do not care about validating this rule for this condition
	if resolvedLHS == nil || resolvedRHS == nil {
		return nil
	}

	operator := condition.Operator.Symbol

	// Validate first that the LHS and RHS types match
	if resolvedLHS.BaseType() != resolvedRHS.BaseType() {
		errors = append(errors,
			errorhandling.NewValidationError(
				errorhandling.ErrorExpressionTypeMismatch,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Operator": operator,
						"LHS":      condition.LHS.ToString(),
						"LHSType":  resolvedLHS.Type(),
						"RHS":      condition.RHS.ToString(),
						"RHSType":  resolvedRHS.Type(),
					},
				},
				condition,
			),
		)
	} else {
		// todo: handle array comparisons
		// e.g single.value in array
	}

	return errors
}

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

func runSideEffectOperandRules(asts []*parser.AST, condition *expressions.Condition, context RuleContext) (errors []error) {
	errors = append(errors, OperandResolutionRule(asts, condition, context)...)

	if len(errors) > 0 {
		return errors
	}

	errors = append(errors, OperandTypesMatchRule(asts, condition, context)...)

	if len(errors) > 0 {
		return errors
	}

	errors = append(errors, InvalidOperatorForOperandsRule(asts, condition, context)...)

	return errors
}
