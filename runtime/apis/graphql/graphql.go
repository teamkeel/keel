package graphql

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/samber/lo"
	"github.com/teamkeel/graphql"
	"github.com/teamkeel/graphql/gqlerrors"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/locale"
	"github.com/teamkeel/keel/schema/parser"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/apis/graphql")

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

func NewHandler(s *proto.Schema, api *proto.Api) common.HandlerFunc {
	var schema *graphql.Schema
	var mutex sync.Mutex

	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "GraphQL")
		defer span.End()

		identity, err := actions.HandleAuthorizationHeader(ctx, s, r.Header)
		if err != nil {
			var extensions map[string]interface{}

			var runtimeErr common.RuntimeError
			if errors.As(err, &runtimeErr) {
				extensions = runtimeErr.Extensions()
			}

			return common.NewJsonResponse(http.StatusOK, graphql.Result{
				Errors: []gqlerrors.FormattedError{
					{
						Message:    "authentication failed",
						Extensions: extensions,
					},
				},
			}, nil)
		}
		if identity != nil {
			ctx = auth.WithIdentity(ctx, identity)
		}

		// handle any Time-Zone headers
		location, err := locale.HandleTimezoneHeader(ctx, r.Header)
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
			return common.NewJsonResponse(http.StatusBadRequest, graphql.Result{
				Errors: []gqlerrors.FormattedError{
					{
						Message: fmt.Sprintf("error setting timezone: %s", err.Error()),
					},
				},
			}, nil)
		}
		ctx = locale.WithTimeLocation(ctx, location)

		// We lazily initialise the GraphQL schema as until there is actually
		// a GraphQL request to handle we don't need it. Also we don't want the
		// whole runtime to crash just because there is an issue with GraphQL.
		// The other API's (JSON-RPC, HTTP-JSON) may work fine.
		if schema == nil {
			mutex.Lock()
			defer mutex.Unlock()

			var err error
			schema, err = NewGraphQLSchema(s, api)
			if err != nil {
				span.RecordError(err, trace.WithStackTrace(true))
				span.SetStatus(codes.Error, err.Error())
				return common.NewJsonResponse(http.StatusInternalServerError, graphql.Result{
					Errors: []gqlerrors.FormattedError{
						{
							Message: fmt.Sprintf("error initialising GraphQL: %s", err.Error()),
						},
					},
				}, nil)
			}
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
			return common.NewJsonResponse(http.StatusBadRequest, graphql.Result{
				Errors: []gqlerrors.FormattedError{
					{
						Message: fmt.Sprintf("error reading body: %s", err.Error()),
					},
				},
			}, nil)
		}

		var params GraphQLRequest
		err = json.Unmarshal(body, &params)
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
			return common.NewJsonResponse(http.StatusBadRequest, graphql.Result{
				Errors: []gqlerrors.FormattedError{
					{
						Message: fmt.Sprintf("invalid request: %s", err.Error()),
					},
				},
			}, nil)
		}

		span.SetAttributes(
			attribute.String("params.query", params.Query),
			attribute.String("params.operationName", params.OperationName),
			attribute.String("api.protocol", "GraphQL"),
		)

		logrus.WithFields(logrus.Fields{
			"query": params.Query,
		}).Debug("graphql")

		// This map can be mutated in the action resolver,
		// allowing us to pass data upwards and into the response.
		headers := map[string][]string{}

		result := graphql.Do(graphql.Params{
			Schema:         *schema,
			Context:        ctx,
			RequestString:  params.Query,
			VariableValues: params.Variables,
			RootObject: map[string]interface{}{
				"headers": headers,
			},
		})

		if result.HasErrors() {
			messages := []string{}
			attr := []attribute.KeyValue{}
			for i, err := range result.Errors {
				messages = append(messages, err.Message)
				attr = append(attr, attribute.String(fmt.Sprintf("error.%d", i), err.Message))
			}
			span.AddEvent("errors", trace.WithAttributes(attr...))
			span.SetStatus(codes.Error, strings.Join(messages, ", "))
		}

		return common.NewJsonResponse(http.StatusOK, result, &common.ResponseMetadata{
			Headers: headers,
		})
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
		types:  make(map[string]graphql.Type),
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
	types    map[string]graphql.Type
	enums    map[string]*graphql.Enum
	globals  map[string]*graphql.Scalar
}

