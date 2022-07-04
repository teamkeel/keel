package resolvers

import (
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
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

func (r *GetOperationResolver) Resolve(p graphql.ResolveParams) (interface{}, error) {
	// Fetch the data to populate the containerToReturn with.
	containerToReturn, err := r.fetch(p)
	if err != nil {
		return nil, err
	}

	return containerToReturn, nil
}

func (r *GetOperationResolver) fetch(queryParams graphql.ResolveParams) (map[string]any, error) {
	// This is where we will talk to the database (or a storage abstraction).
	// But for now - we'll fake it.

	record, err := fetchDbRow(r.model, r.op.WhereExpressions, queryParams)
	if err != nil {
		return nil, err
	}
	return record, nil
}
