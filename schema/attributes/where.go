package attributes

import (
	"encoding/hex"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var wheres = make(map[string]*expressions.Parser)

// defaultWhere will cache the base CEL environment for a schema
func defaultWhere(schema []*parser.AST) (*expressions.Parser, error) {
	var contents string
	for _, s := range schema {
		contents += s.Raw + "\n"
	}
	key := hex.EncodeToString([]byte(contents))

	if parser, exists := wheres[key]; exists {
		return parser, nil
	}

	opts := []expressions.Option{
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithComparisonOperators(),
		options.WithLogicalOperators(),
		options.WithReturnTypeAssertion(parser.FieldTypeBoolean, false),
	}

	parser, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	wheres[key] = parser

	return parser, nil
}

func ValidateWhereExpression(schema []*parser.AST, action *parser.ActionNode, expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
	parser, err := defaultWhere(schema)
	if err != nil {
		return nil, err
	}

	model := query.ActionModel(schema, action.Name.Value)

	opts := []expressions.Option{
		options.WithActionInputs(schema, action),
		options.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value, false),
	}

	p, err := parser.Extend(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}
