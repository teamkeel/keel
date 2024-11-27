package attributes

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func NewPermissionExpressionParser(schema []*parser.AST, action *parser.ActionNode) (*expressions.Parser, error) {
	model := query.ActionModel(schema, action.Name.Value)

	opts := []expressions.Option{
		expressions.WithCtx(),
		expressions.WithSchemaTypes(schema),
		expressions.WithActionInputs(schema, action),
		expressions.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value),
		expressions.WithComparisonOperators(),
		expressions.WithLogicalOperators(),
		expressions.WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func NewPermissionRoleParser(schema []*parser.AST) (*expressions.Parser, error) {
	opts := []expressions.Option{
		expressions.WithSchemaTypes(schema),
		expressions.WithReturnTypeAssertion("_RoleDefinition", true),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func NewPermissionActionsParser() (*expressions.Parser, error) {
	opts := []expressions.Option{
		expressions.WithVariable(parser.ActionTypeGet, "_ActionTypeDefinition"),
		expressions.WithVariable(parser.ActionTypeCreate, "_ActionTypeDefinition"),
		expressions.WithVariable(parser.ActionTypeUpdate, "_ActionTypeDefinition"),
		expressions.WithVariable(parser.ActionTypeList, "_ActionTypeDefinition"),
		expressions.WithVariable(parser.ActionTypeDelete, "_ActionTypeDefinition"),
		expressions.WithReturnTypeAssertion("_ActionTypeDefinition", true),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p, nil
}
