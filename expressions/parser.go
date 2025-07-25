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

// NewParser creates a new expression parser.
func NewParser(options ...Option) (*Parser, error) {
	typeProvider := typing.NewTypeProvider()

	env, err := cel.NewCustomEnv(
		standardKeelLibrary(),
		//options.RegisterAggregationFunctions(),
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

// Extend creates a new parser with the same environment configuration but extended with additional options.
func (p *Parser) Extend(options ...Option) (*Parser, error) {
	env, err := p.CelEnv.Extend([]cel.EnvOption{}...)
	if err != nil {
		return nil, err
	}

	newParser := &Parser{
		CelEnv:             env,
		Provider:           p.Provider,
		ExpectedReturnType: p.ExpectedReturnType,
	}

	for _, opt := range options {
		if err := opt(newParser); err != nil {
			return nil, err
		}
	}

	return newParser, nil
}

// Validate validates an expression and returns a list of validation errors.
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
		if mapType(p.ExpectedReturnType.String()) == typing.TypeBoolean.String() && out == typing.TypeBooleanArray.String() {
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
// For e.g. Markdown can be assigned by a Text value.
func typesAssignable(expected *types.Type, actual *types.Type) bool {
	expectedMapped := mapType(expected.String())
	actualMapped := mapType(actual.String())

	// Define type compatibility rules
	// [key] can be assigned to by [values]
	typeCompatibility := map[string][]string{
		typing.TypeDate.String():      {mapType(typing.TypeDate.String()), mapType(typing.TypeTimestamp.String())},
		typing.TypeTimestamp.String(): {mapType(typing.TypeDate.String()), mapType(typing.TypeTimestamp.String())},
		typing.TypeMarkdown.String():  {mapType(typing.TypeText.String()), mapType(typing.TypeMarkdown.String())},
		typing.TypeID.String():        {mapType(typing.TypeText.String()), mapType(typing.TypeID.String())},
		typing.TypeText.String():      {mapType(typing.TypeText.String()), mapType(typing.TypeMarkdown.String()), mapType(typing.TypeID.String())},
		typing.TypeNumber.String():    {mapType(typing.TypeNumber.String()), mapType(typing.TypeDecimal.String())},
		typing.TypeDecimal.String():   {mapType(typing.TypeNumber.String()), mapType(typing.TypeDecimal.String())},
		typing.TypeDuration.String():  {mapType(typing.TypeDuration.String()), mapType(typing.TypeText.String())},
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
