package associations

import (
	"fmt"
	"strings"

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

func (tree *AssociationTree) PrettyPrint() (ret string) {
	for i, item := range tree.Fragments {
		if i == len(tree.Fragments)-1 {
			ret += strings.ToLower(item.ToString())
		} else {
			ret += fmt.Sprintf("%s.", strings.ToLower(item.ToString()))
		}
	}
	return ret
}

// Takes
func TryResolveIdent(asts []*parser.AST, ident *expressions.Ident) (AssociationTree, error) {
	tree := AssociationTree{}
	var walk func(idx int) (AssociationTree, error)

	walk = func(idx int) (AssociationTree, error) {
		fragment := ident.Fragments[idx-1]
		lookupModel := str.AsTitle(str.Singularize(fragment.Fragment))
		model := query.Model(asts, lookupModel)

		if model == nil {
			tree.Fragments = append(tree.Fragments, &UnresolvedAssociation{
				Node:  ident.Fragments[idx].Node,
				Ident: ident.ToString(),
			})

			return tree, fmt.Errorf("could not find model %s", lookupModel)
		}
		if idx == 1 {
			tree.Fragments = append(tree.Fragments, model)
		}
		lookupField := ident.Fragments[idx].Fragment

		field := query.ModelField(model, lookupField)

		if field == nil {
			tree.Fragments = append(tree.Fragments, &UnresolvedAssociation{
				Node:  ident.Fragments[idx].Node,
				Ident: ident.ToString(),
			})

			return tree, fmt.Errorf("could not find field %s", lookupField)
		}

		tree.Fragments = append(tree.Fragments, field)

		if idx < len(ident.Fragments)-1 {
			return walk(idx + 1)
		} else {
			return tree, nil
		}
	}

	// Start at index 1 so we can look backwards to index 0 for the parent
	return walk(1)
}
