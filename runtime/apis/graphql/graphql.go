package graphql

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
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
// are the API names from the provided proto.Schema.
func NewGraphQLSchema(proto *proto.Schema, api *proto.Api) (*graphql.Schema, error) {
	m := &graphqlSchemaBuilder{
		schema: proto,
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
	schema   *proto.Schema
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
// mk.types.
func (mk *graphqlSchemaBuilder) addModel(model *proto.Model) (*graphql.Object, error) {
	if out, ok := mk.types[fmt.Sprintf("model-%s", model.GetName())]; ok {
		return out.(*graphql.Object), nil
	}

	object := graphql.NewObject(graphql.ObjectConfig{
		Name:   model.GetName(),
		Fields: graphql.Fields{},
	})

	mk.types[fmt.Sprintf("model-%s", model.GetName())] = object

	for _, field := range model.GetFields() {
		// Passwords are omitted from GraphQL responses
		if field.GetType().GetType() == proto.Type_TYPE_PASSWORD {
			continue
		}

		// Vectors are omitted from GraphQL responses
		if field.GetType().GetType() == proto.Type_TYPE_VECTOR {
			continue
		}

		outputType, err := mk.outputTypeForModelField(field)
		if err != nil {
			return nil, err
		}

		if field.GetType().GetType() != proto.Type_TYPE_ENTITY {
			object.AddFieldConfig(field.GetName(), &graphql.Field{
				Name: field.GetName(),
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

		object.AddFieldConfig(field.GetName(), &graphql.Field{
			Name: field.GetName(),
			Type: outputType,
			Args: fieldArgs,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx, span := tracer.Start(p.Context, fmt.Sprintf("Resolve %s.%s", model.GetName(), field.GetName()))
				defer span.End()

				relatedModel := mk.schema.FindModel(field.GetType().GetEntityName().GetValue())

				// Create a new query for the related model
				query := actions.NewQuery(relatedModel)
				query.Select(actions.AllFields())

				foreignKeyField := mk.schema.GetForeignKeyFieldName(field)

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
					return nil, fmt.Errorf("model %s did not have field %s", model.GetName(), parentLookupField)
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

				scope := actions.NewModelScope(ctx, relatedModel, mk.schema)

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

					results, _, pageInfo, err := query.
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
						"results":  results,
						"pageInfo": pageInfo.ToMap(),
					})
					if err != nil {
						span.RecordError(err, trace.WithStackTrace(true))
						span.SetStatus(codes.Error, err.Error())
						return nil, err
					}

					return res, nil
				default:
					return nil, fmt.Errorf("unhandled model relationship configuration for field: %s on model: %s", field.GetName(), field.GetEntityName())
				}
			},
		})
	}

	return object, nil
}

// addOperation generates the graphql field object to represent the given proto.Action.
func (mk *graphqlSchemaBuilder) addAction(
	action *proto.Action,
	schema *proto.Schema) error {
	model := schema.FindModel(action.GetModelName())
	modelType, err := mk.addModel(model)
	if err != nil {
		return err
	}

	field := &graphql.Field{
		Name: action.GetName(),
	}

	// Only add input args if an input field exists.
	if action.GetInputMessageName() != "" && action.GetInputMessageName() != parser.MessageFieldTypeAny {
		operationInputType, allOptionalInputs, err := mk.makeActionInputType(action)
		if err != nil {
			return err
		}

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
	} else if action.GetInputMessageName() == parser.MessageFieldTypeAny {
		// Any type
		field.Args = graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: anyType,
			},
		}
	}

	switch action.GetType() {
	case proto.ActionType_ACTION_TYPE_GET:
		field.Type = modelType
		mk.query.AddFieldConfig(action.GetName(), field)
	case proto.ActionType_ACTION_TYPE_CREATE,
		proto.ActionType_ACTION_TYPE_UPDATE:
		field.Type = graphql.NewNonNull(modelType)
		mk.mutation.AddFieldConfig(action.GetName(), field)
	case proto.ActionType_ACTION_TYPE_DELETE:
		field.Type = deleteResponseType
		mk.mutation.AddFieldConfig(action.GetName(), field)
	case proto.ActionType_ACTION_TYPE_LIST:
		// for list types we need to wrap the output type in the
		// connection type which allows for pagination
		resultInfo := mk.makeResultInfoType(action)
		field.Type = mk.makeConnectionType(modelType, resultInfo)
		mk.query.AddFieldConfig(action.GetName(), field)
	case proto.ActionType_ACTION_TYPE_READ:
		responseMessage := schema.FindMessage(action.GetResponseMessageName())
		if responseMessage == nil {
			return fmt.Errorf("response message does not exist: %s", action.GetResponseMessageName())
		}

		if responseMessage.GetName() == parser.MessageFieldTypeAny {
			field.Type = anyType
		} else {
			field.Type, err = mk.addMessage(responseMessage)
			if err != nil {
				return err
			}
		}

		mk.query.AddFieldConfig(action.GetName(), field)
	case proto.ActionType_ACTION_TYPE_WRITE:
		responseMessage := schema.FindMessage(action.GetResponseMessageName())
		if responseMessage == nil {
			return fmt.Errorf("response message does not exist: %s", action.GetResponseMessageName())
		}
		field.Type, err = mk.addMessage(responseMessage)
		if err != nil {
			return err
		}
		mk.mutation.AddFieldConfig(action.GetName(), field)
	default:
		return fmt.Errorf("addAction() does not yet support this action type: %v", action.GetType())
	}

	field.Resolve = ActionFunc(schema, action)

	return nil
}

