package actions

import (
	"context"

	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

// GenerateFilterQuery visits the expression and adds filter conditions to the provided query builder.
func GenerateFilterQuery(ctx context.Context, query *QueryBuilder, schema *proto.Schema, entity proto.Entity, action *proto.Action, inputs map[string]any) resolve.Visitor[*QueryBuilder] {
	return &baseQueryGen{
		ctx:       ctx,
		query:     query,
		schema:    schema,
		entity:    entity,
		action:    action,
		inputs:    inputs,
		operators: arraystack.New(),
		operands:  arraystack.New(),
		identHandler: func(ctx context.Context, query *QueryBuilder, schema *proto.Schema, ident *parser.ExpressionIdent, operands *arraystack.Stack) error {
			operand, err := generateOperand(ctx, schema, entity, action, inputs, ident.Fragments)
			if err != nil {
				return err
			}

			err = query.AddJoinFromFragments(schema, ident.Fragments)
			if err != nil {
				return err
			}

			operands.Push(operand)

			return nil
		},
	}
}
