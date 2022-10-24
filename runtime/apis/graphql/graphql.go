package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

// A graphqlSchemaBuilder exposes a Make method, that makes a set of graphql.Schema objects - one for each
// of the APIs defined in the keel schema provided at construction time.
type graphqlSchemaBuilder struct {
	proto    *proto.Schema
	query    *graphql.Object
	mutation *graphql.Object
	types    map[string]*graphql.Object
	enums    map[string]*graphql.Enum
}

// NewGraphQLApi creates a map of graphql.Schema objects where the keys
// are the API names from the provided proto.Schema
func NewGraphQLApi(proto *proto.Schema, api *proto.Api) (*graphql.Schema, error) {
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
		types: map[string]*graphql.Object{},
		enums: map[string]*graphql.Enum{},
	}

	return m.build(api, proto)
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

	for _, model := range modelInstances {
		for _, op := range model.Operations {
			err := mk.addOperation(op, schema)
			if err != nil {
				return nil, err
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
		// Passwords are omitted from GraphQL responses
		if field.Type.Type == proto.Type_TYPE_PASSWORD {
			continue
		}

		outputType, err := mk.outputTypeFor(field)
		if err != nil {
			return nil, err
		}
		object.AddFieldConfig(field.Name, &graphql.Field{
			Name: field.Name,
			Type: outputType,
		})
	}

	return object, nil
}

// addOperation generates the graphql field object to represent the given proto.Operation
func (mk *graphqlSchemaBuilder) addOperation(
	op *proto.Operation,
	schema *proto.Schema) error {

	model := proto.FindModel(schema.Models, op.ModelName)

	outputType, err := mk.addModel(model)
	if err != nil {
		return err
	}

	field := &graphql.Field{
		Name: op.Name,
		Type: outputType,
	}

	// Only add args if there are inputs for this operation
	// Unles it's a list and then we need to add pagination
	if len(op.Inputs) > 0 || op.Type == proto.OperationType_OPERATION_TYPE_LIST {
		operationInputType, err := mk.makeOperationInputType(op)
		if err != nil {
			return err
		}

		allOptionalInputs := true
		for _, in := range op.Inputs {
			if !in.Optional {
				allOptionalInputs = false
			}
		}

		if allOptionalInputs {
			field.Args = graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{
					Type: operationInputType,
				},
			}
		} else {
			field.Args = graphql.FieldConfigArgument{
				"input": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(operationInputType),
				},
			}
		}
	}

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_GET:
		field.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"]
			arguments, ok := input.(map[string]any)
			if !ok {
				return nil, errors.New("input not a map")
			}

			var builder actions.GetAction
			scope, err := actions.NewScope(p.Context, op, schema)

			if err != nil {
				return nil, err
			}

			result, err := builder.
				Initialise(scope).
				ApplyImplicitFilters(arguments).
				ApplyExplicitFilters(arguments).
				IsAuthorised(arguments).
				Execute(arguments)

			if result != nil {
				return result.Value.Object, err
			}
			return nil, err
		}
		mk.query.AddFieldConfig(op.Name, field)
	case proto.OperationType_OPERATION_TYPE_CREATE:
		field.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"]
			arguments, ok := input.(map[string]any)
			if !ok {
				return nil, errors.New("input not a map")
			}

			var builder actions.CreateAction

			scope, err := actions.NewScope(p.Context, op, schema)

			if err != nil {
				return nil, err
			}

			result, err := builder.
				Initialise(scope).
				CaptureImplicitWriteInputValues(arguments). // todo: err?
				CaptureSetValues(arguments).
				IsAuthorised(arguments).
				Execute(arguments)

			if result != nil {
				return result.Value.Object, err
			}
			return nil, err
		}
		// create returns a non-null type
		field.Type = graphql.NewNonNull(field.Type)

		mk.mutation.AddFieldConfig(op.Name, field)
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		field.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"]
			arguments, ok := input.(map[string]any)

			if !ok {
				return nil, errors.New("input not a map")
			}

			var builder actions.UpdateAction

			scope, err := actions.NewScope(p.Context, op, schema)

			if err != nil {
				return nil, err
			}

			values, err := runtime.ToArgsMap(arguments, "values")
			if err != nil {
				return nil, err
			}

			wheres, err := runtime.ToArgsMap(arguments, "where")
			if err != nil {
				return nil, err
			}

			result, err := builder.
				Initialise(scope).
				// first capture any implicit inputs
				CaptureImplicitWriteInputValues(values).
				// then capture explicitly used inputs
				CaptureSetValues(values).
				// then apply unique filters
				ApplyImplicitFilters(wheres).
				ApplyExplicitFilters(wheres).
				IsAuthorised(arguments).
				Execute(arguments)

			if result != nil {
				return result.Value.Object, err
			}

			return nil, err
		}

		field.Type = graphql.NewNonNull(field.Type)

		mk.mutation.AddFieldConfig(op.Name, field)
	case proto.OperationType_OPERATION_TYPE_DELETE:
		field.Type = deleteResponseType

		field.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"]
			arguments, ok := input.(map[string]any)

			if !ok {
				return nil, errors.New("input not a map")
			}

			var builder actions.DeleteAction

			scope, err := actions.NewScope(p.Context, op, schema)

			if err != nil {
				return nil, err
			}

			result, err := builder.
				Initialise(scope).
				ApplyImplicitFilters(arguments).
				ApplyExplicitFilters(arguments).
				IsAuthorised(arguments).
				Execute(arguments)

			if result != nil {
				return result.Value.Success, err
			}

			return false, err
		}

		mk.mutation.AddFieldConfig(op.Name, field)
	case proto.OperationType_OPERATION_TYPE_LIST:
		field.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
			input, err := runtime.ToArgsMap(p.Args, "input")

			// If no inputs have been specified then we need to initialise an empty
			// input map with no where conditions
			if err != nil {
				input = map[string]any{
					"where": map[string]any{},
				}
			}

			var builder actions.ListAction

			scope, err := actions.NewScope(p.Context, op, schema)

			if err != nil {
				return nil, err
			}

			where, err := runtime.ToArgsMap(input, "where")

			if err != nil {
				where = map[string]any{}
			}

			result, err := builder.
				Initialise(scope).
				ApplyImplicitFilters(where).
				ApplyExplicitFilters(where).
				IsAuthorised(input).
				Execute(input)

			if err != nil {
				return nil, err
			}

			records := result.Value.Collection

			hasNextPage := result.Value.HasNextPage

			resp, err := connectionResponse(records, hasNextPage)
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
		// for list types we need to wrap the output type in the
		// connection type which allows for pagination
		field.Type = mk.makeConnectionType(outputType)

		mk.query.AddFieldConfig(op.Name, field)
	case proto.OperationType_OPERATION_TYPE_AUTHENTICATE:
		// custom response type as defined in the protobuf schema
		outputTypePrefix := strings.ToUpper(op.Name[0:1]) + op.Name[1:]

		ouput := graphql.NewObject(graphql.ObjectConfig{
			Name:   outputTypePrefix + "Response",
			Fields: graphql.Fields{},
		})

		for _, output := range op.Outputs {

			outputType := protoTypeToGraphQLOutput[output.Type.Type]
			if outputType == nil {
				return fmt.Errorf("cannot yet make output type for: %s", output.Type.Type.String())
			}

			ouput.AddFieldConfig(output.Name, &graphql.Field{
				Type: outputType,
			})
		}

		field.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"]
			inputMap, ok := input.(map[string]any)

			if !ok {
				return nil, errors.New("input not a map")
			}

			authArgs := actions.AuthenticateArgs{
				CreateIfNotExists: inputMap["createIfNotExists"].(bool),
				Email:             inputMap["emailPassword"].(map[string]any)["email"].(string),
				Password:          inputMap["emailPassword"].(map[string]any)["password"].(string),
			}

			identityId, identityCreated, err := actions.Authenticate(p.Context, schema, &authArgs)

			if err != nil {
				return nil, err
			}

			if identityId != nil {
				token, err := runtime.GenerateBearerToken(identityId)

				if err != nil {
					return nil, err
				}

				return map[string]any{
					"identityCreated": identityCreated,
					"token":           token,
				}, nil
			} else {
				return nil, errors.New("failed to authenticate")
			}
		}

		field.Type = ouput

		mk.mutation.AddFieldConfig(op.Name, field)
	default:
		return fmt.Errorf("addOperation() does not yet support this op.Type: %v", op.Type)
	}

	if op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		field.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"]
			inputMap, ok := input.(map[string]any)
			if !ok {
				return nil, errors.New("input not a map")
			}

			res, err := runtime.CallFunction(p.Context, op.Name, op.Type, inputMap)

			if err != nil {
				return nil, err
			}

			return ToGraphQL(p.Context, res, op.Type)
		}
	}

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

