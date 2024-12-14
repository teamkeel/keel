package parser

import (
	"errors"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/teamkeel/keel/schema/node"
)

type Expression struct {
	node.Node
}

func (e *Expression) Parse(lex *lexer.PeekingLexer) error {
	parenCount := 0
	for {
		t := lex.Peek()

		if t.EOF() {
			e.EndPos = t.Pos
			return nil
		}

		if t.Value == ")" || t.Value == "]" {
			parenCount--
			if parenCount < 0 {
				e.EndPos = t.Pos
				return nil
			}
		}

		if t.Value == "(" || t.Value == "[" {
			parenCount++
		}

		if t.Value == "," && parenCount == 0 {
			e.EndPos = t.Pos
			return nil
		}

		t = lex.Next()
		e.Tokens = append(e.Tokens, *t)

		if len(e.Tokens) == 1 {
			e.Pos = t.Pos
		}
	}

}

func (e *Expression) String() string {
	if len(e.Tokens) == 0 {
		return ""
	}

	var result strings.Builder
	firstToken := e.Tokens[0]
	currentLine := e.Pos.Line
	currentColumn := e.Pos.Column

	// Handle first token
	if firstToken.Pos.Line > currentLine {
		// Add necessary newlines
		result.WriteString(strings.Repeat("\n", firstToken.Pos.Line-currentLine))
		// Reset column position for new line
		currentColumn = 0
	}
	// Add spaces to reach the correct column position
	if firstToken.Pos.Column > currentColumn {
		result.WriteString(strings.Repeat(" ", firstToken.Pos.Column-currentColumn))
	}
	result.WriteString(firstToken.Value)
	currentLine = firstToken.Pos.Line
	currentColumn = firstToken.Pos.Column + len(firstToken.Value)

	// Handle subsequent tokens
	for i := 1; i < len(e.Tokens); i++ {
		curr := e.Tokens[i]

		if curr.Pos.Line > currentLine {
			// Add necessary newlines
			result.WriteString(strings.Repeat("\n", curr.Pos.Line-currentLine))
			// Reset column position for new line
			currentColumn = 0
		}

		// Add spaces to reach the correct column position
		if curr.Pos.Column > currentColumn {
			result.WriteString(strings.Repeat(" ", curr.Pos.Column-currentColumn))
		}

		result.WriteString(curr.Value)
		currentLine = curr.Pos.Line
		currentColumn = curr.Pos.Column + len(curr.Value)
	}

	return result.String()
}

func ParseExpression(source string) (*Expression, error) {
	parser, err := participle.Build[Expression]()
	if err != nil {
		return nil, err
	}

	expr, err := parser.ParseString("", source)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

type ExpressionIdent struct {
	node.Node

	Fragments []string
}

func (ident ExpressionIdent) ToString() string {
	idents := []string{}
	for _, v := range ident.Fragments {
		idents = append(idents, v)
	}

	return strings.Join(idents, ".")
}

var ErrInvalidAssignmentExpression = errors.New("expression is not a valid assignment")

// ToAssignmentExpression splits an assignment expression into two separate expressions.
// E.g. the expression `post.age = 1 + 1` will become `post.age` and `1 + 1`
func (expr *Expression) ToAssignmentExpression() (*Expression, *Expression, error) {
	parts := strings.Split(expr.String(), "=")
	if len(parts) != 2 {
		return nil, nil, ErrInvalidAssignmentExpression
	}

	if strings.TrimSpace(parts[0]) == "" {
		return nil, nil, ErrInvalidAssignmentExpression
	}

	if strings.TrimSpace(parts[1]) == "" {
		return nil, nil, ErrInvalidAssignmentExpression
	}

	lhs, err := ParseExpression(parts[0])
	if err != nil {
		return nil, nil, err
	}

	// Set position for left-hand side using original expression's position
	lhs.Pos = expr.Pos
	lhs.EndPos = expr.EndPos

	rhs, err := ParseExpression(parts[1])
	if err != nil {
		return nil, nil, err
	}

	// Set position for right-hand side starting after the equals sign
	rhs.Pos = expr.Pos
	rhs.EndPos = expr.EndPos

	return lhs, rhs, nil
}
