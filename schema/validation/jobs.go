package validation

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func Jobs(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	jobs := []string{}
	jobInputs := []string{}
	attributes := []string{}

	return Visitor{
		EnterJob: func(job *parser.JobNode) {
			if lo.Contains(jobs, job.Name.Value) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.DuplicateDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Job with name '%s' already exists", job.Name.Value),
						Hint:    "Rename the job with a unique name",
					},
					job.Name,
				))
			}

			jobs = append(jobs, job.Name.Value)

			isScheduledJob := false
			var scheduleAttributeNode *parser.AttributeNode
			isAdhocJob := false
			hasInputs := false
			for _, section := range job.Sections {
				if section.Attribute != nil {
					switch section.Attribute.Name.Value {
					case parser.AttributePermission:
						isAdhocJob = true
					case parser.AttributeSchedule:
						isScheduledJob = true
						scheduleAttributeNode = section.Attribute
					}
				}
				if section.Inputs != nil {
					hasInputs = true
				}
			}

			if !isAdhocJob && !isScheduledJob {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.JobDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Job '%s' must be defined with either @schedule or @permission", job.Name.Value),
					},
					job.Name,
				))
			}

			if isScheduledJob && hasInputs {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.JobDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Scheduled job '%s' cannot be defined with inputs", job.Name.Value),
						Hint:    "Remove the inputs section to define a scheduled job",
					},
					scheduleAttributeNode.Name,
				))
			}

			if isScheduledJob && isAdhocJob {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.JobDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Job '%s' must be defined with either @schedule or @permission", job.Name.Value),
					},
					scheduleAttributeNode.Name,
				))
			}

		},
		LeaveJob: func(n *parser.JobNode) {
			jobInputs = []string{}
			attributes = []string{}
		},
		EnterJobInput: func(input *parser.JobInputNode) {
			if lo.Contains(jobInputs, input.Name.Value) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.DuplicateDefinitionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Job input with name '%s' already exists", input.Name.Value),
						Hint:    "Rename the input with a unique name",
					},
					input.Name,
				))
			}

			jobInputs = append(jobInputs, input.Name.Value)
		},
		EnterAttribute: func(n *parser.AttributeNode) {
			attributes = append(attributes, n.Name.Value)
		},
	}

}
