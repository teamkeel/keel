package actions

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var (
	reservedActionNames = []string{
		parser.AuthenticateActionName,
		parser.RequestPasswordResetActionName,
		parser.PasswordResetActionName,
	}
	validActionTypes = []string{
		parser.ActionTypeGet,
		parser.ActionTypeCreate,
		parser.ActionTypeUpdate,
		parser.ActionTypeList,
		parser.ActionTypeDelete,
	}
)

// validate only read+write can be used with returns
// validate returns has to be specified with read+write
func ActionTypesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, function := range query.ModelActions(model, func(a *parser.ActionNode) bool {
			return a.IsFunction()
		}) {
			hasReturns := len(function.Returns) > 0
			validFunctionActionTypes := validActionTypes

			if hasReturns {
				validFunctionActionTypes = []string{parser.ActionTypeRead, parser.ActionTypeWrite}

				if function.Type.Value != parser.ActionTypeRead && function.Type.Value != parser.ActionTypeWrite {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.TypeError,
							errorhandling.ErrorDetails{
								Message: "The 'returns' keyword can only be used with 'read' or 'write' actions",
							},
							function.Type,
						),
					)
					continue
				}

				// Validate multiple values aren't specified in a returns statement
				if len(function.Returns) > 1 {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.TypeError,
							errorhandling.ErrorDetails{
								Message: "Only one type can be specified in a 'returns' statement",
							},
							function.Type,
						),
					)
				}

				// Validate that 'any' (lowercased) is not valid
				if function.Returns[0].Type.Fragments[0].Fragment == "any" {
					errs.AppendError(
						errorhandling.NewValidationErrorWithDetails(
							errorhandling.TypeError,
							errorhandling.ErrorDetails{
								Message: "'any' is not a valid return type",
								Hint:    "Did you mean 'Any'?",
							},
							function.Type,
						),
					)
				}
			}

			if !hasReturns && (function.Type.Value == parser.ActionTypeRead || function.Type.Value == parser.ActionTypeWrite) {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.TypeError,
					errorhandling.ErrorDetails{
						Message: "The 'returns' keyword must be specified when using a 'read' or 'write' action type",
						Hint:    "Try to append 'returns(MyMessageType)'",
					},
					function.Node,
				))

				continue
			}

			// handles case where there is an unknown action type specified for a normal custom function
			if !lo.Contains(validFunctionActionTypes, function.Type.Value) {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.TypeError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("%s is not a valid action type. Valid types are %s", function.Type.Value, formatting.HumanizeList(validFunctionActionTypes, formatting.DelimiterOr)),
							Hint:    fmt.Sprintf("Valid types are %s", formatting.HumanizeList(validFunctionActionTypes, formatting.DelimiterOr)),
						},
						function.Type,
					),
				)
			}
		}

		for _, operation := range query.ModelActions(model, func(a *parser.ActionNode) bool { return !a.IsFunction() }) {
			if operation.Type.Value == parser.ActionTypeRead || operation.Type.Value == parser.ActionTypeWrite {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.TypeError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("The '%s' action type can only be used within a function", operation.Type.Value),
							Hint:    fmt.Sprintf("Did you mean to define '%s' as a function?", operation.Name.Value),
						},
						operation.Type,
					),
				)

				continue
			}

			hasReturns := len(operation.Returns) > 0

			if hasReturns {
				errs.AppendError(errorhandling.NewValidationErrorWithDetails(
					errorhandling.TypeError,
					errorhandling.ErrorDetails{
						Message: "The 'returns' keyword is not valid in an operation",
						Hint:    fmt.Sprintf("Did you mean to create '%s' as a function?", operation.Name.Value),
					},
					operation.Returns[0].Node,
				))

				continue
			}

			if !lo.Contains(validActionTypes, operation.Type.Value) {
				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.TypeError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("%s is not a valid action type. Valid types are %s", operation.Type.Value, formatting.HumanizeList(validActionTypes, formatting.DelimiterOr)),
							Hint:    fmt.Sprintf("Valid types are %s", formatting.HumanizeList(validActionTypes, formatting.DelimiterOr)),
						},
						operation.Type,
					),
				)
			}
		}
	}

	return
}

func UniqueActionNamesRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	actionNames := map[string]bool{}

	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			if _, ok := actionNames[action.Name.Value]; ok {
				errs.Append(errorhandling.ErrorActionUniqueGlobally,
					map[string]string{
						"Model": model.Name.Value,
						"Name":  action.Name.Value,
						"Line":  fmt.Sprint(action.Pos.Line),
					},
					action.Name,
				)
			}
			actionNames[action.Name.Value] = true
		}
	}

	return
}

// ReservedActionNameRule ensures that all actions do not use a reserved name
func ReservedActionNameRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, op := range query.ModelActions(model) {
			if lo.Contains(reservedActionNames, op.Name.Value) {
				errs.Append(errorhandling.ErrorReservedActionName,
					map[string]string{
						"Name":       op.Name.Value,
						"Suggestion": fmt.Sprintf("perform%s", casing.ToCamel(op.Name.Value)),
					},
					op.Name,
				)
			}
		}
	}

	return errs
}

// UpdateOperationUniqueConstraintRule checks that all update operations
// are filtering on unique fields only
func UpdateOperationUniqueConstraintRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {

	for _, model := range query.Models(asts) {
		// Note - this is applied only to Operations, i.e. not Function.
		for _, action := range query.ModelActions(model, func(a *parser.ActionNode) bool { return !a.IsFunction() }) {
			if action.Type.Value != parser.ActionTypeUpdate {
				continue
			}
			errs.Concat(requireUniqueLookup(asts, action, model))
		}
	}

	return
}

func ActionModelInputsRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			allInputs := append(action.Inputs, action.With...)

			for _, input := range allInputs {
				resolvedType := query.ResolveInputType(asts, input, model, action)
				if resolvedType == "" {
					continue
				}

				m := query.Model(asts, resolvedType)
				if m == nil {
					continue
				}

				errs.AppendError(
					errorhandling.NewValidationErrorWithDetails(
						errorhandling.ActionInputError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("'%s' refers to a model which cannot used as an input", input.Type.ToString()),
							Hint:    fmt.Sprintf("Inputs must target fields on models only, e.g %s.id", input.Type.ToString()),
						},
						input.Type,
					),
				)
			}
		}
	}

	return
}

// GetOperationUniqueConstraintRule checks that all get actions
// are filtering on unique fields only
func GetOperationUniqueConstraintRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		// Note - this is applied only to built in Actions, i.e. not Functions.
		for _, action := range query.ModelActions(model, func(a *parser.ActionNode) bool { return !a.IsFunction() }) {
			if action.Type.Value != parser.ActionTypeGet {
				continue
			}

			errs.Concat(requireUniqueLookup(asts, action, model))
		}
	}

	return
}

// DeleteOperationUniqueConstraintRule checks that all get actions
// are filtering on unique fields only
func DeleteOperationUniqueConstraintRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		// Note - this is applied only to Operations, i.e. not Functions.
		for _, action := range query.ModelActions(model, func(a *parser.ActionNode) bool { return !a.IsFunction() }) {
			if action.Type.Value != parser.ActionTypeDelete {
				continue
			}

			errs.Concat(requireUniqueLookup(asts, action, model))
		}
	}

	return
}

