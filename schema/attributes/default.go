package attributes

import (
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/schema/parser"
)

func NewDefaultExpressionParser(schema []*parser.AST, field *parser.FieldNode) (*expressions.Parser, error) {
	opts := []expressions.Option{
		expressions.WithSchemaTypes(schema),
		expressions.WithReturnTypeAssertion(field.Type.Value, field.Repeated),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p, nil
}
