package attribute

import (
	"fmt"
	"strconv"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/formatting"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/schema/validation/rules/expression"
)

// attributeLocationsRule checks that attributes are used in valid places
// For example it's invalid to use a @where attribute inside a model definition
func AttributeLocationsRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, section := range model.Sections {
			if section.Attribute != nil {
				errs.Concat(checkAttributes([]*parser.AttributeNode{section.Attribute}, "model", model.Name.Value))
			}

			if section.Operations != nil {
				for _, op := range section.Operations {
					errs.Concat(checkAttributes(op.Attributes, "operation", op.Name.Value))
				}
			}

			if section.Functions != nil {
				for _, function := range section.Functions {
					errs.Concat(checkAttributes(function.Attributes, "function", function.Name.Value))
				}
			}

			if section.Fields != nil {
				for _, field := range section.Fields {
					errs.Concat(checkAttributes(field.Attributes, "field", field.Name.Value))
				}
			}
		}
	}

	for _, api := range query.APIs(asts) {
		for _, section := range api.Sections {
			if section.Attribute != nil {
				errs.Concat(checkAttributes([]*parser.AttributeNode{section.Attribute}, "api", api.Name.Value))
			}
		}
	}

	return
}

var attributeLocations = map[string][]string{
	parser.KeywordModel: {
		parser.AttributePermission,
		parser.AttributeUnique,
	},
	parser.KeywordApi: {
		parser.AttributeGraphQL,
	},
	parser.KeywordField: {
		parser.AttributeUnique,
		parser.AttributeDefault,
		parser.AttributePrimaryKey,
	},
	parser.KeywordOperation: {
		parser.AttributeSet,
		parser.AttributeWhere,
		parser.AttributePermission,
		parser.AttributeValidate,
	},
	parser.KeywordFunction: {
		parser.AttributePermission,
	},
}

func checkAttributes(attributes []*parser.AttributeNode, definedOn string, parentName string) (errs errorhandling.ValidationErrors) {
	for _, attr := range attributes {
		allowedAttributes := attributeLocations[definedOn]

		if lo.Contains(allowedAttributes, attr.Name.Value) {
			continue
		}

		hintOptions := []string{}

		for _, allowed := range allowedAttributes {
			hintOptions = append(hintOptions, fmt.Sprintf("@%s", allowed))
		}

		hint := errorhandling.NewCorrectionHint(hintOptions, attr.Name.Value)
		suggestions := formatting.HumanizeList(hint.Results, formatting.DelimiterOr)

		errs.Append(errorhandling.ErrorUnsupportedAttributeType,
			map[string]string{
				"Name":        attr.Name.Value,
				"ParentName":  parentName,
				"DefinedOn":   definedOn,
				"Suggestions": suggestions,
			},
			attr.Name,
		)
	}

	return
}

func PermissionAttributeRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, attr := range query.ModelAttributes(model) {
			if attr.Name.Value != parser.AttributePermission {
				continue
			}

			errs.Concat(validatePermissionAttribute(asts, attr, model, nil))
		}

		for _, action := range query.ModelActions(model) {
			for _, attr := range action.Attributes {
				if attr.Name.Value != parser.AttributePermission {
					continue
				}

				errs.Concat(validatePermissionAttribute(asts, attr, model, action))
			}
		}
	}

	return
}

func ValidateAttributeRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			for _, attr := range action.Attributes {
				if attr.Name.Value != parser.AttributeValidate {
					continue
				}

				errs.Concat(
					validateActionAttributeWithExpression(asts, model, action, attr),
				)
			}
		}
	}

	return
}

func SetWhereAttributeRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			for _, attr := range action.Attributes {
				if attr.Name.Value != parser.AttributeSet && attr.Name.Value != parser.AttributeWhere {
					continue
				}

				errs.Concat(
					validateActionAttributeWithExpression(asts, model, action, attr),
				)
			}
		}
	}

	return
}

