package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// FieldOfSameModelType ensures that a model cannot have a field of the same tye
func FieldOfSameModelType(_ []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var model *parser.ModelNode
	model = nil

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			model = m
		},
		LeaveModel: func(m *parser.ModelNode) {
			model = nil
		},
		EnterField: func(f *parser.FieldNode) {
			if model != nil && model.Name.Value == f.Type.Value {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.TypeError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("The model '%s' cannot have a field of its own type.", f.Type.Value),
						},
						f.Type,
					),
				)
			}
		},
	}
}
