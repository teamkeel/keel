package validation

import (
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func TasksValidation(_ []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterModel: func(n *parser.ModelNode) {

		},
		EnterField: func(n *parser.FieldNode) {

		},
		EnterAction: func(n *parser.ActionNode) {

		},
		EnterActionInput: func(n *parser.ActionInputNode) {
			if n.Label == nil {
				return
			}
		},
		EnterEnum: func(n *parser.EnumNode) {

		},
		EnterMessage: func(n *parser.MessageNode) {

		},
		EnterRole: func(n *parser.RoleNode) {

		},
		EnterAPI: func(n *parser.APINode) {

		},
		EnterJob: func(n *parser.JobNode) {

		},
		EnterJobInput: func(n *parser.JobInputNode) {

		},
	}
}
