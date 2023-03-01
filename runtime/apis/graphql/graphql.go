package graphql

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/nleeper/goment"
	"github.com/sirupsen/logrus"

	"github.com/graphql-go/graphql"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/common"
)

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

func NewHandler(s *proto.Schema, api *proto.Api) common.ApiHandlerFunc {

	var schema *graphql.Schema
	var mutex sync.Mutex

	return func(r *http.Request) common.Response {

		// We lazily initialise the GraphQL schema as until there is actually
		// a GraphQL request to handle we don't need it. Also we don't want the
		// whole runtime to crash just because there is an issue with GraphQL.
		// The other API's (JSON-RPC, HTTP-JSON) may work fine.
		if schema == nil {
			mutex.Lock()
			var err error
			schema, err = NewGraphQLSchema(s, api)
			if err != nil {
				return common.Response{
					Status: http.StatusInternalServerError,
					Body:   []byte(`internal server error`),
				}
			}
			// This enables the graphql-go extension for tracing
			schema.AddExtensions(&Tracer{})

			mutex.Unlock()
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return common.Response{
				Status: http.StatusInternalServerError,
				// TODO: make this a valid GraphQL response
				Body: []byte(`internal server error`),
			}
		}

		var params GraphQLRequest
		err = json.Unmarshal(body, &params)
		if err != nil {
			return common.Response{
				Status: http.StatusBadRequest,
				// TODO: make this a valid GraphQL response
				Body: []byte(`invalid JSON body`),
			}
		}

		logrus.WithFields(logrus.Fields{
			"query": params.Query,
		}).Debug("graphql")

		// This map can be mutated in the action resolver,
		// allowing us to pass data upwards and into the response.
		headers := map[string][]string{}

		result := graphql.Do(graphql.Params{
			Schema:         *schema,
			Context:        r.Context(),
			RequestString:  params.Query,
			VariableValues: params.Variables,
			RootObject: map[string]interface{}{
				"headers": headers,
			},
		})

		return common.NewJsonResponse(http.StatusOK, result, headers)
	}
}

// NewGraphQLSchema creates a map of graphql.Schema objects where the keys
// are the API names from the provided proto.Schema
func NewGraphQLSchema(proto *proto.Schema, api *proto.Api) (*graphql.Schema, error) {
	m := &graphqlSchemaBuilder{
		proto: proto,
		query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Query",
			Fields: graphql.Fields{},
		}),
		mutation: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Mutation",
			Fields: graphql.Fields{},
		}),
		inputs: map[string]*graphql.InputObject{},
		types:  map[string]*graphql.Object{},
		enums:  map[string]*graphql.Enum{},
	}

	return m.build(api, proto)
}

// A graphqlSchemaBuilder exposes a Make method, that makes a set of graphql.Schema objects - one for each
// of the APIs defined in the keel schema provided at construction time.
type graphqlSchemaBuilder struct {
	proto    *proto.Schema
	query    *graphql.Object
	mutation *graphql.Object
	inputs   map[string]*graphql.InputObject
	types    map[string]*graphql.Object
	enums    map[string]*graphql.Enum
}

