package resolve

import (
	"github.com/teamkeel/keel/schema/parser"
)

// IdentOperands retrieves all the ident operands in an expression.
func IdentOperands(expression *parser.Expression) ([]*parser.ExpressionIdent, error) {
	ident, err := RunCelVisitor(expression, operands())
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func operands() Visitor[[]*parser.ExpressionIdent] {
	return &operandsResolver{}
}

var _ Visitor[[]*parser.ExpressionIdent] = new(operandsResolver)

type operandsResolver struct {
	idents []*parser.ExpressionIdent
}

func (v *operandsResolver) StartTerm(parenthesis bool) error {
	return nil
}

func (v *operandsResolver) EndTerm(parenthesis bool) error {
	return nil
}

func (v *operandsResolver) StartFunction(name string) error {
	return nil
}

func (v *operandsResolver) EndFunction() error {
	return nil
}

func (v *operandsResolver) StartArgument(num int) error {
	return nil
}

func (v *operandsResolver) EndArgument() error {
	return nil
}

func (v *operandsResolver) VisitAnd() error {
	return nil
}

func (v *operandsResolver) VisitOr() error {
	return nil
}

func (v *operandsResolver) VisitNot() error {
	return nil
}

func (v *operandsResolver) VisitOperator(op string) error {
	return nil
}

func (v *operandsResolver) VisitLiteral(value any) error {
	return nil
}

func (v *operandsResolver) VisitIdent(ident *parser.ExpressionIdent) error {
	for _, id := range v.idents {
		if id.String() == ident.String() {
			// if the ident is already in the list, we don't need to add it again
			return nil
		}
	}

	v.idents = append(v.idents, ident)

	return nil
}

func (v *operandsResolver) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return nil
}

func (v *operandsResolver) Result() ([]*parser.ExpressionIdent, error) {
	return v.idents, nil
}