func (mk *graphqlSchemaBuilder) makeResultInfoType(action *proto.Action) graphql.Output {
	facetFields := proto.FacetFields(mk.schema, action)
	if len(facetFields) == 0 {
		return nil
	}

	fields := graphql.Fields{}
	for _, field := range facetFields {
		fields[field.GetName()] = &graphql.Field{
			Type: protoTypeToFacetType[field.GetType().GetType()],
		}
	}

	return graphql.NewObject(graphql.ObjectConfig{
		Name:   fmt.Sprintf("%sResultInfo", strcase.ToCamel(action.GetName())),
		Fields: fields,
	})
}

func (mk *graphqlSchemaBuilder) makeConnectionType(itemType graphql.Output, resultInfo graphql.Output) graphql.Output {
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

	if resultInfo != nil {
		connection.AddFieldConfig("resultInfo", &graphql.Field{
			Type: resultInfo,
		})
	}

	mk.types[fmt.Sprintf("connection-%s", itemType.Name())] = connection

	return graphql.NewNonNull(connection)
}

func (mk *graphqlSchemaBuilder) addEnum(e *proto.Enum) *graphql.Enum {
	if out, ok := mk.enums[e.GetName()]; ok {
		return out
	}

	values := graphql.EnumValueConfigMap{}

	for _, v := range e.GetValues() {
		values[v.GetName()] = &graphql.EnumValueConfig{
			Value: v.GetName(),
		}
	}

	enum := graphql.NewEnum(graphql.EnumConfig{
		Name:   e.GetName(),
		Values: values,
	})
	mk.enums[e.GetName()] = enum
	return enum
}

// addMessage makes a graphql.Object response output from a proto.Message.
func (mk *graphqlSchemaBuilder) addMessage(message *proto.Message) (graphql.Output, error) {
	if message.GetName() == parser.MessageFieldTypeAny {
		return anyType, nil
	}
	if out, ok := mk.types[message.GetName()]; ok {
		return out, nil
	}

	output := graphql.NewObject(graphql.ObjectConfig{
		Name:   message.GetName(),
		Fields: graphql.Fields{},
	})

	for _, field := range message.GetFields() {
		var fieldType graphql.Output

		switch field.GetType().GetType() {
		case proto.Type_TYPE_MESSAGE:
			fieldMessage := mk.schema.FindMessage(field.GetType().GetMessageName().GetValue())

			var err error
			fieldType, err = mk.addMessage(fieldMessage)
			if err != nil {
				return nil, err
			}
		case proto.Type_TYPE_ENTITY:
			modelMessage := mk.schema.FindModel(field.GetType().GetEntityName().GetValue())

			var err error
			fieldType, err = mk.addModel(modelMessage)
			if err != nil {
				return nil, err
			}
		case proto.Type_TYPE_ENUM:
			enumMessage := proto.FindEnum(mk.schema.GetEnums(), field.GetType().GetEnumName().GetValue())
			fieldType = mk.addEnum(enumMessage)

		default:
			fieldType = protoTypeToGraphQLOutput[field.GetType().GetType()]
			if fieldType == nil {
				return nil, fmt.Errorf("cannot yet make output type for: %s", field.GetType().GetType().String())
			}
		}

		if field.GetType().GetRepeated() {
			fieldType = graphql.NewList(fieldType)
		}
		if !field.GetOptional() {
			fieldType = graphql.NewNonNull(fieldType)
		}

		output.AddFieldConfig(field.GetName(), &graphql.Field{
			Type: fieldType,
		})
	}

	if len(message.GetFields()) == 0 {
		output.AddFieldConfig("success", &graphql.Field{
			Type:    graphql.Boolean,
			Resolve: func(_ graphql.ResolveParams) (interface{}, error) { return true, nil },
		})
	}

	mk.types[message.GetName()] = output

	return output, nil
}