// build returns a graphql.Schema that implements the given API.
func (mk *graphqlSchemaBuilder) build(api *proto.Api, schema *proto.Schema) (*graphql.Schema, error) {
	// The graphql top level query contents will be comprised ONLY of the
	// OPERATIONS from the keel schema. But to find these we have to traverse the
	// schema, first by model, then by said model's operations. As a side effect
	// we must define graphl types for the models involved.

	namesOfModelsUsedByAPI := lo.Map(api.ApiModels, func(m *proto.ApiModel, _ int) string {
		return m.ModelName
	})

	modelInstances := proto.FindModels(mk.proto.Models, namesOfModelsUsedByAPI)

	hasNoQueryOps := true
	for _, model := range modelInstances {
		for _, op := range model.Operations {
			err := mk.addOperation(op, schema)
			if err != nil {
				return nil, err
			}
			if op.Type == proto.OperationType_OPERATION_TYPE_GET || op.Type == proto.OperationType_OPERATION_TYPE_LIST {
				hasNoQueryOps = false
			}
		}
	}

	// The graphql handler cannot manage an empty query object,
	// so if there are no get or list ops, we add the __Empty field
	if hasNoQueryOps {
		mk.query.AddFieldConfig("__Empty", &graphql.Field{
			Type: graphql.Boolean,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return true, nil
			},
		})
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
func (mk *graphqlSchemaBuilder) addModel(model *proto.Model) (*graphql.Object, error) {
	if out, ok := mk.types[fmt.Sprintf("model-%s", model.Name)]; ok {
		return out, nil
	}

	object := graphql.NewObject(graphql.ObjectConfig{
		Name:   model.Name,
		Fields: graphql.Fields{},
	})

	mk.types[fmt.Sprintf("model-%s", model.Name)] = object

	for _, field := range model.Fields {
		field := field

		// Passwords are omitted from GraphQL responses
		if field.Type.Type == proto.Type_TYPE_PASSWORD {
			continue
		}

		outputType, err := mk.outputTypeFor(field)
		if err != nil {
			return nil, err
		}

		if field.Type.Type != proto.Type_TYPE_MODEL {
			object.AddFieldConfig(field.Name, &graphql.Field{
				Name: field.Name,
				Type: outputType,
			})
			continue
		}

		fieldArgs := graphql.FieldConfigArgument{}
		if proto.IsHasMany(field) {
			fieldArgs = graphql.FieldConfigArgument{
				"first": &graphql.ArgumentConfig{
					Type:        graphql.Int,
					Description: "The requested number of nodes for each page.",
				},
				"last": &graphql.ArgumentConfig{
					Type:        graphql.Int,
					Description: "The requested number of nodes for each page.",
				},
				"after": &graphql.ArgumentConfig{
					Type:        graphql.String,
					Description: "The ID cursor to retrieve nodes after in the connection. Typically, you should pass the endCursor of the previous page as after.",
				},
				"before": &graphql.ArgumentConfig{
					Type:        graphql.String,
					Description: "The ID cursor to retrieve nodes before in the connection. Typically, you should pass the startCursor of the previous page as before.",
				},
			}
		}

		object.AddFieldConfig(field.Name, &graphql.Field{
			Name: field.Name,
			Type: outputType,
			Args: fieldArgs,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {

				relatedModel := proto.FindModel(mk.proto.Models, field.Type.ModelName.Value)

				// Create a new query for the related model
				query := actions.NewQuery(relatedModel)
				query.AppendSelect(actions.AllFields())

				foreignKeyField := proto.GetForignKeyFieldName(mk.proto.Models, field)

				// Get the value of the model
				parent, ok := p.Source.(map[string]interface{})
				if !ok {
					return nil, errors.New("graphql source value is not map[string]interface{}")
				}

				// Depending on the relationship type we either need the primary key of this
				// model or a foreign key
				parentLookupField := "id"
				if proto.IsBelongsTo(field) {
					parentLookupField = foreignKeyField
				}

				// Retrieve the value for the lookup
				parentFieldValue, ok := parent[parentLookupField]
				if !ok {
					return nil, fmt.Errorf("model %s did not have field %s", model.Name, parentLookupField)
				}

				// If the value is null (possible if the relationship is not required), then there
				// is no need for a lookup.
				if parentFieldValue == nil {
					return nil, nil
				}

				var leftOperand *actions.QueryOperand
				if proto.IsBelongsTo(field) {
					leftOperand = actions.IdField()
				} else {
					leftOperand = actions.Field(foreignKeyField)
				}

				err = query.Where(leftOperand, actions.Equals, actions.Value(parentFieldValue))
				if err != nil {
					return nil, err
				}

				switch {
				case proto.IsBelongsTo(field), proto.IsHasOne(field):
					result, err := query.
						SelectStatement().
						ExecuteToSingle(p.Context)
					if err != nil {
						return nil, err
					}

					// Return an error if no record if found for the corresponding foreign key
					if result == nil {
						return nil, errors.New("record expected in database but nothing found")
					}

					return result, nil
				case proto.IsHasMany(field):
					page, err := actions.ParsePage(p.Args)
					if err != nil {
						return nil, err
					}

					// Select all columns from this table and distinct on id
					query.AppendDistinctOn(actions.IdField())
					query.AppendSelect(actions.AllFields())
					err = query.ApplyPaging(page)
					if err != nil {
						return nil, err
					}

					results, _, hasNextPage, err := query.
						SelectStatement().
						ExecuteToMany(p.Context)

					if err != nil {
						return nil, err
					}

					res, err := connectionResponse(map[string]any{
						"results":     results,
						"hasNextPage": hasNextPage,
					})

					if err != nil {
						return nil, err
					}

					return res, nil
				default:
					return nil, fmt.Errorf("unhandled model relationship configuration for field: %s on model: %s", field.Name, field.ModelName)
				}
			},
		})
	}

	return object, nil
}

// addOperation generates the graphql field object to represent the given proto.Operation
func (mk *graphqlSchemaBuilder) addOperation(
	op *proto.Operation,
	schema *proto.Schema) error {

	model := proto.FindModel(schema.Models, op.ModelName)
	modelType, err := mk.addModel(model)
	if err != nil {
		return err
	}

	field := &graphql.Field{
		Name: op.Name,
	}

	operationInputType, allOptionalInputs, err := mk.makeOperationInputType(op)
	if err != nil {
		return err
	}

	// Only add input args if an input field exists.
	if len(operationInputType.Fields()) > 0 {
		if allOptionalInputs {
			// Input field is optional if all its fields are optional
			field.Args = graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{
					Type: operationInputType,
				},
			}
		} else {
			// Input field is required if any of its fields are required
			field.Args = graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(operationInputType),
				},
			}
		}
	}

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_GET:
		field.Type = modelType
		mk.query.AddFieldConfig(op.Name, field)
	case proto.OperationType_OPERATION_TYPE_CREATE,
		proto.OperationType_OPERATION_TYPE_UPDATE:
		field.Type = graphql.NewNonNull(modelType)
		mk.mutation.AddFieldConfig(op.Name, field)
	case proto.OperationType_OPERATION_TYPE_DELETE:
		field.Type = deleteResponseType
		mk.mutation.AddFieldConfig(op.Name, field)
	case proto.OperationType_OPERATION_TYPE_LIST:
		// for list types we need to wrap the output type in the
		// connection type which allows for pagination
		field.Type = mk.makeConnectionType(modelType)
		mk.query.AddFieldConfig(op.Name, field)
	case proto.OperationType_OPERATION_TYPE_READ:
		responseMessage := proto.FindMessage(schema.Messages, op.ResponseMessageName)
		if responseMessage == nil {
			return fmt.Errorf("response message does not exist: %s", op.ResponseMessageName)
		}
		field.Type, err = mk.outputTypeFromMessage(responseMessage)
		if err != nil {
			return err
		}
		mk.query.AddFieldConfig(op.Name, field)
	case proto.OperationType_OPERATION_TYPE_WRITE:
		responseMessage := proto.FindMessage(schema.Messages, op.ResponseMessageName)
		if responseMessage == nil {
			return fmt.Errorf("response message does not exist: %s", op.ResponseMessageName)
		}
		field.Type, err = mk.outputTypeFromMessage(responseMessage)
		if err != nil {
			return err
		}
		mk.mutation.AddFieldConfig(op.Name, field)
	default:
		return fmt.Errorf("addOperation() does not yet support this op.Type: %v", op.Type)
	}

	field.Resolve = ActionFunc(schema, op)

	return nil
}

