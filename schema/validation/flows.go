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
	flowHasSchedule := false

	var currentFlow *parser.FlowNode

	return Visitor{
		EnterFlow: func(flow *parser.FlowNode) {
			currentFlow = flow
			flowHasSchedule = false
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
		EnterAttribute: func(n *parser.AttributeNode) {
			if currentFlow == nil {
				return
			}

			if n.Name.Value == "schedule" {
				if flowHasSchedule {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeNotAllowedError,
						errorhandling.ErrorDetails{
							Message: "A flow cannot have more than one @schedule attribute",
						},
						n.Name,
					))
				}

				flowHasSchedule = true
			}
		},
		LeaveFlow: func(n *parser.FlowNode) {
			if flowHasSchedule && len(flowInputs) > 0 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.FlowDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Flow '%s' is scheduled and so cannot also have inputs", n.Name.Value),
					},
					n.Name,
				))
			}
			currentFlow = nil
		},
	}
}
