package gql

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

type Maker struct {
	proto *proto.Schema
}

func NewMaker(proto *proto.Schema) *Maker {
	return &Maker{
		proto: proto,
	}
}

// Make makes a set of graphql.Schema objects - one for each
// of the APIs defined in the given keel schema. It returns these in a map keyed
// on the API name.
func (mk *Maker) Make() (map[string]*graphql.Schema, error) {
	outputSchemas := map[string]*graphql.Schema{}
	for _, api := range mk.proto.Apis {
		gSchema, err := mk.makeSchemaForOneAPI(api)
		if err != nil {
			return nil, err
		}
		outputSchemas[api.Name] = gSchema
	}
	return outputSchemas, nil
}

func (mk *Maker) makeSchemaForOneAPI(api *proto.Api) (*graphql.Schema, error) {
	namesOfModelsUsedByAPI := lo.Map(api.ApiModels, func(m *proto.ApiModel, _ int) string {
		return m.ModelName
	})
	modelInstances := proto.FindModels(mk.proto.Models, namesOfModelsUsedByAPI)

	fieldsUnderConstruction := &fieldsUnderConstruction{
		queries:   graphql.Fields{},
		mutations: graphql.Fields{},
		models:    []graphql.Type{},
	}

	for _, model := range modelInstances {
		modelOutputType, err := mk.addModel(model, fieldsUnderConstruction)
		if err != nil {
			return nil, err
		}
		for _, op := range model.Operations {
			if err := mk.addOperation(op, modelOutputType, model, fieldsUnderConstruction); err != nil {
				return nil, err
			}
		}
	}

	queryType := newObject("Query", fieldsUnderConstruction.queries)
	//mutationType := newObject("Mutation", fieldsUnderConstruction.mutations)
	gSchema, err := graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    queryType,
			Mutation: nil,
			Types:    fieldsUnderConstruction.models,
		},
	)
	if err != nil {
		return nil, err
	}

	return &gSchema, nil
}

func (mk *Maker) addModel(model *proto.Model, addTo *fieldsUnderConstruction) (modelOutputType graphql.Output, err error) {
	// todo - don't add it, if we already did earlier
	fields := graphql.Fields{}
	for _, field := range model.Fields {
		outputType, err := mk.outputTypeFor(field)
		if err != nil {
			return nil, err
		}
		field := newField(field.Name, outputType, NewFieldResolver().Resolve)
		fields[field.Name] = field
	}
	modelOutputType = newObject(model.Name, fields)
	addTo.models = append(addTo.models, modelOutputType)
	return modelOutputType, nil
}

func (mk *Maker) addOperation(
	op *proto.Operation,
	modelOutputType graphql.Output,
	model *proto.Model,
	addTo *fieldsUnderConstruction) error {
	// todo - don't add it, if we already did earlier
	for _, op := range model.Operations {
		if op.Implementation != proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO {
			continue
		}
		switch op.Type {
		case proto.OperationType_OPERATION_TYPE_GET:
			if err := mk.addGetOp(op, modelOutputType, model, addTo); err != nil {
				return err
			}
		default:
			return fmt.Errorf("addOperation() does not yet support this op.Type: %v", op.Type)
		}
	}
	return nil
}

func (mk *Maker) addGetOp(
	op *proto.Operation,
	modelOutputType graphql.Output,
	model *proto.Model,
	addTo *fieldsUnderConstruction) error {
	args, err := mk.makeArgs(op)
	if err != nil {
		return err
	}
	field := newFieldWithArgs(op.Name, args, modelOutputType, NewGetOpResolver().Resolve)
	addTo.queries[op.Name] = field
	return nil
}

func (mk *Maker) outputTypeFor(field *proto.Field) (graphql.Output, error) {
	if outputType, ok := mk.isFieldTypeDirectlyMappableType(field.Type); ok {
		return outputType, nil
	}
	return nil, fmt.Errorf("cannot yet make output type for a: %v", field)
}

func (mk *Maker) inputTypeFor(op *proto.OperationInput) (graphql.Input, error) {
	if inputType, ok := mk.isOperationInputTypeDirectlyMappableType(op.Type); ok {
		return inputType, nil
	}
	return nil, fmt.Errorf("cannot yet make input type for a: %v", op.Type)
}

func (mk *Maker) isFieldTypeDirectlyMappableType(keelType proto.FieldType) (graphql.Output, bool) {
	switch keelType {
	case proto.FieldType_FIELD_TYPE_STRING:
		return graphql.String, true

	case proto.FieldType_FIELD_TYPE_INT:
		return graphql.Int, true

	case proto.FieldType_FIELD_TYPE_ID:
		return graphql.String, true
	}
	return nil, false
}

func (mk *Maker) isOperationInputTypeDirectlyMappableType(keelType proto.OperationInputType) (graphql.Input, bool) {
	switch keelType {
	// Special case, when specifying a field - we expect its name.
	case proto.OperationInputType_OPERATION_INPUT_TYPE_FIELD:
		return graphql.String, true

	// General (scalar) cases.
	case proto.OperationInputType_OPERATION_INPUT_TYPE_BOOL:
		return graphql.Boolean, true
	case proto.OperationInputType_OPERATION_INPUT_TYPE_STRING:
		return graphql.String, true
	}
	return nil, false
}

func (mk *Maker) makeArgs(op *proto.Operation) (graphql.FieldConfigArgument, error) {
	res := graphql.FieldConfigArgument{}
	for _, input := range op.Inputs {
		inputType, err := mk.inputTypeFor(input)
		if err != nil {
			return nil, err
		}
		res[input.Name] = &graphql.ArgumentConfig{
			Type: inputType,
		}
	}
	return res, nil
}

type fieldsUnderConstruction struct {
	queries   graphql.Fields
	mutations graphql.Fields
	models    []graphql.Type
}
