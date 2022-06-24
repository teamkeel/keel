package relationships

import (
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

var (
	TypeModel   = "model"
	TypeInvalid = "not resolvable"
)

type ExpressionScope struct {
	Parent   *ExpressionScope
	Entities []*ExpressionScopeEntity
}

func (a *ExpressionScope) Merge(b *ExpressionScope) *ExpressionScope {
	return &ExpressionScope{
		Entities: append(a.Entities, b.Entities...),
	}
}

type ExpressionObjectEntity struct {
	Name   string
	Fields map[string]*ExpressionScopeEntity
}

type ExpressionScopeEntity struct {
	Object *ExpressionObjectEntity
	// wrap Model in "ExpressionModelEntity" where key = actual name and model is actual model
	Model *parser.ModelNode
	Field *parser.FieldNode

	Literal *expressions.Operand
}

// Type() -> String
// AllowedOperators() -> []string

// person.firstName == 123

// person.age == ctx.ipAddress

func DefaultExpressionScope(asts []*parser.AST) *ExpressionScope {
	stringLiteral := ""
	return &ExpressionScope{
		Entities: []*ExpressionScopeEntity{
			{
				Object: &ExpressionObjectEntity{
					Name: "ctx",
					Fields: map[string]*ExpressionScopeEntity{
						"identity": {
							Model: query.Model(asts, "Identity"),
						},
						"ipAddress": {
							Literal: &expressions.Operand{
								String: &stringLiteral,
							},
						},
					},
				},
			},
		},
	}
}

func scopeFromModel(parent *ExpressionScope, model *parser.ModelNode) *ExpressionScope {
	newEntities := []*ExpressionScopeEntity{}

	for _, field := range query.ModelFields(model) {
		newEntities = append(newEntities, &ExpressionScopeEntity{
			Field: field,
		})
	}

	return &ExpressionScope{
		Entities: newEntities,
		Parent:   parent,
	}
}

func scopeFromObject(parent *ExpressionScope, obj *ExpressionObjectEntity) *ExpressionScope {
	newEntities := []*ExpressionScopeEntity{}

	for _, field := range obj.Fields {
		newEntities = append(newEntities, field)
	}

	return &ExpressionScope{
		Entities: newEntities,
		Parent:   parent,
	}
}

// Given an operand of a condition, tries to resolve the relationships defined within the operand
// e.g if the operand is of type "Ident", and the ident is post.author.name
// then the method will return a Relationships representing each fragment in post.author.name
// along with an error if it hasn't been able to resolve the full path.
func ResolveOperand(asts []*parser.AST, operand *expressions.Operand, scope *ExpressionScope) (*ExpressionScopeEntity, error) {
	if ok, _ := operand.IsValueType(); ok {
		entity := &ExpressionScopeEntity{
			Literal: operand,
		}
		return entity, nil
	}

	var entity *ExpressionScopeEntity

fragments:
	for _, fragment := range operand.Ident.Fragments {
		for _, e := range scope.Entities {
			switch {
			// todo: casing comparison for below
			case e.Model != nil && e.Model.Name.Value == fragment.Fragment:
				entity = e
				scope = scopeFromModel(scope, e.Model)

				continue fragments
			case e.Field != nil && e.Field.Name.Value == fragment.Fragment:
				// handle field e.g person.hobbies
				entity = e

				model := query.Model(asts, e.Field.Type)

				if model == nil {
					scope = &ExpressionScope{}
				} else {
					scope = scopeFromModel(scope, e.Model)
				}
				continue fragments
			case e.Object != nil && e.Object.Name == fragment.Fragment:
				entity = e

				scope = scopeFromObject(scope, e.Object)

				continue fragments
			default:
				// handle anything after unresolvable "thing"
				// including unknown cxt children or unknown model children
				return nil, nil
			}
		}
	}

	return entity, nil
}
