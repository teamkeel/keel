package expression

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/schema/validation/operand"
	"github.com/teamkeel/keel/util/collection"
	"golang.org/x/exp/slices"
)

type RuleContext struct {
	Model       *parser.ModelNode
	ReadInputs  []*parser.ActionInputNode
	WriteInputs []*parser.ActionInputNode
	Attribute   *parser.AttributeNode
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

		errors = append(errors, runSideEffectOperandRules(asts, condition, context, expressions.AssignmentOperators)...)
	}

	return errors
}

// Validates that all conditions in an expression use logical operators
func OperatorLogicalRule(asts []*parser.AST, expression *expressions.Expression, context RuleContext) (errors []error) {
	conditions := expression.Conditions()

	permittedOperators := append(expressions.LogicalOperators, expressions.LogicalOperators...)
	permittedOperators = append(permittedOperators, expressions.ArrayOperators...)
	permittedOperators = append(permittedOperators, expressions.NumericalOperators...)

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

		errors = append(errors, runSideEffectOperandRules(asts, condition, context, permittedOperators)...)
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

func InvalidOperatorForOperandsRule(asts []*parser.AST, condition *expressions.Condition, context RuleContext, permittedOperators []string) (errors []error) {
	resolvedLHS, resolvedRHS, _ := resolveConditionOperands(asts, condition, context)

	// If there is no operator, then we are not interested in validating this rule
	if condition.Operator == nil {
		return nil
	}

	allowedOperatorsLHS := resolvedLHS.AllowedOperators()
	allowedOperatorsRHS := resolvedRHS.AllowedOperators()

	if resolvedLHS.IsRepeated() && resolvedRHS.IsRepeated() {
		return append(errors, errorhandling.NewValidationError(
			errorhandling.ErrorExpressionForbiddenArrayLHS,
			errorhandling.TemplateLiterals{
				Literals: map[string]string{
					"LHS": condition.LHS.ToString(),
				},
			},
			condition.Operator,
		))
	}

	if slices.Equal(allowedOperatorsLHS, allowedOperatorsRHS) {
		if !collection.Contains(allowedOperatorsLHS, condition.Operator.Symbol) {
			collection := lo.Intersect(permittedOperators, allowedOperatorsLHS)
			corrections := errorhandling.NewCorrectionHint(collection, condition.Operator.Symbol)

			errors = append(errors, errorhandling.NewValidationError(
				errorhandling.ErrorForbiddenOperator,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Type":       resolvedLHS.Type(),
						"Operator":   condition.Operator.Symbol,
						"Suggestion": corrections.ToString(),
					},
				},
				condition.Operator,
			))
		}
	} else {
		if resolvedRHS.IsRepeated() {
			if !collection.Contains(allowedOperatorsRHS, condition.Operator.Symbol) {

				errors = append(errors, errorhandling.NewValidationError(
					errorhandling.ErrorExpressionArrayMismatchingOperator,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"RHS":      condition.RHS.ToString(),
							"RHSType":  resolvedRHS.Type(),
							"Operator": condition.Operator.Symbol,
						},
					},
					condition.Operator,
				))
			}
		}
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

	// Simple case: LHS and RHS are the same
	if resolvedLHS.Type() == resolvedRHS.Type() {
		return nil
	}

	// If RHS is an array then we check the type of it's items
	// If they match the LHS then that is ok
	if resolvedRHS.Type() == expressions.TypeArray {
		var arrayType string
		valid := true
		for i, item := range resolvedRHS.Array {
			if i == 0 {
				arrayType = item.Type()
				continue
			}

			if arrayType != item.Type() {
				valid = false
				errors = append(errors,
					errorhandling.NewValidationError(
						errorhandling.ErrorExpressionMixedTypesInArrayLiteral,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"Item": item.Literal.ToString(),
								"Type": arrayType,
							},
						},
						item.Literal,
					),
				)
			}
		}

		if !valid {
			return errors
		}

		// Now we know the RHS is an array of type T we can check if
		// the LHS is also of type T
		if arrayType == resolvedLHS.Type() {
			return nil
		}
	}

	lhsType := resolvedLHS.Type()
	if resolvedLHS.IsRepeated() {
		if resolvedLHS.Array != nil {
			lhsType = "an array of " + resolvedLHS.Array[0].Type()
		} else {
			lhsType = "an array of " + lhsType
		}
	}

	rhsType := resolvedRHS.Type()
	if resolvedRHS.IsRepeated() {
		if resolvedRHS.Array != nil {
			rhsType = "an array of " + resolvedRHS.Array[0].Type()
		} else {
			rhsType = "an array of " + rhsType
		}
	}

	// LHS and RHS types do not match, report error
	errors = append(errors,
		errorhandling.NewValidationError(
			errorhandling.ErrorExpressionTypeMismatch,
			errorhandling.TemplateLiterals{
				Literals: map[string]string{
					"Operator": condition.Operator.Symbol,
					"LHS":      condition.LHS.ToString(),
					"LHSType":  lhsType,
					"RHS":      condition.RHS.ToString(),
					"RHSType":  rhsType,
				},
			},
			condition,
		),
	)

	return errors
}

func resolveConditionOperands(asts []*parser.AST, cond *expressions.Condition, context RuleContext) (resolvedLhs *operand.ExpressionScopeEntity, resolvedRhs *operand.ExpressionScopeEntity, errors []error) {
	lhs := cond.LHS
	rhs := cond.RHS

	scope := &operand.ExpressionScope{
		Entities: []*operand.ExpressionScopeEntity{
			{
				Model: context.Model,
			},
		},
	}

	inputs := append([]*parser.ActionInputNode{}, context.ReadInputs...)
	inputs = append(inputs, context.WriteInputs...)

	for i, input := range inputs {
		// inputs using short-hand syntax that refer to relationships
		// don't get added to the scope
		if input.Label == nil && len(input.Type.Fragments) > 1 {
			continue
		}

		resolvedType := query.ResolveInputType(asts, input, context.Model)
		if resolvedType == "" {
			continue
		}
		scope.Entities = append(scope.Entities, &operand.ExpressionScopeEntity{
			Input: &operand.ExpressionInputEntity{
				Name:       input.Name(),
				Type:       resolvedType,
				AllowWrite: i >= len(context.ReadInputs),
			},
		})
	}

	scope = operand.DefaultExpressionScope(asts).Merge(scope)

	resolvedLhs, lhsErr := operand.ResolveOperand(asts, lhs, scope)

	if lhsErr != nil {
		errors = append(errors, lhsErr)
	}

	if rhs != nil {
		resolvedRhs, rhsErr := operand.ResolveOperand(asts, rhs, scope)

		if rhsErr != nil {
			errors = append(errors, rhsErr)
		}

		return resolvedLhs, resolvedRhs, errors
	}

	return resolvedLhs, nil, errors
}

func runSideEffectOperandRules(asts []*parser.AST, condition *expressions.Condition, context RuleContext, permittedOperators []string) (errors []error) {
	errors = append(errors, OperandResolutionRule(asts, condition, context)...)

	if len(errors) > 0 {
		return errors
	}

	errors = append(errors, OperandTypesMatchRule(asts, condition, context)...)

	if len(errors) > 0 {
		return errors
	}

	errors = append(errors, InvalidOperatorForOperandsRule(asts, condition, context, permittedOperators)...)

	return errors
}