// build returns a graphql.Schema that implements the given API.
func (mk *graphqlSchemaBuilder) build(api *proto.Api, schema *proto.Schema) (*graphql.Schema, error) {
	for _, actionName := range proto.GetActionNamesForApi(schema, api) {
		action := schema.FindAction(actionName)
		err := mk.addAction(action, schema)
		if err != nil {
			return nil, err
		}
	}

	mk.addGlobals()

	// The graphql handler cannot manage an empty query object,
	// so without _health everything would blow up.
	mk.query.AddFieldConfig("_health", &graphql.Field{
		Type: graphql.Boolean,
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			return true, nil
		},
	})

	types := []graphql.Type{}

	for _, global := range mk.globals {
		types = append(types, global)
	}

	mutation := lo.Ternary(len(mk.mutation.Fields()) > 0, mk.mutation, nil)

	gSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: mk.query,
		Types: types,
		// graphql won't accept a mutation object that has zero fields.
		Mutation: mutation,
	})
	if err != nil {
		return nil, err
	}

	return &gSchema, nil
}

func (mk *graphqlSchemaBuilder) addGlobals() {
	mk.globals = map[string]*graphql.Scalar{}
	mk.globals[anyType.Name()] = anyType
	mk.globals[timestampType.Name()] = timestampInputType
}

