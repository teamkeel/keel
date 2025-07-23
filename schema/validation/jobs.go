package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func Jobs(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	jobInputs := []string{}
	hasSchedule := false
	hasPermission := false
	var currentJob *parser.JobNode

	return Visitor{
		EnterJob: func(job *parser.JobNode) {
			currentJob = job
			hasSchedule = false
			hasPermission = false
			jobInputs = []string{}
		},
		LeaveJob: func(n *parser.JobNode) {
			if !hasPermission && !hasSchedule {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.JobDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Job '%s' must be defined with either @schedule or @permission", n.Name.Value),
					},
					n.Name,
				))
			}

			if hasSchedule && len(jobInputs) > 0 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.JobDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Job '%s' is scheduled and so cannot also have inputs", n.Name.Value),
					},
					n.Name,
				))
			}
			currentJob = nil
		},
		EnterJobInput: func(input *parser.JobInputNode) {
			if !parser.IsBuiltInFieldType(input.Type.Value) && !query.IsEnum(asts, input.Type.Value) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.JobDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Job input '%s' is defined with unsupported type %s", input.Name.Value, input.Type.Value),
					},
					input.Name,
				))
			}

			if lo.Contains(jobInputs, input.Name.Value) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.DuplicateDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Job input with name '%s' already exists", input.Name.Value),
					},
					input.Name,
				))
			}

			jobInputs = append(jobInputs, input.Name.Value)
		},
		EnterAttribute: func(n *parser.AttributeNode) {
			if currentJob == nil {
				return
			}

			if n.Name.Value == "schedule" {
				if hasSchedule {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeNotAllowedError,
						errorhandling.ErrorDetails{
							Message: "A job cannot have more than one @schedule attribute",
						},
						n.Name,
					))
				}

				hasSchedule = true
			}

			if n.Name.Value == "permission" {
				hasPermission = true
			}
		},
	}
}
