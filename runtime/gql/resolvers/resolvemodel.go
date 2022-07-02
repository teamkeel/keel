package resolvers

import (
	"github.com/graphql-go/graphql"
)

// A ModelResolver provides a Resolve method that matches the signature needed for
// a graphql.FieldResolveFn. It is implemented as an object that holds state, so that
// individual instances can be constructed with access to the data they need when their
// Resolve method gets called.
type ModelResolver struct {
}

func NewModelResolver() *ModelResolver {
	return &ModelResolver{}
}

func (mr *ModelResolver) Resolve(p graphql.ResolveParams) (interface{}, error) {
	return "Not yet implemented", nil
}