// addModel generates the graphql type to represent the given proto.Model, and inserts it into
// mk.types.
func (mk *graphqlSchemaBuilder) addModelInput(model *proto.Model) (graphql.Input, error) {
	if in, ok := mk.types[fmt.Sprintf("model-%s", model.GetName())]; ok {
		return in.(graphql.Input), nil
	}

	input := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   model.GetName(),
		Fields: graphql.InputObjectConfigFieldMap{},
	})

	for _, field := range model.GetFields() {
		if lo.Contains(parser.FieldNames, field.GetName()) {
			continue
		}

		inputType, err := mk.inputTypeForModelField(field)
		if err != nil {
			return nil, err
		}

		input.AddFieldConfig(field.GetName(), &graphql.InputObjectFieldConfig{
			Type: inputType,
		})
	}

	return input, nil
}

// inputTypeForModelField maps the type in the given proto.Field to a suitable graphql.Input type.
func (mk *graphqlSchemaBuilder) inputTypeForModelField(field *proto.Field) (in graphql.Input, err error) {
	switch field.GetType().GetType() {
	case proto.Type_TYPE_ENUM:
		enumMessage := proto.FindEnum(mk.schema.GetEnums(), field.GetType().GetEnumName().GetValue())
		in = mk.addEnum(enumMessage)
	case proto.Type_TYPE_ENTITY:
		model := mk.schema.FindModel(field.GetType().GetEntityName().GetValue())
		var err error
		in, err = mk.addModelInput(model)
		if err != nil {
			return nil, err
		}
	default:
		var ok bool
		in, ok = protoTypeToGraphQLInput[field.GetType().GetType()]
		if !ok {
			return in, fmt.Errorf("cannot yet make output type for: %s", field.GetType().GetType().String())
		}
	}

	if field.GetType().GetRepeated() {
		if field.GetType().GetType() == proto.Type_TYPE_ENTITY {
			in = mk.makeConnectionType(in, nil)
		} else {
			in = graphql.NewList(in)
			in = graphql.NewNonNull(in)
		}
	} else if !field.GetOptional() {
		in = graphql.NewNonNull(in)
	}

	return in, nil
}

// outputTypeForModelField maps the type in the given proto.Field to a suitable graphql.Output type.
func (mk *graphqlSchemaBuilder) outputTypeForModelField(field *proto.Field) (out graphql.Output, err error) {
	switch field.GetType().GetType() {
	case proto.Type_TYPE_ENUM:
		enumMessage := proto.FindEnum(mk.schema.GetEnums(), field.GetType().GetEnumName().GetValue())
		out = mk.addEnum(enumMessage)
	case proto.Type_TYPE_ENTITY:
		modelMessage := mk.schema.FindModel(field.GetType().GetEntityName().GetValue())
		var err error
		out, err = mk.addModel(modelMessage)
		if err != nil {
			return nil, err
		}
	default:
		var ok bool
		out, ok = protoTypeToGraphQLOutput[field.GetType().GetType()]
		if !ok {
			return out, fmt.Errorf("cannot yet make output type for: %s", field.GetType().GetType().String())
		}
	}

	if field.GetType().GetRepeated() {
		if field.GetType().GetType() == proto.Type_TYPE_ENTITY {
			out = mk.makeConnectionType(out, nil)
		} else {
			out = graphql.NewList(out)
		}
	} else if !field.GetOptional() {
		out = graphql.NewNonNull(out)
	}

	return out, nil
}

