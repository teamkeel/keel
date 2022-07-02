package resolvers

import (
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
)

func fetchDbRow(
	model *proto.Model,
	whereExpressions []*proto.Expression,
	queryParams graphql.ResolveParams) (map[string]any, error) {

	// We will unpack the params to parameterise the db query like this:
	for paramName, paramValue := range queryParams.Args {
		_ = paramName
		_ = paramValue
		// fmt.Printf("row query param: %s == %s\n", paramName, paramValue)
	}

	// We will assemble the where clauses to decide which of the row's columns
	// to return like this:
	for _, where := range whereExpressions {
		// fmt.Printf("where expression is: %v\n", where)
		_ = where
	}

	// Fake it.
	return map[string]any{
		"name": "Harriet",
	}, nil
}
