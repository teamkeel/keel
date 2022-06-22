package expressions

import (
	"fmt"
	"strconv"

	"github.com/teamkeel/keel/schema/node"
)

type Operand struct {
	node.Node

	Number *Number  `  @Int`
	String *String  `| @String`
	Null   *Boolean `| @"null"`
	True   *Boolean `| @"true"`
	False  *Boolean `| @"false"`
	Array  *Array   `| @@`
	Ctx    *Ctx     `| @@`
	Ident  *Ident   `| @@`
}

func (op *Operand) Resolve() (OperandResolution, error) {
	switch {
	case op.Array != nil:
		return op.Array.Resolve()
	case op.Ctx != nil:
		return op.Ctx.Resolve()
	case op.String != nil:
		return op.String.Resolve()
	case op.True != nil:
		return op.True.Resolve()
	case op.False != nil:
		return op.False.Resolve()
	case op.Null != nil:
		return op.Null.Resolve()
	case op.Ident != nil:
		return op.Ident.Resolve()
	case op.Number != nil:
		return op.Null.Resolve()
	default:
		panic("not a known operand type")
	}
}

type OperandPart struct {
	node.Node

	Value      string
	Resolvable bool
	Parent     *OperandPart
	Type       string
}

type OperandResolution struct {
	Parts []OperandPart
}

type Boolean struct {
	node.Node

	Value bool
}

func (b *Boolean) Resolve() (OperandResolution, error) {
	return OperandResolution{
		Parts: []OperandPart{
			{
				Node:       b.Node,
				Value:      strconv.FormatBool(b.Value),
				Resolvable: true,
				Type:       "Boolean",
			},
		},
	}, nil
}

type Number struct {
	node.Node

	Value int64
}

func (n *Number) Resolve() (OperandResolution, error) {
	return OperandResolution{
		Parts: []OperandPart{
			{
				Node:       n.Node,
				Value:      fmt.Sprint(n.Value),
				Resolvable: true,
				Type:       "Number",
			},
		},
	}, nil
}

type String struct {
	node.Node

	Value string
}

func (s *String) Resolve() (OperandResolution, error) {
	return OperandResolution{
		Parts: []OperandPart{
			{
				Node:       s.Node,
				Value:      fmt.Sprint(s.Value),
				Resolvable: true,
				Type:       "String",
			},
		},
	}, nil
}

type Ctx struct {
	node.Node

	Token string `@"ctx" @"." @Ident`
}

func (ctx *Ctx) Resolve() (OperandResolution, error) {
	return OperandResolution{}, nil
}

type Ident struct {
	node.Node

	Fragments []*IdentFragment `( @@ ( "." @@ )* )`
}

func (ident *Ident) Resolve() (OperandResolution, error) {
	return OperandResolution{
		Parts: []OperandPart{},
	}, nil
}

func (ident *Ident) ToString() string {
	ret := ""
	for i, fragment := range ident.Fragments {
		if i == len(ident.Fragments)-1 {
			ret += fragment.Fragment
		} else {
			ret += fmt.Sprintf("%s.", fragment.Fragment)
		}
	}

	return ret
}

type IdentFragment struct {
	node.Node

	Fragment string `@Ident`
}

type Array struct {
	node.Node

	Values []*Operand `"[" @@ ( "," @@ )* "]"`
}

func (ident *Array) Resolve() (OperandResolution, error) {
	return OperandResolution{
		Parts: []OperandPart{},
	}, nil
}

func (v *Operand) ToString() string {
	if v == nil {
		return ""
	}

	switch v.Type() {
	case "Number":
		return fmt.Sprintf("%d", *v.Number)
	case "Text":
		return v.String.Value
	case "Null":
		return "null"
	case "False":
		return "false"
	case "True":
		return "true"
	case "Array":
		r := "["
		for i, el := range v.Array.Values {
			if i > 0 {
				r += ", "
			}
			r += el.ToString()
		}
		return r + "]"
	case "Ident":
		return v.Ident.ToString()
	case "Ctx":
		return v.Ctx.Token
	default:
		return ""
	}
}

func (v *Operand) Type() string {
	switch {
	case v.Number != nil:
		return "Number"
	case v.String != nil:
		return "Text"
	case v.Null != nil:
		return "Null"
	case v.False != nil:
		return "False"
	case v.True != nil:
		return "True"
	case v.Array != nil:
		return "Array"
	case v.Ident != nil && len(v.Ident.Fragments) > 0:
		return "Ident"
	case v.Ctx != nil:
		return "Ctx"
	default:
		return ""
	}
}
