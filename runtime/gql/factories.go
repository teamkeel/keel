package gql

import "github.com/graphql-go/graphql"

// newField creates a graphql.Field. It exists to handle some of the boiler plate,
// and help enforce the required arguments.
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

// newFieldWithArgs is an extension of newField that allows input arguments to
// be specified.
func newFieldWithArgs(
	name string,
	args graphql.FieldConfigArgument,
	outputType graphql.Output,
	resolver graphql.FieldResolveFn) *graphql.Field {
	return &graphql.Field{
		Name:    name,
		Args:    args,
		Type:    outputType,
		Resolve: resolver,
	}
}

// newObject creates a graphql.Object. It exists to handle some of the boiler plate,
// and help enforce the required arguments.
func newObject(name string, fields graphql.Fields) *graphql.Object {
	objectConfig := graphql.ObjectConfig{
		Name:   name,
		Fields: fields,
	}
	return graphql.NewObject(objectConfig)
}