func (mk *graphqlSchemaBuilder) makeConnectionType(itemType graphql.Output) graphql.Output {
	if out, found := mk.types[fmt.Sprintf("connection-%s", itemType.Name())]; found {
		return graphql.NewNonNull(out)
	}

	edgeType := graphql.NewObject(graphql.ObjectConfig{
		Name: itemType.Name() + "Edge",
		Fields: graphql.Fields{
			"node": &graphql.Field{
				Type: graphql.NewNonNull(
					itemType,
				),
			},
		},
	})

	connection := graphql.NewObject(graphql.ObjectConfig{
		Name: itemType.Name() + "Connection",
		Fields: graphql.Fields{
			"edges": &graphql.Field{
				Type: graphql.NewNonNull(
					graphql.NewList(
						graphql.NewNonNull(edgeType),
					),
				),
			},
			"pageInfo": &graphql.Field{
				Type: graphql.NewNonNull(pageInfoType),
			},
		},
	})

	mk.types[fmt.Sprintf("connection-%s", itemType.Name())] = connection

	return graphql.NewNonNull(connection)
}

func (mk *graphqlSchemaBuilder) addEnum(e *proto.Enum) *graphql.Enum {
	if out, ok := mk.enums[e.Name]; ok {
		return out
	}

	values := graphql.EnumValueConfigMap{}

	for _, v := range e.Values {
		values[v.Name] = &graphql.EnumValueConfig{
			Value: v.Name,
		}
	}

	enum := graphql.NewEnum(graphql.EnumConfig{
		Name:   e.Name,
		Values: values,
	})
	mk.enums[e.Name] = enum
	return enum
}

