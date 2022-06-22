package relationships

import (
	"github.com/teamkeel/keel/schema/node"
)

// Represents one fragment of a relationship
// e.g in the expression operand post.author.name
// each fragment is separated by dots
type RelationshipFragment struct {
	node.Node

	Current    string
	Type       string
	Resolvable bool
	Parent     string
}

type Relationships struct {
	Fragments []RelationshipFragment
}

func (t *Relationships) LastFragment() *RelationshipFragment {
	return &t.Fragments[len(t.Fragments)-1]
}

func (t *Relationships) UnresolvedFragment() *RelationshipFragment {
	for _, item := range t.Fragments {
		if item.Resolvable {
			continue
		}

		return &item
	}

	return nil
}

var (
	TypeModel   = "model"
	TypeInvalid = "not resolvable"
)

// Given an operand of a condition, tries to resolve the relationships defined within the operand
// e.g if the operand is of type "Ident", and the ident is post.author.name
// then the method will return a Relationships representing each fragment in post.author.name
// along with an error if it hasn't been able to resolve the full path.
// func TryResolveIdent(asts []*parser.AST, operand *expressions.Operand) (*Relationships, error) {

// }
