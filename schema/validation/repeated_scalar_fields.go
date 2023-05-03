package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// RepeatedScalarFieldRule validates that you cannot define a repeated scalar field
// This will be temporary until we can support repeated fields at the database level
func RepeatedScalarFieldRule(_ []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	isModel := false

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			isModel = true
		},
		LeaveModel: func(m *parser.ModelNode) {
			isModel = false
		},
		EnterField: func(f *parser.FieldNode) {
			// because parser.FieldNode's are reused across fields inside of models
			// and fields inside of Messages, we want to return early if the field is defined
			// inside of a message as they are perfectly acceptable to be used in arbitrary functions
			if !isModel {
				return
			}

			if f.Repeated && f.IsScalar() {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.TypeError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("Repeated fields of type '%s' are not supported", f.Type.Value),
							Hint:    fmt.Sprintf("If this was a mistake, remove [] from '%s[]'", f.Type.Value),
						},
						f.Type,
					),
				)
			}
		},
	}
}
