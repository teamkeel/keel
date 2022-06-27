package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
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
			for _, input := range action.Arguments {
				field := query.ModelField(model, input.Name.Value)
				if field != nil {
					continue
				}

				fieldNames := []string{}
				for _, field := range query.ModelFields(model) {
					fieldNames = append(fieldNames, field.Name.Value)
				}

				hint := errorhandling.NewCorrectionHint(fieldNames, input.Name.Value)

				errors = append(
					errors,
					errorhandling.NewValidationError(
						errorhandling.ErrorInvalidActionInput,
						errorhandling.TemplateLiterals{
							Literals: map[string]string{
								"Input":     input.Name.Value,
								"Suggested": hint.ToString(),
							},
						},
						input.Name,
					),
				)

			}

		}
	}

	return errors
}

// GET operations must take a unique field as an input or filter on a unique field
// using @where
func GetOperationUniqueLookupRule(asts []*parser.AST) []error {
	var errors []error

	for _, model := range query.Models(asts) {

	actions:
		for _, action := range query.ModelActions(model) {
			if action.Type != parser.ActionTypeGet {
				continue
			}

			for _, arg := range action.Arguments {
				field := query.ModelField(model, arg.Name.Value)
				if field == nil {
					continue
				}

				// action has a unique field, go to next action
				if query.FieldIsUnique(field) {
					continue actions
				}

			}

			// no input was for a unique field so we need to check if there is a @where
			// attribute with a LHS that is for a unique field
			for _, attr := range action.Attributes {
				if attr.Name.Value != parser.AttributeWhere {
					continue
				}

				if len(attr.Arguments) != 1 {
					continue
				}

				if attr.Arguments[0].Expression == nil {
					continue
				}

				conds := attr.Arguments[0].Expression.Conditions()

				for _, condition := range conds {
					if condition.RHS == nil {
						continue
					}

					if condition.LHS.Ident == nil {
						continue
					}

					for _, op := range []*expressions.Operand{condition.LHS, condition.RHS} {
						if len(op.Ident.Fragments) != 2 {
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

						// action has a @where filtering on a unique field - go to next action
						if query.FieldIsUnique(field) {
							continue actions
						}
					}
				}
			}

			// we did not find a unique field - this action is invalid
			errors = append(
				errors,
				errorhandling.NewValidationError(errorhandling.ErrorOperationInputFieldNotUnique,
					errorhandling.TemplateLiterals{
						Literals: map[string]string{
							"Name": action.Name.Value,
						},
					},
					action.Name,
				),
			)
		}

	}

	return errors
}
