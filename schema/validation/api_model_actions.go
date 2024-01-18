package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ApiModelActions(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
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