// addModel generates the graphql type to represent the given proto.Model, and inserts it into
// mk.types
func (mk *graphqlSchemaBuilder) addModel(model *proto.Model) (*graphql.Object, error) {
	if out, ok := mk.types[fmt.Sprintf("model-%s", model.Name)]; ok {
		return out.(*graphql.Object), nil
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

		// Vectors are omitted from GraphQL responses
		if field.Type.Type == proto.Type_TYPE_VECTOR {
			continue
		}

		outputType, err := mk.outputTypeForModelField(field)
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
		if field.IsHasMany() {
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
				ctx, span := tracer.Start(p.Context, fmt.Sprintf("Resolve %s.%s", model.Name, field.Name))
				defer span.End()

				relatedModel := mk.proto.FindModel(field.Type.ModelName.Value)

				// Create a new query for the related model
				query := actions.NewQuery(relatedModel)
				query.Select(actions.AllFields())

				foreignKeyField := proto.GetForeignKeyFieldName(mk.proto.Models, field)

				// Get the value of the model
				parent, ok := p.Source.(map[string]interface{})
				if !ok {
					return nil, errors.New("graphql source value is not map[string]interface{}")
				}

				// Depending on the relationship type we either need the primary key of this
				// model or a foreign key
				parentLookupField := "id"
				if field.IsBelongsTo() {
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
				if field.IsBelongsTo() {
					leftOperand = actions.IdField()
				} else {
					leftOperand = actions.Field(foreignKeyField)
				}

				err = query.Where(leftOperand, actions.Equals, actions.Value(parentFieldValue))
				if err != nil {
					return nil, err
				}

				scope := actions.NewModelScope(ctx, relatedModel, mk.proto)

				switch {
				case field.IsBelongsTo(), field.IsHasOne():
					result, err := query.
						SelectStatement().
						ExecuteToSingle(ctx)
					if err != nil {
						span.RecordError(err, trace.WithStackTrace(true))
						span.SetStatus(codes.Error, err.Error())
						return nil, err
					}

					authorised, err := actions.AuthoriseForActionType(scope, proto.ActionType_ACTION_TYPE_GET, []map[string]any{result})
					if err != nil {
						span.RecordError(err, trace.WithStackTrace(true))
						span.SetStatus(codes.Error, err.Error())
						return nil, err
					}

					if !authorised {
						return nil, common.NewPermissionError()
					}

					// Return an error if no record if found for the corresponding foreign key
					if result == nil {
						return nil, errors.New("record expected in database but nothing found")
					}

					return result, nil
				case field.IsHasMany():
					page, err := actions.ParsePage(p.Args)
					if err != nil {
						span.RecordError(err, trace.WithStackTrace(true))
						span.SetStatus(codes.Error, err.Error())
						return nil, err
					}

					// Select all columns from this table and distinct on id
					query.DistinctOn(actions.IdField())
					query.Select(actions.AllFields())
					err = query.ApplyPaging(page)
					if err != nil {
						span.RecordError(err, trace.WithStackTrace(true))
						span.SetStatus(codes.Error, err.Error())
						return nil, err
					}

					results, resultInfo, pageInfo, err := query.
						SelectStatement().
						ExecuteToMany(ctx, &page)
					if err != nil {
						span.RecordError(err, trace.WithStackTrace(true))
						span.SetStatus(codes.Error, err.Error())
						return nil, err
					}

					authorised, err := actions.AuthoriseForActionType(scope, proto.ActionType_ACTION_TYPE_LIST, results)
					if err != nil {
						span.RecordError(err, trace.WithStackTrace(true))
						span.SetStatus(codes.Error, err.Error())
						return nil, err
					}

					if !authorised {
						return nil, common.NewPermissionError()
					}

					res, err := connectionResponse(map[string]any{
						"results":    results,
						"resultInfo": resultInfo,
						"pageInfo":   pageInfo.ToMap(),
					})
					if err != nil {
						span.RecordError(err, trace.WithStackTrace(true))
						span.SetStatus(codes.Error, err.Error())
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

// addOperation generates the graphql field object to represent the given proto.Action
func (mk *graphqlSchemaBuilder) addAction(
	action *proto.Action,
	schema *proto.Schema) error {
	model := schema.FindModel(action.ModelName)
	modelType, err := mk.addModel(model)
	if err != nil {
		return err
	}

	field := &graphql.Field{
		Name: action.Name,
	}

	operationInputType, allOptionalInputs, err := mk.makeActionInputType(action)
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
	} else if action.InputMessageName == parser.MessageFieldTypeAny {
		// Any type
		field.Args = graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: anyType,
			},
		}
	}

	switch action.Type {
	case proto.ActionType_ACTION_TYPE_GET:
		field.Type = modelType
		mk.query.AddFieldConfig(action.Name, field)
	case proto.ActionType_ACTION_TYPE_CREATE,
		proto.ActionType_ACTION_TYPE_UPDATE:
		field.Type = graphql.NewNonNull(modelType)
		mk.mutation.AddFieldConfig(action.Name, field)
	case proto.ActionType_ACTION_TYPE_DELETE:
		field.Type = deleteResponseType
		mk.mutation.AddFieldConfig(action.Name, field)
	case proto.ActionType_ACTION_TYPE_LIST:
		// for list types we need to wrap the output type in the
		// connection type which allows for pagination
		field.Type = mk.makeConnectionType(modelType)
		mk.query.AddFieldConfig(action.Name, field)
	case proto.ActionType_ACTION_TYPE_READ:
		responseMessage := schema.FindMessage(action.ResponseMessageName)
		if responseMessage == nil {
			return fmt.Errorf("response message does not exist: %s", action.ResponseMessageName)
		}

		if responseMessage.Name == parser.MessageFieldTypeAny {
			field.Type = anyType
		} else {
			field.Type, err = mk.addMessage(responseMessage)
			if err != nil {
				return err
			}
		}

		mk.query.AddFieldConfig(action.Name, field)
	case proto.ActionType_ACTION_TYPE_WRITE:
		responseMessage := schema.FindMessage(action.ResponseMessageName)
		if responseMessage == nil {
			return fmt.Errorf("response message does not exist: %s", action.ResponseMessageName)
		}
		field.Type, err = mk.addMessage(responseMessage)
		if err != nil {
			return err
		}
		mk.mutation.AddFieldConfig(action.Name, field)
	default:
		return fmt.Errorf("addAction() does not yet support this action type: %v", action.Type)
	}

	field.Resolve = ActionFunc(schema, action)

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

// addMessage makes a graphql.Object response output from a proto.Message.
func (mk *graphqlSchemaBuilder) addMessage(message *proto.Message) (graphql.Output, error) {
	if message.Name == parser.MessageFieldTypeAny {
		return anyType, nil
	}
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
			fieldMessage := mk.proto.FindMessage(field.Type.MessageName.Value)

			var err error
			fieldType, err = mk.addMessage(fieldMessage)
			if err != nil {
				return nil, err
			}
		case proto.Type_TYPE_MODEL:
			modelMessage := mk.proto.FindModel(field.Type.ModelName.Value)

			var err error
			fieldType, err = mk.addModel(modelMessage)
			if err != nil {
				return nil, err
			}
		case proto.Type_TYPE_ENUM:
			enumMessage := proto.FindEnum(mk.proto.Enums, field.Type.EnumName.Value)
			fieldType = mk.addEnum(enumMessage)

		default:
			fieldType = protoTypeToGraphQLOutput[field.Type.Type]
			if fieldType == nil {
				return nil, fmt.Errorf("cannot yet make output type for: %s", field.Type.Type.String())
			}
		}

		if field.Type.Repeated {
			fieldType = graphql.NewList(fieldType)
		}
		if !field.Optional {
			fieldType = graphql.NewNonNull(fieldType)
		}

		output.AddFieldConfig(field.Name, &graphql.Field{
			Type: fieldType,
		})
	}

	if len(message.Fields) == 0 {
		output.AddFieldConfig("success", &graphql.Field{
			Type:    graphql.Boolean,
			Resolve: func(_ graphql.ResolveParams) (interface{}, error) { return true, nil },
		})
	}

	mk.types[message.Name] = output

	return output, nil
}

// addModel generates the graphql type to represent the given proto.Model, and inserts it into
// mk.types
func (mk *graphqlSchemaBuilder) addModelInput(model *proto.Model) (graphql.Input, error) {
	if in, ok := mk.types[fmt.Sprintf("model-%s", model.Name)]; ok {
		return in.(graphql.Input), nil
	}

	input := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   model.Name,
		Fields: graphql.InputObjectConfigFieldMap{},
	})

	for _, field := range model.Fields {
		if lo.Contains(parser.FieldNames, field.Name) {
			continue
		}

		inputType, err := mk.inputTypeForModelField(field)
		if err != nil {
			return nil, err
		}

		input.AddFieldConfig(field.Name, &graphql.InputObjectFieldConfig{
			Type: inputType,
		})
	}

	return input, nil
}

