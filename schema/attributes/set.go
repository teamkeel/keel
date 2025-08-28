package attributes

import (
	"fmt"

	"github.com/iancoleman/strcase"

	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/options"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func ValidateSetExpression(schema []*parser.AST, action *parser.ActionNode, lhs *parser.Expression, rhs *parser.Expression) ([]*errorhandling.ValidationError, error) {
	model := query.ActionModel(schema, action.Name.Value)

	lhsOpts := []expressions.Option{
		options.WithSchemaTypes(schema),
		options.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value, false),
	}

	lhsParser, err := expressions.NewParser(lhsOpts...)
	if err != nil {
		return nil, err
	}

	issues, err := lhsParser.Validate(lhs)
	if err != nil {
		return nil, err
	}

	if len(issues) > 0 {
		return issues, err
	}

	targetField, err := resolve.AsIdent(lhs)
	if err != nil {
		return nil, err
	}

	if len(targetField.Fragments) < 2 {
		return nil, fmt.Errorf("lhs operand is less than two fragments")
	}

	if targetField.Fragments[0] != strcase.ToLowerCamel(model.Name.Value) {
		return nil, fmt.Errorf("wrong model")
	}

	var field *parser.FieldNode
	currModel := model
	for i, fragment := range targetField.Fragments {
		if i == 0 {
			continue
		}
		field = currModel.Field(fragment)
		if i < len(targetField.Fragments)-1 {
			currModel = query.Model(schema, field.Type.Value)
		}
	}

	rhsOpts := []expressions.Option{
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value, false),
		options.WithActionInputs(schema, action),
		options.WithReturnTypeAssertion(field.Type.Value, field.Repeated),
	}

	rhsParser, err := expressions.NewParser(rhsOpts...)
	if err != nil {
		return nil, err
	}

	return rhsParser.Validate(rhs)
}
