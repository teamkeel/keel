package relationships

import (
	"fmt"

	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/util/str"
)

var (
	TypeModel   = "model"
	TypeInvalid = "not resolvable"
)

// Given an operand of a condition, tries to resolve the relationships defined within the operand
// e.g if the operand is of type "Ident", and the ident is post.author.name
// then the method will return a Relationships representing each fragment in post.author.name
// along with an error if it hasn't been able to resolve the full path.
func TryResolveIdent(asts []*parser.AST, operand *expressions.Operand) (*expressions.OperandResolution, []error) {
	ident := operand.Ident
	errors := []error{}
	res := expressions.OperandResolution{}

	var resolvePart func(idx int) (*expressions.OperandResolution, []error)

	resolvePart = func(idx int) (*expressions.OperandResolution, []error) {
		// if its index 0, then do root model resolution
		if idx == 0 {
			lookupModel := str.AsTitle(ident.Fragments[idx].Fragment)

			rootModel := query.Model(asts, lookupModel)
			resolvableRoot := rootModel != nil

			resolvedType := ""

			if !resolvableRoot {
				errors = append(errors, fmt.Errorf("could not find root model %s", lookupModel))

				resolvedType = TypeInvalid
			} else {
				resolvedType = rootModel.Name.Value
			}

			res.Parts = append(res.Parts, expressions.OperandPart{
				Node:       ident.Fragments[idx].Node,
				Resolvable: resolvableRoot,
				Model:      lookupModel,
				Value:      ident.Fragments[idx].Fragment,
				Type:       resolvedType,
			})
		} else {
			// we're dealing with fields on a parent
			// although the parent may not have been resolved
			parent := res.Parts[idx-1]

			if parent.Resolvable {
				parentModel := query.Model(asts, parent.Model)

				lookupField := ident.Fragments[idx].Fragment

				field := query.ModelField(parentModel, lookupField)

				resolvableField := field != nil

				resolvableType := ""

				if !resolvableField {
					errors = append(errors, fmt.Errorf("unresolvable field %s", lookupField))
					resolvableType = TypeInvalid
				} else {
					resolvableType = field.Type
				}

				res.Parts = append(res.Parts, expressions.OperandPart{
					Node:       ident.Fragments[idx].Node,
					Resolvable: resolvableField,
					Value:      ident.Fragments[idx].Fragment,
					Model:      str.AsTitle(ident.Fragments[idx].Fragment),
					Parent:     &parent,
					Type:       resolvableType,
				})
			} else {
				errors = append(errors, fmt.Errorf("unresolvable field %s", ident.Fragments[idx].Fragment))

				res.Parts = append(res.Parts, expressions.OperandPart{
					Node:       ident.Fragments[idx].Node,
					Resolvable: parent.Resolvable,
					Value:      ident.Fragments[idx].Fragment,
					Model:      str.AsTitle(ident.Fragments[idx].Fragment),
					Parent:     &parent,
					Type:       TypeInvalid,
				})
			}
		}

		// continue resolving the next fragment as we haven't reached the end yet
		if idx < len(ident.Fragments)-1 {
			return resolvePart(idx + 1)
		}

		// relationship path has been fully resolved
		return &res, errors
	}

	// Start at index 1 so we can look backwards to index 0 for the parent
	if len(ident.Fragments) > 0 {
		return resolvePart(0)
	}

	panic("no fragments")
}
