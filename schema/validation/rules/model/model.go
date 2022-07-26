package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var (
	reservedModelNames = []string{"query"}
)

func ModelNamingRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		// todo - these MustCompile regex would be better at module scope, to
		// make the MustCompile panic a load-time thing rather than a runtime thing.
		reg := regexp.MustCompile("([A-Z][a-z0-9]+)+")

		if reg.FindString(model.Name.Value) != model.Name.Value {
			suggested := strcase.ToCamel(strings.ToLower(model.Name.Value))

			errors = append(
				errors,
				errorhandling.NewValidationError(
					errorhandling.ErrorUpperCamel,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Model":     model.Name.Value,
							"Suggested": suggested,
						},
					},
					model.Name,
				),
			)
		}

	}

	return errors
}

func ReservedModelNamesRule(asts []*parser.AST) []error {
	var errors []error

	for _, model := range query.Models(asts) {
		for _, name := range reservedModelNames {
			if strings.EqualFold(name, model.Name.Value) {
				errors = append(
					errors,
					errorhandling.NewValidationError(errorhandling.ErrorReservedModelName,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"Name":       model.Name.Value,
								"Suggestion": fmt.Sprintf("%ser", model.Name.Value),
							},
						},
						model.Name,
					),
				)
			}
		}
	}

	return errors
}

func UniqueModelNamesRule(asts []*parser.AST) (errors []error) {
	seenModelNames := map[string]bool{}

	for _, model := range query.Models(asts) {
		if _, ok := seenModelNames[model.Name.Value]; ok {
			errors = append(
				errors,
				errorhandling.NewValidationError(errorhandling.ErrorUniqueModelsGlobally,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Name": model.Name.Value,
						},
					},
					model.Name,
				),
			)

			continue
		}
		seenModelNames[model.Name.Value] = true
	}

	return errors
}

func ActionNamingRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			if strcase.ToLowerCamel(action.Name.Value) != action.Name.Value {
				errors = append(
					errors,
					errorhandling.NewValidationError(errorhandling.ErrorActionNameLowerCamel,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"Name":      action.Name.Value,
								"Suggested": strcase.ToLowerCamel(strings.ToLower(action.Name.Value)),
							},
						},
						action.Name,
					),
				)
			}
		}
	}

	return errors
}

var validActionTypes = []string{
	parser.ActionTypeGet,
	parser.ActionTypeCreate,
	parser.ActionTypeUpdate,
	parser.ActionTypeList,
	parser.ActionTypeDelete,
}

func ActionTypesRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			if !lo.Contains(validActionTypes, action.Type.Value) {
				errors = append(
					errors,
					errorhandling.NewValidationError(errorhandling.ErrorInvalidActionType,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"Type":       action.Type.Value,
								"ValidTypes": formatting.HumanizeList(validActionTypes, formatting.DelimiterOr),
							},
						},
						action.Type,
					),
				)
			}
		}
	}

	return errors
}

func UniqueOperationNamesRule(asts []*parser.AST) (errors []error) {
	operationNames := map[string]bool{}

	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			if _, ok := operationNames[action.Name.Value]; ok {
				errors = append(
					errors,
					errorhandling.NewValidationError(errorhandling.ErrorOperationsUniqueGlobally,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"Model": model.Name.Value,
								"Name":  action.Name.Value,
								"Line":  fmt.Sprint(action.Pos.Line),
							},
						},
						action.Name,
					),
				)
			}
			operationNames[action.Name.Value] = true
		}
	}

	return errors
}

func ValidActionInputsRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			for _, input := range action.Inputs {
				err := validateInput(asts, input, model, action)
				if err != nil {
					errors = append(errors, err)
				}
			}
			for _, input := range action.With {
				err := validateInput(asts, input, model, action)
				if err != nil {
					errors = append(errors, err)
				}
			}
		}
	}

	return errors
}

func validateInput(asts []*parser.AST, input *parser.ActionInputNode, model *parser.ModelNode, action *parser.ActionNode) error {
	resolvedType := query.ResolveInputType(asts, input, model)

	// If type cannot be resolved report error
	if resolvedType == "" {
		fieldNames := []string{}
		for _, field := range query.ModelFields(model) {
			fieldNames = append(fieldNames, field.Name.Value)
		}

		hint := errorhandling.NewCorrectionHint(fieldNames, input.Type.ToString())

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

	// if not explicitly labelled then we don't need to check for the input being used
	// as inputs using short-hand syntax are implicitly used
	if input.Label == nil {
		return nil
	}

	isUsed := false

	for _, attr := range action.Attributes {
		if !lo.Contains([]string{parser.AttributeWhere, parser.AttributeSet}, attr.Name.Value) {
			continue
		}

		if len(attr.Arguments) == 0 {
			continue
		}

		expr := attr.Arguments[0].Expression
		if expr == nil {
			continue
		}

		for _, cond := range expr.Conditions() {
			for _, operand := range []*expressions.Operand{cond.LHS, cond.RHS} {
				if operand.Ident != nil && operand.ToString() == input.Label.Value {
					// we've found a usage of the input
					isUsed = true
				}
			}
		}
	}

	if isUsed {
		return nil
	}

	// No usages of the input - report error
	return errorhandling.NewValidationError(
		errorhandling.ErrorUnusedInput,
		errorhandling.TemplateLiterals{
			Literals: map[string]string{
				"InputName": input.Label.Value,
			},
		},
		input.Label,
	)
}

// CreateOperationNoReadInputsRule validates that create actions don't accept
// any read-only inputs
func CreateOperationNoReadInputsRule(asts []*parser.AST) (errors []error) {
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
				errors = append(
					errors,
					errorhandling.NewValidationError(errorhandling.ErrorCreateOperationNoInputs,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"Input": name,
							},
						},
						i,
					),
				)
			}
		}
	}

	return errors
}

