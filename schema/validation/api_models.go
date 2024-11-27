package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ApiModelActionsRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currModel *parser.ModelNode

	return Visitor{
		EnterAPIModel: func(m *parser.APIModelNode) {
			currModel = query.Model(asts, m.Name.Value)
		},
		LeaveAPIModel: func(n *parser.APIModelNode) {
			currModel = nil
		},
		EnterAPIModelAction: func(action *parser.APIModelActionNode) {
			has := false
			for _, a := range query.ModelActions(currModel) {
				if a.Name.Value == action.Name.Value {
					has = true
				}
			}

			if !has {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("%s does not exist as an action on the %s model", action.Name.Value, currModel.Name.Value),
					},
					action,
				))
			}
		},
	}
}

func ApiDuplicateModelNamesRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currAPI *parser.APINode
	return Visitor{
		EnterAPI: func(n *parser.APINode) {
			currAPI = n
		},
		LeaveAPI: func(n *parser.APINode) {
			currAPI = nil
		},
		EnterAPIModel: func(n *parser.APIModelNode) {
			for _, model := range query.APIModelNodes(currAPI) {
				if n == model {
					continue
				}
				if n.Name.Value == model.Name.Value {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.DuplicateDefinitionError,
							errorhandling.ErrorDetails{
								Message: fmt.Sprintf("The model '%s' has already been included in the '%s' API", n.Name.Value, currAPI.Name.Value),
								Hint:    "Remove one of the duplicates",
							},
							n.Name,
						),
					)
					break
				}
			}
		},
	}
}
