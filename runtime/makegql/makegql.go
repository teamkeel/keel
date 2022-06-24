package makegql

import (
	"github.com/graphql-go/graphql"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

// MakeGQLSchemas makes a set of graphql.Schema objects - one for each
// of the APIs defined in the given keel schema. It returns these in a map keyed
// on the API name.
func MakeGQLSchemas(kSchema *proto.Schema) map[string]*graphql.Schema {
	res := map[string]*graphql.Schema{}
	for _, api := range kSchema.Apis {
		gSchema := makeSchemaForOneAPI(api, kSchema)
		res[api.Name] = gSchema
	}
	return res
}

func makeSchemaForOneAPI(api *proto.Api, kSchema *proto.Schema) *graphql.Schema {
	namesOfModelsUsedByAPI := lo.Map(api.ApiModels, func(m *proto.ApiModel, _ int) string {
		return m.ModelName
	})
	modelInstances := proto.FindModels(kSchema.Models, namesOfModelsUsedByAPI)

	fieldsToRepresentAPIModels := graphql.Fields{}
	for _, model := range modelInstances {
		gField := makeModelAsField(model, kSchema)
		fieldsToRepresentAPIModels[model.Name] = gField
	}

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: fieldsToRepresentAPIModels,
	})

	gSchema, err := graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    rootQuery,
			Mutation: nil,
		},
	)
	if err != nil {
		a := 42
		_ = a
		panic(err.Error())
	}

	return &gSchema
}

func makeModelAsField(model *proto.Model, kSchema *proto.Schema) *graphql.Field {
	outputType := SimpleMapAsGQLOutput{}
	modelAsAField := newField(model.Name, outputType, placeholderResolver)
	return modelAsAField
}

func outputTypeFor(field *proto.Field) graphql.Output {
	if outputType, ok := isDirectlyMappableType(field.Type); ok {
		return outputType
	}
	if outputType, ok := canMakeWellKnownCompoundType(field.Type); ok {
		return outputType
	}
	panic("Other output types not catered for yet")
}

var _ any = outputTypeFor // Preserve it from "unused" compile error

func placeholderResolver(p graphql.ResolveParams) (interface{}, error) {
	return 42, nil
}

func isDirectlyMappableType(keelType proto.FieldType) (graphql.Output, bool) {
	switch keelType {
	case proto.FieldType_FIELD_TYPE_STRING:
		return graphql.String, true

	case proto.FieldType_FIELD_TYPE_INT:
		return graphql.Int, true
		// todo put in the other directly mappable types
	}
	return nil, false
}

func canMakeWellKnownCompoundType(keelType proto.FieldType) (graphql.Output, bool) {
	switch keelType {
	case proto.FieldType_FIELD_TYPE_DATETIME:
		return DateTimeAsGQLOutput{}, true
	}
	return nil, false
}
