package attributes

import (
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ValidateCompositeUnique(entity parser.Entity, expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
	opts := []expressions.Option{
		options.WithReturnTypeAssertion("_FieldName", true),
	}

	for _, f := range entity.Fields() {
		opts = append(opts, options.WithConstant(f.Name.Value, "_FieldName"))
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}
