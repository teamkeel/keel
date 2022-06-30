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

	if resolvedLHS.BaseType() == expressions.TypeArray && resolvedRHS.BaseType() == expressions.TypeArray {
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
		if resolvedRHS.BaseType() == expressions.TypeArray {
			if !collection.Contains(allowedOperatorsRHS, condition.Operator.Symbol) {
				corrections := errorhandling.NewCorrectionHint(allowedOperatorsRHS, condition.Operator.Symbol)

				errors = append(errors, errorhandling.NewValidationError(
					errorhandling.ErrorExpressionArrayMismatchingOperator,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"LHS":        resolvedLHS.Value(),
							"RHS":        condition.RHS.ToString(),
							"RHSType":    resolvedRHS.Type(),
							"Operator":   condition.Operator.Symbol,
							"Suggestion": corrections.ToString(),
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

	operator := condition.Operator.Symbol

	// Validate first that the LHS and RHS types match
	if resolvedLHS.BaseType() != resolvedRHS.BaseType() {
		// BaseType returns Array for literal arrays and fields which are repeated
		if resolvedRHS.BaseType() == expressions.TypeArray {
			// check that the rhs array is an array of T where T is LHS base type
			allMatching := true

			if resolvedRHS.Field != nil && resolvedRHS.Field.Repeated {
				// non literal arrays
				if resolvedRHS.Field.Type == parser.FieldTypeText && resolvedLHS.BaseType() == expressions.TypeString {
					allMatching = true
				} else if resolvedRHS.Type() != resolvedLHS.BaseType() {
					allMatching = false
				}
			} else {
				// literal arrays
				for _, item := range resolvedRHS.Array {
					if item.BaseType() == resolvedLHS.BaseType() {
						continue
					}

					allMatching = false
				}

				arrayType := ""

				mixedTypes := false

				for _, item := range resolvedRHS.Array {
					if arrayType != "" {
						if item.Type() != arrayType {
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

							mixedTypes = true
							break
						}
					}
					arrayType = item.Type()
				}

				if mixedTypes {
					return errors
				}
			}

			if !allMatching {
				errors = append(errors,
					errorhandling.NewValidationError(
						errorhandling.ErrorExpressionArrayWrongType,
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
			}
		} else {
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
		}
	}

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
