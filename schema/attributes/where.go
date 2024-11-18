package attributes

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func NewWhereExpressionParser(schema []*parser.AST, action *parser.ActionNode) (*expressions.ExpressionParser, error) {
	model := query.ActionModel(schema, action.Name.Value)

	opts := []expressions.Option{
		expressions.WithCtx(),
		expressions.WithSchema(schema),
		expressions.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value),
		expressions.WithComparisonOperators(),
		expressions.WithLogicalOperators(),
		expressions.WithReturnTypeAssertion(parser.FieldTypeBoolean),
	}

	// Add filter inputs as variables
	for _, f := range action.Inputs {
		t := query.ResolveInputType(schema, f, model, action)
		opts = append(opts, expressions.WithVariable(f.Name(), t))
	}

	// Add with inputs as variables
	for _, f := range action.With {
		t := query.ResolveInputType(schema, f, model, action)
		opts = append(opts, expressions.WithVariable(f.Name(), t))
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p, nil
}
