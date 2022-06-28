package gql

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
)

type FieldResolver struct {
}

func NewFieldResolver() *FieldResolver {
	return &FieldResolver{}
}

func (r *FieldResolver) Resolve(p graphql.ResolveParams) (interface{}, error) {
	fmt.Printf("XXXX field resolver fired\n")
	return "Not yet implemented", nil
}

type ModelResolver struct {
}

func NewModelResolver() *ModelResolver {
	return &ModelResolver{}
}

func (mr *ModelResolver) Resolve(p graphql.ResolveParams) (interface{}, error) {
	fmt.Printf("XXXX model resolver fired\n")
	return "Not yet implemented", nil
}

type GetOpResolver struct {
	op *proto.Operation
}

func NewGetOpResolver(op *proto.Operation) *GetOpResolver {
	return &GetOpResolver{
		op: op,
	}
}

func (r *GetOpResolver) Resolve(p graphql.ResolveParams) (interface{}, error) {
	// For the moment - just illustrate how this resolver has the info it needs
	// to anticpate, which inputs are expected, and to fetch the corresponding
	// values from the incoming ResolveParams object.
	for _, input := range r.op.Inputs {
		paramValue, ok := p.Args[input.Name]
		if !ok {
			return nil, fmt.Errorf("the input named: %s is missing", input.Name)
		}
		fmt.Printf("XXXX found value: %v for param: %s\n", paramValue, input.Name)
	}

	return "Not yet implemented", nil
}
