package resolvers

import (
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

// A FieldResolver provides a Resolve method that matches the signature needed for
// a graphql.FieldResolveFn. It is implemented as an object that holds state, so that
// individual instances can be constructed with access to the data they need when their
// Resolve method gets called.
type FieldResolver struct {
	field *proto.Field
}

func NewFieldResolver(field *proto.Field) *FieldResolver {
	return &FieldResolver{
		field: field,
	}
}

func (r *FieldResolver) Resolve(p graphql.ResolveParams) (interface{}, error) {
	// Expects to be able to retreive the value for this field from the source object.
	field, err := actions.Field(p.Info.FieldName, p.Source)
	if err != nil {
		return nil, err
	}
	return field, nil
}
