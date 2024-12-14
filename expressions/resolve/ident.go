package resolve

import (
	"errors"

	"github.com/teamkeel/keel/expressions/visitor"
	"github.com/teamkeel/keel/schema/parser"
)

var ErrExpressionNotValidIdent = errors.New("expression is not an ident")

// AsIdent expects and retrieves a single ident operand in an expression
func AsIdent(expression *parser.Expression) (*parser.ExpressionIdent, error) {
	ident, err := visitor.RunCelVisitor(expression, ident())
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func ident() visitor.Visitor[*parser.ExpressionIdent] {
	return &identGen{}
}

var _ visitor.Visitor[*parser.ExpressionIdent] = new(identGen)

type identGen struct {
	ident *parser.ExpressionIdent
}

func (v *identGen) StartCondition(parenthesis bool) error {
	return nil
}

func (v *identGen) EndCondition(parenthesis bool) error {
	return nil
}

func (v *identGen) VisitAnd() error {
	return ErrExpressionNotValidIdent
}

func (v *identGen) VisitOr() error {
	return ErrExpressionNotValidIdent
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
