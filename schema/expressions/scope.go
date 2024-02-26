package expressions

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
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
	Field     *parser.FieldNode
}

type ResolutionError struct {
	scope    *ExpressionScope
	fragment *parser.IdentFragment
	parent   string
	operand  *parser.Operand
}

func (e *ResolutionError) InScopeEntities() []string {
	return lo.Map(e.scope.Entities, func(e *ExpressionScopeEntity, _ int) string {
		return e.Name
	})
}

func (e *ResolutionError) Error() string {
	return fmt.Sprintf("Could not resolve %s in %s", e.fragment.Fragment, e.operand.ToString())
}

func (e *ResolutionError) ToValidationError() *errorhandling.ValidationError {
	suggestions := errorhandling.NewCorrectionHint(e.InScopeEntities(), e.fragment.Fragment)

	literals := map[string]string{
		"Fragment": e.fragment.Fragment,
		"Parent":   e.parent,
	}

	if len(suggestions.Results) > 0 {
		literals["Suggestion"] = suggestions.ToString()
	}

	return errorhandling.NewValidationError(
		errorhandling.ErrorUnresolvableExpression,
		errorhandling.TemplateLiterals{
			Literals: literals,
		},
		e.fragment,
	)
}

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

func buildRootExpressionScope(asts []*parser.AST, context *ExpressionContext) *ExpressionScope {
	if context.Model == nil {
		return DefaultExpressionScope(asts)
	}

	contextualScope := &ExpressionScope{
		Entities: []*ExpressionScopeEntity{
			{
				Name:  casing.ToLowerCamel(context.Model.Name.Value),
				Model: context.Model,
			},
		},
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

	// Type can be things like "Text", "Boolean" etc.. but can also be "Enum".
	// If the value is "Enum" then the Field attribute will be populated
	// with a model field that is an enum
	Type string

	Parent *ExpressionScopeEntity
}

func (e *ExpressionScopeEntity) IsNull() bool {
	return e.Literal != nil && e.Literal.Null
}

func (e *ExpressionScopeEntity) IsOptional() bool {
	return e.Field != nil && e.Field.Optional
}

func (e *ExpressionScopeEntity) IsEnumField() bool {
	return e.Field != nil && e.Type == parser.TypeEnum
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
		return e.Field.Type.Value
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

func (e *ExpressionScopeEntity) AllowedOperators(asts []*parser.AST) []string {
	t := e.GetType()

	arrayEntity := e.IsRepeated()

	if !arrayEntity && e.Model != nil {
		t = parser.TypeModel
	}

	// When the field is of type model
	if !arrayEntity && e.Field != nil && query.Model(asts, e.Field.Type.Value) != nil {
		t = parser.TypeModel
	}

	if arrayEntity {
		t = parser.TypeArray
	}

	if (e.IsEnumField() || e.IsEnumValue()) && !arrayEntity {
		t = parser.TypeEnum
	}

	return operatorsForType[t]
}

func DefaultExpressionScope(asts []*parser.AST) *ExpressionScope {
	var envVarEntities []*ExpressionScopeEntity
	var secretsEntities []*ExpressionScopeEntity
	for _, ast := range asts {
		for _, envVar := range ast.EnvironmentVariables {
			envVarEntities = append(envVarEntities, &ExpressionScopeEntity{
				Name: envVar,
				Type: parser.FieldTypeText,
			})
		}
		for _, secret := range ast.Secrets {
			secretsEntities = append(secretsEntities, &ExpressionScopeEntity{
				Name: secret,
				Type: parser.FieldTypeText,
			})
		}
	}

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
					{
						Name: "env",
						Object: &ExpressionObjectEntity{
							Name:   "Environment Variables",
							Fields: envVarEntities,
						},
					},
					{
						Name: "headers",
						Type: TypeStringMap,
					},
					{
						Name: "secrets",
						Object: &ExpressionObjectEntity{
							Name:   "Secrets",
							Fields: secretsEntities,
						},
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

func applyAdditionalOperandScopes(asts []*parser.AST, scope *ExpressionScope, context *ExpressionContext) *ExpressionScope {
	additionalScope := &ExpressionScope{}

	attribute := context.Attribute
	action := context.Action

	// If there is no action, then we dont want to do anything
	if action == nil {
		return scope
	}

	switch attribute.Name.Value {
	case parser.AttributePermission:
		// no inputs allowed in permissions
	default:
		scope = applyInputsInScope(asts, context, scope)
	}

	return scope.Merge(additionalScope)
}

func applyInputsInScope(asts []*parser.AST, context *ExpressionContext, scope *ExpressionScope) *ExpressionScope {
	additionalScope := &ExpressionScope{}

	inputs := []*parser.ActionInputNode{}
	inputs = append(inputs, context.Action.Inputs...)
	inputs = append(inputs, context.Action.With...)

	for _, input := range inputs {
		// inputs using short-hand syntax that refer to relationships
		// don't get added to the scope
		if input.Label == nil && len(input.Type.Fragments) > 1 {
			continue
		}

		resolvedType := query.ResolveInputType(asts, input, context.Model, context.Action)
		if resolvedType == "" {
			continue
		}
		additionalScope.Entities = append(additionalScope.Entities, &ExpressionScopeEntity{
			Name: input.Name(),
			Type: resolvedType,
		})
	}

	return scope.Merge(additionalScope)
}