// inputTypeForModelField maps the type in the given proto.Field to a suitable graphql.Input type.
func (mk *graphqlSchemaBuilder) inputTypeForModelField(field *proto.Field) (in graphql.Input, err error) {
	switch field.Type.Type {
	case proto.Type_TYPE_ENUM:
		enumMessage := proto.FindEnum(mk.proto.Enums, field.Type.EnumName.Value)
		in = mk.addEnum(enumMessage)
	case proto.Type_TYPE_MODEL:
		model := mk.proto.FindModel(field.Type.ModelName.Value)
		var err error
		in, err = mk.addModelInput(model)
		if err != nil {
			return nil, err
		}
	default:
		var ok bool
		in, ok = protoTypeToGraphQLInput[field.Type.Type]
		if !ok {
			return in, fmt.Errorf("cannot yet make output type for: %s", field.Type.Type.String())
		}
	}

	if field.Type.Repeated {
		if field.Type.Type == proto.Type_TYPE_MODEL {
			in = mk.makeConnectionType(in)
		} else {
			in = graphql.NewList(in)
			in = graphql.NewNonNull(in)
		}
	} else if !field.Optional {
		in = graphql.NewNonNull(in)
	}

	return in, nil
}

// outputTypeForModelField maps the type in the given proto.Field to a suitable graphql.Output type.
func (mk *graphqlSchemaBuilder) outputTypeForModelField(field *proto.Field) (out graphql.Output, err error) {
	switch field.Type.Type {
	case proto.Type_TYPE_ENUM:
		enumMessage := proto.FindEnum(mk.proto.Enums, field.Type.EnumName.Value)
		out = mk.addEnum(enumMessage)
	case proto.Type_TYPE_MODEL:
		modelMessage := mk.proto.FindModel(field.Type.ModelName.Value)
		var err error
		out, err = mk.addModel(modelMessage)
		if err != nil {
			return nil, err
		}
	default:
		var ok bool
		out, ok = protoTypeToGraphQLOutput[field.Type.Type]
		if !ok {
			return out, fmt.Errorf("cannot yet make output type for: %s", field.Type.Type.String())
		}
	}

	if field.Type.Repeated {
		if field.Type.Type == proto.Type_TYPE_MODEL {
			out = mk.makeConnectionType(out)
		} else {
			out = graphql.NewList(out)
		}
	} else if !field.Optional {
		out = graphql.NewNonNull(out)
	}

	return out, nil
}

