package resolve_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/schema/parser"
)

func TestGenerateExpression(t *testing.T) {
	expression, err := parser.ParseExpression("1 + 1 * 10 + 10 < 0 AND true")
	assert.NoError(t, err)

	sql, err := resolve.RunCelVisitor(expression, generator())
	assert.NoError(t, err)

	expected := `
1 + 
	1 * 10 + 
10
`

	assert.Equal(t, expected, sql)
}

func generator() resolve.Visitor[string] {
	return &expressionVisitor{}
}

var _ resolve.Visitor[string] = new(expressionVisitor)

type expressionVisitor struct {
	expression string
	level      int
}

func (v *expressionVisitor) printLevel() string {
	return strings.Repeat(" ", v.level)
}

func (v *expressionVisitor) StartTerm(nested bool) error {

	v.level++

	if nested {
		v.expression += v.printLevel() + "("
	}

	v.expression += "\n"

	return nil
}

func (v *expressionVisitor) EndTerm(nested bool) error {
	v.level--

	if nested {
		v.expression += v.printLevel() + ")"
	}

	return nil
}

func (v *expressionVisitor) StartFunction(name string) error {
	v.expression += v.printLevel() + name + "("
	return nil
}

func (v *expressionVisitor) EndFunction() error {
	v.expression += v.printLevel() + ")"
	return nil
}

func (v *expressionVisitor) StartArgument(num int) error {
	return nil
}

func (v *expressionVisitor) EndArgument() error {
	return nil
}

func (v *expressionVisitor) VisitAnd() error {
	v.expression += " AND "
	return nil
}

func (v *expressionVisitor) VisitOr() error {
	v.expression += " OR "
	return nil
}

func (v *expressionVisitor) VisitNot() error {
	v.expression += " NOT "
	return nil
}

func (v *expressionVisitor) VisitOperator(op string) error {

	v.expression += " " + strings.Trim(op, "_") + " "
	return nil
}

func (v *expressionVisitor) VisitLiteral(value any) error {
	v.expression += fmt.Sprintf("%v", value)
	return nil
}

func (v *expressionVisitor) VisitIdent(ident *parser.ExpressionIdent) error {
	v.expression += ident.String()
	return nil
}

func (v *expressionVisitor) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return nil
}

func (v *expressionVisitor) Result() (string, error) {
	return v.expression, nil
}
