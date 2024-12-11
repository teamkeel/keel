package attributes

import (
	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ValidateCompositeUnique(model *parser.ModelNode, expression *parser.Expression) ([]*errorhandling.ValidationError, error) {
	opts := []expressions.Option{
		options.WithReturnTypeAssertion("_FieldName", true),
	}

	for _, f := range query.ModelFields(model) {
		if f.Repeated {
			continue
		}
		if f.Type.Value == parser.FieldTypeTimestamp {
			continue
		}
		opts = append(opts, options.WithConstant(f.Name.Value, "_FieldName"))
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression)
}
