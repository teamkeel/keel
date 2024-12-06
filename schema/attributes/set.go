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

func ValidateSetExpression(schema []*parser.AST, action *parser.ActionNode, lhs *parser.Expression, rhs *parser.Expression) ([]expressions.ValidationError, error) {
	model := query.ActionModel(schema, action.Name.Value)

	lhsOpts := []expressions.Option{
		options.WithSchemaTypes(schema),
		options.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value),
	}

	lhsParser, err := expressions.NewParser(lhsOpts...)
	if err != nil {
		return nil, err
	}

	issues, err := lhsParser.Validate(lhs.String())
	if err != nil {
		return nil, err
	}

	if len(issues) > 0 {
		projectIssuesToPosition(lhs.Node, issues)
		return issues, nil
	}

	targetField, err := resolve.AsIdent(lhs.String())
	if err != nil {
		return nil, fmt.Errorf("lhs operand incorrect")
	}

	if len(targetField) < 2 {
		return nil, fmt.Errorf("lhs operand incorrect")
	}

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

	rhsOpts := []expressions.Option{
		options.WithCtx(),
		options.WithSchemaTypes(schema),
		options.WithVariable(strcase.ToLowerCamel(model.Name.Value), model.Name.Value),
		options.WithActionInputs(schema, action),
		options.WithReturnTypeAssertion(field.Type.Value, field.Repeated),
	}

	rhsParser, err := expressions.NewParser(rhsOpts...)
	if err != nil {
		return nil, err
	}

	issues, err = rhsParser.Validate(rhs.String())
	if err != nil {
		return nil, err
	}

	projectIssuesToPosition(lhs.Node, issues)
	return issues, nil
}
