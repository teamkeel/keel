package attributes

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
)

func ValidatePermissionExpression(schema []*parser.AST, model *parser.ModelNode, action *parser.ActionNode, job *parser.JobNode, expression *parser.Expression) ([]expressions.ValidationError, error) {
	opts := []expressions.Option{
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithComparisonOperators(),
		options.WithLogicalOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	}

	if action != nil {
		opts = append(opts, options.WithActionInputs(schema, action))
	}

	if model != nil {
		opts = append(opts, options.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value))
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression.String())
}

func ValidatePermissionRoles(schema []*parser.AST, expression *parser.Expression) ([]expressions.ValidationError, error) {
	opts := []expressions.Option{
		options.WithSchemaTypes(schema),
		options.WithReturnTypeAssertion("_Role", true),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression.String())
}

func ValidatePermissionActions(expression *parser.Expression) ([]expressions.ValidationError, error) {
	opts := []expressions.Option{
		options.WithVariable(parser.ActionTypeGet, "_ActionType"),
		options.WithVariable(parser.ActionTypeCreate, "_ActionType"),
		options.WithVariable(parser.ActionTypeUpdate, "_ActionType"),
		options.WithVariable(parser.ActionTypeList, "_ActionType"),
		options.WithVariable(parser.ActionTypeDelete, "_ActionType"),
		options.WithReturnTypeAssertion("_ActionType", true),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression.String())
}
