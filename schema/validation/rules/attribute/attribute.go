package attribute

import (
	"fmt"

	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/util/collection"
)

// attributeLocationsRule checks that attributes are used in valid places
// For example it's invalid to use a @where attribute inside a model definition
func AttributeLocationsRule(asts []*parser.AST) []error {
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
		if collection.Contains(builtIns[definedOn], attr.Name.Value) {
			continue
		}

		if !collection.Contains(supportedAttributes[definedOn], attr.Name.Value) {
			hintOptions := supportedAttributes[definedOn]

			for i, hint := range hintOptions {
				hintOptions[i] = fmt.Sprintf("@%s", hint)
			}

			hint := errorhandling.NewCorrectionHint(hintOptions, attr.Name.Value)
			suggestions := formatting.HumanizeList(hint.Results, formatting.DelimiterOr)

			errors = append(
				errors,
				errorhandling.NewValidationError(errorhandling.ErrorUnsupportedAttributeType,
					errorhandling.TemplateLiterals{
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

func PermissionAttributeRule(asts []*parser.AST) (errors []error) {
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
				errors = append(errors, errorhandling.NewValidationError(errorhandling.ErrorInvalidAttributeArgument,
					errorhandling.TemplateLiterals{
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
				errors = append(errors, errorhandling.NewValidationError(errorhandling.ErrorAttributeRequiresNamedArguments,
					errorhandling.TemplateLiterals{
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
				errors = append(errors, errorhandling.NewValidationError(errorhandling.ErrorInvalidAttributeArgument,
					errorhandling.TemplateLiterals{
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
		errors = append(errors, errorhandling.NewValidationError(errorhandling.ErrorAttributeMissingRequiredArgument,
			errorhandling.TemplateLiterals{
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
		errors = append(errors, errorhandling.NewValidationError(errorhandling.ErrorAttributeMissingRequiredArgument,
			errorhandling.TemplateLiterals{
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
		errors = append(errors, errorhandling.NewValidationError(errorhandling.ErrorInvalidValue,
			errorhandling.TemplateLiterals{
				Literals: map[string]string{
					"Expected": expected,
				},
			},
			expr,
		))
		return
	}

	for _, item := range value.Array.Values {
		if item.Ident == nil {
			continue
		}
		// Each item should be a singular ident e.g. "foo" and not "foo.baz.bop"
		valid := len(item.Ident.Fragments) == 1
		if valid {
			// If it is a single ident check it's an allowed value
			name := item.Ident.Fragments[0].Fragment
			valid = collection.Contains(allowedIdents, name)
		}

		if !valid {
			expected := ""
			if len(allowedIdents) > 0 {
				expected = "any of the following identifiers - " + formatting.HumanizeList(allowedIdents, formatting.DelimiterOr)
			}
			errors = append(errors, errorhandling.NewValidationError(errorhandling.ErrorInvalidValue,
				errorhandling.TemplateLiterals{
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
