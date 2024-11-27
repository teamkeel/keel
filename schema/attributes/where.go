package attributes

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func NewWhereExpressionParser(schema []*parser.AST, action *parser.ActionNode) (*expressions.Parser, error) {
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
