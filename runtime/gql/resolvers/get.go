package resolvers

import (
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

// A GetOperationResolver provides a Resolve method that matches the signature needed for
// a graphql.FieldResolveFn. It is implemented as an object that holds state, so that
// individual instances can be constructed with access to the data they need when their
// Resolve method gets called.
type GetOperationResolver struct {
	op    *proto.Operation
	model *proto.Model
}

func NewGetOperationResolver(op *proto.Operation, model *proto.Model) *GetOperationResolver {
	return &GetOperationResolver{
		op:    op,
		model: model,
	}
}

func (r *GetOperationResolver) Resolve(p graphql.ResolveParams) (any, error) {
	res, err := actions.Get(p.Context, r.model, p.Args, r.op.WhereExpressions)
	if err != nil {
		return nil, err
	}
	return res, nil
}
