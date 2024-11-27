package attributes

import (
	"fmt"

	"github.com/iancoleman/strcase"

	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func NewSetExpressionParser(schema []*parser.AST, targetField *parser.Ident, action *parser.ActionNode) (*expressions.Parser, error) {
	if len(targetField.Fragments) < 2 {
		return nil, fmt.Errorf("lhs operand incorrect")
	}

	model := query.ActionModel(schema, action.Name.Value)

	if targetField.Fragments[0].Fragment != strcase.ToLowerCamel(model.Name.Value) {
		return nil, fmt.Errorf("wrong model")
	}

	var field *parser.FieldNode
	currModel := model
	for i, fragment := range targetField.Fragments {
		if i == 0 {
			continue
		}
		field = query.Field(currModel, fragment.Fragment)
		if i < len(targetField.Fragments)-1 {
			currModel = query.Model(schema, field.Type.Value)
		}
	}

	opts := []expressions.Option{
		expressions.WithCtx(),
		expressions.WithSchemaTypes(schema),
		expressions.WithActionInputs(schema, action),
		expressions.WithReturnTypeAssertion(field.Type.Value, field.Repeated),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p, nil

}
