package parser

import (
	"fmt"

	"github.com/teamkeel/keel/schema/node"
)

type Operand struct {
	node.Node

	Number *int64  `  @('-'? Int)`
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

func (ident *Ident) LastFragment() string {
	return ident.Fragments[len(ident.Fragments)-1].Fragment
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

func (ident *Ident) IsContext() bool {
	return ident != nil && ident.Fragments[0].Fragment == "ctx"
}

func (ident *Ident) IsContextIdentity() bool {
	if !ident.IsContext() {
		return false
	}

	if len(ident.Fragments) > 1 && ident.Fragments[1].Fragment == "identity" {
		return true
	}

	return false
}

func (ident *Ident) IsContextIdentityId() bool {
	if !ident.IsContextIdentity() {
		return false
	}

	if len(ident.Fragments) == 2 {
		return true
	}

	if len(ident.Fragments) == 3 && ident.Fragments[2].Fragment == "id" {
		return true
	}

	return false
}

func (ident *Ident) IsContextIsAuthenticatedField() bool {
	if ident.IsContext() && len(ident.Fragments) == 2 {
		return ident.Fragments[1].Fragment == "isAuthenticated"
	}
	return false
}

func (ident *Ident) IsContextNowField() bool {
	if ident.IsContext() && len(ident.Fragments) == 2 {
		return ident.Fragments[1].Fragment == "now"
	}
	return false
}

func (ident *Ident) IsContextHeadersField() bool {
	if ident.IsContext() && len(ident.Fragments) == 3 {
		return ident.Fragments[1].Fragment == "headers"
	}
	return false
}

func (ident *Ident) IsContextEnvField() bool {
	if ident.IsContext() && len(ident.Fragments) == 3 {
		return ident.Fragments[1].Fragment == "env"
	}
	return false
}

func (ident *Ident) IsContextSecretField() bool {
	if ident.IsContext() && len(ident.Fragments) == 3 {
		return ident.Fragments[1].Fragment == "secrets"
	}
	return false
}

type IdentFragment struct {
	node.Node

	Fragment string `@Ident`
}

type Array struct {
	node.Node

	Values []*Operand `"[" @@ ( "," @@ )* "]"`
}
