package expressions

import (
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

type ConditionResolver struct {
	condition *parser.Condition
	context   *ExpressionContext
	asts      []*parser.AST
}

// func (c *ConditionResolver) ToSql(data map[string]any){}
func (c *ConditionResolver) Resolve() (resolvedLhs *ExpressionScopeEntity, resolvedRhs *ExpressionScopeEntity, errors []error) {
	lhs := NewOperandResolver(
		c.condition.LHS,
		c.asts,
		c.context,
		OperandPositionLhs,
	)
	rhs := NewOperandResolver(
		c.condition.RHS,
		c.asts,
		c.context,
		OperandPositionRhs,
	)

	resolvedLhs, lhsErr := lhs.Resolve()

	if lhsErr != nil {
		errors = append(errors, lhsErr.ToValidationError())
	}

	if rhs != nil {
		resolvedRhs, rhsErr := rhs.Resolve()

		if rhsErr != nil {
			errors = append(errors, rhsErr.ToValidationError())
		}

		return resolvedLhs, resolvedRhs, errors
	}

	return resolvedLhs, nil, errors
}

func NewConditionResolver(condition *parser.Condition, asts []*parser.AST, context *ExpressionContext) *ConditionResolver {
	return &ConditionResolver{
		condition: condition,
		context:   context,
		asts:      asts,
	}
}

type OperandResolver struct {
	operand  *parser.Operand
	asts     []*parser.AST
	context  *ExpressionContext
	scope    *ExpressionScope
	position OperandPosition
}

func NewOperandResolver(operand *parser.Operand, asts []*parser.AST, context *ExpressionContext, position OperandPosition) *OperandResolver {
	return &OperandResolver{
		operand:  operand,
		asts:     asts,
		context:  context,
		scope:    &ExpressionScope{},
		position: position,
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
func (o *OperandResolver) Resolve() (entity *ExpressionScopeEntity, err *ResolutionError) {
	// build the default expression scope for all expressions
	o.scope = buildRootExpressionScope(o.asts, o.context)
	// build additional root scopes based on position of operand
	// and also what type attribute the expression is used in.
	o.scope = applyAdditionalOperandScopes(o.asts, o.scope, o.context, o.position)

	// If it is a literal then handle differently.
	if ok, _ := o.operand.IsLiteralType(); ok {
		if o.operand.Type() == parser.TypeArray {
			array := []*ExpressionScopeEntity{}

			for _, item := range o.operand.Array.Values {
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
				Literal: o.operand,
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
	for _, fragment := range o.operand.Ident.Fragments {
		for _, e := range o.scope.Entities {
			if e.Name != fragment.Fragment {
				continue
			}

			switch {
			case e.Model != nil:
				// crucially, scope is redefined for the next iteration of the outer loop
				// so that we check the subsequent fragment against the models fields
				// e.g post.field - fragment at idx 0 would be post, so scopeFromModel finds the fields on the
				// Post model and populates the new scope with them.
				o.scope = scopeFromModel(o.scope, e, e.Model)
			case e.Field != nil:
				// covers fields which are associations to other models:
				// e.g post.association.associationField
				// repopulates the scope for the third fragment to be the fields of 'association'
				model := query.Model(o.asts, e.Field.Type)

				// if no model is found for the field type, then we need to check other potential matches
				// e.g enums in the schema
				if model == nil {
					// try matching the field name to a known enum instead
					enum := query.Enum(o.asts, e.Field.Type)

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
						o.scope = &ExpressionScope{
							Parent: o.scope,
						}
					}
				} else {
					// move onto the associations' fields so we populate the new scope with them
					o.scope = scopeFromModel(o.scope, e, model)
				}
			case e.Object != nil:
				// object is a special wrapper type to describe entities we want
				// to be in scope that aren't models. It is a more flexible type that
				// allows us to add fields to an object at our choosing
				// Mostly used for ctx
				o.scope = scopeFromObject(o.scope, e)
			case e.Enum != nil:
				// if the first fragment of the Ident matches an Enum name, then we want to proceed populating the scope for the next fragment
				// with all of the potential values of the enum
				o.scope = scopeFromEnum(o.scope, e)
			case e.EnumValue != nil:
				// if we are evaluating an EnumValue, then there are no more
				// child entities to append as an EnumValue is a termination point.
				// e.g EnumName.EnumValue.SomethingElse isnt a thing.
				o.scope = &ExpressionScope{
					Parent: o.scope,
				}
			case e.Type != "":
				// Otherwise, the scope is empty of any new entities
				o.scope = &ExpressionScope{
					Parent: o.scope,
				}
			}

			entity = e
			continue fragments
		}

		parent := ""

		if entity != nil {
			parent = entity.GetType()
		}

		return nil, &ResolutionError{
			fragment: fragment,
			parent:   parent,
			scope:    o.scope,
			operand:  o.operand,
		}
	}

	return entity, nil
}
