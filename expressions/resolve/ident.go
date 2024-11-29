package resolve

import (
	"errors"

	"github.com/teamkeel/keel/expressions/visitor"
)

// AsIdent expects and retrieves a single ident operand in an expression
func AsIdent(expression string) ([]string, error) {
	ident, err := visitor.RunCelVisitor(expression, ident())
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func ident() visitor.Visitor[[]string] {
	return &identGen{}
}

var _ visitor.Visitor[[]string] = new(identGen)

type identGen struct {
	ident []string
}

func (v *identGen) StartCondition(parenthesis bool) error {
	return nil
}

func (v *identGen) EndCondition(parenthesis bool) error {
	return nil
}

func (v *identGen) VisitAnd() error {
	return errors.New("expression with operators cannot be resolved to a single ident")
}

func (v *identGen) VisitOr() error {
	return errors.New("expression with operators cannot be resolved to a single ident")
}

func (v *identGen) VisitOperator(op string) error {
	return errors.New("expression with operators cannot be resolved to a single ident")
}

func (v *identGen) VisitLiteral(value any) error {
	return errors.New("expression with literals cannot be resolved to a single ident")
}

func (v *identGen) VisitVariable(name string) error {
	v.ident = []string{name}

	return nil
}

func (v *identGen) VisitField(fragments []string) error {
	v.ident = fragments

	return nil
}

func (v *identGen) ModelName() string {
	return ""
}

func (v *identGen) Result() []string {
	return v.ident
}
