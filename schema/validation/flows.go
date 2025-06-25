package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func Flows(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	flowInputs := []string{}

	return Visitor{
		EnterFlow: func(job *parser.FlowNode) {
			flowInputs = []string{}
		},
		EnterFlowInput: func(input *parser.FlowInputNode) {
			if !parser.IsBuiltInFieldType(input.Type.Value) && !query.IsEnum(asts, input.Type.Value) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.FlowDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Flow input '%s' is defined with unsupported type %s", input.Name.Value, input.Type.Value),
					},
					input.Name,
				))
			}

			if lo.Contains(flowInputs, input.Name.Value) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.DuplicateDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Flow input with name '%s' already exists", input.Name.Value),
					},
					input.Name,
				))
			}

			flowInputs = append(flowInputs, input.Name.Value)
		},
	}
}
