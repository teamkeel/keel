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
			return nil
		}

		if t.Value == ")" {
			parenCount--
			if parenCount < 0 {
				return nil
			}
		}

		if t.Value == "(" {
			parenCount++
		}

		t = lex.Next()
		e.Tokens = append(e.Tokens, *t)

		if len(e.Tokens) == 1 {
			e.Pos = t.Pos
		}

		e.EndPos = t.Pos
	}
}

func (e *Expression) String() string {
	v := ""
	for i, t := range e.Tokens {
		if i == 0 {
			v += t.Value
			continue
		}
		last := e.Tokens[i-1]
		hasWhitespace := (last.Pos.Offset + len(last.Value)) < t.Pos.Offset
		if hasWhitespace {
			v += " "
		}
		v += t.Value
	}
	return v
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

var ErrInvalidAssignmentExpression = errors.New("assignment expression is not valid")

func (expr *Expression) ToAssignmentExpression() ([]string, string, error) {

	parts := strings.Split(expr.String(), "=")

	if len(parts) != 2 {
		return nil, "", ErrInvalidAssignmentExpression
	}

	return strings.Split(parts[0], "."), parts[1], nil
}