// CreateOperationRequiredFieldsRule validates that all create actions
// accept all required fields (that don't have default values) as write inputs
func CreateOperationRequiredFieldsRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {

		requiredFields := []*parser.FieldNode{}
		for _, field := range query.ModelFields(model) {
			// Optional and repeated fields are not required
			if field.Optional || field.Repeated {
				continue
			}
			if query.FieldHasAttribute(field, parser.AttributeDefault) {
				continue
			}
			requiredFields = append(requiredFields, field)
		}

		for _, action := range query.ModelActions(model) {
			if action.Type.Value != parser.ActionTypeCreate {
				continue
			}

		fields:
			for _, requiredField := range requiredFields {
				for _, input := range action.With {
					// short-hand syntax that matches the required field
					if input.Label == nil && input.Type.ToString() == requiredField.Name.Value {
						continue fields
					}
				}

				// if no explicitly an input we need to look for a @set() that references the field
				for _, attr := range action.Attributes {
					if attr.Name.Value != parser.AttributeSet {
						continue
					}

					if len(attr.Arguments) == 0 {
						continue
					}

					if attr.Arguments[0].Expression == nil {
						continue
					}

					assignment, err := expressions.ToAssignmentCondition(attr.Arguments[0].Expression)
					if err != nil {
						continue
					}

					lhs := assignment.LHS

					if len(lhs.Ident.Fragments) != 2 {
						continue
					}

					modelName, fieldName := lhs.Ident.Fragments[0].Fragment, lhs.Ident.Fragments[1].Fragment

					if modelName != strcase.ToLowerCamel(model.Name.Value) {
						continue
					}

					// If we've found an assignment expression with a left hand side like:
					//   {modelName}.{fieldName}
					// then we are satisfied that this required field is being set
					// We're not validating the RHS of the expression here as that is handled
					// by the @set attribute rules
					if fieldName == requiredField.Name.Value {
						continue fields
					}
				}

				// we didn't find an input or a @set that satisfies this required field
				// so that's an error
				errors = append(
					errors,
					errorhandling.NewValidationError(errorhandling.ErrorCreateOperationMissingInput,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"FieldName": requiredField.Name.Value,
							},
						},
						action.Name,
					),
				)
			}
		}
	}

	return errors
}

// UpdateOperationUniqueConstraintRule checks that all update operations
// are filtering on unique fields only
func UpdateOperationUniqueConstraintRule(asts []*parser.AST) []error {
	var errors []error

	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			if action.Type.Value != parser.ActionTypeUpdate {
				continue
			}
			errs := requireUniqueLookup(asts, action, model)
			errors = append(errors, errs...)
		}
	}

	return errors
}

// GetOperationUniqueConstraintRule checks that all get operations
// are filtering on unique fields only
func GetOperationUniqueConstraintRule(asts []*parser.AST) []error {
	var errors []error

	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			if action.Type.Value != parser.ActionTypeGet {
				continue
			}
			errs := requireUniqueLookup(asts, action, model)
			errors = append(errors, errs...)
		}
	}

	return errors
}

func requireUniqueLookup(asts []*parser.AST, action *parser.ActionNode, model *parser.ModelNode) (errors []error) {

	hasUniqueLookup := false

	// check for inputs that refer to non-unique fields
	for _, arg := range action.Inputs {
		field := query.ResolveInputField(asts, arg, model)

		// if input type cannot be resolved to a field e.g. it's a built-in
		// type like Text then we can't check for uniqueness
		if field == nil {
			continue
		}

		// If the input refers to a unique field then we're all good
		if query.FieldIsUnique(field) {
			hasUniqueLookup = true
			continue
		}

		// input refers to a non-unique field - this is an error
		errors = append(
			errors,
			errorhandling.NewValidationError(errorhandling.ErrorOperationInputNotUnique,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Input":         arg.Type.ToString(),
						"OperationType": action.Type.Value,
					},
				},
				arg,
			),
		)
	}

	// check for @where attributes that filter on non-unique fields
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
			if condition.Type() != expressions.LogicalCondition {
				continue
			}

			operator := condition.Operator.Symbol

			// Only "==" and "in" are direct comparison operators, anything else
			// doesn't make sense for a unique lookup e.g. age > 5
			if operator != expressions.OperatorEquals && operator != expressions.OperatorIn {
				errors = append(
					errors,
					errorhandling.NewValidationError(errorhandling.ErrorNonDirectComparisonOperatorUsed,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"Operator":      operator,
								"OperationType": action.Type.Value,
							},
						},
						condition.Operator,
					),
				)
				continue
			}

			// we always check the LHS
			operands := []*expressions.Operand{condition.LHS}

			// if it's an equal operator we can check both sides
			if operator == expressions.OperatorEquals {
				operands = append(operands, condition.RHS)
			}

			for _, op := range operands {
				if op.Ident == nil || len(op.Ident.Fragments) != 2 {
					continue
				}

				modelName, fieldName := op.Ident.Fragments[0].Fragment, op.Ident.Fragments[1].Fragment

				if modelName != strcase.ToLowerCamel(model.Name.Value) {
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
				errors = append(
					errors,
					errorhandling.NewValidationError(errorhandling.ErrorOperationWhereNotUnique,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"Ident":         op.Ident.ToString(),
								"OperationType": action.Type.Value,
							},
						},
						op.Ident,
					),
				)

			}

		}
	}

	// we did not find a unique field - this action is invalid
	if !hasUniqueLookup {
		errors = append(
			errors,
			errorhandling.NewValidationError(errorhandling.ErrorOperationMissingUniqueInput,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Name": action.Name.Value,
					},
				},
				action.Name,
			),
		)
	}

	return errors
}
