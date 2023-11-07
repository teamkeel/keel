package validation

import (
	"strings"

	"github.com/teamkeel/keel/cron"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ScheduleAttributeRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterAttribute: func(attribute *parser.AttributeNode) {
			if attribute.Name.Value != parser.AttributeSchedule {
				return
			}

			if len(attribute.Arguments) != 1 {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "@schedule must have exactly one argument",
					},
					attribute.Name,
				))
				return
			}

			arg := attribute.Arguments[0]
			if arg.Label != nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "argument to @schedule cannot be labelled",
					},
					arg.Label,
				))
				return
			}

			op, err := arg.Expression.ToValue()
			if err != nil || op.String == nil {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeArgumentError,
					errorhandling.ErrorDetails{
						Message: "argument must be a string",
						Hint:    "e.g. @schedule(\"every 10 minutes\")",
					},
					arg.Expression,
				))
				return
			}

			src := strings.TrimPrefix(*op.String, `"`)
			src = strings.TrimSuffix(src, `"`)

			_, err = cron.Parse(src)
			if err != nil {
				cronError, ok := cron.ToError(err)
				if !ok || cronError.Token == nil {
					errs.AppendError(errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: err.Error(),
						},
						arg.Expression,
					))
					return
				}

				start, end := arg.Expression.GetPositionRange()
				tok := cronError.Token
				endOffset := (len(*op.String) - tok.End)

				errs.AppendError(&errorhandling.ValidationError{
					Code: string(errorhandling.AttributeArgumentError),
					ErrorDetails: &errorhandling.ErrorDetails{
						Message: cronError.Message,
					},
					Pos: errorhandling.LexerPos{
						Filename: start.Filename,
						Offset:   start.Offset + tok.Start,
						Line:     start.Line,
						Column:   start.Column + tok.Start,
					},
					EndPos: errorhandling.LexerPos{
						Filename: end.Filename,
						Offset:   end.Offset - endOffset,
						Line:     end.Line,
						Column:   end.Column - endOffset,
					},
				})
			}
		},
	}

}