var fromNowType = graphql.Field{
	Name: "fromNow",
	Type: graphql.NewNonNull(graphql.String),
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		t, ok := p.Source.(time.Time)

		if !ok {
			return nil, fmt.Errorf("not a valid time")
		}

		g, err := goment.New(t)

		if err != nil {
			return nil, err
		}

		return g.FromNow(), nil
	},
}

// outputTypeFromMessage makes a graphql.Object response output from a proto.Message.
func (mk *graphqlSchemaBuilder) outputTypeFromMessage(message *proto.Message) (graphql.Output, error) {
	if out, ok := mk.types[message.Name]; ok {
		return out, nil
	}

	output := graphql.NewObject(graphql.ObjectConfig{
		Name:   message.Name,
		Fields: graphql.Fields{},
	})

	for _, field := range message.Fields {
		var fieldType graphql.Output

		switch field.Type.Type {
		case proto.Type_TYPE_MESSAGE:
			fieldMessage := proto.FindMessage(mk.proto.Messages, field.Type.MessageName.Value)

			var err error
			fieldType, err = mk.outputTypeFromMessage(fieldMessage)
			if err != nil {
				return nil, err
			}
		case proto.Type_TYPE_MODEL:
			// todo: https://linear.app/keel/issue/BLD-319/model-type-field-in-message-type
			return nil, errors.New("not supporting nested models just yet")
		default:
			fieldType = protoTypeToGraphQLOutput[field.Type.Type]
			if fieldType == nil {
				return nil, fmt.Errorf("cannot yet make output type for: %s", field.Type.Type.String())
			}
		}

		if !field.Optional {
			fieldType = graphql.NewNonNull(fieldType)
		}

		output.AddFieldConfig(field.Name, &graphql.Field{
			Type: fieldType,
		})
	}

	mk.types[message.Name] = output

	return output, nil
}

