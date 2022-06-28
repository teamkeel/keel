package gql

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

type ModelResolver struct {
}

func NewModelResolver() *ModelResolver {
	return &ModelResolver{}
}

func (mr *ModelResolver) Resolve(p graphql.ResolveParams) (interface{}, error) {
	fmt.Printf("XXXX model resolver fired\n")
	return "Not yet implemented", nil
}
