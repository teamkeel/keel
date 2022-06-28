package gql

import "github.com/graphql-go/graphql"

func newField(
	name string,
	outputType graphql.Output,
	resolver graphql.FieldResolveFn) *graphql.Field {
	return &graphql.Field{
		Name:    name,
		Type:    outputType,
		Resolve: resolver,
	}
}

func newObject(name string, fields graphql.Fields) *graphql.Object {
	objectConfig := graphql.ObjectConfig{
		Name:   name,
		Fields: fields,
	}
	return graphql.NewObject(objectConfig)
}
