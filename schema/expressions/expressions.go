package expressions

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

type ConditionResolver struct {
	condition *parser.Condition
	context   *ExpressionContext
	asts      []*parser.AST
}

func (c *ConditionResolver) Resolve() (resolvedLhs *ExpressionScopeEntity, resolvedRhs *ExpressionScopeEntity, errors []error) {
	lhs := NewOperandResolver(
		c.condition.LHS,
		c.asts,
		c.context,
		OperandPositionLhs,
	)

	resolvedLhs, lhsErr := lhs.Resolve()
	if lhsErr != nil {
		errors = append(errors, lhsErr.ToValidationError())
	}

	// Check RHS only if it exists
	if c.condition.RHS != nil {
		rhs := NewOperandResolver(
			c.condition.RHS,
			c.asts,
			c.context,
			OperandPositionRhs,
		)

		resolvedRhs, rhsErr := rhs.Resolve()
		if rhsErr != nil {
			errors = append(errors, rhsErr.ToValidationError())
		}

		return resolvedLhs, resolvedRhs, errors
	} else if resolvedLhs != nil && resolvedLhs.GetType() != parser.FieldTypeBoolean {
		errors = append(errors,
			errorhandling.NewValidationError(
				errorhandling.ErrorExpressionSingleConditionNotBoolean,
				errorhandling.TemplateLiterals{
					Literals: map[string]string{
						"Value":      lhs.operand.ToString(),
						"Attribute":  fmt.Sprintf("@%s", c.context.Attribute.Name.Value),
						"Suggestion": fmt.Sprintf("%s == xxx", lhs.operand.ToString()),
					},
				},
				lhs.operand.Node,
			),
		)
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
// - post.author
// - MyEnum.ValueName
// - "123"
// - true
// - ctx.identity.account
// - ctx.identity.account.name
// All of these types above are checked / attempted to be resolved in this method.
func (o *OperandResolver) Resolve() (entity *ExpressionScopeEntity, err *ResolutionError) {
	// build the default expression scope for all expressions
	o.scope = buildRootExpressionScope(o.asts, o.context)
	// build additional root scopes based on what type attribute the expression is used in.
	o.scope = applyAdditionalOperandScopes(o.asts, o.scope, o.context)

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
		if entity != nil && entity.Type == TypeStringMap {
			o.scope = &ExpressionScope{
				Parent: o.scope,
			}

			entity = &ExpressionScopeEntity{
				Type: parser.TypeText,
			}

			continue
		}

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

				model := query.Model(o.asts, e.Field.Type.Value)
				enum := query.Enum(o.asts, e.Field.Type.Value)

				if model != nil {
					// If the field type is a model the scope is now that models fields
					o.scope = scopeFromModel(o.scope, e, model)
				} else {
					// For enums we add some extra context to the scope entity so we can
					// resolve it properly later on
					if enum != nil {
						e.Type = parser.TypeEnum
					}

					// Non-model fields have no sub-properties, so the scope is now empty
					o.scope = &ExpressionScope{
						Parent: o.scope,
					}
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
