package resolve

import (
	"errors"

	"github.com/teamkeel/keel/expressions/visitor"
)

var ErrExpressionNotValidIdentArray = errors.New("expression is not an ident array")

// AsIdentArray expects and retrieves an array of idents
func AsIdentArray(expression string) ([]Ident, error) {
	ident, err := visitor.RunCelVisitor(expression, identArray())
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func identArray() visitor.Visitor[[]Ident] {
	return &identArrayGen{}
}

var _ visitor.Visitor[[]Ident] = new(identArrayGen)

type identArrayGen struct {
	ident []Ident
}

func (v *identArrayGen) StartCondition(parenthesis bool) error {
	return nil
}

func (v *identArrayGen) EndCondition(parenthesis bool) error {
	return nil
}

func (v *identArrayGen) VisitAnd() error {
	return ErrExpressionNotValidIdentArray
}

func (v *identArrayGen) VisitOr() error {
	return ErrExpressionNotValidIdentArray
}

func (v *identArrayGen) VisitOperator(op string) error {
	return ErrExpressionNotValidIdentArray
}

func (v *identArrayGen) VisitLiteral(value any) error {
	return ErrExpressionNotValidIdentArray
}

func (v *identArrayGen) VisitVariable(name string) error {
	v.ident = append(v.ident, []string{name})

	return nil
}

func (v *identArrayGen) VisitField(fragments []string) error {
	v.ident = append(v.ident, fragments)

	return nil
}

func (v *identArrayGen) ModelName() string {
	return ""
}

func (v *identArrayGen) Result() ([]Ident, error) {
	return v.ident, nil
}