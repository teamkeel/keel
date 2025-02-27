package resolve

import (
	"github.com/google/cel-go/common/operators"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/parser"
)

// FieldLookups retrieves groups of ident lookups using equals comparison which could apply as a filter
func FieldLookups(model *parser.ModelNode, expression *parser.Expression) ([][]*parser.ExpressionIdent, error) {
	ident, err := RunCelVisitor(expression, fieldLookups(model))
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func fieldLookups(model *parser.ModelNode) Visitor[[][]*parser.ExpressionIdent] {
	return &fieldLookupsGen{
		uniqueLookupGroups: [][]*parser.ExpressionIdent{},
		current:            0,
		modelName:          model.Name.Value,
		anyNull:            false,
	}
}

var _ Visitor[[][]*parser.ExpressionIdent] = new(fieldLookupsGen)

type fieldLookupsGen struct {
	uniqueLookupGroups [][]*parser.ExpressionIdent
	operands           []*parser.ExpressionIdent
	operator           string
	current            int
	modelName          string
	anyNull            bool
}

func (v *fieldLookupsGen) StartTerm(parenthesis bool) error {
	return nil
}

func (v *fieldLookupsGen) EndTerm(parenthesis bool) error {
	if v.operator == operators.Equals && !v.anyNull {
		if v.operands != nil {
			if len(v.uniqueLookupGroups) == 0 {
				v.uniqueLookupGroups = make([][]*parser.ExpressionIdent, 1)
			}

			v.uniqueLookupGroups[v.current] = append(v.uniqueLookupGroups[v.current], v.operands...)
		}
	}

	v.operands = nil
	v.operator = ""
	v.anyNull = false

	return nil
}

func (v *fieldLookupsGen) StartFunction(name string) error {
	return nil
}

func (v *fieldLookupsGen) EndFunction() error {
	return nil
}

func (v *fieldLookupsGen) VisitAnd() error {
	return nil
}

func (v *fieldLookupsGen) VisitOr() error {
	v.uniqueLookupGroups = append(v.uniqueLookupGroups, []*parser.ExpressionIdent{})

	v.current++
	return nil
}

func (v *fieldLookupsGen) VisitNot() error {
	return nil
}

func (v *fieldLookupsGen) VisitOperator(op string) error {
	v.operator = op
	return nil
}

func (v *fieldLookupsGen) VisitLiteral(value any) error {
	if value == nil {
		v.anyNull = true
	}
	return nil
}

func (v *fieldLookupsGen) VisitIdent(ident *parser.ExpressionIdent) error {
	if ident.Fragments[0] == strcase.ToLowerCamel(v.modelName) {
		if len(ident.Fragments) == 1 {
			ident.Fragments = append(ident.Fragments, "id")
		}
		v.operands = append(v.operands, ident)
	}
	return nil
}

func (v *fieldLookupsGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return nil
}

func (v *fieldLookupsGen) ModelName() string {
	return v.modelName
}

func (v *fieldLookupsGen) Result() ([][]*parser.ExpressionIdent, error) {
	return v.uniqueLookupGroups, nil
}
