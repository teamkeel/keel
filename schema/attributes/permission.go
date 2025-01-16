package attributes

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/expressions/typing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ValidatePermissionExpression(schema []*parser.AST, model *parser.ModelNode, action *parser.ActionNode, job *parser.JobNode, expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
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
		opts = append(opts, options.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value, false))
		opts = append(opts, options.WithVariable(parser.ThisVariable, model.Name.Value, false))
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}

func ValidatePermissionRoles(schema []*parser.AST, expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
	opts := []expressions.Option{
		options.WithSchemaTypes(schema),
		options.WithReturnTypeAssertion(typing.Role.String(), true),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}

func ValidatePermissionActions(expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
	opts := []expressions.Option{
		options.WithConstant(parser.ActionTypeGet, "_ActionType"),
		options.WithConstant(parser.ActionTypeCreate, "_ActionType"),
		options.WithConstant(parser.ActionTypeUpdate, "_ActionType"),
		options.WithConstant(parser.ActionTypeList, "_ActionType"),
		options.WithConstant(parser.ActionTypeDelete, "_ActionType"),
		options.WithReturnTypeAssertion("_ActionType", true),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}
