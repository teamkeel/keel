package actions

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// ValidActionInputLabelRule makes sure that labels like "foo" in this case <getPerson(foo: ID)>
// are only used when the type of the input has a built-in type (ID is a built-in type, as are Text, Bool etc.)
//
// Conversely, using an alias is not allowed when the input references a
// field on THIS model, because that is simultaneously providing
// us with a name for the input AND its type. E.g. <getPerson(id)>. "id" is a field on the
// model, as could be "name" etc.
func ValidActionInputLabelRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			for _, input := range action.Inputs {
				errs.AppendError(validateInputLabel(asts, input))
			}
		}
	}
	return errs
}

// validateInputLabel executes this rule on ONE particular action input.
func validateInputLabel(
	asts []*parser.AST,
	input *parser.ActionInputNode) *errorhandling.ValidationError {

	if input.Label == nil {
		return nil
	}
	if parser.IsBuiltInFieldType(input.Type.ToString()) {
		return nil
	}
	if query.IsEnum(asts, input.Type.ToString()) {
		return nil
	}
	// Otherwise invalid
	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.ActionInputError,
		errorhandling.ErrorDetails{
			Message: "You're only allowed to use the name:type form for an input if the type is a built-in type (like Text), or an enum",
			Hint:    fmt.Sprintf("Just use (%s) on its own", input.Type.ToString()),
		},
		input.Node,
	)
}
