package attributes

import (
	"fmt"

	"github.com/iancoleman/strcase"

	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func ValidateSetExpression(schema []*parser.AST, action *parser.ActionNode, target *parser.Expression, expression *parser.Expression) ([]expressions.ValidationError, error) {
	targetField, err := resolve.AsIdent(target.String())
	if err != nil {
		return nil, fmt.Errorf("lhs operand incorrect")
	}

	if len(targetField) < 2 {
		return nil, fmt.Errorf("lhs operand incorrect")
	}

	model := query.ActionModel(schema, action.Name.Value)

	if targetField[0] != strcase.ToLowerCamel(model.Name.Value) {
		return nil, fmt.Errorf("wrong model")
	}

	var field *parser.FieldNode
	currModel := model
	for i, fragment := range targetField {
		if i == 0 {
			continue
		}
		field = query.Field(currModel, fragment)
		if i < len(targetField)-1 {
			currModel = query.Model(schema, field.Type.Value)
		}
	}

	opts := []expressions.Option{
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value),
		options.WithActionInputs(schema, action),
		options.WithReturnTypeAssertion(field.Type.Value, field.Repeated),
	}

	p, err := expressions.NewParser(opts...)
	if err != nil {
		return nil, err
	}

	return p.Validate(expression.String())
}
