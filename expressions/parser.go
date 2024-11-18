package expressions

import (
	"fmt"

	"github.com/google/cel-go/cel"

	"github.com/google/cel-go/common/types"
)

type ExpressionParser struct {
	celEnv             *cel.Env
	provider           *typeProvider
	expectedReturnType *types.Type
}

// NewParser creates a new expression parser with all the options applied
func NewParser(options ...Option) (*ExpressionParser, error) {
	typeProvider := NewTypeProvider()

	env, err := cel.NewCustomEnv(
		standardKeelLibrary(),
		cel.ClearMacros(),
		cel.CustomTypeProvider(typeProvider),
		cel.EagerlyValidateDeclarations(true),
	)
	if err != nil {
		return nil, fmt.Errorf("program setup err: %s", err)
	}

	parser := &ExpressionParser{
		celEnv:   env,
		provider: typeProvider,
	}

	for _, opt := range options {
		if err := opt(parser); err != nil {
			return nil, err
		}
	}

	return parser, nil
}

// Validate parses and validates the expression
func (p *ExpressionParser) Validate(expression string) ([]string, error) {
	ast, issues := p.celEnv.Compile(expression)
	if issues != nil && issues.Err() != nil {
		validationErrors := []string{}
		for _, e := range issues.Errors() {
			validationErrors = append(validationErrors, e.Message)
		}
		return validationErrors, nil
	}

	if p.expectedReturnType != nil {
		if !ast.OutputType().IsExactType(p.expectedReturnType) {
			return []string{fmt.Sprintf("expression expected to resolve to type '%s' but it is '%s'", p.expectedReturnType.String(), ast.OutputType().String())}, nil
		}
	}

	// Valid expression
	return nil, nil
}

// // Build will construct a SQL statement for the expression
// func (p *Parser) Build(query *actions.QueryBuilder, expression string, input map[string]any) error {
// 	checkedExpr, err := cel.AstToCheckedExpr(p.ast)
// 	if err != nil {
// 		return err
// 	}

// 	un := &builder{
// 		query: query,
// 	}
// 	if err := un.visit(checkedExpr.Expr); err != nil {
// 		return err
// 	}

// 	return nil
// }
