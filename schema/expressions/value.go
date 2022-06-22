package expressions

import (
	"fmt"

	"github.com/teamkeel/keel/schema/node"
)

type Operand struct {
	node.Node

	Number *int64  `  @Int`
	String *string `| @String`
	Null   bool    `| @"null"`
	True   bool    `| @"true"`
	False  bool    `| @"false"`
	Array  *Array  `| @@`
	Ident  *Ident  `| @@`
}

func (o *Operand) ToString() string {
	if o == nil {
		return ""
	}

	switch o.Type() {
	case "Number":
		return fmt.Sprintf("%d", *o.Number)
	case "String":
		return *o.String
	case "Null":
		return "null"
	case "Boolean":
		if o.False {
			return "false"
		} else {
			return "true"
		}
	case "Array":
		r := "["
		for i, el := range o.Array.Values {
			if i > 0 {
				r += ", "
			}
			r += el.ToString()
		}
		return r + "]"
	case "Ident":
		return o.Ident.ToString()
	default:
		return ""
	}
}

var (
	TypeNumber  = "Number"
	TypeString  = "String"
	TypeText    = "Text"
	TypeNull    = "Null"
	TypeBoolean = "Boolean"
	TypeArray   = "Array"
	TypeIdent   = "Ident"
)

func (o *Operand) Type() string {
	switch {
	case o.Number != nil:
		return "Number"
	case o.String != nil:
		return "String"
	case o.Null:
		return "Null"
	case o.False:
		return "Boolean"
	case o.True:
		return "Boolean"
	case o.Array != nil:
		return "Array"
	case o.Ident != nil && len(o.Ident.Fragments) > 0:
		return "Ident"
	default:
		return ""
	}
}

func (o *Operand) IsLiteralType() (bool, string) {
	switch {
	case o.Number != nil:
		return true, o.ToString()
	case o.String != nil:
		return true, o.ToString()
	case o.Null:
		return true, o.ToString()
	case o.False:
		return true, o.ToString()
	case o.True:
		return true, o.ToString()
	case o.Array != nil:
		return false, o.ToString() // todo: arrays containing idents?
	case o.Ident != nil && len(o.Ident.Fragments) > 0:
		return false, o.ToString()
	default:
		return true, o.ToString()
	}
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
