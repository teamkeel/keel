package operand

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
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
	Fields []*ExpressionScopeEntity
}

type ExpressionScopeEntity struct {
	Name string

	Object    *ExpressionObjectEntity
	Model     *parser.ModelNode
	Field     *parser.FieldNode
	Literal   *expressions.Operand
	Enum      *parser.EnumNode
	EnumValue *parser.EnumValueNode
	Array     []*ExpressionScopeEntity
	Type      string

	Parent *ExpressionScopeEntity
}

func (e *ExpressionScopeEntity) IsNull() bool {
	return e.Literal != nil && e.Literal.Null
}

func (e *ExpressionScopeEntity) IsOptional() bool {
	return e.Field != nil && e.Field.Optional
}

func (e *ExpressionScopeEntity) IsEnumField() bool {
	return e.Enum != nil
}

func (e *ExpressionScopeEntity) IsEnumValue() bool {
	return e.Parent != nil && e.Parent.Enum != nil && e.EnumValue != nil
}

func (e *ExpressionScopeEntity) GetType() string {
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

	if e.Enum != nil {
		return e.Enum.Name.Value
	}

	if e.EnumValue != nil {
		return e.Parent.Enum.Name.Value
	}

	if e.Array != nil {
		return expressions.TypeArray
	}

	if e.Type != "" {
		return e.Type
	}

	return ""
}

var operatorsForType = map[string][]string{
	parser.FieldTypeText: {
		expressions.OperatorEquals,
		expressions.OperatorNotEquals,
		expressions.OperatorAssignment,
	},
	parser.FieldTypeID: {
		expressions.OperatorEquals,
		expressions.OperatorNotEquals,
		expressions.OperatorAssignment,
	},
	parser.FieldTypeNumber: {
		expressions.OperatorEquals,
		expressions.OperatorNotEquals,
		expressions.OperatorGreaterThan,
		expressions.OperatorGreaterThanOrEqualTo,
		expressions.OperatorLessThan,
		expressions.OperatorLessThanOrEqualTo,
		expressions.OperatorAssignment,
		expressions.OperatorIncrement,
		expressions.OperatorDecrement,
	},
	parser.FieldTypeBoolean: {
		expressions.OperatorAssignment,
		expressions.OperatorEquals,
		expressions.OperatorNotEquals,
	},
	parser.FieldTypeDate: {
		expressions.OperatorEquals,
		expressions.OperatorNotEquals,
		expressions.OperatorGreaterThan,
		expressions.OperatorGreaterThanOrEqualTo,
		expressions.OperatorLessThan,
		expressions.OperatorLessThanOrEqualTo,
		expressions.OperatorAssignment,
	},
	parser.FieldTypeDatetime: {
		expressions.OperatorEquals,
		expressions.OperatorNotEquals,
		expressions.OperatorGreaterThan,
		expressions.OperatorGreaterThanOrEqualTo,
		expressions.OperatorLessThan,
		expressions.OperatorLessThanOrEqualTo,
		expressions.OperatorAssignment,
	},
	expressions.TypeEnum: {
		expressions.OperatorEquals,
		expressions.OperatorNotEquals,
		expressions.OperatorAssignment,
	},
	expressions.TypeArray: {
		expressions.OperatorIn,
		expressions.OperatorNotIn,
	},
}

func (e *ExpressionScopeEntity) AllowedOperators() []string {
	t := e.GetType()

	arrayEntity := e.IsRepeated()

	if e.Model != nil || (e.Field != nil && !arrayEntity) {
		return []string{
			expressions.OperatorEquals,
			expressions.OperatorNotEquals,
			expressions.OperatorAssignment,
		}
	}

	if arrayEntity {
		t = expressions.TypeArray
	}

	if e.IsEnumField() || e.IsEnumValue() {
		t = expressions.TypeEnum
	}

	return operatorsForType[t]
}

func DefaultExpressionScope(asts []*parser.AST) *ExpressionScope {
	entities := []*ExpressionScopeEntity{
		{
			Name: "ctx",
			Object: &ExpressionObjectEntity{
				Name: "Context",
				Fields: []*ExpressionScopeEntity{
					{
						Name:  "identity",
						Model: query.Model(asts, "Identity"),
					},
					{
						Name: "now",
						Type: parser.FieldTypeDatetime,
					},
				},
			},
		},
	}

	for _, enum := range query.Enums(asts) {
		entities = append(entities, &ExpressionScopeEntity{
			Name: enum.Name.Value,
			Enum: enum,
		})
	}

	return &ExpressionScope{
		Entities: entities,
	}
}

