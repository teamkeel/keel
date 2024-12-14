package resolve

import (
	"github.com/teamkeel/keel/expressions/visitor"
	"github.com/teamkeel/keel/schema/parser"
)

// IdentOperands retrieves all the ident operands in an expression as a slice
func IdentOperands(expression *parser.Expression) ([]*parser.ExpressionIdent, error) {
	ident, err := visitor.RunCelVisitor(expression, operands())
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func operands() visitor.Visitor[[]*parser.ExpressionIdent] {
	return &operandsResolver{}
}

var _ visitor.Visitor[[]*parser.ExpressionIdent] = new(operandsResolver)

type operandsResolver struct {
	idents []*parser.ExpressionIdent
}

func (v *operandsResolver) StartCondition(parenthesis bool) error {
	return nil
}

func (v *operandsResolver) EndCondition(parenthesis bool) error {
	return nil
}

func (v *operandsResolver) VisitAnd() error {
	return nil
}

func (v *operandsResolver) VisitOr() error {
	return nil
}

func (v *operandsResolver) VisitOperator(op string) error {
	return nil
}

func (v *operandsResolver) VisitLiteral(value any) error {
	return nil
}

func (v *operandsResolver) VisitIdent(ident *parser.ExpressionIdent) error {
	v.idents = append(v.idents, ident)

	return nil
}

func (v *operandsResolver) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return nil
}

func (v *operandsResolver) Result() ([]*parser.ExpressionIdent, error) {
	return v.idents, nil
}
