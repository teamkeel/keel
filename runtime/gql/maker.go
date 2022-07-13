package gql

import (
	"fmt"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/gql/resolvers"
)

// A maker exposes a Make method, that makes a set of graphql.Schema objects - one for each
// of the APIs defined in the keel schema provided at construction time.
type maker struct {
	proto *proto.Schema
}

func newMaker(proto *proto.Schema) *maker {
	return &maker{
		proto: proto,
	}
}

// The graphql.Schema(s) are returned in a map, keyed on the name of the
// API name they belong to.
func (mk *maker) make() (map[string]*graphql.Schema, error) {
	outputSchemas := map[string]*graphql.Schema{}
	for _, api := range mk.proto.Apis {
		gSchema, err := mk.newSchema(api)
		if err != nil {
			return nil, err
		}
		outputSchemas[api.Name] = gSchema
	}
	return outputSchemas, nil
}

// newSchema returns a graphql.Schema that implements the given API.
func (mk *maker) newSchema(api *proto.Api) (*graphql.Schema, error) {
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

	gSchema, err := graphql.NewSchema(
		graphql.SchemaConfig{
			Query: newObject("Query", fieldsUnderConstruction.queries),
			// graphql won't accept a mutation object that has zero fields.
			Mutation: lo.Ternary(len(fieldsUnderConstruction.mutations) > 0, newObject("Mutation", fieldsUnderConstruction.mutations), nil),
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
func (mk *maker) addModel(model *proto.Model, addTo *fieldsUnderConstruction) (modelOutputType graphql.Output, err error) {
	// todo - don't add it, if we already did earlier
	fields := graphql.Fields{}
	for _, field := range model.Fields {
		outputType, err := mk.outputTypeFor(field)
		if err != nil {
			return nil, err
		}
		field := newField(field.Name, outputType, resolvers.NewFieldResolver(field).Resolve)
		fields[field.Name] = field
	}
	modelOutputType = newObject(model.Name, fields)
	addTo.models = append(addTo.models, modelOutputType)
	return modelOutputType, nil
}

// addOperation generates the graphql field object to represent the given proto.Operation.
// This field will eventually live in the top level graphql Query type, but at this stage
// the function just accumulates them in the given fieldsUnderConstruction container.
func (mk *maker) addOperation(
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
	case proto.OperationType_OPERATION_TYPE_CREATE:
		if err := mk.addCreateOp(op, modelOutputType, model, addTo); err != nil {
			return err
		}
	default:
		return fmt.Errorf("addOperation() does not yet support this op.Type: %v", op.Type)
	}
	return nil
}

// addGetOp is just a helper for addOperation - that is dedicated to operations of type GET.
func (mk *maker) addGetOp(
	op *proto.Operation,
	modelOutputType graphql.Output,
	model *proto.Model,
	addTo *fieldsUnderConstruction) error {

	operationInputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   strings.ToUpper(op.Name[0:1]) + op.Name[1:] + "Input",
		Fields: graphql.InputObjectConfigFieldMap{},
	})

	for _, in := range op.Inputs {
		inputType, err := mk.inputTypeFor(in)
		if err != nil {
			return err
		}

		operationInputType.AddFieldConfig(in.Name, &graphql.InputObjectFieldConfig{
			Type: inputType,
		})
	}

	args := graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(operationInputType),
		},
	}

	field := newFieldWithArgs(op.Name, args, modelOutputType, resolvers.NewGetOperationResolver(op, model).Resolve)
	addTo.queries[op.Name] = field
	return nil
}

// addCreateOp is just a helper for addOperation - that is dedicated to operations of type CREATE.
func (mk *maker) addCreateOp(
	// todo see if the family of addXXXOp methods diverge more than they do now - and if so DRY up the code.
	// At the moment the only difference between the two we have is which element of the addTo container
	// they append to.
	op *proto.Operation,
	modelOutputType graphql.Output,
	model *proto.Model,
	addTo *fieldsUnderConstruction) error {
	args, err := mk.makeArgs(op)
	if err != nil {
		return err
	}
	field := newFieldWithArgs(
		op.Name,
		args, modelOutputType,
		resolvers.NewCreateOperationResolver(op, model).Resolve)
	addTo.mutations[op.Name] = field
	return nil
}

func (mk *maker) addEnum(e *proto.Enum) *graphql.Enum {
	values := graphql.EnumValueConfigMap{}

	for _, v := range e.Values {
		values[v.Name] = &graphql.EnumValueConfig{
			Value: v.Name,
		}
	}

	return graphql.NewEnum(graphql.EnumConfig{
		Name:   e.Name,
		Values: values,
	})
}

var timestampType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Timestamp",
	Fields: graphql.Fields{
		"seconds": &graphql.Field{
			Name: "seconds",
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(graphql.ResolveParams) (interface{}, error) {
				// TODO: implement this
				panic("not implemented")
			},
		},
		// TODO: add `fromNow` and `formatted` fields
	},
})

var protoTypeToGraphQLOutput = map[proto.Type]graphql.Output{
	proto.Type_TYPE_ID:       graphql.ID,
	proto.Type_TYPE_STRING:   graphql.String,
	proto.Type_TYPE_INT:      graphql.Int,
	proto.Type_TYPE_BOOL:     graphql.Boolean,
	proto.Type_TYPE_DATETIME: timestampType,
}

// outputTypeFor maps the type in the given proto.Field to a suitable graphql.Output type.
func (mk *maker) outputTypeFor(field *proto.Field) (out graphql.Output, err error) {
	switch field.Type.Type {

	case proto.Type_TYPE_ENUM:
		for _, e := range mk.proto.Enums {
			if e.Name == field.Type.EnumName.Value {
				out = mk.addEnum(e)
				break
			}
		}

	default:
		var ok bool
		out, ok = protoTypeToGraphQLOutput[field.Type.Type]
		if !ok {
			return out, fmt.Errorf("cannot yet make output type for: %s", field.Type.Type.String())
		}
	}

	if !field.Optional {
		out = graphql.NewNonNull(out)
	}

	if field.Type.Repeated {
		out = graphql.NewList(out)
		out = graphql.NewNonNull(out)
	}

	return out, nil
}

var protoTypeToGraphQLInput = map[proto.Type]graphql.Input{
	proto.Type_TYPE_ID:     graphql.ID,
	proto.Type_TYPE_STRING: graphql.String,
	proto.Type_TYPE_INT:    graphql.Int,
	proto.Type_TYPE_BOOL:   graphql.Boolean,
}

// inputTypeFor maps the type in the given proto.OperationInput to a suitable graphql.Input type.
func (mk *maker) inputTypeFor(op *proto.OperationInput) (graphql.Input, error) {
	var in graphql.Input

	in, ok := protoTypeToGraphQLInput[op.Type.Type]
	if !ok {
		return nil, fmt.Errorf("cannot yet make input type for a: %v", op.Type)
	}

	if !op.Optional {
		in = graphql.NewNonNull(in)
	}

	if op.Type.Repeated {
		in = graphql.NewList(in)
		in = graphql.NewNonNull(in)
	}

	return in, nil
}

// makeArgs generates a graphql.FieldConfigArgument to reflect the inputs of the given
// proto.Operation - which can be used as the Args field in a graphql.Field.
func (mk *maker) makeArgs(op *proto.Operation) (graphql.FieldConfigArgument, error) {
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