// validateActionAttributeWithExpression validates attributes that have the
// signature @attributeName(expression) and exist inside an action. This applies
// to @set, @where, and @validate attributes
func validateActionAttributeWithExpression(
	asts []*parser.AST,
	model *parser.ModelNode,
	action *parser.ActionNode,
	attr *parser.AttributeNode,
) (errs errorhandling.ValidationErrors) {
	argLength := len(attr.Arguments)

	if argLength == 0 || argLength >= 2 {
		errs.Append(
			errorhandling.ErrorIncorrectArguments,
			map[string]string{
				"AttributeName":     attr.Name.Value,
				"ActualArgsCount":   strconv.FormatInt(int64(argLength), 10),
				"ExpectedArgsCount": "1",
				"Signature":         "(expression)",
			},
			attr,
		)
		return
	}

	expr := attr.Arguments[0].Expression

	rules := []expression.Rule{expression.PreventValueConditionRule}

	if attr.Name.Value == parser.AttributeSet {
		rules = append(rules, expression.OperatorAssignmentRule)
	} else {
		rules = append(rules, expression.OperatorLogicalRule)
	}

	err := expression.ValidateExpression(
		asts,
		expr,
		rules,
		expression.RuleContext{
			Model:     model,
			Attribute: attr,
			Action:    action,
		},
	)
	for _, e := range err {
		// TODO: remove case when expression.ValidateExpression returns correct type
		errs.AppendError(e.(*errorhandling.ValidationError))
	}

	return
}

var validActionKeywords = []string{
	parser.ActionTypeGet,
	parser.ActionTypeCreate,
	parser.ActionTypeUpdate,
	parser.ActionTypeList,
	parser.ActionTypeDelete,
}

func validatePermissionAttribute(asts []*parser.AST, attr *parser.AttributeNode, model *parser.ModelNode, action *parser.ActionNode) (errs errorhandling.ValidationErrors) {
	hasActions := false
	hasExpression := false
	hasRoles := false

	for _, arg := range attr.Arguments {
		if arg.Label == nil || arg.Label.Value == "" {
			// All arguments to @permission should have a label
			errs.Append(errorhandling.ErrorAttributeRequiresNamedArguments,
				map[string]string{
					"AttributeName":      "permission",
					"ValidArgumentNames": "actions, expression, or roles",
				},
				arg,
			)
			continue
		}

		switch arg.Label.Value {
		case "actions":
			// The 'actions' argument should not be provided if the permission attribute
			// is defined inside an action as that implicitly means the permission only
			// applies to that action.
			if action == nil {
				allowedIdents := append([]string{}, validActionKeywords...)
				for _, action := range query.ModelActions(model) {
					allowedIdents = append(allowedIdents, action.Name.Value)
				}
				errs.Concat(validateIdentArray(arg.Expression, allowedIdents))

			} else {
				errs.Append(errorhandling.ErrorInvalidAttributeArgument,
					map[string]string{
						"AttributeName": "permission",
						"ArgumentName":  "actions",
						"Location":      "action",
					},
					arg,
				)
			}
			hasActions = true
		case "expression":
			hasExpression = true

			expressionErrors := expression.ValidateExpression(
				asts,
				arg.Expression,
				[]expression.Rule{
					expression.OperatorLogicalRule,
				},
				expression.RuleContext{
					Model:     model,
					Attribute: attr,
					Action:    action,
				},
			)
			for _, err := range expressionErrors {
				// TODO: remove cast when expression.ValidateExpression returns correct type
				errs.AppendError(err.(*errorhandling.ValidationError))
			}
		case "roles":
			hasRoles = true
			allowedIdents := []string{}
			for _, role := range query.Roles(asts) {
				allowedIdents = append(allowedIdents, role.Name.Value)
			}
			errs.Concat(validateIdentArray(arg.Expression, allowedIdents))
		default:
			// Unknown argument
			errs.Append(errorhandling.ErrorInvalidAttributeArgument,
				map[string]string{
					"AttributeName":      "permission",
					"ArgumentName":       arg.Label.Value,
					"ValidArgumentNames": "actions, expression, or roles",
				},
				arg.Label,
			)
		}
	}

	// Missing actions argument which is required
	if action == nil && !hasActions {
		errs.Append(errorhandling.ErrorAttributeMissingRequiredArgument,
			map[string]string{
				"AttributeName": "permission",
				"ArgumentName":  "actions",
			},
			attr.Name,
		)
	}

	// One of expression or roles must be provided
	if !hasExpression && !hasRoles {
		errs.Append(errorhandling.ErrorAttributeMissingRequiredArgument,
			map[string]string{
				"AttributeName": "permission",
				"ArgumentName":  `"expression" or "roles"`,
			},
			attr.Name,
		)
	}

	return
}

