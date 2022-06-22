package expressions

import (
	"fmt"

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

type Number struct {
	node.Node

	Value int64
}

type String struct {
	node.Node

	Value string
}

type Ctx struct {
	node.Node

	Token string `@"ctx" @"." @Ident`
}

type Ident struct {
	node.Node

	Fragments []*IdentFragment `( @@ ( "." @@ )* )`
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

func (v *Operand) ToString() string {
	if v == nil {
		return ""
	}

	switch v.Type() {
	case "Number":
		return fmt.Sprintf("%d", &v.Number)
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
