package operand

import (
	"github.com/iancoleman/strcase"
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
	Fields map[string]*ExpressionScopeEntity
}

type ExpressionScopeEntity struct {
	Object    *ExpressionObjectEntity
	Model     *parser.ModelNode
	Field     *parser.FieldNode
	Input     *ExpressionInputEntity
	Literal   *expressions.Operand
	Enum      *parser.EnumNode
	EnumValue *parser.EnumValueNode
	Array     []*ExpressionScopeEntity

	Parent *ExpressionScopeEntity
}

type ExpressionInputEntity struct {
	Name       string // The name of the input as it can be referenced in an expression
	Type       string // will be a valid field type e.g. Text, Boolean, Number etc...
	AllowWrite bool   // true if this input value can be used to write values
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

	if e.EnumValue != nil {
		return e.Parent.Enum.Name.Value
	}

	if e.Array != nil {
		return expressions.TypeArray
	}

	if e.Input != nil {
		return e.Input.Type
	}

	return ""
}

func (e *ExpressionScopeEntity) AllowedOperators() (operators []string) {
	if e.IsRepeated() {
		operators = append(operators, expressions.ArrayOperators...)
		return operators
	}

	switch {
	case e.Literal != nil:
		t := e.Literal.Type()

		switch t {
		case expressions.TypeBoolean:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
		case expressions.TypeNumber:
			operators = append(operators, expressions.LogicalOperators...)
			operators = append(operators, expressions.OperatorAssignment)
		case expressions.TypeNull:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.AssignmentCondition)
		case expressions.TypeText:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
		case expressions.TypeArray:
			operators = append(operators, expressions.ArrayOperators...)
		}
	case e.Model != nil:
		operators = append(operators, expressions.OperatorEquals)
		operators = append(operators, expressions.OperatorAssignment)
	case e.Field != nil || e.Input != nil:
		switch e.Type() {
		case expressions.TypeText, parser.FieldTypeText:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
		case expressions.TypeBoolean:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
		case expressions.TypeNumber:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
			operators = append(operators, expressions.NumericalOperators...)
		default:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
		}
	case e.Object != nil:
		operators = append(operators, expressions.OperatorEquals)
		operators = append(operators, expressions.OperatorAssignment)
	}

	return operators
}

func DefaultExpressionScope(asts []*parser.AST) *ExpressionScope {
	entities := []*ExpressionScopeEntity{
		{
			Object: &ExpressionObjectEntity{
				Name: "ctx",
				Fields: map[string]*ExpressionScopeEntity{
					"identity": {
						Model: query.Model(asts, "Identity"),
					},
				},
			},
		},
	}

	for _, enum := range query.Enums(asts) {
		entities = append(entities, &ExpressionScopeEntity{
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

	for _, field := range parentEntity.Object.Fields {
		newEntities = append(newEntities, &ExpressionScopeEntity{
			// copy all fields across
			Model:     field.Model,
			Object:    field.Object,
			Field:     field.Field,
			Input:     field.Input,
			Literal:   field.Literal,
			Enum:      field.Enum,
			EnumValue: field.EnumValue,
			Array:     field.Array,

			// set parent
			Parent: parentEntity,
		})
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
			switch {
			case e.Model != nil && strcase.ToLowerCamel(e.Model.Name.Value) == fragment.Fragment:
				entity = e

				scope = scopeFromModel(scope, e, e.Model)

				continue fragments
			case e.Field != nil && e.Field.Name.Value == fragment.Fragment:
				entity = e

				model := query.Model(asts, e.Field.Type)

				if model == nil {
					// Did not find the model matching the field
					scope = &ExpressionScope{
						Parent: scope,
					}
				} else {
					scope = scopeFromModel(scope, e, model)
				}

				continue fragments
			case e.Object != nil && e.Object.Name == fragment.Fragment:
				entity = e

				scope = scopeFromObject(scope, e)

				continue fragments
			case e.Enum != nil && e.Enum.Name.Value == fragment.Fragment:
				entity = e

				scope = scopeFromEnum(scope, e)

				continue fragments
			case e.EnumValue != nil && e.EnumValue.Name.Value == fragment.Fragment:
				entity = e

				scope = &ExpressionScope{
					Parent: scope,
				}

				continue fragments
			case e.Input != nil && e.Input.Name == fragment.Fragment:
				entity = e
				scope = &ExpressionScope{
					Parent: scope,
				}
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
