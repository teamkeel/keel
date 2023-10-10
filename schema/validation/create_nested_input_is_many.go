package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func CreateNestedInputIsMany(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var currentModel *parser.ModelNode
	isCreateAction := false

	return Visitor{
		EnterModel: func(model *parser.ModelNode) {
			currentModel = model
		},
		LeaveModel: func(model *parser.ModelNode) {
			currentModel = nil
		},
		EnterAction: func(n *parser.ActionNode) {
			if currentModel == nil {
				return
			}

			if n.Type.Value == parser.ActionTypeCreate {
				isCreateAction = true
			}
		},
		LeaveAction: func(n *parser.ActionNode) {
			isCreateAction = false
		},
		EnterActionInput: func(input *parser.ActionInputNode) {
			if !isCreateAction {
				return
			}

			if parser.IsBuiltInFieldType(input.Type.ToString()) {
				return
			}

			var field *parser.FieldNode
			model := currentModel
			toMany := false
			for i, fragment := range input.Type.Fragments {
				if model == nil {
					return
				}
				field = query.ModelField(model, fragment.Fragment)
				if field == nil {
					return
				}

				if toMany && field.Name.Value == "id" {
					fields := query.ModelFields(model, func(f *parser.FieldNode) bool {
						return !f.BuiltIn && query.Model(asts, f.Type.Value) == nil
					})

					exampleFieldName := "someField"
					if len(fields) > 0 {
						exampleFieldName = fields[0].Name.Value
					}

					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.ActionInputError,
						errorhandling.ErrorDetails{
							Message: "Cannot provide the id of nested records which do not exist yet",
							Hint:    fmt.Sprintf("Rather create the nested models by providing their field names as inputs. For example, %s.%s", input.Type.Fragments[i-1].Fragment, exampleFieldName),
						},
						input,
					))
				}

				toMany = field.Repeated
				model = query.Model(asts, field.Type.Value)
			}
		},
	}
}
