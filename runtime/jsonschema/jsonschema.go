package jsonschema

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/xeipuuv/gojsonschema"
)

type JSONSchema struct {
	// Type is generally just a string, but when we need a type to be
	// null it is a list containing the type and the string "null".
	// In JSON output for most cases we just want a string, not a list
	// of one string, so we use any here so it can be either
	Type any `json:"type,omitempty"`

	Format               string                `json:"format,omitempty"`
	Properties           map[string]JSONSchema `json:"properties,omitempty"`
	AdditionalProperties *bool                 `json:"additionalProperties,omitempty"`
	AnyOf                []JSONSchema          `json:"anyOf,omitempty"`
	Enum                 []string              `json:"enum,omitempty"`
	Items                *JSONSchema           `json:"items,omitempty"`
	Required             []string              `json:"required,omitempty"`
	Ref                  string                `json:"$ref,omitempty"`
	Defs                 map[string]JSONSchema `json:"$defs,omitempty"`
}

// ValidateRequest validates that the input is valid for the given operation and schema.
// If validation errors are found they will be contained in the returned result. If an error
// is returned then validation could not be completed, likely to do an invalid JSON schema
// being created.
func ValidateRequest(ctx context.Context, schema *proto.Schema, op *proto.Operation, input map[string]any) (*gojsonschema.Result, error) {
	requestType := JSONSchemaForOperation(ctx, schema, op)
	return gojsonschema.Validate(gojsonschema.NewGoLoader(requestType), gojsonschema.NewGoLoader(input))
}

func JSONSchemaForOperation(ctx context.Context, schema *proto.Schema, op *proto.Operation) JSONSchema {
	// TODO: implement proper support for authenticate once it's been re-done using
	// arbitrary functions
	if op.Type == proto.OperationType_OPERATION_TYPE_AUTHENTICATE {
		return JSONSchema{
			Type:                 "object",
			AdditionalProperties: boolPtr(true),
		}
	}

	root := JSONSchema{
		Type:                 "object",
		Properties:           map[string]JSONSchema{},
		AdditionalProperties: boolPtr(false),
		Defs:                 map[string]JSONSchema{},
	}

	where := JSONSchema{
		Type:                 "object",
		Properties:           map[string]JSONSchema{},
		AdditionalProperties: boolPtr(false),
	}

	values := JSONSchema{
		Type:                 "object",
		Properties:           map[string]JSONSchema{},
		AdditionalProperties: boolPtr(false),
	}

	isUpdate := op.Type == proto.OperationType_OPERATION_TYPE_UPDATE
	isList := op.Type == proto.OperationType_OPERATION_TYPE_LIST

	for _, input := range op.Inputs {
		name, prop := jsonSchemaForInput(ctx, op, input, schema)

		isWrite := input.Mode == proto.InputMode_INPUT_MODE_WRITE
		isRead := input.Mode == proto.InputMode_INPUT_MODE_READ

		var obj *JSONSchema

		switch {
		case isUpdate && isRead, isList:
			obj = &where
		case isUpdate && isWrite:
			obj = &values
		default:
			obj = &root
		}

		if name != "" {
			root.Defs[name] = prop
			obj.Properties[input.Name] = JSONSchema{Ref: fmt.Sprintf("#/$defs/%s", name)}
		} else {
			obj.Properties[input.Name] = prop
		}

		// If the input is not optional then mark it required in the JSON schema
		if !input.Optional {
			obj.Required = append(obj.Required, input.Name)
		}
	}

	if isUpdate || isList {
		// Always add the "where" prop but only make it required if has any properties
		typeName := strcase.ToCamel(op.Name) + "WhereInput"
		root.Properties["where"] = JSONSchema{Ref: fmt.Sprintf("#/$defs/%s", typeName)}
		root.Defs[typeName] = where
		if len(where.Properties) > 0 {
			root.Required = append(root.Required, "where")
		}
	}

	if isUpdate {
		// Always add the "values" prop but only make it required if has any properties
		typeName := strcase.ToCamel(op.Name) + "ValuesInput"
		root.Properties["values"] = JSONSchema{Ref: fmt.Sprintf("#/$defs/%s", typeName)}
		root.Defs[typeName] = values
		if len(values.Properties) > 0 {
			root.Required = append(root.Required, "values")
		}
	}

	return root
}

