package expressions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/teamkeel/keel/expressions/typing"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
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

func (p *Parser) Validate(expression *parser.Expression) ([]*errorhandling.ValidationError, error) {

	expr := expression.String()
	expr = strings.ReplaceAll(expr, " and ", " && ")
	expr = strings.ReplaceAll(expr, " or ", " || ")

	ast, issues := p.CelEnv.Compile(expr)

	if issues != nil && issues.Err() != nil {
		validationErrors := []*errorhandling.ValidationError{}

		for _, e := range issues.Errors() {
			msg := e.Message
			for _, match := range matches {
				pattern, err := regexp.Compile(match.Regex)
				if err != nil {
					return nil, err
				}
				if matches := pattern.FindStringSubmatch(e.Message); matches != nil {
					msg = match.Construct(matches[1:])
					break
				}
			}

			parsed, _ := p.CelEnv.Parse(expr)
			offsets := parsed.NativeRep().SourceInfo().OffsetRanges()[e.ExprID]
			start := parsed.NativeRep().SourceInfo().GetStartLocation(e.ExprID)
			end := parsed.NativeRep().SourceInfo().GetStopLocation(e.ExprID)

			pos := lexer.Position{
				Offset: int(offsets.Start),
				Line:   start.Line(),
				Column: start.Column(),
			}
			endPos := lexer.Position{
				Offset: int(offsets.Stop),
				Line:   end.Line(),
				Column: end.Column(),
			}

			node := node.Node{
				Pos:    pos.Add(expression.Pos),
				EndPos: endPos.Add(expression.Pos),
			}

			validationErrors = append(validationErrors,
				errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeExpressionError,
					errorhandling.ErrorDetails{
						Message: msg,
					},
					node,
				))
		}

		return validationErrors, nil
	}

	if p.ExpectedReturnType != nil {
		if ast.OutputType() != types.NullType && !ast.OutputType().IsExactType(p.ExpectedReturnType) {
			return []*errorhandling.ValidationError{
				errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeExpressionError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("expression expected to resolve to type %s but it is %s", mapType(p.ExpectedReturnType.String()), mapType(ast.OutputType().String())),
					},
					expression),
			}, nil
		}
	}

	return nil, nil
}
