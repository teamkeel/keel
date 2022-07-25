package actions

import (
	"context"

	"github.com/teamkeel/keel/proto"
)

func Get(ctx context.Context, model *proto.Model, schema *proto.Schema, args map[string]any, where []*proto.Expression) (interface{}, error) {

	// We will use the where clauses to filter the rows
	// to return like this:
	for _, where := range where {
		// fmt.Printf("where expression is: %v\n", where)
		_ = where
	}

	// We also use the ResolveParams to filter the rows.
	for paramName, paramValue := range args {
		//fmt.Printf("XXXX paramName: %s, paramValue: %v\n", paramName, paramValue)
		_ = paramName
		_ = paramValue
	}

	// Fake a row for now
	row, err := fakeRow(model, schema.Enums)
	if err != nil {
		return nil, err
	}
	return row, nil
}
