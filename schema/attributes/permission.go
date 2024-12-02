package attributes

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func ValidatePermissionExpression(schema []*parser.AST, action *parser.ActionNode, expression string) ([]expressions.ValidationError, error) {
	model := query.ActionModel(schema, action.Name.Value)

	opts := []expressions.Option{
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithActionInputs(schema, action),
		options.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value),
		options.WithComparisonOperators(),
		options.WithLogicalOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}

func ValidatePermissionRoles(schema []*parser.AST, expression string) ([]expressions.ValidationError, error) {
	opts := []expressions.Option{
		options.WithSchemaTypes(schema),
		options.WithReturnTypeAssertion("_RoleDefinition", true),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}

func ValidatePermissionActions(expression string) ([]expressions.ValidationError, error) {
	opts := []expressions.Option{
		options.WithVariable(parser.ActionTypeGet, "_ActionTypeDefinition"),
		options.WithVariable(parser.ActionTypeCreate, "_ActionTypeDefinition"),
		options.WithVariable(parser.ActionTypeUpdate, "_ActionTypeDefinition"),
		options.WithVariable(parser.ActionTypeList, "_ActionTypeDefinition"),
		options.WithVariable(parser.ActionTypeDelete, "_ActionTypeDefinition"),
		options.WithReturnTypeAssertion("_ActionTypeDefinition", true),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}
