package gql

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

type FieldResolver struct {
}

func NewFieldResolver() *FieldResolver {
	return &FieldResolver{}
}

func (r *FieldResolver) Resolve(p graphql.ResolveParams) (interface{}, error) {
	fmt.Printf("XXXX field resolver fired, with params: %+v\n", p)
	return "Not yet implemented", nil
}

type ModelResolver struct {
}

func NewModelResolver() *ModelResolver {
	return &ModelResolver{}
}

func (mr *ModelResolver) Resolve(p graphql.ResolveParams) (interface{}, error) {
	fmt.Printf("XXXX model resolver fired, with params: %+v\n", p)
	return "Not yet implemented", nil
}

type GetOpResolver struct {
}

func NewGetOpResolver() *GetOpResolver {
	return &GetOpResolver{}
}

func (r *GetOpResolver) Resolve(p graphql.ResolveParams) (interface{}, error) {
	fmt.Printf("XXXX operation resolver fired, with params: %+v\n", p)
	return "Not yet implemented", nil
}
