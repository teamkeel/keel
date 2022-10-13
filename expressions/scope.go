package expressions

import (
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// ExpressionContext represents all of the metadata that we need to know about
// to resolve an expression.
// For example, we need to know the parent constructs in the schema such as the
// current model, the current attribute or the current action in order to determine
// what fragments are expected in an expression
type ExpressionContext struct {
	Model     *parser.ModelNode
	Action    *parser.ActionNode
	Attribute *parser.AttributeNode
}

type OperandPosition = string

const (
	OperandPositionLhs OperandPosition = "lhs"
	OperandPositionRhs OperandPosition = "rhs"
)

// ExpressionScope is used to represent things that should be in the scope
// of an expression.
// Operands in an expression are composed of fragments,
// which are dot separated identifiers:
// e.g post.title
// The base scope that is constructed before we start evaluating the first
// fragment contains things like ctx, any input parameters, the current model etc
type ExpressionScope struct {
	Parent   *ExpressionScope
	Entities []*ExpressionScopeEntity
}

func BuildRootExpressionScope(asts []*parser.AST, context ExpressionContext) *ExpressionScope {
	contextualScope := &ExpressionScope{
		Entities: []*ExpressionScopeEntity{
			{
				Name:  strcase.ToLowerCamel(context.Model.Name.Value),
				Model: context.Model,
			},
		},
	}

	if context.Action != nil {
		// todo: this isnt right
		// the scope logic for inputs should be:
		// if lhs, suggest read and write ONLY for @permission expression, otherwise, dont suggest anything
		// if rhs, suggest write inputs only

		for _, input := range context.Action.AllInputs() {
			// inputs using short-hand syntax that refer to relationships
			// don't get added to the scope
			if input.Label == nil && len(input.Type.Fragments) > 1 {
				continue
			}

			resolvedType := query.ResolveInputType(asts, input, context.Model)
			if resolvedType == "" {
				continue
			}
			contextualScope.Entities = append(contextualScope.Entities, &ExpressionScopeEntity{
				Name: input.Name(),
				Type: resolvedType,
			})
		}
	}

	return DefaultExpressionScope(asts).Merge(contextualScope)
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

// An ExpressionScopeEntity is an individual item that is inserted into an
// expression scope. So a scope might have multiple entities of different types in it
// at one single time:
// example:
// &ExpressionScope{Entities: []*ExpressionScopeEntity{{ Name: "ctx": Object: {....} }}, Parent: nil}
// Parent is used to provide useful metadata about any upper scopes (e.g previous fragments that were evaluated)
type ExpressionScopeEntity struct {
	Name string

	Object    *ExpressionObjectEntity
	Model     *parser.ModelNode
	Field     *parser.FieldNode
	Literal   *parser.Operand
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
	return (e.Field != nil && e.Field.Optional) || (e.Enum != nil && e.Enum.Optional)
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
		return parser.TypeArray
	}

	if e.Type != "" {
		return e.Type
	}

	return ""
}

var operatorsForType = map[string][]string{
	parser.FieldTypeText: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorAssignment,
	},
	parser.FieldTypeID: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorAssignment,
	},
	parser.FieldTypeNumber: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorGreaterThan,
		parser.OperatorGreaterThanOrEqualTo,
		parser.OperatorLessThan,
		parser.OperatorLessThanOrEqualTo,
		parser.OperatorAssignment,
		parser.OperatorIncrement,
		parser.OperatorDecrement,
	},
	parser.FieldTypeBoolean: {
		parser.OperatorAssignment,
		parser.OperatorEquals,
		parser.OperatorNotEquals,
	},
	parser.FieldTypeDate: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorGreaterThan,
		parser.OperatorGreaterThanOrEqualTo,
		parser.OperatorLessThan,
		parser.OperatorLessThanOrEqualTo,
		parser.OperatorAssignment,
	},
	parser.FieldTypeDatetime: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorGreaterThan,
		parser.OperatorGreaterThanOrEqualTo,
		parser.OperatorLessThan,
		parser.OperatorLessThanOrEqualTo,
		parser.OperatorAssignment,
	},
	parser.TypeEnum: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorAssignment,
	},
	parser.TypeArray: {
		parser.OperatorIn,
		parser.OperatorNotIn,
	},
	parser.TypeNull: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorAssignment,
	},
}

func (e *ExpressionScopeEntity) AllowedOperators() []string {
	t := e.GetType()

	arrayEntity := e.IsRepeated()

	if e.Model != nil || (e.Field != nil && !arrayEntity) {
		return []string{
			parser.OperatorEquals,
			parser.OperatorNotEquals,
			parser.OperatorAssignment,
		}
	}

	if arrayEntity {
		t = parser.TypeArray
	}

	if e.IsEnumField() || e.IsEnumValue() {
		t = parser.TypeEnum
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
						Name: "isAuthenticated",
						Type: parser.FieldTypeBoolean,
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

// A condition is composed of a LHS operand (and an operator, and a RHS operand if not a value only condition like expression: true)
// Given an operand of a condition, tries to resolve all of the fragments defined within the operand
// an operand might be:
// - post.author.name
// - MyEnum.ValueName
// - "123"
// - true
// All of these types above are checked / attempted to be resolved in this method.
func ResolveOperand(asts []*parser.AST, operand *parser.Operand, scope *ExpressionScope, operandPosition OperandPosition) (entity *ExpressionScopeEntity, err error) {
	// build additional root scopes based on position of operand
	// and also what type attribute the expression is used in.

	// If it is a literal then handle differently.
	if ok, _ := operand.IsLiteralType(); ok {
		if operand.Type() == parser.TypeArray {
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
				// crucially, scope is redefined for the next iteration of the outer loop
				// so that we check the subsequent fragment against the models fields
				// e.g post.field - fragment at idx 0 would be post, so scopeFromModel finds the fields on the
				// Post model and populates the new scope with them.
				scope = scopeFromModel(scope, e, e.Model)
			case e.Field != nil:
				// covers fields which are associations to other models:
				// e.g post.association.associationField
				// repopulates the scope for the third fragment to be the fields of 'association'
				model := query.Model(asts, e.Field.Type)

				// if no model is found for the field type, then we need to check other potential matches
				// e.g enums in the schema
				if model == nil {
					// try matching the field name to a known enum instead
					enum := query.Enum(asts, e.Field.Type)

					if enum != nil {
						// enum definitions aren't optional when they are defined
						// declaratively in the schema, but instead you mark an enum's
						// optionality at field level, so we need to attach it here based on the field's
						// optional value.
						enum.Optional = e.Field.Optional

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
					// move onto the associations' fields so we populate the new scope with them
					scope = scopeFromModel(scope, e, model)
				}
			case e.Object != nil:
				// object is a special wrapper type to describe entities we want
				// to be in scope that aren't models. It is a more flexible type that
				// allows us to add fields to an object at our choosing
				// Mostly used for ctx
				scope = scopeFromObject(scope, e)
			case e.Enum != nil:
				// if the first fragment of the Ident matches an Enum name, then we want to proceed populating the scope for the next fragment
				// with all of the potential values of the enum
				scope = scopeFromEnum(scope, e)
			case e.EnumValue != nil:
				scope = &ExpressionScope{
					Parent: scope,
				}
			case e.Type != "":
				// Otherwise, the scope is empty of any new entities
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
