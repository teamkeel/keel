package actions

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// Updates the query with all set attributes defined on the operation.
func (query *QueryBuilder) captureSetValues(scope *Scope, args ValueArgs) error {
	for _, setExpression := range scope.operation.SetExpressions {
		expression, err := parser.ParseExpression(setExpression.Source)
		if err != nil {
			return err
		}

		assignment, err := expression.ToAssignmentCondition()
		if err != nil {
			return err
		}

		lhsResolver := NewOperandResolver(scope.context, scope.schema, scope.operation, assignment.LHS)
		rhsResolver := NewOperandResolver(scope.context, scope.schema, scope.operation, assignment.RHS)
		operandType, err := lhsResolver.GetOperandType()
		if err != nil {
			return err
		}

		if !(lhsResolver.IsDatabaseColumn() || lhsResolver.IsWriteValue()) {
			return fmt.Errorf("lhs operand of assignment expression must be a model field")
		}

		value, err := rhsResolver.ResolveValue(args, query.writeValues)
		if err != nil {
			return err
		}

		fieldName := assignment.LHS.Ident.Fragments[1].Fragment

		// If targeting the field of a nested model, then combine into a camelCase field name.
		// For example, post.author.id will become authorId.
		if len(assignment.LHS.Ident.Fragments) == 3 {
			fieldName = fmt.Sprintf("%s%s", fieldName, strcase.ToCamel(assignment.LHS.Ident.Fragments[2].Fragment))
		}

		// If targeting the nested model (without a field), then set the foreign key with the "id" of the assigning model.
		// For example, @set(post.user = ctx.identity) will set post.userId with ctx.identity.id.
		if operandType == proto.Type_TYPE_MODEL && value != nil {
			fieldName = fmt.Sprintf("%sId", fieldName)
		}

		// Add a value to be written during an INSERT or UPDATE
		query.AddWriteValue(fieldName, value)
	}
	return nil
}

// Updates the query with all inputs defined on the operation.
func (query *QueryBuilder) captureWriteValues(scope *Scope, args ValueArgs) error {
	for _, input := range scope.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		if input.Mode != proto.InputMode_INPUT_MODE_WRITE {
			continue
		}

		fieldName := input.Target[0]

		// If targeting the field of a nested model, then combine into a camelCase field name.
		// For example, author.id will become authorId.
		if len(input.Target) == 2 {
			fieldName = fmt.Sprintf("%s%s", fieldName, strcase.ToCamel(input.Target[1]))
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
