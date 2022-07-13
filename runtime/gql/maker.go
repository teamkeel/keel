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
	proto    *proto.Schema
	query    *graphql.Object
	mutation *graphql.Object
	types    map[string]*graphql.Object
}

func newMaker(proto *proto.Schema) *maker {
	return &maker{
		proto: proto,
		query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Query",
			Fields: graphql.Fields{},
		}),
		mutation: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Mutation",
			Fields: graphql.Fields{},
		}),
		types: map[string]*graphql.Object{},
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

	for _, model := range modelInstances {
		for _, op := range model.Operations {
			switch op.Implementation {
			case proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO:
				err := mk.addOperation(model, op)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("operations with implementation %s are not supported", op.Implementation.String())
			}
		}
	}

	gSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: mk.query,

		// graphql won't accept a mutation object that has zero fields.
		Mutation: lo.Ternary(len(mk.mutation.Fields()) > 0, mk.mutation, nil),
	})
	if err != nil {
		return nil, err
	}

	return &gSchema, nil
}

// addModel generates the graphql type to represent the given proto.Model, and inserts it into
// the given fieldsUnderConstruction container.
func (mk *maker) addModel(model *proto.Model) (*graphql.Object, error) {
	if out, ok := mk.types[model.Name]; ok {
		return out, nil
	}

	object := graphql.NewObject(graphql.ObjectConfig{
		Name:   model.Name,
		Fields: graphql.Fields{},
	})
	mk.types[model.Name] = object

	for _, field := range model.Fields {
		outputType, err := mk.outputTypeFor(field)
		if err != nil {
			return nil, err
		}
		object.AddFieldConfig(field.Name, &graphql.Field{
			Name:    field.Name,
			Type:    outputType,
			Resolve: resolvers.NewFieldResolver(field).Resolve,
		})
	}

	return object, nil
}

// addOperation generates the graphql field object to represent the given proto.Operation
func (mk *maker) addOperation(model *proto.Model, op *proto.Operation) error {
	operationInputType, err := mk.makeInputType(op)
	if err != nil {
		return err
	}

	args := graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(operationInputType),
		},
	}

	outputType, err := mk.addModel(model)
	if err != nil {
		return err
	}

	resolver := resolvers.NewGetOperationResolver(op, model).Resolve

	field := &graphql.Field{
		Name:    op.Name,
		Args:    args,
		Type:    outputType,
		Resolve: resolver,
	}

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_GET:
		mk.query.AddFieldConfig(op.Name, field)
	case proto.OperationType_OPERATION_TYPE_CREATE:
		// create returns a non-null type
		field.Type = graphql.NewNonNull(field.Type)

		mk.mutation.AddFieldConfig(op.Name, field)
	default:
		return fmt.Errorf("addOperation() does not yet support this op.Type: %v", op.Type)
	}

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

// makeInputType generates an input type to reflect the inputs of the given
// proto.Operation - which can be used as the Args field in a graphql.Field.
func (mk *maker) makeInputType(op *proto.Operation) (*graphql.InputObject, error) {

	operationInputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   strings.ToUpper(op.Name[0:1]) + op.Name[1:] + "Input",
		Fields: graphql.InputObjectConfigFieldMap{},
	})

	for _, in := range op.Inputs {
		inputType, err := mk.inputTypeFor(in)
		if err != nil {
			return nil, err
		}

		operationInputType.AddFieldConfig(in.Name, &graphql.InputObjectFieldConfig{
			Type: inputType,
		})
	}

	return operationInputType, nil
}
