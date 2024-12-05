package resolve

import (
	"github.com/google/cel-go/common/operators"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions/visitor"
	"github.com/teamkeel/keel/schema/parser"
)

// FieldLookups retrieves all the ident lookups using equals comparison which are certain to apply as a filter
func FieldLookups(model *parser.ModelNode, expression string) ([][]string, error) {
	ident, err := visitor.RunCelVisitor(expression, fieldLookups(model))
	if err != nil {
		return nil, err
	}

	return ident, nil
}

func fieldLookups(model *parser.ModelNode) visitor.Visitor[[][]string] {
	return &fieldLookupsGen{
		hasOr:     false,
		modelName: model.Name.Value,
	}
}

var _ visitor.Visitor[[][]string] = new(fieldLookupsGen)

type fieldLookupsGen struct {
	idents    [][]string
	operands  [][]string
	operator  string
	hasOr     bool
	modelName string
}

func (v *fieldLookupsGen) StartCondition(parenthesis bool) error {
	return nil
}

func (v *fieldLookupsGen) EndCondition(parenthesis bool) error {
	if v.operator == operators.Equals {
		if v.operands != nil {
			v.idents = append(v.idents, v.operands...)
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
	v.hasOr = true
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

func (v *fieldLookupsGen) ModelName() string {
	return v.modelName
}

func (v *fieldLookupsGen) Result() ([][]string, error) {
	if v.hasOr {
		return nil, nil
	}
	return v.idents, nil
}