func requireUniqueLookup(asts []*parser.AST, action *parser.ActionNode, model *parser.ModelNode) (errs errorhandling.ValidationErrors) {

	hasUniqueLookup := false

	// check for inputs that refer to non-unique fields
	for _, arg := range action.Inputs {
		isUnique, err := validateInputIsUnique(asts, action, arg, model)
		if err != nil {
			errs.AppendError(err)
		}
		if isUnique {
			hasUniqueLookup = true
		}
	}

	// check for @where attributes that filter on non-unique fields
	// only when the inputs are non-unique
	if !hasUniqueLookup {
		for _, attr := range action.Attributes {

			if attr.Name.Value != parser.AttributeWhere {
				continue
			}

			if len(attr.Arguments) == 0 {
				continue
			}

			if attr.Arguments[0].Expression == nil {
				continue
			}

			conds := attr.Arguments[0].Expression.Conditions()

			for _, condition := range conds {
				// If it's not a logical condition it will be caught by the
				// @where attribute validation
				if condition.Type() != parser.LogicalCondition {
					continue
				}

				operator := condition.Operator.Symbol

				// Only "==" and "in" are direct comparison operators, anything else
				// doesn't make sense for a unique lookup e.g. age > 5
				if operator != parser.OperatorEquals && operator != parser.OperatorIn {
					errs.Append(errorhandling.ErrorNonDirectComparisonOperatorUsed,
						map[string]string{
							"Operator":   operator,
							"ActionType": action.Type.Value,
						},
						condition.Operator,
					)
					continue
				}

				// we always check the LHS
				operands := []*parser.Operand{condition.LHS}

				// if it's an equal operator we can check both sides
				if operator == parser.OperatorEquals {
					operands = append(operands, condition.RHS)
				}

				for _, op := range operands {
					if op.Ident == nil || len(op.Ident.Fragments) != 2 {
						continue
					}

					modelName, fieldName := op.Ident.Fragments[0].Fragment, op.Ident.Fragments[1].Fragment

					if modelName != casing.ToLowerCamel(model.Name.Value) {
						continue
					}

					field := query.ModelField(model, fieldName)
					if field == nil {
						continue
					}

					// we've found a @where that is filtering on a unique
					// field using a direct comparison operator
					if query.FieldIsUnique(field) {
						hasUniqueLookup = true
						continue
					}

					// @where attribute that has a condition on a non-unique field
					// this is an error
					errs.Append(errorhandling.ErrorActionWhereNotUnique,
						map[string]string{
							"Ident":      op.Ident.ToString(),
							"ActionType": action.Type.Value,
						},
						op.Ident,
					)
				}
			}
		}
	}

	// If a unique lookup was found, then drop all errors found for any
	// non-unique lookups found
	if hasUniqueLookup {
		errs = errorhandling.ValidationErrors{}
	}

	// If we did not find a unique field make sure there is an error on the
	// action. This might happen if the action is defined with no inputs or
	// @where clauses e.g. `get getMyThing()`
	if !hasUniqueLookup && len(errs.Errors) == 0 {
		errs.Append(errorhandling.ErrorActionMissingUniqueInput,
			map[string]string{
				"Name": action.Name.Value,
			},
			action.Name,
		)
	}

	return
}

// CreateOperationNoReadInputsRule validates that create actions don't accept
// any read-only inputs
func CreateOperationNoReadInputsRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			if action.Type.Value != parser.ActionTypeCreate {
				continue
			}

			if len(action.Inputs) == 0 {
				continue
			}

			for _, i := range action.Inputs {
				var name string
				if i.Label != nil {
					name = i.Label.Value
				} else {
					name = i.Type.ToString()
				}
				errs.Append(errorhandling.ErrorCreateActionNoInputs,
					map[string]string{
						"Input": name,
					},
					i,
				)
			}
		}
	}

	return
}

func validateInputIsUnique(asts []*parser.AST, action *parser.ActionNode, input *parser.ActionInputNode, model *parser.ModelNode) (isUnique bool, err *errorhandling.ValidationError) {
	// handle built-in type e.g. not referencing a field name
	// for example `get getMyThing(name: Text)`
	if parser.IsBuiltInFieldType(input.Type.ToString()) {
		return false, nil
	}

	var field *parser.FieldNode

	for _, fragment := range input.Type.Fragments {
		if model == nil {
			return false, nil
		}
		field = query.ModelField(model, fragment.Fragment)
		if field == nil {
			return false, nil
		}
		if !query.FieldIsUnique(field) {
			// input refers to a non-unique field - this is an error
			return false, errorhandling.NewValidationError(errorhandling.ErrorActionInputNotUnique,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Input":      fragment.Fragment,
						"ActionType": action.Type.Value,
					},
				},
				fragment,
			)
		}
		model = query.Model(asts, field.Type.Value)
	}

	return true, nil
}
