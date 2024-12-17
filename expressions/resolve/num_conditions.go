package resolve

import (
	"github.com/teamkeel/keel/schema/parser"
)

// AsIdent expects and retrieves a single ident operand in an expression
func NumConditions(expression *parser.Expression) (int, error) {
	ident, err := RunCelVisitor(expression, numCond())
	if err != nil {
		return 0, err
	}

	return ident, nil
}

func numCond() Visitor[int] {
	return &numConditions{}
}

var _ Visitor[int] = new(numConditions)

type numConditions struct {
	count int
}

func (v *numConditions) StartCondition(parenthesis bool) error {
	if v.count == 0 {
		v.count++
	}
	return nil
}

func (v *numConditions) EndCondition(parenthesis bool) error {
	return nil
}

func (v *numConditions) VisitAnd() error {
	v.count++
	return nil
}

func (v *numConditions) VisitOr() error {
	v.count++
	return nil
}

func (v *numConditions) VisitNot() error {
	return nil
}

func (v *numConditions) VisitOperator(op string) error {
	return nil
}

func (v *numConditions) VisitLiteral(value any) error {
	return nil
}

func (v *numConditions) VisitIdent(ident *parser.ExpressionIdent) error {
	return nil
}

func (v *numConditions) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return nil
}

func (v *numConditions) Result() (int, error) {
	return v.count, nil
}
