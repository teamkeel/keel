package attributes

import (
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/node"
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
	if err != nil {
		return nil, err
	}

	for i, issue := range issues {
		msg, err := ConvertMessage(issue.Message)
		if err != nil {
			return nil, err
		}
		issues[i].Message = msg
	}

	projectIssuesToPosition(expression.Node, issues)

	return issues, err
}

func projectIssuesToPosition(expressionPosition node.Node, issues []expressions.ValidationError) {
	// TODO: this is not working correctly yet when expressions span multiple lines
	for i, _ := range issues {
		if issues[i].Pos != *new(lexer.Position) || issues[i].EndPos != *new(lexer.Position) {
			issues[i].Pos = expressionPosition.Pos.Add(issues[i].Pos)
			issues[i].EndPos = expressionPosition.Pos.Add(issues[i].EndPos)
		} else {
			issues[i].Pos = expressionPosition.Pos
			issues[i].EndPos = expressionPosition.EndPos
		}
	}
}
