package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// UpdateActionNestedInputsRule checks that the action inputs for update aren't referencing any relationship fields
// apart from the foreign key
func UpdateActionNestedInputsRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var action *parser.ActionNode
	var currentModel *parser.ModelNode

	return Visitor{
		EnterModel: func(n *parser.ModelNode) {
			currentModel = n
		},
		LeaveModel: func(n *parser.ModelNode) {
			currentModel = nil
		},
		EnterAction: func(n *parser.ActionNode) {
			if currentModel == nil {
				return
			}
			if n.Type.Value == parser.ActionTypeUpdate {
				action = n
			}
		},
		LeaveAction: func(n *parser.ActionNode) {
			action = nil
		},
		EnterActionInput: func(input *parser.ActionInputNode) {
			if action == nil || currentModel == nil {
				return
			}

			if !lo.Contains(action.With, input) {
				return
			}

			var field *parser.FieldNode
			model := currentModel
			relationField := false
			for _, fragment := range input.Type.Fragments {
				if model == nil {
					return
				}
				field = query.ModelField(model, fragment.Fragment)
				if field == nil {
					return
				}

				if relationField && field.Name.Value != "id" {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.ActionInputError,
						errorhandling.ErrorDetails{
							Message: "Update actions cannot perform field updates on nested models.",
							Hint:    fmt.Sprintf("A %s's fields cannot be updated directly from a %s's update action.", model.Name.Value, currentModel.Name.Value),
						},
						input,
					))
				}

				if model = query.Model(asts, field.Type.Value); model != nil {
					relationField = true
				}
			}

			return
		},
	}
}
