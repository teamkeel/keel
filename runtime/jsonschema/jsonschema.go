package jsonschema

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/xeipuuv/gojsonschema"
)

var (
	AnyTypes = []string{"string", "object", "array", "integer", "number", "boolean", "null"}
)

type JSONSchema struct {
	// Type is generally just a string, but when we need a type to be
	// null it is a list containing the type and the string "null".
	// In JSON output for most cases we just want a string, not a list
	// of one string, so we use any here so it can be either
	Type any `json:"type,omitempty"`

	// The enum field needs to be able to contains strings and null,
	// so we use *string here
	Enum []*string `json:"enum,omitempty"`

	// Validation for strings
	Format string `json:"format,omitempty"`

	// Validation for objects
	Properties           map[string]JSONSchema `json:"properties,omitempty"`
	AdditionalProperties *bool                 `json:"additionalProperties,omitempty"`
	Required             []string              `json:"required,omitempty"`

	// For arrays
	Items *JSONSchema `json:"items,omitempty"`

	// Used to link to a type defined in the root $defs
	Ref string `json:"$ref,omitempty"`

	// Only used in the root JSONSchema object to define types that
	// can then be referenced using $ref
	Components *Components `json:"components,omitempty"`
}

type Components struct {
	Schemas map[string]JSONSchema `json:"schemas"`
}

// ValidateRequest validates that the input is valid for the given operation and schema.
// If validation errors are found they will be contained in the returned result. If an error
// is returned then validation could not be completed, likely to do an invalid JSON schema
// being created.
func ValidateRequest(ctx context.Context, schema *proto.Schema, op *proto.Operation, input any) (*gojsonschema.Result, error) {
	requestType := JSONSchemaForOperationInput(ctx, schema, op)
	return gojsonschema.Validate(gojsonschema.NewGoLoader(requestType), gojsonschema.NewGoLoader(input))
}

func JSONSchemaForOperationInput(ctx context.Context, schema *proto.Schema, op *proto.Operation) JSONSchema {
	inputMessage := proto.FindMessage(schema.Messages, op.InputMessageName)
	return JSONSchemaForMessage(ctx, schema, op, inputMessage)
}

func JSONSchemaForOperationResponse(ctx context.Context, schema *proto.Schema, op *proto.Operation) JSONSchema {
	if op.ResponseMessageName != "" {
		responseMsg := proto.FindMessage(schema.Messages, op.ResponseMessageName)

		return JSONSchemaForMessage(ctx, schema, op, responseMsg)
	}

	if op.Implementation != proto.OperationImplementation_OPERATION_IMPLEMENTATION_AUTO {
		panic("unexpected implementation type for response schema definition")
	}

	// If we've reached this point then we know that we are dealing with built-in operations

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE, proto.OperationType_OPERATION_TYPE_GET, proto.OperationType_OPERATION_TYPE_UPDATE:
		// model

		model := proto.FindModel(schema.Models, op.ModelName)

		return jsonSchemaForModel(ctx, schema, model, false)
	case proto.OperationType_OPERATION_TYPE_LIST:
		// array of models

		return JSONSchema{}
	case proto.OperationType_OPERATION_TYPE_DELETE:
		// string id of deleted record

		return JSONSchema{}
	default:
		panic("unexpected operation type " + op.Type.String())
	}
}

// Generates JSONSchema for an operation by generating properties for the root input message.
// Any subsequent nested messages are referenced.
func JSONSchemaForMessage(ctx context.Context, schema *proto.Schema, op *proto.Operation, message *proto.Message) JSONSchema {
	components := Components{
		Schemas: map[string]JSONSchema{},
	}

	messageIsNil := message == nil
	isAny := !messageIsNil && message.Name == parser.MessageFieldTypeAny

	root := JSONSchema{
		Type:                 "object",
		Properties:           map[string]JSONSchema{},
		AdditionalProperties: boolPtr(isAny),
	}

	if isAny {
		root.Type = AnyTypes
	}

	if !isAny {
		for _, field := range message.Fields {
			prop := jsonSchemaForField(ctx, schema, op, field.Type, field.Optional)

			// Merge components from this request schema into OpenAPI components
			if prop.Components != nil {
				for name, comp := range prop.Components.Schemas {
					components.Schemas[name] = comp
				}
				prop.Components = nil
			}

			root.Properties[field.Name] = prop

			// If the input is not optional then mark it required in the JSON schema
			if !field.Optional {
				root.Required = append(root.Required, field.Name)
			}
		}
	}

	if len(components.Schemas) > 0 {
		root.Components = &components
	}

	return root
}

