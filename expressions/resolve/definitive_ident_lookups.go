package resolve

import (
	"github.com/teamkeel/keel/expressions/visitor"
)

// DefinitiveLookups retrieves all the ident lookups using equals comparison which are certain to apply as a filter
func DefinitiveLookups(expression string) ([][]string, error) {
	ident, err := visitor.RunCelVisitor(expression, definitiveLookups())
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func definitiveLookups() visitor.Visitor[[][]string] {
	return &operandsResolver{}
}

var _ visitor.Visitor[[][]string] = new(definitiveLookupsGen)

type definitiveLookupsGen struct {
	idents [][]string
}

func (v *definitiveLookupsGen) StartCondition(parenthesis bool) error {
	return nil
}

func (v *definitiveLookupsGen) EndCondition(parenthesis bool) error {
	return nil
}

func (v *definitiveLookupsGen) VisitAnd() error {
	return nil
}

func (v *definitiveLookupsGen) VisitOr() error {
	return nil
}

func (v *definitiveLookupsGen) VisitOperator(op string) error {
	return nil
}

func (v *definitiveLookupsGen) VisitLiteral(value any) error {
	return nil
}

func (v *definitiveLookupsGen) VisitVariable(name string) error {
	v.idents = append(v.idents, []string{name})

	return nil
}

func (v *definitiveLookupsGen) VisitField(fragments []string) error {
	v.idents = append(v.idents, fragments)

	return nil
}

func (v *definitiveLookupsGen) ModelName() string {
	return ""
}

func (v *definitiveLookupsGen) Result() ([][]string, error) {
	return v.idents, nil
}