var protoTypeToGraphQLOutput = map[proto.Type]graphql.Output{
	proto.Type_TYPE_ID:       graphql.ID,
	proto.Type_TYPE_STRING:   graphql.String,
	proto.Type_TYPE_INT:      graphql.Int,
	proto.Type_TYPE_BOOL:     graphql.Boolean,
	proto.Type_TYPE_DATETIME: timestampType,
	proto.Type_TYPE_DATE:     dateType,
	proto.Type_TYPE_SECRET:   graphql.String,
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

	case proto.Type_TYPE_MODEL, proto.Type_TYPE_IDENTITY:
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

var protoTypeToGraphQLInput = map[proto.Type]graphql.Input{
	proto.Type_TYPE_ID:        graphql.ID,
	proto.Type_TYPE_STRING:    graphql.String,
	proto.Type_TYPE_INT:       graphql.Int,
	proto.Type_TYPE_BOOL:      graphql.Boolean,
	proto.Type_TYPE_TIMESTAMP: timestampInputType,
	proto.Type_TYPE_DATETIME:  timestampInputType,
	proto.Type_TYPE_DATE:      dateInputType,
	proto.Type_TYPE_SECRET:    graphql.String,
	proto.Type_TYPE_PASSWORD:  graphql.String,
}

// inputTypeFor maps the type in the given proto.OperationInput to a suitable graphql.Input type.
func (mk *graphqlSchemaBuilder) inputTypeFor(op *proto.OperationInput) (graphql.Input, error) {
	var in graphql.Input

	switch op.Type.Type {
	case proto.Type_TYPE_ENUM:
		enum, _ := lo.Find(mk.proto.Enums, func(e *proto.Enum) bool {
			return e.Name == op.Type.EnumName.Value
		})
		in = mk.addEnum(enum)
	case proto.Type_TYPE_OBJECT:
		operationNamePrefix := strings.ToUpper(op.OperationName[0:1]) + op.OperationName[1:]
		inputObjectName := strings.ToUpper(op.Name[0:1]) + op.Name[1:]

		inputObject := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   operationNamePrefix + inputObjectName + "Input",
			Fields: graphql.InputObjectConfigFieldMap{},
		})

		for _, input := range op.Inputs {
			inputField, err := mk.inputTypeFor(input)

			if err != nil {
				return nil, err
			}

			inputObject.AddFieldConfig(input.Name, &graphql.InputObjectFieldConfig{
				Type: inputField,
			})
		}
		in = inputObject
	default:
		var ok bool
		in, ok = protoTypeToGraphQLInput[op.Type.Type]

		if !ok {
			return nil, fmt.Errorf("operation %s has unsupposed input type %s", op.OperationName, op.Type.Type.String())
		}

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

// queryInputTypeFor maps the type in the given proto.OperationInput to a suitable graphql.Input type.
func (mk *graphqlSchemaBuilder) queryInputTypeFor(op *proto.OperationInput) (graphql.Input, error) {
	var in graphql.Input

	switch op.Type.Type {
	case proto.Type_TYPE_ENUM:
		enum, _ := lo.Find(mk.proto.Enums, func(e *proto.Enum) bool {
			return e.Name == op.Type.EnumName.Value
		})
		enumType := mk.addEnum(enum)
		in = graphql.NewInputObject(graphql.InputObjectConfig{
			Name: enum.Name + "QueryInput",
			Fields: graphql.InputObjectConfigFieldMap{
				"equals": &graphql.InputObjectFieldConfig{
					Type: enumType,
				},
				"oneOf": &graphql.InputObjectFieldConfig{
					Type: graphql.NewList(graphql.NewNonNull(enumType)),
				},
			},
		})
	default:
		var ok bool
		in, ok = protoTypeToGraphQLQueryInput[op.Type.Type]
		if !ok {
			return nil, fmt.Errorf("operation %s has unsupported input type %s", op.OperationName, op.Type)
		}
	}

	if !op.Optional {
		in = graphql.NewNonNull(in)
	}

	return in, nil
}

// makeOperationInputType generates an input type to reflect the inputs of the given
// proto.Operation - which can be used as the Args field in a graphql.Field.
func (mk *graphqlSchemaBuilder) makeOperationInputType(op *proto.Operation) (*graphql.InputObject, error) {

	inputTypePrefix := strings.ToUpper(op.Name[0:1]) + op.Name[1:]

	inputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   inputTypePrefix + "Input",
		Fields: graphql.InputObjectConfigFieldMap{},
	})

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_GET,
		proto.OperationType_OPERATION_TYPE_CREATE,
		proto.OperationType_OPERATION_TYPE_DELETE,
		proto.OperationType_OPERATION_TYPE_AUTHENTICATE:

		for _, in := range op.Inputs {
			fieldType, err := mk.inputTypeFor(in)
			if err != nil {
				return nil, err
			}

			inputType.AddFieldConfig(in.Name, &graphql.InputObjectFieldConfig{
				Type: fieldType,
			})
		}
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		where := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   inputTypePrefix + "QueryInput",
			Fields: graphql.InputObjectConfigFieldMap{},
		})
		values := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   inputTypePrefix + "ValuesInput",
			Fields: graphql.InputObjectConfigFieldMap{},
		})

		// Update operations could have no read or no write inputs if the filtering
		// and updating is happening in @where or @set expressions
		hasReadInputs := false
		hasWriteInputs := false

		for _, in := range op.Inputs {
			fieldType, err := mk.inputTypeFor(in)
			if err != nil {
				return nil, err
			}

			field := &graphql.InputObjectFieldConfig{
				Type: fieldType,
			}

			switch in.Mode {
			case proto.InputMode_INPUT_MODE_READ:
				hasReadInputs = true
				where.AddFieldConfig(in.Name, field)
			case proto.InputMode_INPUT_MODE_WRITE:
				hasWriteInputs = true
				values.AddFieldConfig(in.Name, field)
			}
		}

		if hasReadInputs {
			inputType.AddFieldConfig("where", &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(where),
			})
		}

		if hasWriteInputs {
			inputType.AddFieldConfig("values", &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(values),
			})
		}
	case proto.OperationType_OPERATION_TYPE_LIST:
		where := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   inputTypePrefix + "QueryInput",
			Fields: graphql.InputObjectConfigFieldMap{},
		})

		allOptionalInputs := true
		for _, in := range op.Inputs {
			var fieldType graphql.Input
			var err error

			if !in.Optional {
				allOptionalInputs = false
			}

			if in.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
				fieldType, err = mk.queryInputTypeFor(in)
			} else {
				fieldType, err = mk.inputTypeFor(in)
			}

			if err != nil {
				return nil, err
			}

			where.AddFieldConfig(in.Name, &graphql.InputObjectFieldConfig{
				Type: fieldType,
			})
		}

		inputType.AddFieldConfig("first", &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		})
		inputType.AddFieldConfig("after", &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		})

		if len(op.Inputs) > 0 {
			// Nullable if all inputs are optional
			if allOptionalInputs {
				inputType.AddFieldConfig("where", &graphql.InputObjectFieldConfig{
					Type: where,
				})
			} else {
				inputType.AddFieldConfig("where", &graphql.InputObjectFieldConfig{
					Type: graphql.NewNonNull(where),
				})
			}
		}

	}

	return inputType, nil
}

