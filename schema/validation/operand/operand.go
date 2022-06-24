package operand

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/util/str"
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
func ResolveOperand(asts []*parser.AST, operand *expressions.Operand, scope *ExpressionScope) (entity *ExpressionScopeEntity, err error) {
	if ok, _ := operand.IsLiteralType(); ok {
		entity = &ExpressionScopeEntity{
			Literal: operand,
		}
		return entity, nil
	}

fragments:
	for _, fragment := range operand.Ident.Fragments {

		for _, e := range scope.Entities {
			switch {
			case e.Model != nil && e.Model.Name.Value == str.AsTitle(str.Singularize(fragment.Fragment)):
				entity = e

				scope = scopeFromModel(scope, e.Model)

				continue fragments
			case e.Field != nil && e.Field.Name.Value == fragment.Fragment:
				entity = e

				model := query.Model(asts, e.Field.Type)

				if model == nil {
					scope = &ExpressionScope{}
				} else {

					scope = scopeFromModel(scope, model)
				}
				continue fragments
			case e.Object != nil && e.Object.Name == fragment.Fragment:
				entity = e

				scope = scopeFromObject(scope, e.Object)

				continue fragments
			}

		}

		// entity in this case is the last resolved parent
		if entity == nil {
			// unresolvable root model

			inScope := []string{}

			for _, entity := range scope.Entities {
				if entity.Model != nil {
					inScope = append(inScope, strcase.ToLowerCamel(entity.Model.Name.Value))
				}

				if entity.Object != nil {
					inScope = append(inScope, entity.Object.Name)
				}

				// todo: input parameters + genericize
			}

			hint := errorhandling.NewCorrectionHint(inScope, fragment.Fragment)
			err = errorhandling.NewValidationError(
				errorhandling.ErrorUnresolvedRootModel,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Root":        fragment.Fragment,
						"Suggestions": hint.ToString(),
					},
				},
				fragment,
			)
		} else if entity.Model != nil {
			fieldNames := query.ModelFieldNames(entity.Model)
			suggestions := errorhandling.NewCorrectionHint(fieldNames, fragment.Fragment)
			err = errorhandling.NewValidationError(
				errorhandling.ErrorUnresolvableExpression,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Fragment":   fragment.Fragment,
						"Parent":     entity.Model.Name.Value,
						"Suggestion": suggestions.ToString(),
					},
				},
				fragment,
			)
		} else if entity.Object != nil {
			fieldNames := []string{}

			for key := range entity.Object.Fields {
				fieldNames = append(fieldNames, key)
			}
			suggestions := errorhandling.NewCorrectionHint(fieldNames, fragment.Fragment)
			err = errorhandling.NewValidationError(
				errorhandling.ErrorUnresolvableExpression,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Fragment":   fragment.Fragment,
						"Parent":     entity.Object.Name,
						"Suggestion": suggestions.ToString(),
					},
				},
				fragment,
			)
		} else if entity.Field != nil {
			parentModel := query.Model(asts, entity.Field.Type)
			fieldNames := query.ModelFieldNames(parentModel)
			suggestions := errorhandling.NewCorrectionHint(fieldNames, fragment.Fragment)
			err = errorhandling.NewValidationError(
				errorhandling.ErrorUnresolvableExpression,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Fragment":   fragment.Fragment,
						"Parent":     entity.Field.Type,
						"Suggestion": suggestions.ToString(),
					},
				},
				fragment,
			)
		}

		return nil, err
	}

	return entity, nil
}