// outputTypeFor maps the type in the given proto.Field to a suitable graphql.Output type.
func (mk *graphqlSchemaBuilder) outputTypeFor(field *proto.Field) (out graphql.Output, err error) {
	switch field.Type.Type {
	case proto.Type_TYPE_ENUM:
		for _, e := range mk.proto.Enums {
			if e.Name == field.Type.EnumName.Value {
				out = mk.addEnum(e)
				break
			}
		}
	case proto.Type_TYPE_MODEL:
		for _, m := range mk.proto.Models {
			if m.Name == field.Type.ModelName.Value {
				out, err = mk.addModel(m)
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

	if err != nil {
		return out, err
	}

	if field.Type.Repeated {
		if field.Type.Type == proto.Type_TYPE_MODEL {
			out = mk.makeConnectionType(out)
		} else {
			out = graphql.NewList(out)
			out = graphql.NewNonNull(out)
		}
	} else if !field.Optional {
		out = graphql.NewNonNull(out)
	}

	return out, nil
}

// inputTypeFromMessageField maps the type in the given proto.MessageField to a suitable graphql.Input type.
func (mk *graphqlSchemaBuilder) inputTypeFromMessageField(field *proto.MessageField, op *proto.Operation) (graphql.Input, error) {
	var in graphql.Input

	switch {
	case field.Type.Type == proto.Type_TYPE_MESSAGE:
		inputObjectName := field.Type.MessageName.Value
		message := proto.FindMessage(mk.proto.Messages, inputObjectName)

		if len(message.Fields) == 0 {
			break
		}

		if out, ok := mk.inputs[inputObjectName]; ok {
			in = out
			break
		}

		inputObject := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   inputObjectName,
			Fields: graphql.InputObjectConfigFieldMap{},
		})

		for _, input := range message.Fields {
			inputField, err := mk.inputTypeFromMessageField(input, op)
			if err != nil {
				return nil, err
			}

			inputObject.AddFieldConfig(input.Name, &graphql.InputObjectFieldConfig{
				Type: inputField,
			})
		}

		mk.inputs[inputObjectName] = inputObject

		in = inputObject
	default:
		var err error
		if in, err = mk.inputTypeFor(field); err != nil {
			return nil, err
		}
	}

	if !field.Optional {
		in = graphql.NewNonNull(in)
	}

	if field.Type.Repeated {
		in = graphql.NewList(in)
		if !field.Optional {
			in = graphql.NewNonNull(in)
		}
	}

	return in, nil
}

// inputTypeFor creates a graphql.Input for non-list operation input types.
func (mk *graphqlSchemaBuilder) inputTypeFor(field *proto.MessageField) (graphql.Input, error) {
	var in graphql.Input
	if field.Type.Type == proto.Type_TYPE_ENUM {
		enum, _ := lo.Find(mk.proto.Enums, func(e *proto.Enum) bool {
			return e.Name == field.Type.EnumName.Value
		})
		in = mk.addEnum(enum)
	} else {
		var ok bool
		if in, ok = protoTypeToGraphQLInput[field.Type.Type]; !ok {
			return nil, fmt.Errorf("message %s has unsupported message field type: %s", field.MessageName, field.Type.Type.String())
		}
	}
	return in, nil
}

// makeOperationInputType generates an input type to reflect the inputs of the given
// proto.Operation - which can be used as the Args field in a graphql.Field.
func (mk *graphqlSchemaBuilder) makeOperationInputType(op *proto.Operation) (*graphql.InputObject, bool, error) {
	message := proto.FindMessage(mk.proto.Messages, op.InputMessageName)
	allOptionalInputs := true

	inputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   message.Name,
		Fields: graphql.InputObjectConfigFieldMap{},
	})

	for _, field := range message.Fields {
		fieldType, err := mk.inputTypeFromMessageField(field, op)
		if err != nil {
			return nil, false, err
		}

		if fieldType != nil {
			if !field.Optional {
				allOptionalInputs = false
			}
			inputType.AddFieldConfig(field.Name, &graphql.InputObjectFieldConfig{
				Type: fieldType,
			})
		}
	}

	return inputType, allOptionalInputs, nil
}