// IsRepeated returns true if the entity is a repeated value
// This can be because it is a literal array e.g. [1,2,3]
// or because it's a repeated field or at least one parent
// entity is a repeated field e.g. order.items.product.price
// would be a list of prices (assuming order.items is an
// array of items)
func (e *ExpressionScopeEntity) IsRepeated() bool {
	entity := e
	if len(entity.Array) > 0 {
		return true
	}
	if entity.Field != nil && entity.Field.Repeated {
		return true
	}
	for entity.Parent != nil {
		entity = entity.Parent
		if entity.Field != nil && entity.Field.Repeated {
			return true
		}
	}
	return false
}

func scopeFromModel(parentScope *ExpressionScope, parentEntity *ExpressionScopeEntity, model *parser.ModelNode) *ExpressionScope {
	newEntities := []*ExpressionScopeEntity{}

	for _, field := range query.ModelFields(model) {
		newEntities = append(newEntities, &ExpressionScopeEntity{
			Name:   field.Name.Value,
			Field:  field,
			Parent: parentEntity,
		})
	}

	return &ExpressionScope{
		Entities: newEntities,
		Parent:   parentScope,
	}
}

func scopeFromObject(parentScope *ExpressionScope, parentEntity *ExpressionScopeEntity) *ExpressionScope {
	newEntities := []*ExpressionScopeEntity{}

	for _, entity := range parentEntity.Object.Fields {
		// create a shallow copy by getting the _value_ of entity
		entityCopy := *entity
		// update parent (this does _not_ mutate entity)
		entityCopy.Parent = parentEntity
		// then add a pointer to the _copy_
		newEntities = append(newEntities, &entityCopy)
	}

	return &ExpressionScope{
		Entities: newEntities,
		Parent:   parentScope,
	}
}

func scopeFromEnum(parentScope *ExpressionScope, parentEntity *ExpressionScopeEntity) *ExpressionScope {
	newEntities := []*ExpressionScopeEntity{}

	for _, value := range parentEntity.Enum.Values {
		newEntities = append(newEntities, &ExpressionScopeEntity{
			Name:      value.Name.Value,
			EnumValue: value,
			Parent:    parentEntity,
		})
	}

	return &ExpressionScope{
		Entities: newEntities,
		Parent:   parentScope,
	}
}

// Given an operand of a condition, tries to resolve the relationships defined within the operand
// e.g if the operand is of type "Ident", and the ident is post.author.name
func ResolveOperand(asts []*parser.AST, operand *expressions.Operand, scope *ExpressionScope) (entity *ExpressionScopeEntity, err error) {
	if ok, _ := operand.IsLiteralType(); ok {

		// If it is an array literal then handle differently.
		if operand.Type() == expressions.TypeArray {

			array := []*ExpressionScopeEntity{}

			for _, item := range operand.Array.Values {
				array = append(array,
					&ExpressionScopeEntity{
						Literal: item,
					},
				)
			}

			entity = &ExpressionScopeEntity{
				Array: array,
			}

			return entity, nil
		} else {
			entity = &ExpressionScopeEntity{
				Literal: operand,
			}
			return entity, nil
		}

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
			if e.Name != fragment.Fragment {
				continue
			}

			switch {
			case e.Model != nil:
				scope = scopeFromModel(scope, e, e.Model)
			case e.Field != nil:
				model := query.Model(asts, e.Field.Type)

				if model == nil {
					// try matching the field name to a known enum instead
					enum := query.Enum(asts, e.Field.Type)

					if enum != nil {
						// if we've reached a field that is an enum
						// then we want to return the enum as the resolved
						// scope entity. There will be no further nested entities
						// added to the scope for enum types because you can't compare
						// enum values if you are doing a field comparison
						// e.g  expression: post.enumField.EnumValue == post.anotherEnumField.EnumValue
						// doesnt make sense
						return &ExpressionScopeEntity{
							Enum: enum,
						}, nil

					} else {
						// Did not find the model matching the field
						scope = &ExpressionScope{
							Parent: scope,
						}
					}
				} else {
					scope = scopeFromModel(scope, e, model)
				}
			case e.Object != nil:
				scope = scopeFromObject(scope, e)
			case e.Enum != nil:
				scope = scopeFromEnum(scope, e)
			case e.EnumValue != nil:
				scope = &ExpressionScope{
					Parent: scope,
				}
			case e.Type != "":
				scope = &ExpressionScope{
					Parent: scope,
				}
			}

			entity = e
			continue fragments
		}

		// Suggest the names of all things in scope
		inScope := lo.Map(scope.Entities, func(e *ExpressionScopeEntity, _ int) string {
			return e.Name
		})

		suggestions := errorhandling.NewCorrectionHint(inScope, fragment.Fragment)

		parent := ""
		if entity != nil {
			parent = entity.GetType()
		}

		err = errorhandling.NewValidationError(
			errorhandling.ErrorUnresolvableExpression,
			errorhandling.TemplateLiterals{
				Literals: map[string]string{
					"Fragment":   fragment.Fragment,
					"Parent":     parent,
					"Suggestion": suggestions.ToString(),
				},
			},
			fragment,
		)

		return nil, err
	}

	return entity, nil
}
