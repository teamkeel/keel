package orderby_expression

import (
	"fmt"

	"github.com/google/cel-go/cel"

	"github.com/teamkeel/keel/proto"
)

type OrderByExpressionParser struct {
	env *cel.Env
	ast *cel.Ast
}

type OrderByExpression struct {
}

func NewOrderByExpressionParser(schema *proto.Schema, model *proto.Model) (*OrderByExpressionParser, error) {
	typeProvider := NewTypeProvider(schema, model)

	env, err := cel.NewCustomEnv(
		orderByAttributeLib(),
		cel.ClearMacros(),
		cel.CustomTypeProvider(typeProvider),
		cel.EagerlyValidateDeclarations(true),
	)
	if err != nil {
		return nil, fmt.Errorf("program setup err: %s", err)
	}

	return &OrderByExpressionParser{
		env: env,
	}, nil
}

// Validate parses and validates the expression
func (p *OrderByExpressionParser) Validate(expression string) ([]string, error) {
	ast, issues := p.env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		validationErrors := []string{}
		for _, e := range issues.Errors() {
			validationErrors = append(validationErrors, e.Message)
		}
		return validationErrors, nil
	}

	// if ast.OutputType() !=   {
	// 	return []string{fmt.Sprintf("expression expected to resolve to type '%s'", expectedOutoutType.GetType())}, nil
	// }

	p.ast = ast

	// Valid expression
	return nil, nil
}