func jsonSchemaForModel(ctx context.Context, schema *proto.Schema, model *proto.Model, isRepeated bool) JSONSchema {
	definitionSchema := JSONSchema{
		Properties: map[string]JSONSchema{},
	}

	s := JSONSchema{}

	name := ModelComponentName(model)

	if isRepeated {
		s.Type = "array"
		s.Items = &JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", name)}
	} else {
		s = JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", name)}
	}

	for _, field := range model.Fields {
		definitionSchema.Properties[field.Name] = jsonSchemaForField(ctx, schema, nil, field.Type, field.Optional)
	}

	schemas := map[string]JSONSchema{}

	schemas[name] = definitionSchema

	s.Components = &Components{
		Schemas: schemas,
	}

	return s
}

func jsonSchemaForField(ctx context.Context, schema *proto.Schema, op *proto.Operation, t *proto.TypeInfo, isOptional bool) JSONSchema {
	components := &Components{
		Schemas: map[string]JSONSchema{},
	}
	prop := JSONSchema{}
	nullable := isOptional

	switch t.Type {
	case proto.Type_TYPE_ANY:
		prop.Type = AnyTypes
	case proto.Type_TYPE_MESSAGE:
		if op == nil {
			panic("operation not in context")
		}
		// Add the nested message to schema components.
		message := proto.FindMessage(schema.Messages, t.MessageName.Value)
		component := JSONSchemaForMessage(ctx, schema, op, message)

		// If that nested message component has ref fields itself, then its components must be bundled.
		if component.Components != nil {
			for cName, comp := range component.Components.Schemas {
				components.Schemas[cName] = comp
			}
			component.Components = nil
		}

		name := t.MessageName.Value
		if nullable {
			component.allowNull()
			name = "nullable_" + name
		}

		if t.Repeated {
			prop.Type = "array"
			prop.Items = &JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", name)}
		} else {
			prop = JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", name)}
		}

		components.Schemas[name] = component
	case proto.Type_TYPE_ID, proto.Type_TYPE_STRING:
		prop.Type = "string"
	case proto.Type_TYPE_BOOL:
		prop.Type = "boolean"
	case proto.Type_TYPE_INT:
		prop.Type = "number"
	case proto.Type_TYPE_MODEL:
		model := proto.FindModel(schema.Models, t.ModelName.Value)
		schema := jsonSchemaForModel(ctx, schema, model, t.Repeated)
		name := ModelComponentName(model)

		if t.Repeated {
			prop.Type = "array"
			prop.Items = &JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", name)}
		} else {
			prop = JSONSchema{Ref: fmt.Sprintf("#/components/schemas/%s", name)}
		}

		components.Schemas[name] = schema
	case proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		// date-time format allows both YYYY-MM-DD and full ISO8601/RFC3339 format
		prop.Type = "string"
		prop.Format = "date-time"
	case proto.Type_TYPE_ENUM:
		// For enum's we actually don't need to set the `type` field at all
		enum, _ := lo.Find(schema.Enums, func(e *proto.Enum) bool {
			return e.Name == t.EnumName.Value
		})

		for _, v := range enum.Values {
			prop.Enum = append(prop.Enum, &v.Name)
		}

		if nullable {
			prop.allowNull()
		}
	}

	if t.Repeated && t.Type != proto.Type_TYPE_MESSAGE {
		prop.Items = &JSONSchema{Type: prop.Type, Enum: prop.Enum}
		prop.Enum = nil
		prop.Type = "array"
	}

	if nullable {
		prop.allowNull()
	}

	if len(components.Schemas) > 0 {
		prop.Components = components
	}

	return prop
}

// allowNull makes sure that s allows null, either by modifying
// the type field or the enum field
//
// This is an area where OpenAPI differs from JSON Schema, from
// the OpenAPI spec:
//
//	| Note that there is no null type; instead, the nullable
//	| attribute is used as a modifier of the base type.
//
// We currently only support JSON schema
func (s *JSONSchema) allowNull() {
	t := s.Type
	switch t := t.(type) {
	case string:
		s.Type = []string{t, "null"}
	case []string:
		if lo.Contains(t, "null") {
			return
		}
		t = append(t, "null")
		s.Type = t
	}

	if len(s.Enum) > 0 && !lo.Contains(s.Enum, nil) {
		s.Enum = append(s.Enum, nil)
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func ModelComponentName(m *proto.Model) string {
	return fmt.Sprintf("%s_model", m.Name)
}
