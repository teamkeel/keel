package resolve

import (
	"github.com/google/cel-go/common/operators"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions/visitor"
	"github.com/teamkeel/keel/schema/parser"
)

// FieldLookups retrieves groups of ident lookups using equals comparison which could apply as a filter
func FieldLookups(model *parser.ModelNode, expression string) ([][]Ident, error) {
	ident, err := visitor.RunCelVisitor(expression, fieldLookups(model))
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func fieldLookups(model *parser.ModelNode) visitor.Visitor[[][]Ident] {
	return &fieldLookupsGen{
		uniqueLookupGroups: [][]Ident{},
		current:            0,
		modelName:          model.Name.Value,
	}
}

var _ visitor.Visitor[[][]Ident] = new(fieldLookupsGen)

type fieldLookupsGen struct {
	uniqueLookupGroups [][]Ident
	operands           []Ident
	operator           string
	current            int
	modelName          string
}

func (v *fieldLookupsGen) StartCondition(parenthesis bool) error {
	return nil
}

func (v *fieldLookupsGen) EndCondition(parenthesis bool) error {
	if v.operator == operators.Equals {
		if v.operands != nil {
			if len(v.uniqueLookupGroups) == 0 {
				v.uniqueLookupGroups = make([][]Ident, 1)
			}

			v.uniqueLookupGroups[v.current] = append(v.uniqueLookupGroups[v.current], v.operands...)
		}
	}

	v.operands = nil
	v.operator = ""

	return nil
}
func (v *fieldLookupsGen) VisitAnd() error {
	return nil
}

func (v *fieldLookupsGen) VisitOr() error {
	v.uniqueLookupGroups = append(v.uniqueLookupGroups, []Ident{})

	v.current++
	return nil
}

func (v *fieldLookupsGen) VisitOperator(op string) error {
	v.operator = op
	return nil
}

func (v *fieldLookupsGen) VisitLiteral(value any) error {
	return nil
}

func (v *fieldLookupsGen) VisitVariable(name string) error {
	return nil
}

func (v *fieldLookupsGen) VisitField(fragments []string) error {
	if fragments[0] == strcase.ToLowerCamel(v.modelName) {
		v.operands = append(v.operands, fragments)
	}
	return nil
}

func (v *fieldLookupsGen) VisitIdentArray(fragments [][]string) error {
	return nil
}

func (v *fieldLookupsGen) ModelName() string {
	return v.modelName
}

func (v *fieldLookupsGen) Result() ([][]Ident, error) {
	return v.uniqueLookupGroups, nil
}
