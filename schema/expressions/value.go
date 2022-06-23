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
	Ctx    *Ctx    `| @@`
	Ident  *Ident  `| @@`
}

type OperandPart struct {
	node.Node

	Value      string
	Resolvable bool
	Model      string // Modelised representation of fragment
	Parent     *OperandPart
	Type       string
}

type OperandResolution struct {
	Parts []OperandPart
}

func (res *OperandResolution) LastFragment() *OperandPart {
	if len(res.Parts) < 1 {
		return nil
	}

	return &res.Parts[len(res.Parts)-1]
}

func (res *OperandResolution) UnresolvedFragments() []OperandPart {
	unresolvable := []OperandPart{}
	parts := res.Parts

	for _, part := range parts {
		if !part.Resolvable {
			unresolvable = append(unresolvable, part)
		}
	}

	return unresolvable
}

func (a *OperandResolution) TypesMatch(b *OperandResolution) bool {
	if a == nil || b == nil {
		return false
	}

	lhs := a.LastFragment()
	rhs := b.LastFragment()

	return lhs.Type != rhs.Type
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

func (o *Operand) ToString() string {
	if o == nil {
		return ""
	}

	switch o.Type() {
	case "Number":
		return fmt.Sprintf("%d", &o.Number)
	case "Text":
		return *o.String
	case "Null":
		return "null"
	case "False":
		return "false"
	case "True":
		return "true"
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
	case "Ctx":
		return o.Ctx.Token
	default:
		return ""
	}
}

func (o *Operand) Type() string {
	switch {
	case o.Number != nil:
		return "Number"
	case o.String != nil:
		return "Text"
	case o.Null:
		return "Null"
	case o.False:
		return "False"
	case o.True:
		return "True"
	case o.Array != nil:
		return "Array"
	case o.Ident != nil && len(o.Ident.Fragments) > 0:
		return "Ident"
	case o.Ctx != nil:
		return "Ctx"
	default:
		return ""
	}
}

func (o *Operand) IsValueType() (bool, string) {
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
	case o.Ctx != nil:
		return false, o.ToString()
	default:
		return true, o.ToString()
	}
}
