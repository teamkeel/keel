package associations

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
type AssociationFragment struct {
	node.Node

	Current    string
	Resolvable bool
	Parent     string
}

type Association struct {
	Fragments []AssociationFragment
}

func (t *Association) UnresolvedFragment() (frag *AssociationFragment) {
	for _, item := range t.Fragments {
		if item.Resolvable {
			continue
		}

		return &item
	}

	return nil
}

// Given an operand of a condition, tries to resolve the associations defined within the operand
// e.g if the operand is of type "Ident", and the ident is post.author.name
// then the method will return a AssociationTree representing each fragment in post.author.name
// along with an error if it hasn't been able to resolve the full path.
func TryResolveOperand(asts []*parser.AST, operand *expressions.Operand) (*Association, error) {
	// If the operand is of a different type (e.g string, bool etc),
	// then return early.
	if operand.Ident == nil {
		return nil, nil
	}

	ident := operand.Ident

	association := Association{}
	var walk func(previousModel *parser.ModelNode, idx int) (*Association, error)

	walk = func(previousModel *parser.ModelNode, idx int) (*Association, error) {
		// If we are at the first index passed to this method,
		// add the parent model to the fragment tree
		if idx == 1 {
			association.Fragments = append(association.Fragments,
				AssociationFragment{
					Current:    previousModel.Name.Value,
					Resolvable: true,
					Node:       previousModel.Node,
				},
			)
		}

		lookupField := ident.Fragments[idx].Fragment

		field := query.ModelField(previousModel, lookupField)

		if field == nil {
			association.Fragments = append(association.Fragments,
				AssociationFragment{
					Node:       ident.Fragments[idx].Node,
					Resolvable: false,
					Current:    ident.Fragments[idx].Fragment,
					Parent:     previousModel.Name.Value,
				},
			)

			return &association, fmt.Errorf("could not find field %s", lookupField)
		}

		association.Fragments = append(association.Fragments, AssociationFragment{
			Node:       ident.Fragments[idx].Node,
			Resolvable: true,
			Current:    ident.Fragments[idx].Fragment,
			Parent:     previousModel.Name.Value,
		})

		if idx < len(ident.Fragments)-1 {
			nextModel := query.ModelForAssociationField(asts, field)
			return walk(nextModel, idx+1)
		} else {
			// association path has been fully resolved
			return &association, nil
		}
	}

	lookupModel := str.AsTitle(ident.Fragments[0].Fragment)
	rootModel := query.Model(asts, lookupModel)

	if rootModel == nil {
		association.Fragments = append(association.Fragments, AssociationFragment{
			Node:       ident.Node,
			Resolvable: false,
			Current:    ident.Fragments[0].Fragment,
		})

		return &association, fmt.Errorf("could not find model %s", lookupModel)
	}

	// Start at index 1 so we can look backwards to index 0 for the parent
	if len(ident.Fragments) > 1 {
		return walk(rootModel, 1)
	}

	panic("no fragments")
}
