package actions

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/expressions"
	"github.com/teamkeel/keel/schema/parser"
)

// Updates the query with all set attributes defined on the operation.
func (query *QueryBuilder) captureSetValues(scope *Scope, args map[string]any) error {
	for _, setExpression := range scope.operation.SetExpressions {
		expression, err := parser.ParseExpression(setExpression.Source)
		if err != nil {
			return err
		}

		assignment, err := expression.ToAssignmentCondition()
		if err != nil {
			return err
		}

		lhsResolver := expressions.NewOperandResolver(scope.context, scope.schema, scope.operation, assignment.LHS)
		rhsResolver := expressions.NewOperandResolver(scope.context, scope.schema, scope.operation, assignment.RHS)
		operandType, err := lhsResolver.GetOperandType()
		if err != nil {
			return err
		}

		if !lhsResolver.IsDatabaseColumn() {
			return fmt.Errorf("lhs operand of assignment expression must be a model field")
		}

		value, err := rhsResolver.ResolveValue(args)
		if err != nil {
			return err
		}

		fieldName := assignment.LHS.Ident.Fragments[1].Fragment

		// Currently we only support 3 fragments in an set expression operand if it is targeting an "id" field.
		// If so, we generate the foreign key field name from the fragments.
		// For example, post.author.id will become authorId.
		if len(assignment.LHS.Ident.Fragments) == 3 {
			if assignment.LHS.Ident.Fragments[2].Fragment != "id" {
				return errors.New("currently only support 'id' as a third fragment in a set expression")
			}
			fieldName = fmt.Sprintf("%sId", fieldName)
		}

		// If targeting the nested model (without a field), then set the foreign key with the "id" of the assigning model.
		// For example, @set(post.user = ctx.identity) will set post.userId with ctx.identity.id.
		if operandType == proto.Type_TYPE_MODEL {
			fieldName = fmt.Sprintf("%sId", fieldName)
		}

		// Add a value to be written during an INSERT or UPDATE
		query.AddWriteValue(fieldName, value)
	}
	return nil
}

// Updates the query with all inputs defined on the operation.
func (query *QueryBuilder) captureWriteValues(scope *Scope, args map[string]any) error {
	message := proto.FindValuesInputMessage(scope.schema, scope.operation.Name)
	if message == nil {
		return nil
	}

	for _, input := range message.Fields {
		if !input.IsModelField() {
			continue
		}

		fieldName := input.Target[0]

		// Currently we only support a single-fragment implicit input EXCEPT when the 'id' of a model is targeted.
		// If so, we generate the foreign key field name from the fragments.
		// For example, author.id will become authorId.
		if len(input.Target) == 2 {
			if input.Target[1] != "id" {
				return errors.New("currently only support 'id' as a second fragment in an implicit input")
			}
			fieldName = fmt.Sprintf("%sId", fieldName)
		} else if len(input.Target) > 2 {
			return errors.New("nested implicit input not supported")
		}

		value, ok := args[fieldName]
		if !ok {
			continue
		}

		// Add a value to be written during an INSERT or UPDATE
		query.AddWriteValue(fieldName, value)
	}
	return nil
}
