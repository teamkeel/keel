package validation

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

var (
	reservedFieldNames = []string{"id", "createdAt", "updatedAt"}
	reservedModelNames = []string{"query"}
	builtInFieldTypes  = map[string]bool{
		"Text":             true,
		"Date":             true,
		"Timestamp":        true,
		"Image":            true,
		"Boolean":          true,
		"Identity":         true,
		parser.FieldTypeID: true,
	}
)

// A Validator knows how to validate a parsed Keel schema.
//
// Conceptually we are validating a single schema.
// But the Validator supports it being "delivered" as a collection
// of *parser.Schema objects - to match up with a user's schema likely
// being written across N files.

type Validator struct {
	asts []*parser.AST
}

func NewValidator(asts []*parser.AST) *Validator {
	return &Validator{
		asts: asts,
	}
}

type validationFunc func([]*parser.AST) []error

var validatorFuncs = []validationFunc{
	reservedFieldNamesRule,
	reservedModelNamesRule,
	modelNamingRule,
	fieldNamingRule,
	actionNamingRule,
	validFieldTypesRule,
	uniqueFieldNamesRule,
	uniqueOperationNamesRule,
	validActionInputsRule,
	getOperationUniqueLookupRule,
	uniqueModelNamesRule,
	attributeLocationsRule,
	permissionAttributeRule,
}

func (v *Validator) RunAllValidators() error {
	var errors []*ValidationError

	for _, vf := range validatorFuncs {
		err := vf(v.asts)

		for _, e := range err {
			if verrs, ok := e.(*ValidationError); ok {
				errors = append(errors, verrs)
			}
		}
	}

	if len(errors) > 0 {
		errors := ValidationErrors{Errors: errors}
		return errors
	}

	return nil
}

func modelNamingRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		// todo - these MustCompile regex would be better at module scope, to
		// make the MustCompile panic a load-time thing rather than a runtime thing.
		reg := regexp.MustCompile("([A-Z][a-z0-9]+)+")

		if reg.FindString(model.Name.Value) != model.Name.Value {
			suggested := strcase.ToCamel(strings.ToLower(model.Name.Value))

			errors = append(
				errors,
				validationError(
					ErrorUpperCamel,
					TemplateLiterals{
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

func fieldNamingRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {
			if field.BuiltIn {
				continue
			}
			if strcase.ToLowerCamel(field.Name.Value) != field.Name.Value {
				errors = append(
					errors,
					validationError(ErrorFieldNameLowerCamel,
						TemplateLiterals{
							Literals: map[string]string{
								"Name":      field.Name.Value,
								"Suggested": strcase.ToLowerCamel(strings.ToLower(field.Name.Value)),
							},
						},
						field.Name,
					),
				)
			}
		}
	}

	return errors
}

func actionNamingRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			if strcase.ToLowerCamel(action.Name.Value) != action.Name.Value {
				errors = append(
					errors,
					validationError(ErrorActionNameLowerCamel,
						TemplateLiterals{
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

func uniqueFieldNamesRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		fieldNames := map[string]bool{}
		for _, field := range query.ModelFields(model) {
			// Ignore built in fields as usage of these field names is handled
			// by reservedFieldNamesRule
			if field.BuiltIn {
				continue
			}
			if _, ok := fieldNames[field.Name.Value]; ok {
				errors = append(
					errors,
					validationError(ErrorFieldNamesUniqueInModel,
						TemplateLiterals{
							Literals: map[string]string{
								"Name": field.Name.Value,
								"Line": fmt.Sprint(field.Name.Pos.Line),
							},
						},
						field.Name,
					),
				)
			}

			fieldNames[field.Name.Value] = true
		}
	}

	return errors
}

func uniqueOperationNamesRule(asts []*parser.AST) (errors []error) {
	operationNames := map[string]bool{}

	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			if _, ok := operationNames[action.Name.Value]; ok {
				errors = append(
					errors,
					validationError(ErrorOperationsUniqueGlobally,
						TemplateLiterals{
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

func validActionInputsRule(asts []*parser.AST) (errors []error) {

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

				hint := NewCorrectionHint(fieldNames, input.Name.Value)

				suggestions := formatting.HumanizeList(hint.Results, formatting.DelimiterOr)

				errors = append(
					errors,
					validationError(ErrorInvalidActionInput,
						TemplateLiterals{
							Literals: map[string]string{
								"Input":     input.Name.Value,
								"Suggested": suggestions,
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

func reservedFieldNamesRule(asts []*parser.AST) []error {
	var errors []error

	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {

			if field.BuiltIn {
				continue
			}

			for _, reserved := range reservedFieldNames {
				if strings.EqualFold(reserved, field.Name.Value) {
					errors = append(
						errors,
						validationError(ErrorReservedFieldName,
							TemplateLiterals{
								Literals: map[string]string{
									"Name":       field.Name.Value,
									"Suggestion": fmt.Sprintf("%ser", field.Name.Value),
								},
							},
							field.Name,
						),
					)

				}
			}
		}
	}

	return errors
}

func reservedModelNamesRule(asts []*parser.AST) []error {
	var errors []error

	for _, model := range query.Models(asts) {
		for _, name := range reservedModelNames {
			if strings.EqualFold(name, model.Name.Value) {
				errors = append(
					errors,
					validationError(ErrorReservedModelName,
						TemplateLiterals{
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

// GET operations must take a unique field as an input or filter on a unique field
// using @where
func getOperationUniqueLookupRule(asts []*parser.AST) []error {
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

				condition, err := expressions.ToAssignmentCondition(attr.Arguments[0].Expression)
				if err != nil {
					continue
				}

				if len(condition.LHS.Ident) != 2 {
					continue
				}

				modelName, fieldName := condition.LHS.Ident[0], condition.LHS.Ident[1]

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

			// we did not find a unique field - this action is invalid
			errors = append(
				errors,
				validationError(ErrorOperationInputFieldNotUnique,
					TemplateLiterals{
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

func validFieldTypesRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {

			if _, ok := builtInFieldTypes[field.Type]; ok {
				continue
			}

			if query.IsUserDefinedType(asts, field.Type) {
				continue
			}

			validTypes := query.UserDefinedTypes(asts)
			for t := range builtInFieldTypes {
				validTypes = append(validTypes, t)
			}

			// todo feed hint suggestions into validation error somehow.
			sort.Strings(validTypes)

			hint := NewCorrectionHint(validTypes, field.Type)

			suggestions := formatting.HumanizeList(hint.Results, formatting.DelimiterOr)

			errors = append(
				errors,
				validationError(ErrorUnsupportedFieldType,
					TemplateLiterals{
						Literals: map[string]string{
							"Name":        field.Name.Value,
							"Type":        field.Type,
							"Suggestions": suggestions,
						},
					},
					field.Name,
				),
			)
		}
	}

	return errors
}

func uniqueModelNamesRule(asts []*parser.AST) (errors []error) {
	seenModelNames := map[string]bool{}

	for _, model := range query.Models(asts) {
		if _, ok := seenModelNames[model.Name.Value]; ok {
			errors = append(
				errors,
				validationError(ErrorUniqueModelsGlobally,
					TemplateLiterals{
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

// attributeLocationsRule checks that attributes are used in valid places
// For example it's invalid to use a @where attribute inside a model definition
func attributeLocationsRule(asts []*parser.AST) []error {
	var errors []error

	for _, model := range query.Models(asts) {
		for _, section := range model.Sections {
			if section.Attribute != nil {
				errors = append(errors, checkAttributes([]*parser.AttributeNode{section.Attribute}, "model", model.Name.Value)...)
			}

			if section.Operations != nil {
				for _, op := range section.Operations {
					errors = append(errors, checkAttributes(op.Attributes, "operation", op.Name.Value)...)
				}
			}

			if section.Functions != nil {
				for _, function := range section.Functions {
					errors = append(errors, checkAttributes(function.Attributes, "function", function.Name.Value)...)
				}
			}

			if section.Fields != nil {
				for _, field := range section.Fields {
					errors = append(errors, checkAttributes(field.Attributes, "field", field.Name.Value)...)
				}
			}
		}
	}

	for _, api := range query.APIs(asts) {
		for _, section := range api.Sections {
			if section.Attribute != nil {
				errors = append(errors, checkAttributes([]*parser.AttributeNode{section.Attribute}, "api", api.Name.Value)...)
			}
		}
	}

	return errors
}

func checkAttributes(attributes []*parser.AttributeNode, definedOn string, parentName string) []error {
	var supportedAttributes = map[string][]string{
		parser.KeywordModel:     {parser.AttributePermission},
		parser.KeywordApi:       {parser.AttributeGraphQL},
		parser.KeywordField:     {parser.AttributeUnique, parser.AttributeOptional},
		parser.KeywordOperation: {parser.AttributeSet, parser.AttributeWhere, parser.AttributePermission},
		parser.KeywordFunction:  {parser.AttributePermission},
	}

	var builtIns = map[string][]string{
		parser.KeywordModel:     {},
		parser.KeywordApi:       {},
		parser.KeywordOperation: {},
		parser.KeywordFunction:  {},
		parser.KeywordField:     {parser.AttributePrimaryKey},
	}

	errors := make([]error, 0)

	for _, attr := range attributes {
		if contains(builtIns[definedOn], attr.Name.Value) {
			continue
		}

		if !contains(supportedAttributes[definedOn], attr.Name.Value) {
			hintOptions := supportedAttributes[definedOn]

			for i, hint := range hintOptions {
				hintOptions[i] = fmt.Sprintf("@%s", hint)
			}

			hint := NewCorrectionHint(hintOptions, attr.Name.Value)
			suggestions := formatting.HumanizeList(hint.Results, formatting.DelimiterOr)

			errors = append(
				errors,
				validationError(ErrorUnsupportedAttributeType,
					TemplateLiterals{
						Literals: map[string]string{
							"Name":        attr.Name.Value,
							"ParentName":  parentName,
							"DefinedOn":   definedOn,
							"Suggestions": suggestions,
						},
					},
					attr.Name,
				),
			)
		}
	}

	return errors
}

func permissionAttributeRule(asts []*parser.AST) (errors []error) {
	for _, model := range query.Models(asts) {
		for _, attr := range query.ModelAttributes(model) {
			if attr.Name.Value != parser.AttributePermission {
				continue
			}

			errors = append(errors, validatePermissionAttribute(asts, attr, model, nil)...)
		}

		for _, action := range query.ModelActions(model) {
			for _, attr := range action.Attributes {
				if attr.Name.Value != parser.AttributePermission {
					continue
				}

				errors = append(errors, validatePermissionAttribute(asts, attr, model, action)...)
			}
		}
	}

	return errors
}

var validActionKeywords = []string{
	parser.ActionTypeGet,
	parser.ActionTypeCreate,
	parser.ActionTypeUpdate,
	parser.ActionTypeList,
	parser.ActionTypeDelete,
}

func validatePermissionAttribute(asts []*parser.AST, attr *parser.AttributeNode, model *parser.ModelNode, action *parser.ActionNode) (errors []error) {
	hasActions := false
	hasExpression := false
	hasRoles := false

	for _, arg := range attr.Arguments {
		switch arg.Name.Value {
		case "actions":
			// The 'actions' argument should not be provided if the permission attribute
			// is defined inside an action as that implicitly means the permission only
			// applies to that action.
			if action == nil {
				allowedIdents := append([]string{}, validActionKeywords...)
				for _, action := range query.ModelActions(model) {
					allowedIdents = append(allowedIdents, action.Name.Value)
				}
				errors = append(errors, validateIdentArray(model, arg.Expression, allowedIdents)...)
			} else {
				errors = append(errors, validationError(ErrorInvalidAttributeArgument,
					TemplateLiterals{
						Literals: map[string]string{
							"AttributeName": "permission",
							"ArgumentName":  "actions",
							"Location":      "action",
						},
					},
					arg,
				))
			}
			hasActions = true
		case "expression":
			hasExpression = true
			// TODO: validate expression
		case "roles":
			hasRoles = true
			allowedIdents := []string{}
			for _, role := range query.Roles(asts) {
				allowedIdents = append(allowedIdents, role.Name.Value)
			}
			errors = append(errors, validateIdentArray(model, arg.Expression, allowedIdents)...)
		default:
			if arg.Name.Value == "" {
				// All arguments to @permission should have a label
				errors = append(errors, validationError(ErrorAttributeRequiresNamedArguments,
					TemplateLiterals{
						Literals: map[string]string{
							"AttributeName":      "permission",
							"ValidArgumentNames": "actions, expression, or roles",
						},
					},
					arg,
				))
				continue
			} else {
				// Unknown argument
				errors = append(errors, validationError(ErrorInvalidAttributeArgument,
					TemplateLiterals{
						Literals: map[string]string{
							"AttributeName":      "permission",
							"ArgumentName":       arg.Name.Value,
							"ValidArgumentNames": "actions, expression, or roles",
						},
					},
					arg.Name,
				))
			}
		}
	}

	// Missing actions argument which is required
	if action == nil && !hasActions {
		errors = append(errors, validationError(ErrorAttributeMissingRequiredArgument,
			TemplateLiterals{
				Literals: map[string]string{
					"AttributeName": "permission",
					"ArgumentName":  "actions",
				},
			},
			attr.Name,
		))
	}

	// One of expression or roles must be provided
	if !hasExpression && !hasRoles {
		errors = append(errors, validationError(ErrorAttributeMissingRequiredArgument,
			TemplateLiterals{
				Literals: map[string]string{
					"AttributeName": "permission",
					"ArgumentName":  `"expression" or "roles"`,
				},
			},
			attr.Name,
		))
	}

	return errors
}

func validateIdentArray(model *parser.ModelNode, expr *expressions.Expression, allowedIdents []string) (errors []error) {
	value, err := expressions.ToValue(expr)
	if err != nil || value.Array == nil {
		expected := ""
		if len(allowedIdents) > 0 {
			expected = "any of the following identifiers - " + formatting.HumanizeList(allowedIdents, formatting.DelimiterOr)
		}
		// Check expression is an array
		errors = append(errors, validationError(ErrorInvalidValue,
			TemplateLiterals{
				Literals: map[string]string{
					"Expected": expected,
				},
			},
			expr,
		))
		return
	}

	for _, item := range value.Array.Values {

		// Each item should be a singular ident e.g. "foo" and not "foo.baz.bop"
		valid := len(item.Ident) == 1
		if valid {
			// If it is a single ident check it's an allowed value
			name := item.Ident[0]
			valid = contains(allowedIdents, name)
		}

		if !valid {
			expected := ""
			if len(allowedIdents) > 0 {
				expected = "any of the following identifiers - " + formatting.HumanizeList(allowedIdents, formatting.DelimiterOr)
			}
			errors = append(errors, validationError(ErrorInvalidValue,
				TemplateLiterals{
					Literals: map[string]string{
						"Expected": expected,
					},
				},
				item,
			))
		}
	}

	return errors
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}
