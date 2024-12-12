package resolve

import (
	"github.com/teamkeel/keel/expressions/visitor"
)

// IdentOperands retrieves all the ident operands in an expression as a slice
func IdentOperands(expression string) ([]Ident, error) {
	ident, err := visitor.RunCelVisitor(expression, operands())
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func operands() visitor.Visitor[[]Ident] {
	return &operandsResolver{}
}

var _ visitor.Visitor[[]Ident] = new(operandsResolver)

type operandsResolver struct {
	idents []Ident
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

func (v *operandsResolver) VisitVariable(name string) error {
	v.idents = append(v.idents, []string{name})

	return nil
}

func (v *operandsResolver) VisitField(fragments []string) error {
	v.idents = append(v.idents, fragments)

	return nil
}

func (v *operandsResolver) VisitIdentArray(fragments [][]string) error {
	return nil
}

func (v *operandsResolver) ModelName() string {
	return ""
}

func (v *operandsResolver) Result() ([]Ident, error) {
	return v.idents, nil
}
