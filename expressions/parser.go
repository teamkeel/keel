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

// NewParser creates a new expression parser
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
	ast, issues := p.CelEnv.Compile(expr)

	if issues != nil && issues.Err() != nil {
		validationErrors := []*errorhandling.ValidationError{}

		for _, e := range issues.Errors() {
			msg := e.Message
			for _, match := range messageConverters {
				pattern, err := regexp.Compile(match.Regex)
				if err != nil {
					return nil, err
				}
				if matches := pattern.FindStringSubmatch(e.Message); matches != nil {
					msg = match.Construct(p.ExpectedReturnType, matches[1:])
					break
				}
			}

			var n node.Node
			if e.ExprID == 0 {
				// Synax errors means the expression could not be parsed, which means there are no expr nodes
				n = expression.Node
			} else {
				parsed, _ := p.CelEnv.Parse(expr)
				offset := parsed.NativeRep().SourceInfo().OffsetRanges()[e.ExprID]
				start := parsed.NativeRep().SourceInfo().GetStartLocation(e.ExprID)
				end := parsed.NativeRep().SourceInfo().GetStopLocation(e.ExprID)

				pos := lexer.Position{
					Offset: int(offset.Start),
					Line:   start.Line(),
					Column: start.Column(),
				}
				endPos := lexer.Position{
					Offset: int(offset.Stop),
					Line:   end.Line(),
					Column: end.Column(),
				}

				n = node.Node{
					Pos: lexer.Position{
						Filename: expression.Pos.Filename,
						Line:     expression.Pos.Line + pos.Line - 1,
						Column:   expression.Pos.Column + pos.Column,
						Offset:   expression.Pos.Offset + pos.Offset,
					},
					EndPos: lexer.Position{
						Filename: expression.Pos.Filename,
						Line:     expression.Pos.Line + endPos.Line - 1,
						Column:   expression.Pos.Column + endPos.Column,
						Offset:   expression.Pos.Offset + endPos.Offset,
					},
				}
			}

			validationErrors = append(validationErrors,
				errorhandling.NewValidationErrorWithDetails(
					errorhandling.AttributeExpressionError,
					errorhandling.ErrorDetails{
						Message: msg,
					},
					n,
				))

			// For syntax errors (i.e. unparseable expressions), we only need to show the first error.
			if strings.HasPrefix(e.Message, "Syntax error:") {
				break
			}
		}

		return validationErrors, nil
	}

	if p.ExpectedReturnType != nil && ast.OutputType() != types.NullType {
		out := mapType(ast.OutputType().String())

		// Backwards compatibility for relationships expressions which is actually performing an "ANY" query
		// For example, @where(supplier.products.brand.isActive)
		if mapType(p.ExpectedReturnType.String()) == typing.Boolean.String() && out == typing.BooleanArray.String() {
			return nil, nil
		}

		if out != "dyn[]" && !typesAssignable(p.ExpectedReturnType, ast.OutputType()) {
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

// typesAssignable defines the compatible assignable types
// For e.g. Markdown can be assigned by a Text value
func typesAssignable(expected *types.Type, actual *types.Type) bool {
	expectedMapped := mapType(expected.String())
	actualMapped := mapType(actual.String())

	// Define type compatibility rules
	// [key] can be assigned to by [values]
	typeCompatibility := map[string][]string{
		typing.Date.String():      {mapType(typing.Date.String()), mapType(typing.Timestamp.String())},
		typing.Timestamp.String(): {mapType(typing.Date.String()), mapType(typing.Timestamp.String())},
		typing.Markdown.String():  {mapType(typing.Text.String()), mapType(typing.Markdown.String())},
		typing.ID.String():        {mapType(typing.Text.String()), mapType(typing.ID.String())},
		typing.Text.String():      {mapType(typing.Text.String()), mapType(typing.Markdown.String()), mapType(typing.ID.String())},
		typing.Number.String():    {mapType(typing.Number.String()), mapType(typing.Decimal.String())},
		typing.Decimal.String():   {mapType(typing.Number.String()), mapType(typing.Decimal.String())},
		typing.Duration.String():  {mapType(typing.Duration.String()), mapType(typing.Text.String())},
	}

	// Check if there are specific compatibility rules for the expected type
	if compatibleTypes, exists := typeCompatibility[expected.String()]; exists {
		for _, compatibleType := range compatibleTypes {
			if actualMapped == compatibleType {
				return true
			}
		}
		return false
	}

	// Default case: types must match exactly
	return expectedMapped == actualMapped
}
