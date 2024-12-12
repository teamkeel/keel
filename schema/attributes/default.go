package attributes

import (
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ValidateDefaultExpression(schema []*parser.AST, field *parser.FieldNode, expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
	returnType := field.Type.Value
	if field.Type.Value == parser.FieldTypeID {
		returnType = parser.FieldTypeText
	}

	opts := []expressions.Option{
		options.WithSchemaTypes(schema),
		options.WithReturnTypeAssertion(returnType, field.Repeated),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}
