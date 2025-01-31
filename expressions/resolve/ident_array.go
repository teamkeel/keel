package resolve

import (
	"errors"
	"reflect"

	"github.com/teamkeel/keel/schema/parser"
)

var ErrExpressionNotValidIdentArray = errors.New("expression is not an ident array")

// AsIdentArray expects and retrieves the array of idents
func AsIdentArray(expression *parser.Expression) ([]*parser.ExpressionIdent, error) {
	ident, err := RunCelVisitor(expression, identArray())
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func identArray() Visitor[[]*parser.ExpressionIdent] {
	return &identArrayGen{}
}

var _ Visitor[[]*parser.ExpressionIdent] = new(identArrayGen)

type identArrayGen struct {
	idents []*parser.ExpressionIdent
}

func (v *identArrayGen) StartTerm(parenthesis bool) error {
	return nil
}

func (v *identArrayGen) EndTerm(parenthesis bool) error {
	return nil
}

func (v *identArrayGen) StartFunction(name string) error {
	return nil
}

func (v *identArrayGen) EndFunction() error {
	return nil
}

func (v *identArrayGen) VisitAnd() error {
	return ErrExpressionNotValidIdentArray
}

func (v *identArrayGen) VisitOr() error {
	return ErrExpressionNotValidIdentArray
}

func (v *identArrayGen) VisitNot() error {
	return nil
}

func (v *identArrayGen) VisitOperator(op string) error {
	return ErrExpressionNotValidIdentArray
}

func (v *identArrayGen) VisitLiteral(value any) error {
	// Check if the array is empty
	if t := reflect.TypeOf(value); t.Kind() == reflect.Slice && reflect.ValueOf(value).Len() == 0 {
		v.idents = []*parser.ExpressionIdent{}
	} else {
		return ErrExpressionNotValidIdentArray
	}
	return nil
}

func (v *identArrayGen) VisitIdent(ident *parser.ExpressionIdent) error {
	return ErrExpressionNotValidIdentArray
}

func (v *identArrayGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	if v.idents != nil {
		return ErrExpressionNotValidIdentArray
	}

	v.idents = idents
	return nil
}

func (v *identArrayGen) Result() ([]*parser.ExpressionIdent, error) {
	return v.idents, nil
}
