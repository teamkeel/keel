package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// FunctionDisallowedBehavioursRule will validate against usages of @set, @where and nested inputs
// for any actions marked with the @function attribute as we do not support these sets of
// functionality in @function's
func FunctionDisallowedBehavioursRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterAction: func(n *parser.ActionNode) {
			if !n.IsFunction() {
				return
			}

			if found, attr := checkPresenceOfAttribute(n, parser.AttributeSet); found {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeNotAllowedError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("@%s attributes are not supported when using the @function attribute", parser.AttributeSet),
							Hint:    fmt.Sprintf("Remove @%s if you would like to continue writing the function yourself.", parser.AttributeSet),
						},
						attr,
					),
				)
			}

			if found, attr := checkPresenceOfAttribute(n, parser.AttributeWhere); found {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeNotAllowedError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("@%s attributes are not supported when using the @function attribute", parser.AttributeWhere),
							Hint:    fmt.Sprintf("Remove @%s if you would like to continue writing the function yourself.", parser.AttributeWhere),
						},
						attr,
					),
				)
			}
		},
	}
}

func checkPresenceOfAttribute(n *parser.ActionNode, attributeName string) (bool, *parser.AttributeNode) {
	attr, found := lo.Find(n.Attributes, func(a *parser.AttributeNode) bool {
		return a.Name.Value == attributeName
	})

	return found, attr
}
