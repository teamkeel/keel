package resolve

import (
	"errors"

	"github.com/teamkeel/keel/expressions/visitor"
)

// AsIdent expects and retrieves a single ident operand in an expression
func AsIdentArray(expression string) ([][]string, error) {
	ident, err := visitor.RunCelVisitor(expression, identArray())
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func identArray() visitor.Visitor[[][]string] {
	return &identArrayGen{}
}

var _ visitor.Visitor[[][]string] = new(identArrayGen)

type identArrayGen struct {
	ident [][]string
}

func (v *identArrayGen) StartCondition(parenthesis bool) error {
	return nil
}

func (v *identArrayGen) EndCondition(parenthesis bool) error {
	return nil
}

func (v *identArrayGen) VisitAnd() error {
	return errors.New("expression with operators cannot be resolved to a single ident")
}

func (v *identArrayGen) VisitOr() error {
	return errors.New("expression with operators cannot be resolved to a single ident")
}

func (v *identArrayGen) VisitOperator(op string) error {
	return errors.New("expression with operators cannot be resolved to a single ident")
}

func (v *identArrayGen) VisitLiteral(value any) error {
	return errors.New("expression with literals cannot be resolved to a single ident")
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

func (v *identArrayGen) Result() [][]string {
	return v.ident
}
