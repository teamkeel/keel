package expressions

import (
	"fmt"
	"regexp"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/teamkeel/keel/expressions/typing"
	"github.com/teamkeel/keel/schema/node"
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
func (p *Parser) Validate(expression string) ([]ValidationError, error) {
	ast, issues := p.CelEnv.Compile(expression)

	if issues != nil && issues.Err() != nil {
		validationErrors := []ValidationError{}

		for _, e := range issues.Errors() {
			parsed, _ := p.CelEnv.Parse(expression)
			offsets := parsed.NativeRep().SourceInfo().OffsetRanges()[e.ExprID]
			start := parsed.NativeRep().SourceInfo().GetStartLocation(e.ExprID)
			end := parsed.NativeRep().SourceInfo().GetStopLocation(e.ExprID)

			validationErrors = append(validationErrors, ValidationError{
				Message: convertMessage(e.Message),
				Node: node.Node{
					Pos: lexer.Position{
						Offset: int(offsets.Start),
						Line:   start.Line(),
						Column: start.Column() + 1,
					},
					EndPos: lexer.Position{
						Offset: int(offsets.Stop),
						Line:   end.Line(),
						Column: end.Column() + 1,
					},
				},
			})
		}

		return validationErrors, nil
	}

	if p.ExpectedReturnType != nil {
		if !ast.OutputType().IsAssignableType(p.ExpectedReturnType) {
			return []ValidationError{{
				Message: fmt.Sprintf("expression expected to resolve to type '%s' but it is '%s'", p.ExpectedReturnType.String(), ast.OutputType().String()),
			}}, nil
		}
	}

	// Valid expression

	return nil, nil
}

func convertMessage(message string) string {
	pattern := regexp.MustCompile(`found no matching overload for '([^']+)' applied to '\(([^,]+),\s*([^)]+)\)'`)
	if matches := pattern.FindStringSubmatch(message); matches != nil {
		return fmt.Sprintf("cannot use operator %s between %s and %s", matches[1], matches[2], matches[3])
	}

	return message
	// switch message {
	// 	case
	// }
}

type ValidationError struct {
	node.Node
	Message string
}
