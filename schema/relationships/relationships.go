package relationships

import (
	"fmt"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/util/str"
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
	TypeInvalid = "invalid"
)

// Given an operand of a condition, tries to resolve the relationships defined within the operand
// e.g if the operand is of type "Ident", and the ident is post.author.name
// then the method will return a Relationships representing each fragment in post.author.name
// along with an error if it hasn't been able to resolve the full path.
func TryResolveOperand(asts []*parser.AST, operand *expressions.Operand) (*Relationships, error) {
	// If the operand is of a different type (e.g string, bool etc),
	// then return early.
	relationships := Relationships{}

	if operand.Ident == nil {
		relationships.Fragments = append(relationships.Fragments, RelationshipFragment{
			Current: operand.ToString(),
			Type:    operand.Type(),
		})

		return &relationships, nil
	}

	ident := operand.Ident

	var walk func(previousModel *parser.ModelNode, idx int) (*Relationships, error)

	walk = func(previousModel *parser.ModelNode, idx int) (*Relationships, error) {
		// If we are at the first index passed to this method,
		// add the parent model to the fragment tree
		if idx == 1 {
			relationships.Fragments = append(relationships.Fragments,
				RelationshipFragment{
					Current:    previousModel.Name.Value,
					Resolvable: true,
					Node:       previousModel.Node,
					Type:       TypeModel,
				},
			)
		}

		lookupField := ident.Fragments[idx].Fragment

		field := query.ModelField(previousModel, lookupField)

		if field == nil {
			relationships.Fragments = append(relationships.Fragments,
				RelationshipFragment{
					Node:       ident.Fragments[idx].Node,
					Resolvable: false,
					Current:    ident.Fragments[idx].Fragment,
					Parent:     previousModel.Name.Value,
					Type:       TypeInvalid,
				},
			)

			return &relationships, fmt.Errorf("could not find field %s", lookupField)
		}

		relationships.Fragments = append(relationships.Fragments, RelationshipFragment{
			Node:       ident.Fragments[idx].Node,
			Resolvable: true,
			Current:    ident.Fragments[idx].Fragment,
			Parent:     previousModel.Name.Value,
			Type:       field.Type,
		})

		if idx < len(ident.Fragments)-1 {
			nextModel := query.Model(asts, field.Type)
			return walk(nextModel, idx+1)
		}

		// relationship path has been fully resolved
		return &relationships, nil
	}

	lookupModel := str.AsTitle(ident.Fragments[0].Fragment)
	rootModel := query.Model(asts, lookupModel)

	if rootModel == nil {
		relationships.Fragments = append(relationships.Fragments, RelationshipFragment{
			Node:       ident.Node,
			Resolvable: false,
			Current:    ident.Fragments[0].Fragment,
			Type:       TypeInvalid,
		})

		return &relationships, fmt.Errorf("could not find model %s", lookupModel)
	}

	// Start at index 1 so we can look backwards to index 0 for the parent
	if len(ident.Fragments) > 1 {
		return walk(rootModel, 1)
	}

	panic("no fragments")
}
