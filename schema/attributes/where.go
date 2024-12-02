package attributes

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func ValidateWhereExpression(schema []*parser.AST, action *parser.ActionNode, expression *parser.Expression) ([]expressions.ValidationError, error) {
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

	issues, err := p.Validate(expression.String())

	// TODO: this is not working correctly yet when expressions span multiple lines
	for i, _ := range issues {
		issues[i].Pos = expression.Pos.Add(issues[i].Pos)
		issues[i].EndPos = expression.Pos.Add(issues[i].EndPos)
	}

	return issues, err
}