// ToGraphQLSchemaLanguage converts the result of an introspection query
// into a GraphQL schema string
// Note: this implementation is not complete and only covers cases
// that are relevant to us, for example directives are not handled
func ToGraphQLSchemaLanguage(response *runtime.Response) string {

	// First we have the marshal the response bytes back into
	// a graphql.Result
	var result graphql.Result
	json.Unmarshal(response.Body, &result)

	// Then we pull out the data contained in the result and convert
	// back into JSON
	b, _ := json.Marshal(result.Data)

	// Finally we marshal that JSON into the IntrospectionQueryResult
	// type... urgh
	var r IntrospectionQueryResult
	json.Unmarshal(b, &r)

	definitions := []string{}

	sort.Slice(r.Schema.Types, func(a, b int) bool {
		aType := r.Schema.Types[a]
		bType := r.Schema.Types[b]

		// Make sure Query and Mutation come at the top of the generated
		// schema with Query first and Mutation second
		typeNameOrder := []string{"Mutation", "Query"}
		aIndex := lo.IndexOf(typeNameOrder, aType.Name)
		bIndex := lo.IndexOf(typeNameOrder, bType.Name)
		if aIndex != -1 || bIndex != -1 {
			return aIndex > bIndex
		}

		// Then order by input types, types, and enums
		kindOrder := []string{"ENUM", "OBJECT", "INPUT_OBJECT"}
		aIndex = lo.IndexOf(kindOrder, aType.Kind)
		bIndex = lo.IndexOf(kindOrder, bType.Kind)
		if aIndex != bIndex {
			return aIndex > bIndex
		}

		// Order same kind by name
		return aType.Name < bType.Name
	})

	for _, t := range r.Schema.Types {
		if t.Kind == "SCALAR" {
			continue
		}
		if strings.HasPrefix(t.Name, "__") {
			continue
		}

		keyword, ok := map[string]string{
			"OBJECT":       "type",
			"INPUT_OBJECT": "input",
			"ENUM":         "enum",
		}[t.Kind]
		if !ok {
			continue
		}

		b := strings.Builder{}
		b.WriteString(keyword)
		b.WriteString(" ")
		b.WriteString(t.Name)
		b.WriteString(" {\n")

		if t.Kind == "ENUM" {
			values := t.EnumValues
			sort.Slice(values, func(i, j int) bool {
				return values[i].Name < values[j].Name
			})

			for _, v := range values {
				b.WriteString("  ")
				b.WriteString(v.Name)
				b.WriteString("\n")
			}
		} else {
			fields := t.Fields
			if t.Kind == "INPUT_OBJECT" {
				fields = t.InputFields
			}

			sort.Slice(fields, func(i, j int) bool {
				return fields[i].Name < fields[j].Name
			})

			for _, field := range fields {
				b.WriteString("  ")
				b.WriteString(field.Name)

				sort.Slice(field.Args, func(i, j int) bool {
					return field.Args[i].Name < field.Args[j].Name
				})

				if len(field.Args) > 0 {
					b.WriteString("(")
					for i, arg := range field.Args {
						if i > 0 {
							b.WriteString(", ")
						}
						b.WriteString(arg.Name)
						b.WriteString(": ")
						b.WriteString(arg.Type.String())
					}
					b.WriteString(")")
				}

				b.WriteString(": ")
				b.WriteString(field.Type.String())
				b.WriteString("\n")
			}
		}

		b.WriteString("}")

		definitions = append(definitions, b.String())
	}

	return strings.Join(definitions, "\n\n") + "\n"
}

func ToGraphQL(ctx context.Context, response any, opType proto.OperationType) (interface{}, error) {
	client, ok := ctx.Value(functionsClientKey).(FunctionsClient)
	if !ok {
		return nil, errors.New("no functions client in context")
	}

	return client.ToGraphQL(ctx, response, opType)
}
