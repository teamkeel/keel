package actions

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"golang.org/x/exp/slices"
)

// ValidActionInputTypesRule makes sure the inputs specified for all the actions
// in the schema (i.e. operations and functions), are well formed and conform to various rules.
func ValidActionInputTypesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			allInputs := append(slices.Clone(action.Inputs), action.With...)
			for _, input := range allInputs {
				errs.AppendError(validateInputType(asts, input, len(allInputs), model, action))
			}
		}
	}
	return errs
}

func ValidArbitraryFunctionReturns(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			if !action.IsArbitraryFunction() {
				continue
			}

			returns := action.Returns

			switch {
			case len(returns) < 1:
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.ActionInputError,
					errorhandling.ErrorDetails{
						Message: "read and write functions must return exactly one message-based response",
						Hint:    fmt.Sprintf("Add a return type to %s.", action.Name.Value),
					},
					action,
				))
			case len(returns) > 1:
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.ActionInputError,
					errorhandling.ErrorDetails{
						Message: "read and write functions must return exactly one message-based response",
						Hint:    fmt.Sprintf("'%s' can be the only response to '%s'. Additional returns are not permitted.", returns[0].Type.ToString(), action.Name.Value),
					},
					returns[0].Type,
				))
			case query.Message(asts, returns[0].Type.ToString()) == nil && returns[0].Type.ToString() != parser.MessageFieldTypeAny:
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.ActionInputError,
					errorhandling.ErrorDetails{
						Message: "read and write functions must return a message-based response, or Any",
						Hint:    fmt.Sprintf("'%s' was not recognised as a known message.", returns[0].Type.ToString()),
					},
					returns[0].Type,
				))
			}
		}
	}
	return errs
}

// validateInputType makes sure that one particular action input
// is well formed and conforms to various rules.
func validateInputType(
	asts []*parser.AST,
	input *parser.ActionInputNode,
	numberOfInputs int,
	model *parser.ModelNode,
	action *parser.ActionNode) *errorhandling.ValidationError {

	// It's makes things simpler and clearer if we treat the validation
	// for Message input types separately from the other types.
	inputIsMsg := query.Message(asts, input.Type.ToString()) != nil

	if inputIsMsg {
		return validateInputMessage(input, numberOfInputs, action)
	} else {
		// I.e. we expect the input type to specify a built-in type like "Text",
		// or the name of a field on this or a related model, or an enum.
		if query.ResolveInputType(asts, input, model, action) == "" {
			return unresolvedTypeError(asts, input, model)
		}
	}
	return nil
}

// validateInputMessage makes sure that an input that is already known to
// refer to a Message obeys the rules.
func validateInputMessage(
	input *parser.ActionInputNode,
	numberOfInputs int,
	action *parser.ActionNode) *errorhandling.ValidationError {

	messageName := input.Type.ToString()

	if !action.IsArbitraryFunction() {
		return messageNotAllowedForNonArbitraryFunctionErr(input, messageName)
	}

	if numberOfInputs != 1 {
		return messageMustBeOnlyInputErr(messageName, action.Name.Value, input)
	}
	return nil
}

// a ValidationError convenience constructor.
func messageNotAllowedForNonArbitraryFunctionErr(
	input *parser.ActionInputNode,
	messageName string) *errorhandling.ValidationError {

	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.ActionInputError,
		errorhandling.ErrorDetails{
			Message: fmt.Sprintf("You are only allowed to use message %s in a read or write function", messageName),
			Hint:    "Messages can only be used in read/write functions",
		},
		input.Node,
	)
}

// a ValidationError convenience constructor.
func messageMustBeOnlyInputErr(messageName string, actionName string, input *parser.ActionInputNode) *errorhandling.ValidationError {
	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.ActionInputError,
		errorhandling.ErrorDetails{
			Message: "read and write functions must receive exactly one message-based input",
			Hint:    fmt.Sprintf("'%s' can be the only input to '%s'. Additional inputs are not permitted.", messageName, actionName),
		},
		input,
	)
}

// a ValidationError convenience constructor.
func unresolvedTypeError(
	asts []*parser.AST,
	input *parser.ActionInputNode,
	model *parser.ModelNode) *errorhandling.ValidationError {

	types := []string{}
	for _, field := range query.ModelFields(model) {
		types = append(types, field.Name.Value)
	}

	types = append(types, query.MessageNames(asts)...)

	// todo:
	// if there is no label, suggest model field names
	// if there is no label and only first input and isFunction, suggest message types
	// if there is a label, then suggest built ins

	hint := errorhandling.NewCorrectionHint(types, input.Type.ToString())

	return errorhandling.NewValidationError(
		errorhandling.ErrorInvalidActionInput,
		errorhandling.TemplateLiterals{
			Literals: map[string]string{
				"Input":     input.Type.ToString(),
				"Suggested": hint.ToString(),
			},
		},
		input.Type,
	)
}