func jsonSchemaForInput(ctx context.Context, op *proto.Operation, input *proto.OperationInput, schema *proto.Schema) (string, JSONSchema) {
	if op.Type == proto.OperationType_OPERATION_TYPE_LIST && input.IsModelField() {
		return jsonSchemaForQueryObject(ctx, op, input, schema)
	}

	isImplicit := input.IsModelField()
	isWrite := input.Mode == proto.InputMode_INPUT_MODE_WRITE

	name := ""
	prop := JSONSchema{}

	switch input.Type.Type {
	case proto.Type_TYPE_ID:
		prop.Type = "string"
	case proto.Type_TYPE_STRING:
		prop.Type = "string"
	case proto.Type_TYPE_BOOL:
		prop.Type = "boolean"
	case proto.Type_TYPE_INT:
		prop.Type = "number"

	// date-time format allows both YYYY-MM-DD and full ISO8601/RFC3339 format
	case proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		prop.Type = "string"
		prop.Format = "date-time"

	case proto.Type_TYPE_ENUM:
		prop.Type = "string"
		enum, _ := lo.Find(schema.Enums, func(e *proto.Enum) bool {
			return e.Name == input.Type.EnumName.Value
		})
		for _, v := range enum.Values {
			prop.Enum = append(prop.Enum, v.Name)
		}
		name = enum.Name
	}

	// An input is allowed to be null if the field it relates to is marked
	// optional. This is because "optional" has a slightly different meaning
	// between model fields and action inputs:
	//     - model field:  "can be null"
	//     - action input: "can be ommitted"
	if isImplicit && isWrite {
		field := proto.FindField(schema.Models, op.ModelName, input.Target[0])
		if field.Optional {
			// OpenAPI differs from JSON Schema in that:
			//   | Note that there is no null type; instead, the nullable
			//   | attribute is used as a modifier of the base type.
			// TODO: if we want to support OpenAPI generation we'll need
			// to support both styles. For now we only support JSON Schema.
			prop.Type = []any{prop.Type, "null"}
		}
	}

	return name, prop
}

func jsonSchemaForQueryObject(ctx context.Context, op *proto.Operation, input *proto.OperationInput, schema *proto.Schema) (string, JSONSchema) {
	switch input.Type.Type {
	case proto.Type_TYPE_ID:
		t := JSONSchema{
			Type: "string",
		}
		return "IDQueryInput", JSONSchema{
			Type: []string{"object"},
			Properties: map[string]JSONSchema{
				"equals": t,
				"oneOf":  {Type: "array", Items: &t},
			},
			AdditionalProperties: boolPtr(false),
		}
	case proto.Type_TYPE_STRING:
		t := JSONSchema{
			Type: "string",
		}
		return "StringQueryInput", JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"equals":     t,
				"startsWith": t,
				"endsWith":   t,
				"contains":   t,
				"oneOf":      {Type: "array", Items: &t},
			},
			AdditionalProperties: boolPtr(false),
		}
	case proto.Type_TYPE_DATE:
		t := JSONSchema{
			Type:   "string",
			Format: "date-time",
		}
		return "DateQueryInput", JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"equals":     t,
				"before":     t,
				"onOrBefore": t,
				"after":      t,
				"onOrAfter":  t,
			},
			AdditionalProperties: boolPtr(false),
		}
	case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		t := JSONSchema{
			Type:   "string",
			Format: "date-time",
		}
		return "TimestampQueryInput", JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"before": t,
				"after":  t,
			},
			AdditionalProperties: boolPtr(false),
		}
	case proto.Type_TYPE_BOOL:
		return "BooleanQueryInput", JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"equals": {Type: "boolean"},
			},
			AdditionalProperties: boolPtr(false),
		}
	case proto.Type_TYPE_INT:
		t := JSONSchema{
			Type: "number",
		}
		return "IntQueryInput", JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"equals":              t,
				"lessThan":            t,
				"lessThanOrEquals":    t,
				"greaterThan":         t,
				"greaterThanOrEquals": t,
			},
			AdditionalProperties: boolPtr(false),
		}
	case proto.Type_TYPE_ENUM:
		t := JSONSchema{
			Type: "string",
		}
		enum, _ := lo.Find(schema.Enums, func(e *proto.Enum) bool {
			return e.Name == input.Type.EnumName.Value
		})
		for _, v := range enum.Values {
			t.Enum = append(t.Enum, v.Name)
		}
		return fmt.Sprintf("%sQueryInput", enum.Name), JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"equals": t,
				"oneOf": {
					Type:  "array",
					Items: &t,
				},
			},
			AdditionalProperties: boolPtr(false),
		}
	}

	return "", JSONSchema{}
}

func boolPtr(b bool) *bool {
	return &b
}
