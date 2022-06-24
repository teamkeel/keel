package operand

import (
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/util/str"
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
	Object  *ExpressionObjectEntity
	Model   *parser.ModelNode
	Field   *parser.FieldNode
	Literal *expressions.Operand
}

func (e *ExpressionScopeEntity) Type() string {
	if e.Object != nil {
		return e.Object.Name
	}

	if e.Model != nil {
		return e.Model.Name.Value
	}

	if e.Field != nil {
		return e.Field.Type
	}

	if e.Literal != nil {
		return e.Literal.Type()
	}

	return ""
}

func (e *ExpressionScopeEntity) BaseType() string {
	if e.Object != nil {
		return e.Object.Name
	}

	if e.Model != nil {
		return e.Model.Name.Value
	}

	if e.Field != nil {
		if e.Field.Type == "Text" {
			return "String"
		}
		return e.Field.Type
	}

	if e.Literal != nil {
		return e.Literal.Type()
	}

	return ""
}

func (e *ExpressionScopeEntity) AllowedOperators() []string {
	return []string{}
}

func (e *ExpressionScopeEntity) Value() string {
	if e.Object != nil {
		return e.Object.Name
	}

	if e.Model != nil {
		return e.Model.Name.Value
	}

	if e.Field != nil {
		return e.Field.Name.Value
	}

	if e.Literal != nil {
		return e.Literal.ToString()
	}

	return ""
}

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
func ResolveOperand(asts []*parser.AST, operand *expressions.Operand, scope *ExpressionScope) (entity *ExpressionScopeEntity, err error) {
	if ok, _ := operand.IsLiteralType(); ok {
		entity = &ExpressionScopeEntity{
			Literal: operand,
		}
		return entity, nil
	}

	// We want to loop over every fragment in the Ident, each time checking if the Ident matches anything
	// stored in the expression scope.
	// e.g if the first ident fragment is "ctx", and the ExpressionScope has a matching key
	// (which it does if you use the DefaultExpressionScope)
	// then it will continue onto the next fragment, setting the new scope to Ctx
	// so that the next fragment can be compared to fields that exist on the Ctx object
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
			// The first fragment didn't match anything in the scope
			inScope := []string{}

			// Suggest all of the top level things that are in the scope, e.g ctx, {modelName}, any input parameters
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
