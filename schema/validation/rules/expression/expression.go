package expression

import (
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

type Rules func(asts []*parser.AST, expression *expressions.Expression) []error

func ValidateExpression(asts []*parser.AST, expression *expressions.Expression, customRules []Rules) (errors []error) {
	baseRules := []Rules{
		OperandResolutionRule,
		MismatchedTypesRule,
		CtxResolutionRule,
	}

	baseRules = append(baseRules, customRules...)

	for _, rule := range baseRules {
		errs := rule(asts, expression)

		for _, err := range errs {
			if verrs, ok := err.(*errorhandling.ValidationError); ok {
				errors = append(errors, verrs)
			}
		}
	}

	return errors
}

// Validates that all conditions in an expression use equality
func OperatorAssignmentRule(asts []*parser.AST, expression *expressions.Expression) []error {
	return nil
}

// Validates that no value conditions are used
// e.g just true or false as a condition with  would not be permitted
func PreventValueConditionRule(asts []*parser.AST, expression *expressions.Expression) []error {
	return nil
}

// Validates that all conditions in an expression use equality
func OperatorLogicalRule(asts []*parser.AST, expression *expressions.Expression) []error {
	return nil
}

// Validates that only one condition has been used in an expression
func OnlyOneConditionRule(asts []*parser.AST, expression *expressions.Expression) []error {
	return nil
}

// Validates that all operands resolve correctly
// This handles operands of all types including operands such as model.associationA.associationB
// as well as simple value types such as string, number, bool etc
func OperandResolutionRule(asts []*parser.AST, expression *expressions.Expression) []error {
	return nil
}

// Validates ctx access
func CtxResolutionRule(asts []*parser.AST, expression *expressions.Expression) []error {
	return nil
}

// Validates that all lhs and rhs operands of each condition in an expression match
func MismatchedTypesRule(asts []*parser.AST, expression *expressions.Expression) []error {
	return nil
}

// Validates inverse traversal in a relationship based expression
// e.g walking backwards (model => association => model) is not permitted
func PreventInverseTraversalRule(asts []*parser.AST, expression *expressions.Expression) []error {
	return nil
}
