package attributes

import (
	"encoding/hex"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var computed = make(map[string]*expressions.Parser)

func defaultComputed(schema []*parser.AST) (*expressions.Parser, error) {
	mutex.Lock()
	defer mutex.Unlock()

	var contents string
	for _, s := range schema {
		contents += s.Raw + "\n"
	}
	key := hex.EncodeToString([]byte(contents))

	if parser, exists := computed[key]; exists {
		return parser, nil
	}

	opts := []expressions.Option{
		options.WithSchemaTypes(schema),
		options.WithComparisonOperators(),
		options.WithLogicalOperators(),
		options.WithArithmeticOperators(),
		options.WithFunctions(),
	}

	parser, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	computed[key] = parser

	return parser, nil
}

func ValidateComputedExpression(schema []*parser.AST, model *parser.ModelNode, field *parser.FieldNode, expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
	parser, err := defaultComputed(schema)
	if err != nil {
		return nil, err
	}

	opts := []expressions.Option{
		options.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value, false),
		options.WithReturnTypeAssertion(field.Type.Value, field.Repeated),
	}

	p, err := parser.Extend(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}
