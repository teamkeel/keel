package attributes

import (
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
)

func ValidateDefaultExpression(schema []*parser.AST, field *parser.FieldNode, expression string) ([]string, error) {
	opts := []expressions.Option{
		options.WithSchemaTypes(schema),
		options.WithReturnTypeAssertion(field.Type.Value, field.Repeated),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}