// inputTypeFromMessageField maps the type in the given proto.MessageField to a suitable graphql.Input type.
func (mk *graphqlSchemaBuilder) inputTypeFromMessageField(field *proto.MessageField) (graphql.Input, error) {
	var in graphql.Input
	var err error

	switch {
	case field.Type.Type == proto.Type_TYPE_MESSAGE:
		messageName := field.Type.MessageName.Value
		message := mk.proto.FindMessage(messageName)
		if message == nil {
			return nil, fmt.Errorf("message does not exist: %s", messageName)
		}

		if len(message.Fields) == 0 {
			break
		}

		if out, ok := mk.inputs[messageName]; ok {
			in = out
			break
		}

		inputObject := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   mk.makeUniqueInputMessageName(messageName),
			Fields: graphql.InputObjectConfigFieldMap{},
		})

		for _, input := range message.Fields {
			inputField, err := mk.inputTypeFromMessageField(input)
			if err != nil {
				return nil, err
			}

			inputObject.AddFieldConfig(input.Name, &graphql.InputObjectFieldConfig{
				Type: inputField,
			})
		}

		mk.inputs[messageName] = inputObject
		in = inputObject
	case field.Type.Type == proto.Type_TYPE_UNION:
		// GraphQL doesn't support union type or the concept of oneOf for inputs _yet_,
		// so we will rather compile all the fields from the union types into one message,
		// make all the fields optional, and rely on runtime validation.
		messageName := fmt.Sprintf("%sOrderBy", field.MessageName)
		inputObject := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   mk.makeUniqueInputMessageName(messageName),
			Fields: graphql.InputObjectConfigFieldMap{},
		})

		for _, typeName := range field.Type.UnionNames {
			fieldMessage := mk.proto.FindMessage(typeName.Value)
			for _, typeField := range fieldMessage.Fields {
				typeField.Optional = true
				typeField.Nullable = true
				inputField, err := mk.inputTypeFromMessageField(typeField)
				if err != nil {
					return nil, err
				}

				inputObject.AddFieldConfig(typeField.Name, &graphql.InputObjectFieldConfig{
					Type: inputField,
				})
			}
		}

		mk.inputs[messageName] = inputObject
		in = inputObject
	case field.Type.Type == proto.Type_TYPE_MODEL:
		model := mk.proto.FindModel(field.Type.ModelName.Value)
		in, err = mk.addModelInput(model)
		if err != nil {
			return nil, err
		}
	default:
		if in, err = mk.inputTypeFor(field); err != nil {
			return nil, err
		}
	}

	if !field.Optional && !field.Nullable {
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

// makeActionInputType generates an input type to reflect the inputs of the given
// proto.Action - which can be used as the Args field in a graphql.Field.
func (mk *graphqlSchemaBuilder) makeActionInputType(action *proto.Action) (*graphql.InputObject, bool, error) {
	message := mk.proto.FindMessage(action.InputMessageName)
	allOptionalInputs := true

	inputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   mk.makeUniqueInputMessageName(message.Name),
		Fields: graphql.InputObjectConfigFieldMap{},
	})

	for _, field := range message.Fields {
		fieldType, err := mk.inputTypeFromMessageField(field)
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

func (mk *graphqlSchemaBuilder) makeUniqueInputMessageName(name string) string {
	if proto.MessageUsedAsResponse(mk.proto, name) {
		return fmt.Sprintf("%sInput", name)
	}
	return name
}
