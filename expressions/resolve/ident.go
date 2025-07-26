package resolve

import (
	"errors"

	"github.com/teamkeel/keel/schema/parser"
)

var ErrExpressionNotValidIdent = errors.New("expression is not an ident")

// AsIdent expects and retrieves the single ident operand in an expression.
func AsIdent(expression *parser.Expression) (*parser.ExpressionIdent, error) {
	ident, err := RunCelVisitor(expression, ident())
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func ident() Visitor[*parser.ExpressionIdent] {
	return &identGen{}
}

var _ Visitor[*parser.ExpressionIdent] = new(identGen)

type identGen struct {
	ident *parser.ExpressionIdent
}

func (v *identGen) StartTerm(parenthesis bool) error {
	return nil
}

func (v *identGen) EndTerm(parenthesis bool) error {
	return nil
}

func (v *identGen) StartFunction(name string) error {
	return nil
}

func (v *identGen) EndFunction() error {
	return nil
}

func (v *identGen) StartArgument(num int) error {
	return nil
}

func (v *identGen) EndArgument() error {
	return nil
}

func (v *identGen) VisitAnd() error {
	return ErrExpressionNotValidIdent
}

func (v *identGen) VisitOr() error {
	return ErrExpressionNotValidIdent
}

func (v *identGen) VisitNot() error {
	return nil
}

func (v *identGen) VisitOperator(op string) error {
	return ErrExpressionNotValidIdent
}

func (v *identGen) VisitLiteral(value any) error {
	return ErrExpressionNotValidIdent
}

func (v *identGen) VisitIdent(ident *parser.ExpressionIdent) error {
	v.ident = ident

	return nil
}

func (v *identGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return nil
}

func (v *identGen) Result() (*parser.ExpressionIdent, error) {
	if v.ident == nil {
		return nil, ErrExpressionNotValidIdent
	}

	return v.ident, nil
}
