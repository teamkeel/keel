package associations

import (
	"fmt"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/util/str"
)

type AssociationPart interface {
	ToString() string
}

type UnresolvedAssociation struct {
	node.Node

	Ident string
}

func (a *UnresolvedAssociation) ToString() string {
	return a.Ident
}

type AssociationTree struct {
	Fragments []AssociationPart
}

func (tree *AssociationTree) ErrorFragment() *UnresolvedAssociation {
	for _, frag := range tree.Fragments {
		if err, ok := frag.(*UnresolvedAssociation); ok {
			return err
		}
	}

	return nil
}

func TryResolveIdent(asts []*parser.AST, ident *expressions.Ident) (AssociationTree, error) {
	tree := AssociationTree{}
	var walk func(previousModel *parser.ModelNode, idx int) (AssociationTree, error)

	walk = func(previousModel *parser.ModelNode, idx int) (AssociationTree, error) {
		if idx == 1 {
			tree.Fragments = append(tree.Fragments, previousModel)
		}
		lookupField := ident.Fragments[idx].Fragment

		field := query.ModelField(previousModel, lookupField)

		if field == nil {
			tree.Fragments = append(tree.Fragments, &UnresolvedAssociation{
				Node:  ident.Fragments[idx].Node,
				Ident: ident.Fragments[idx].Fragment,
			})

			return tree, fmt.Errorf("could not find field %s", lookupField)
		}

		tree.Fragments = append(tree.Fragments, field)

		if idx < len(ident.Fragments)-1 {
			nextModel := query.ModelForAssociationField(asts, field)
			return walk(nextModel, idx+1)
		} else {
			// Tree has been fully resolved
			return tree, nil
		}
	}

	lookupModel := str.AsTitle(str.Singularize(ident.Fragments[0].Fragment))
	rootModel := query.Model(asts, lookupModel)

	if rootModel == nil {
		tree.Fragments = append(tree.Fragments, &UnresolvedAssociation{
			Node:  ident.Fragments[0].Node,
			Ident: ident.Fragments[0].Fragment,
		})

		return tree, fmt.Errorf("could not find model %s", lookupModel)
	}

	// Start at index 1 so we can look backwards to index 0 for the parent
	if len(ident.Fragments) > 1 {
		return walk(rootModel, 1)
	}

	panic("no fragments")
}
