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
	case TypeNumber:
		return fmt.Sprintf("%d", *o.Number)
	case TypeText:
		return *o.String
	case TypeNull:
		return "null"
	case TypeBoolean:
		if o.False {
			return "false"
		} else {
			return "true"
		}
	case TypeArray:
		r := "["
		for i, el := range o.Array.Values {
			if i > 0 {
				r += ", "
			}
			r += el.ToString()
		}
		return r + "]"
	case TypeIdent:
		return o.Ident.ToString()
	default:
		return ""
	}
}

var (
	// These intentionally match the parser field types
	// TODO: maybe refactor so we can use the same constants
	//       refactoring required as the parser depends on this package
	//       so this package can't depend on the parser consts
	TypeNumber  = "Number"
	TypeText    = "Text"
	TypeBoolean = "Boolean"

	// These are unique to expressions
	TypeNull  = "Null"
	TypeArray = "Array"
	TypeIdent = "Ident"
)

func (o *Operand) Type() string {
	switch {
	case o.Number != nil:
		return TypeNumber
	case o.String != nil:
		return TypeText
	case o.Null:
		return TypeNull
	case o.False:
		return TypeBoolean
	case o.True:
		return TypeBoolean
	case o.Array != nil:
		return TypeArray
	case o.Ident != nil && len(o.Ident.Fragments) > 0:
		return TypeIdent
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
		allLiterals := true

		for _, item := range o.Array.Values {
			if ok, _ := item.IsLiteralType(); ok {
				continue
			}

			allLiterals = false
		}

		return allLiterals, o.ToString()
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
