package attributes

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func ValidateWhereExpression(schema []*parser.AST, action *parser.ActionNode, expression string) ([]string, error) {
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
