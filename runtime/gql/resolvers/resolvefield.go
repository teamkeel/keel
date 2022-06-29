package resolvers

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
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
	fmt.Printf("XXXX field resolver fired\n")
	// Expects to be able to retreive the value for this field from the source object.
	fmt.Printf("XXXX source obj given as: \n%v\n", p.Source)

	asMap, ok := p.Source.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot coerce the source object to map[string]any")
	}
	value, ok := asMap[r.field.Name]
	if !ok {
		return nil, fmt.Errorf("the source map does not contain field: %s", r.field.Name)
	}
	return value, nil
}
