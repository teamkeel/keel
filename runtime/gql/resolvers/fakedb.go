package resolvers

import (
	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
)

func fetchDbRow(
	model *proto.Model,
	whereExpressions []*proto.Expression,
	queryParams graphql.ResolveParams) (map[string]any, error) {

	// We will use the where clauses to filter the rows
	// to return like this:
	for _, where := range whereExpressions {
		// fmt.Printf("where expression is: %v\n", where)
		_ = where
	}

	// We also use the ResolveParams to filter the rows.
	for paramName, paramValue := range queryParams.Args {
		//fmt.Printf("XXXX paramName: %s, paramValue: %v\n", paramName, paramValue)
		_ = paramName
		_ = paramValue
	}

	// Fake a row for now
	row, err := fakeRow(model)
	if err != nil {
		return nil, err
	}
	return row, nil
}
