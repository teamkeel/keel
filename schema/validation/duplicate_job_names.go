package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func DuplicateJobNames(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterJob: func(n *parser.JobNode) {
			for _, job := range query.Jobs(asts) {
				if n == job {
					continue
				}

				if job.Name.Value == n.Name.Value {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.NamingError,
							errorhandling.ErrorDetails{
								Message: fmt.Sprintf("There already exists a job with the name '%s'", n.Name.Value),
							},
							n.Name,
						),
					)
					break
				}
			}
		},
	}
}
