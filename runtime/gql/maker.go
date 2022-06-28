package gql

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

// A Maker exposes a Make method, that makes a set of graphql.Schema objects - one for each
// of the APIs defined in the keel schema provided at construction time.
type Maker struct {
	proto *proto.Schema
}

func NewMaker(proto *proto.Schema) *Maker {
	return &Maker{
		proto: proto,
	}
}

// The graphql.Schema(s) are returned in a map, keyed on the name of the
// API name they belong to.
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

// makeSchemaForOneAPI returns a graphql.Schema that implements the given API.
func (mk *Maker) makeSchemaForOneAPI(api *proto.Api) (*graphql.Schema, error) {
	// The graphql top level query contents will be comprised ONLY of the
	// OPERATIONS from the keel schema. But to find these we have to traverse the
	// schema, first by model, then by said model's operations. As a side effect
	// we must define graphl types for the models involved.

	namesOfModelsUsedByAPI := lo.Map(api.ApiModels, func(m *proto.ApiModel, _ int) string {
		return m.ModelName
	})
	modelInstances := proto.FindModels(mk.proto.Models, namesOfModelsUsedByAPI)

	// This is a container that the function call stack below populates as it goes.
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

// addModel generates the graphql type to represent the given proto.Model, and inserts it into
// the given fieldsUnderConstruction container.
func (mk *Maker) addModel(model *proto.Model, addTo *fieldsUnderConstruction) (modelOutputType graphql.Output, err error) {
	// todo - don't add it, if we already did earlier
	fields := graphql.Fields{}
	for _, field := range model.Fields {
		outputType, err := mk.outputTypeFor(field)
		if err != nil {
			return nil, err
		}
		field := newField(field.Name, outputType, NewFieldResolver(field).Resolve)
		fields[field.Name] = field
	}
	modelOutputType = newObject(model.Name, fields)
	addTo.models = append(addTo.models, modelOutputType)
	return modelOutputType, nil
}

// addOperation generates the graphql field object to represent the given proto.Operation.
// This field will eventually live in the top level graphql Query type, but at this stage
// the function just accumulates them in the given fieldsUnderConstruction container.
func (mk *Maker) addOperation(
	op *proto.Operation,
	modelOutputType graphql.Output,
	model *proto.Model,
	addTo *fieldsUnderConstruction) error {
	// todo - don't add it, if we already did earlier
	if op.Implementation != proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO {
		return nil
	}
	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_GET:
		if err := mk.addGetOp(op, modelOutputType, model, addTo); err != nil {
			return err
		}
	default:
		return fmt.Errorf("addOperation() does not yet support this op.Type: %v", op.Type)
	}
	return nil
}

// addGetOp is just a helper for addOperation - that is dedicated to operations of type GET.
func (mk *Maker) addGetOp(
	op *proto.Operation,
	modelOutputType graphql.Output,
	model *proto.Model,
	addTo *fieldsUnderConstruction) error {
	args, err := mk.makeArgs(op)
	if err != nil {
		return err
	}
	field := newFieldWithArgs(op.Name, args, modelOutputType, NewGetOperationResolver(op).Resolve)
	addTo.queries[op.Name] = field
	return nil
}

// outputTypeFor maps the type in the given proto.Field to a suitable graphql.Output type.
func (mk *Maker) outputTypeFor(field *proto.Field) (graphql.Output, error) {
	if outputType, ok := mk.isFieldTypeDirectlyMappableType(field.Type); ok {
		return outputType, nil
	}
	return nil, fmt.Errorf("cannot yet make output type for a: %v", field)
}

// inputTypeFor maps the type in the given proto.OperationInput to a suitable graphql.Input type.
func (mk *Maker) inputTypeFor(op *proto.OperationInput) (graphql.Input, error) {
	if inputType, ok := mk.isOperationInputTypeDirectlyMappableType(op.Type); ok {
		return inputType, nil
	}
	return nil, fmt.Errorf("cannot yet make input type for a: %v", op.Type)
}

// isFieldTypeDirectlyMappableType attempts to map the field type in the given proto.FieldType
// to a suitable built-in graphql.Output type. It returns a boolean to indicate if such a mapping
// can be found, and so, it also returns that mapped type.
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

// isOperationInputTypeDirectlyMappableType attempts to map the type in the given proto.OperationInputType
// to a suitable built-in graphql.Input type. It returns a boolean to indicate if such a mapping
// can be found, and so, it also returns that mapped type.
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

// makeArgs generates a graphql.FieldConfigArgument to reflect the inputs of the given
// proto.Operation - which can be used as the Args field in a graphql.Field.
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

// A fieldsUnderConstruction is a container to carry graphql.Fields and
// graphql.Type(s) that can be used later to compose a graphql.Schema.
// We intend the queries bucket to be the fields that should be added to
// the top level graphql Query. Simiarly for mutations. The models are different; these
// are graphql.Type(s) which are intented to populate the graphql.Schema's Types attribute.
type fieldsUnderConstruction struct {
	queries   graphql.Fields
	mutations graphql.Fields
	models    []graphql.Type
}
