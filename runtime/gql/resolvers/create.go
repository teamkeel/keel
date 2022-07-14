package resolvers

import (
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

// A CreateOperationResolver provides a Resolve method that matches the signature needed for
// a graphql.FieldResolveFn. It is implemented as an object that holds state, so that
// individual instances can be constructed with access to the data they need when their
// Resolve method gets called.
type CreateOperationResolver struct {
	op    *proto.Operation
	model *proto.Model
}

func NewCreateOperationResolver(op *proto.Operation, model *proto.Model) *CreateOperationResolver {
	return &CreateOperationResolver{
		op:    op,
		model: model,
	}
}

func (r *CreateOperationResolver) Resolve(p graphql.ResolveParams) (any, error) {
	res, err := actions.Create(p.Context, r.op, p.Args)
	if err != nil {
		return nil, err
	}
	return res, nil
}
