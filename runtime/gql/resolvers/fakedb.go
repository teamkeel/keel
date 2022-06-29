package resolvers

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
)

func fetchDbRowThatMatchParamsAndFilters(
	model *proto.Model,
	whereExpressions []*proto.Expression,
	queryParams graphql.ResolveParams) (map[string]any, error) {

	// We will unpack the params to parameterise the db query like this:
	for paramName, paramValue := range queryParams.Args {
		fmt.Printf("XXXX row query param: %s == %s\n", paramName, paramValue)
	}

	// We will assemble the where clauses to decide which of the row's columns
	// to return like this:
	for _, where := range whereExpressions {
		fmt.Printf("XXXX where expression is: %v\n", where)
	}

	// Fake it.
	return map[string]any{
		"name": "Harriet",
	}, nil
}