// inputTypeFromMessageField maps the type in the given proto.MessageField to a suitable graphql.Input type.
func (mk *graphqlSchemaBuilder) inputTypeFromMessageField(field *proto.MessageField) (graphql.Input, error) {
	var in graphql.Input
	var err error

	switch {
	case field.GetType().GetType() == proto.Type_TYPE_MESSAGE:
		messageName := field.GetType().GetMessageName().GetValue()
		message := mk.schema.FindMessage(messageName)
		if message == nil {
			return nil, fmt.Errorf("message does not exist: %s", messageName)
		}

		if len(message.GetFields()) == 0 {
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

		for _, input := range message.GetFields() {
			inputField, err := mk.inputTypeFromMessageField(input)
			if err != nil {
				return nil, err
			}

			inputObject.AddFieldConfig(input.GetName(), &graphql.InputObjectFieldConfig{
				Type: inputField,
			})
		}

		mk.inputs[messageName] = inputObject
		in = inputObject
	case field.GetType().GetType() == proto.Type_TYPE_UNION:
		// GraphQL doesn't support union type or the concept of oneOf for inputs _yet_,
		// so we will rather compile all the fields from the union types into one message,
		// make all the fields optional, and rely on runtime validation.
		messageName := fmt.Sprintf("%sOrderBy", field.GetMessageName())
		inputObject := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   mk.makeUniqueInputMessageName(messageName),
			Fields: graphql.InputObjectConfigFieldMap{},
		})

		for _, typeName := range field.GetType().GetUnionNames() {
			fieldMessage := mk.schema.FindMessage(typeName.GetValue())
			for _, typeField := range fieldMessage.GetFields() {
				typeField.Optional = true
				typeField.Nullable = true
				inputField, err := mk.inputTypeFromMessageField(typeField)
				if err != nil {
					return nil, err
				}

				inputObject.AddFieldConfig(typeField.GetName(), &graphql.InputObjectFieldConfig{
					Type: inputField,
				})
			}
		}

		mk.inputs[messageName] = inputObject
		in = inputObject
	case field.GetType().GetType() == proto.Type_TYPE_ENTITY:
		model := mk.schema.FindModel(field.GetType().GetEntityName().GetValue())
		in, err = mk.addModelInput(model)
		if err != nil {
			return nil, err
		}
	default:
		if in, err = mk.inputTypeFor(field); err != nil {
			return nil, err
		}
	}

	if !field.GetOptional() && !field.GetNullable() {
		in = graphql.NewNonNull(in)
	}

	if field.GetType().GetRepeated() {
		in = graphql.NewList(in)
		if !field.GetOptional() {
			in = graphql.NewNonNull(in)
		}
	}

	return in, nil
}

// inputTypeFor creates a graphql.Input for non-list operation input types.
func (mk *graphqlSchemaBuilder) inputTypeFor(field *proto.MessageField) (graphql.Input, error) {
	var in graphql.Input
	if field.GetType().GetType() == proto.Type_TYPE_ENUM {
		enum, _ := lo.Find(mk.schema.GetEnums(), func(e *proto.Enum) bool {
			return e.GetName() == field.GetType().GetEnumName().GetValue()
		})
		in = mk.addEnum(enum)
	} else {
		var ok bool
		if in, ok = protoTypeToGraphQLInput[field.GetType().GetType()]; !ok {
			return nil, fmt.Errorf("message %s has unsupported message field type: %s", field.GetMessageName(), field.GetType().GetType().String())
		}
	}
	return in, nil
}

// makeActionInputType generates an input type to reflect the inputs of the given
// proto.Action - which can be used as the Args field in a graphql.Field.
func (mk *graphqlSchemaBuilder) makeActionInputType(action *proto.Action) (*graphql.InputObject, bool, error) {
	if action.GetInputMessageName() == "" {
		return nil, true, nil
	}

	message := mk.schema.FindMessage(action.GetInputMessageName())
	allOptionalInputs := true

	inputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   mk.makeUniqueInputMessageName(message.GetName()),
		Fields: graphql.InputObjectConfigFieldMap{},
	})

	for _, field := range message.GetFields() {
		fieldType, err := mk.inputTypeFromMessageField(field)
		if err != nil {
			return nil, false, err
		}

		if fieldType != nil {
			if !field.GetOptional() {
				allOptionalInputs = false
			}
			inputType.AddFieldConfig(field.GetName(), &graphql.InputObjectFieldConfig{
				Type: fieldType,
			})
		}
	}

	return inputType, allOptionalInputs, nil
}

func (mk *graphqlSchemaBuilder) makeUniqueInputMessageName(name string) string {
	if proto.MessageUsedAsResponse(mk.schema, name) {
		return fmt.Sprintf("%sInput", name)
	}
	return name
}
