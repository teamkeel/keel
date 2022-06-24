package makegql

import "github.com/graphql-go/graphql"

func newArgumentConfig(inputType graphql.Input) *graphql.ArgumentConfig {
	return &graphql.ArgumentConfig{
		Type: inputType,
	}
}

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

func newSchema(
	query *graphql.Object,
	mutation *graphql.Object,
) *graphql.Schema {
	config := graphql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	}
	schema, err := graphql.NewSchema(config)
	if err != nil {
		panic(err.Error())
	}
	return &schema
}
