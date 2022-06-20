package expression

import (
	"fmt"

	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

type ResolvedValue struct {
	*node.Node

	Type string
}

func ValidateExpressionRule(asts []*parser.AST) []error {
	errs := make([]error, 0)

	for _, model := range query.Models(asts) {
		attrs := query.ModelAttributes(model)

		for _, attr := range attrs {
			for _, arg := range attr.Arguments {
				// get all of the nested conditions in the expression
				conditions := arg.Expression.Conditions()

				for _, condition := range conditions {
					if condition == nil {
						continue
					}
					conditionType := condition.Type()
					lhs, operator, rhs := condition.ToFragments()

					fmt.Printf("LHS: %s O: %s RHS: %s (type: %s) %s", lhs.ToString(), operator.ToString(), rhs.ToString(), conditionType, "\n")
				}
			}
		}
	}

	return errs
}

// func checkExpressionConditionSide(asts []*parser.AST, contextModel *parser.ModelNode, value *expressions.Operand) (*ResolvedValue, error) {
// 	if value.Ident != nil {
// 		fragments := value.Ident.ToArray()

// 		// Handle special case where an ident refers to the ctx object, which is not a model.
// 		if fragments[0] == "ctx" {
// 			return &ResolvedValue{
// 				Type: "ctx",
// 			}, nil
// 		}

// 		rootModel := query.FuzzyFindModel(asts, fragments[0])

// 		if rootModel == nil {
// 			suggested := str.Pluralize(strings.ToLower(contextModel.Name.Value))
// 			mutatedValue := value
// 			mutatedValue.EndPos.Column = mutatedValue.Pos.Column + len(fragments[0])
// 			mutatedValue.Tokens = []lexer.Token{}

// 			return nil, errorhandling.NewValidationError(
// 				errorhandling.ErrorUnresolvedRootCondition,
// 				errorhandling.TemplateLiterals{
// 					Literals: map[string]string{
// 						"Root":       fragments[0],
// 						"Type":       "model",
// 						"Suggestion": suggested,
// 					},
// 				},
// 				mutatedValue,
// 			)
// 		}

// 		// Try to resolve the association based on the contextModel
// 		// e.g contextModel will be "modelName" in the path fragment modelName.associationA.associationB
// 		_, err := tryAssociation(asts, contextModel, value.Ident.ToArray())

// 		if err != nil {

// 			resolutionError := err.(*query.AssociationResolutionError)
// 			// todo: fix this check levenstein distance for ctx (e.g user writes context) and return suggestion hint

// 			errModel := resolutionError.ContextModel
// 			allModelFields := query.ModelFieldNames(errModel)

// 			suggestions := errorhandling.NewCorrectionHint(allModelFields, resolutionError.ErrorFragment)

// 			mutatedValue := value

// 			// Set the start and end column values to the length of the erroring token
// 			mutatedValue.Pos.Column = mutatedValue.Pos.Column + resolutionError.StartCol
// 			mutatedValue.EndPos.Column = mutatedValue.Pos.Column + len(resolutionError.ErrorFragment)

// 			// Clear out the old tokens which are used by the GetPositionRange function to calculate the error underlining
// 			// With the old tokens (for the whole expression string) in place, the wrong portion of the string is highlighted
// 			// todo: A long term fix for this is to change the tokenization so that it tokenizes each fragment of an expression condition with regex
// 			mutatedValue.Tokens = []lexer.Token{}

// 			return nil, errorhandling.NewValidationError(
// 				errorhandling.ErrorUnresolvableExpression,
// 				errorhandling.TemplateLiterals{
// 					Literals: map[string]string{
// 						"Suggestions": suggestions.ToString(),
// 						"Fragment":    resolutionError.ErrorFragment,
// 						"Type":        resolutionError.Type,
// 						"Parent":      resolutionError.Parent,
// 					},
// 				},
// 				mutatedValue,
// 			)
// 		}
// 	}

// 	return &ResolvedValue{
// 		Type: value.Type(),
// 	}, nil
// }

// func tryAssociation(asts []*parser.AST, contextModel *parser.ModelNode, fragments []string) (*ResolvedValue, error) {
// 	n, err := query.ResolveAssociation(asts, contextModel, fragments, 1)

// 	if err == nil {
// 		return &ResolvedValue{
// 			Node: n,
// 			Type: "association",
// 		}, nil
// 	}

// 	return nil, err
// }
