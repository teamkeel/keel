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

			if section.Actions != nil {
				for _, function := range section.Actions {
					errs.Concat(checkAttributes(function.Attributes, parser.KeywordActions, function.Name.Value))
				}
			}

			if section.Fields != nil {
				for _, field := range section.Fields {
					errs.Concat(checkAttributes(field.Attributes, "field", field.Name.Value))
				}
			}
		}
	}

	for _, job := range query.Jobs(asts) {
		for _, section := range job.Sections {
			if section.Attribute != nil {
				errs.Concat(checkAttributes([]*parser.AttributeNode{section.Attribute}, "job", job.Name.Value))
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
		parser.AttributeOn,
	},
	parser.KeywordField: {
		parser.AttributeUnique,
		parser.AttributeDefault,
		parser.AttributePrimaryKey,
		parser.AttributeRelation,
	},
	parser.KeywordActions: {
		parser.AttributeSet,
		parser.AttributeWhere,
		parser.AttributePermission,
		parser.AttributeValidate,
		parser.AttributeOrderBy,
		parser.AttributeSortable,
		parser.AttributeFunction,
	},
	parser.KeywordJob: {
		parser.AttributePermission,
		parser.AttributeSchedule,
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

func ValidateFieldAttributeRule(asts []*parser.AST) (errs errorhandling.ValidationErrors) {
	for _, model := range query.Models(asts) {
		for _, field := range query.ModelFields(model) {
			for _, attr := range field.Attributes {
				if attr.Name.Value != parser.AttributeDefault {
					continue
				}

				errs.Concat(
					validateModelFieldDefaultAttribute(asts, model, field, attr),
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
func validateModelFieldDefaultAttribute(
	asts []*parser.AST,
	model *parser.ModelNode,
	field *parser.FieldNode,
	attr *parser.AttributeNode,
) (errs errorhandling.ValidationErrors) {
	expressionContext := expressions.ExpressionContext{
		Model:     model,
		Attribute: attr,
		Field:     field,
	}

	argLength := len(attr.Arguments)

	if argLength == 0 {
		err := expression.DefaultCanUseZeroValueRule(asts, attr, expressionContext)
		for _, e := range err {
			errs.AppendError(e)
		}
		return
	}

	if argLength >= 2 {
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

	rules := []expression.Rule{expression.ValueTypechecksRule}

	err := expression.ValidateExpression(
		asts,
		expr,
		rules,
		expressionContext,
	)
	for _, e := range err {
		// TODO: remove case when expression.ValidateExpression returns correct type
		errs.AppendError(e.(*errorhandling.ValidationError))
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

	rules := []expression.Rule{}

	if attr.Name.Value == parser.AttributeSet {
		rules = append(rules, expression.OperatorAssignmentRule)
	} else {
		rules = append(rules, expression.OperatorLogicalRule)
	}

	err := expression.ValidateExpression(
		asts,
		expr,
		rules,
		expressions.ExpressionContext{
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

func validateIdentArray(expr *parser.Expression, allowedIdents []string) (errs errorhandling.ValidationErrors) {
	value, err := expr.ToValue()
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
		// we dont want to validate built in models
		if model.BuiltIn {
			continue
		}
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

			value, _ := attr.Arguments[0].Expression.ToValue()
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