func validateIdentArray(expr *expressions.Expression, allowedIdents []string) (errs errorhandling.ValidationErrors) {
	value, err := expressions.ToValue(expr)
	if err != nil || value.Array == nil {
		expected := ""
		if len(allowedIdents) > 0 {
			expected = "an array containing any of the following identifiers - " + formatting.HumanizeList(allowedIdents, formatting.DelimiterOr)
		}
		// Check expression is an array
		errs.Append(errorhandling.ErrorInvalidValue,
			map[string]string{
				"Expected": expected,
			},
			expr,
		)
		return
	}

	for _, item := range value.Array.Values {
		// Each item should be a singular ident e.g. "foo" and not "foo.baz.bop"
		// String literal idents e.g ["thisisinvalid"] are assumed not to be invalid
		valid := false

		if item.Ident != nil {
			valid = len(item.Ident.Fragments) == 1
		}

		if valid {
			// If it is a single ident check it's an allowed value
			name := item.Ident.Fragments[0].Fragment
			valid = lo.Contains(allowedIdents, name)
		}

		if !valid {
			expected := ""
			if len(allowedIdents) > 0 {
				expected = "any of the following identifiers - " + formatting.HumanizeList(allowedIdents, formatting.DelimiterOr)
			}
			errs.Append(errorhandling.ErrorInvalidValue,

				map[string]string{
					"Expected": expected,
				},

				item,
			)
		}
	}

	return
}

func UniqueAttributeArgsRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {

	for _, model := range query.Models(asts) {

		// field level e.g. @unique
		for _, field := range query.ModelFields(model) {
			for _, attr := range field.Attributes {
				if attr.Name.Value != parser.AttributeUnique {
					continue
				}

				if len(attr.Arguments) > 0 {
					errs.Append(errorhandling.ErrorIncorrectArguments,
						map[string]string{
							"AttributeName":     attr.Name.Value,
							"ActualArgsCount":   strconv.FormatInt(int64(len(attr.Arguments)), 10),
							"ExpectedArgsCount": "0",
							"Signature":         "()",
						},
						attr,
					)
				}
			}
		}

		// model level e.g. @unique([fieldOne, fieldTwo])
		for _, attr := range query.ModelAttributes(model) {
			if attr.Name.Value != parser.AttributeUnique {
				continue
			}

			if len(attr.Arguments) != 1 {
				errs.Append(errorhandling.ErrorIncorrectArguments,
					map[string]string{
						"AttributeName":     attr.Name.Value,
						"ActualArgsCount":   strconv.FormatInt(int64(len(attr.Arguments)), 10),
						"ExpectedArgsCount": "1",
						"Signature":         "([fieldName, otherFieldName])",
					},
					attr.Name,
				)
				continue
			}

			e := validateIdentArray(attr.Arguments[0].Expression, query.ModelFieldNames(model))
			errs.Concat(e)
			if len(e.Errors) > 0 {
				continue
			}

			value, _ := expressions.ToValue(attr.Arguments[0].Expression)
			if len(value.Array.Values) < 2 {
				errs.Append(errorhandling.ErrorInvalidValue,
					map[string]string{
						"Expected": "at least two field names to be provided",
					},
					attr.Arguments[0].Expression,
				)
			}
		}
	}

	return
}
