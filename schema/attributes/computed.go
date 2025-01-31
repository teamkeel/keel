package attributes

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ValidateComputedExpression(schema []*parser.AST, model *parser.ModelNode, field *parser.FieldNode, expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
	opts := []expressions.Option{
		options.WithSchemaTypes(schema),
		options.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value, false),
		options.WithComparisonOperators(),
		options.WithLogicalOperators(),
		options.WithArithmeticOperators(),
		options.WithFunctions(),
		options.WithReturnTypeAssertion(field.Type.Value, field.Repeated),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}
