package expressions

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/teamkeel/keel/expressions/typing"
)

type Parser struct {
	CelEnv             *cel.Env
	Provider           *typing.TypeProvider
	ExpectedReturnType *types.Type
}

type Option func(*Parser) error

// NewParser creates a new expression parser with all the options applied
func NewParser(options ...Option) (*Parser, error) {
	typeProvider := typing.NewTypeProvider()

	env, err := cel.NewCustomEnv(
		standardKeelLibrary(),
		cel.ClearMacros(),
		cel.CustomTypeProvider(typeProvider),
		cel.EagerlyValidateDeclarations(true),
	)
	if err != nil {
		return nil, fmt.Errorf("program setup err: %s", err)
	}

	parser := &Parser{
		CelEnv:   env,
		Provider: typeProvider,
	}

	for _, opt := range options {
		if err := opt(parser); err != nil {
			return nil, err
		}
	}

	return parser, nil
}

// Validate parses and validates the expression
func (p *Parser) Validate(expression string) ([]string, error) {
	ast, issues := p.CelEnv.Compile(expression)
	if issues != nil && issues.Err() != nil {
		validationErrors := []string{}
		for _, e := range issues.Errors() {
			validationErrors = append(validationErrors, e.Message)
		}
		return validationErrors, nil
	}

	if p.ExpectedReturnType != nil {
		if !ast.OutputType().IsAssignableType(p.ExpectedReturnType) {
			return []string{fmt.Sprintf("expression expected to resolve to type '%s' but it is '%s'", p.ExpectedReturnType.String(), ast.OutputType().String())}, nil
		}
	}

	// Valid expression
	return nil, nil
}

type ValidationError struct {
	Message string
}
