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

// expressionHasUniqueLookup will work through the logical expression syntax to determine if
// either a unique look is possible.
func expressionHasUniqueLookup(asts []*parser.AST, expression *parser.Expression, fieldsInCompositeUnique map[*parser.ModelNode][]*parser.FieldNode) bool {
	hasUniqueLookup := false

	for _, or := range expression.Or {
		for _, and := range or.And {
			if and.Expression != nil {
				hasUniqueLookup = expressionHasUniqueLookup(asts, and.Expression, fieldsInCompositeUnique)
			}

			if and.Condition != nil {
				if and.Condition.Type() != parser.LogicalCondition {
					continue
				}

				operator := and.Condition.Operator.Symbol

				if operator != parser.OperatorEquals {
					continue
				}

				// we always check the LHS
				operands := []*parser.Operand{and.Condition.LHS}

				// if it's an equal operator we can check both sides
				if operator == parser.OperatorEquals {
					operands = append(operands, and.Condition.RHS)
				}

				for _, op := range operands {
					if op.Null == true {
						hasUniqueLookup = false
						break
					}

					if op.Ident == nil {
						continue
					}

					modelName := op.Ident.Fragments[0].Fragment
					model := query.Model(asts, casing.ToCamel(modelName))

					if model == nil {
						// For example, ctx, or an explicit input
						continue
					}

					hasUniqueLookup = true

					for i, fragment := range op.Ident.Fragments[1:] {
						field := query.ModelField(model, fragment.Fragment)
						if field == nil {
							hasUniqueLookup = false
							continue
						}

						isComposite, _ := query.FieldIsInCompositeUnique(model, field)
						if isComposite {
							fieldsInCompositeUnique[model] = append(fieldsInCompositeUnique[model], field)
						}

						if !(query.FieldIsUnique(field) || query.IsBelongsToModelField(asts, model, field)) {
							hasUniqueLookup = false
							//continue
						}

						if i < len(op.Ident.Fragments)-2 {
							model = query.Model(asts, field.Type.Value)
							if model == nil {
								hasUniqueLookup = false
								continue
							}
						}
					}
				}
			}

			// Once we find a unique lookup between ANDs,
			// then we know the expression is a unique lookup
			if hasUniqueLookup {
				break
			}

		}

		for m, fields := range fieldsInCompositeUnique {
			for _, attribute := range query.ModelAttributes(m) {
				if attribute.Name.Value != parser.AttributeUnique {
					continue
				}

				uniqueFields, _ := query.CompositeUniqueFields(m, attribute)
				diff, _ := lo.Difference(uniqueFields, fields)
				if len(diff) == 0 {
					hasUniqueLookup = true
				}
			}
		}

		// There is no point checking further conditions in this expression
		// because all ORed conditions need to be unique lookup
		if !hasUniqueLookup {
			return false
		}

	}

	return hasUniqueLookup
}

func requireUniqueLookup(asts []*parser.AST, action *parser.ActionNode, model *parser.ModelNode) (errs errorhandling.ValidationErrors) {

	hasUniqueLookup := false

	fieldsInCompositeUnique := map[*parser.ModelNode][]*parser.FieldNode{}

	// check for inputs that refer to non-unique fields
	for _, input := range action.Inputs {
		currentModel := model

		// ignore if it's a named input
		// for example `get getMyThing(name: Text)`
		if query.ResolveInputField(asts, input, currentModel) == nil {
			continue
		}

		var candidateCompositeModel *parser.ModelNode
		var candidateCompositeField *parser.FieldNode

		// Step through the fragments of this input to check for:
		//  - does is form a unique lookup?
		//  - does is form a part of a composite unique?
		// In both cases, we require that all fragments are either:
		//  - a relationship field

		for i, fragment := range input.Type.Fragments {
			hasUniqueLookup = true

			field := query.ModelField(currentModel, fragment.Fragment)
			if field == nil {
				// Input field does not exist
				hasUniqueLookup = false
				continue
			}

			isComposite, _ := query.FieldIsInCompositeUnique(currentModel, field)
			if isComposite && candidateCompositeModel == nil && candidateCompositeField == nil {
				candidateCompositeModel = currentModel
				candidateCompositeField = field

			}

			if !(query.FieldIsUnique(field) || query.IsBelongsToModelField(asts, currentModel, field)) {
				hasUniqueLookup = false
			}

			if i == len(input.Type.Fragments)-1 && (hasUniqueLookup || isComposite) {
				if candidateCompositeModel != nil && candidateCompositeField != nil {
					fieldsInCompositeUnique[candidateCompositeModel] = append(fieldsInCompositeUnique[candidateCompositeModel], candidateCompositeField)
				}
			}

			if i < len(input.Type.Fragments)-1 {
				currentModel = query.Model(asts, field.Type.Value)
				if currentModel == nil {
					hasUniqueLookup = false
					break
				}
			}
		}

		// If any single input is unique, then we know this is a unique lookup
		if hasUniqueLookup {
			break
		}
	}

	for m, fields := range fieldsInCompositeUnique {
		for _, attribute := range query.ModelAttributes(m) {
			if attribute.Name.Value != parser.AttributeUnique {
				continue
			}

			uniqueFields, _ := query.CompositeUniqueFields(m, attribute)
			diff, _ := lo.Difference(uniqueFields, fields)
			if len(diff) == 0 {
				hasUniqueLookup = true
			}
		}
	}

	// check for @where attributes that filter on non-unique fields
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

			hasUniqueLookup = expressionHasUniqueLookup(asts, attr.Arguments[0].Expression, fieldsInCompositeUnique)

			// There is no point checking further attributes because only
			//  one of the ANDed attributes need to be unique lookup
			if hasUniqueLookup {
				break
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
		errs.AppendError(errorhandling.NewValidationErrorWithDetails(
			errorhandling.ActionInputError,
			errorhandling.ErrorDetails{
				Message: fmt.Sprintf("The action '%s' is only permitted to %s a single record and therefore the inputs or @where attribute must filter by unique fields", action.Name.Value, action.Type.Value),
				Hint:    "Did you mean to add 'id' or some other unique field as an input?",
			},
			action.Name,
		))
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

// func validateInputIsUnique(asts []*parser.AST, action *parser.ActionNode, input *parser.ActionInputNode, model *parser.ModelNode) (isUnique bool, err *errorhandling.ValidationError) {
// 	// handle built-in type e.g. not referencing a field name
// 	// for example `get getMyThing(name: Text)`
// 	if parser.IsBuiltInFieldType(input.Type.ToString()) {
// 		return false, nil
// 	}

// 	var field *parser.FieldNode

// 	for _, fragment := range input.Type.Fragments {
// 		if model == nil {
// 			return false, nil
// 		}
// 		field = query.ModelField(model, fragment.Fragment)
// 		if field == nil {
// 			return false, nil
// 		}
// 		if !query.FieldIsUnique(field) {
// 			// input refers to a non-unique field - this is an error
// 			return false, errorhandling.NewValidationError(errorhandling.ErrorActionInputNotUnique,
// 				errorhandling.TemplateLiterals{
// 					Literals: map[string]string{
// 						"Input":      fragment.Fragment,
// 						"ActionType": action.Type.Value,
// 					},
// 				},
// 				fragment,
// 			)
// 		}
// 		model = query.Model(asts, field.Type.Value)
// 	}

// 	return true, nil
// }
