package associations

import (
	"fmt"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/util/str"
)

type AssociationFragment struct {
	node.Node

	Current    string
	Resolvable bool
	Parent     string
}

type AssociationTree struct {
	Fragments []AssociationFragment
}

func (t *AssociationTree) UnresolvedFragment() (frag *AssociationFragment) {
	for _, item := range t.Fragments {
		if item.Resolvable {
			continue
		}

		return &item
	}

	return nil
}

func TryResolveOperand(asts []*parser.AST, operand *expressions.Operand) (*AssociationTree, error) {
	// If the operand is of a different type (e.g string, bool etc),
	// then return early.
	if operand.Ident == nil {
		return nil, nil
	}

	ident := operand.Ident

	tree := AssociationTree{}
	var walk func(previousModel *parser.ModelNode, idx int) (*AssociationTree, error)

	walk = func(previousModel *parser.ModelNode, idx int) (*AssociationTree, error) {
		// If we are at the first index passed to this method,
		// add the parent model to the fragment tree
		if idx == 1 {
			tree.Fragments = append(tree.Fragments,
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
			tree.Fragments = append(tree.Fragments,
				AssociationFragment{
					Node:       ident.Fragments[idx].Node,
					Resolvable: false,
					Current:    ident.Fragments[idx].Fragment,
					Parent:     previousModel.Name.Value,
				},
			)

			return &tree, fmt.Errorf("could not find field %s", lookupField)
		}

		tree.Fragments = append(tree.Fragments, AssociationFragment{
			Node:       ident.Fragments[idx].Node,
			Resolvable: true,
			Current:    ident.Fragments[idx].Fragment,
			Parent:     previousModel.Name.Value,
		})

		if idx < len(ident.Fragments)-1 {
			nextModel := query.ModelForAssociationField(asts, field)
			return walk(nextModel, idx+1)
		} else {
			// Tree has been fully resolved
			return &tree, nil
		}
	}

	lookupModel := str.AsTitle(ident.Fragments[0].Fragment)
	rootModel := query.Model(asts, lookupModel)

	if rootModel == nil {
		tree.Fragments = append(tree.Fragments, AssociationFragment{
			Node:       ident.Node,
			Resolvable: false,
			Current:    ident.Fragments[0].Fragment,
		})

		return &tree, fmt.Errorf("could not find model %s", lookupModel)
	}

	// Start at index 1 so we can look backwards to index 0 for the parent
	if len(ident.Fragments) > 1 {
		return walk(rootModel, 1)
	}

	panic("no fragments")
}
