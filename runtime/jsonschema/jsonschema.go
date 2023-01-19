package jsonschema

import (
	"context"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/xeipuuv/gojsonschema"
)

type JSONSchema struct {
	Type                 any                   `json:"type"`
	Format               string                `json:"format,omitempty"`
	AdditionalProperties bool                  `json:"additionalProperties"`
	AnyOf                []JSONSchema          `json:"anyOf,omitempty"`
	Enum                 []string              `json:"enum,omitempty"`
	Items                *JSONSchema           `json:"items,omitempty"`
	Properties           map[string]JSONSchema `json:"properties,omitempty"`
	Required             []string              `json:"required,omitempty"`
}

// ValidateRequest validates that the input is valid for the given operation and schema.
// If validation errors are found they will be contained in the returned result. If an error
// is returned then validation could not be completed, likely to do an invalid JSON schema
// being created.
func ValidateRequest(ctx context.Context, schema *proto.Schema, op *proto.Operation, input map[string]any) (*gojsonschema.Result, error) {
	requestType := jsonSchemaForOperation(ctx, schema, op)
	return gojsonschema.Validate(gojsonschema.NewGoLoader(requestType), gojsonschema.NewGoLoader(input))
}

func jsonSchemaForOperation(ctx context.Context, schema *proto.Schema, op *proto.Operation) JSONSchema {
	root := JSONSchema{
		Type:       "object",
		Properties: map[string]JSONSchema{},
	}

	where := JSONSchema{
		Type:       "object",
		Properties: map[string]JSONSchema{},
	}

	values := JSONSchema{
		Type:       "object",
		Properties: map[string]JSONSchema{},
	}

	isUpdate := op.Type == proto.OperationType_OPERATION_TYPE_UPDATE
	isList := op.Type == proto.OperationType_OPERATION_TYPE_LIST

	for _, input := range op.Inputs {
		prop := jsonSchemaForInput(ctx, op, input, schema)

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

		obj.Properties[input.Name] = prop

		// If the input is not optional then mark it required in the JSON schema
		if !input.Optional {
			obj.Required = append(obj.Required, input.Name)
		}
	}

	if isUpdate || isList {
		// Always add the "where" prop but only make it required if has any properties
		root.Properties["where"] = where
		if len(where.Properties) > 0 {
			root.Required = append(root.Required, "where")
		}
	}

	if isUpdate {
		// Always add the "values" prop but only make it required if has any properties
		root.Properties["values"] = values
		if len(values.Properties) > 0 {
			root.Required = append(root.Required, "values")
		}
	}

	return root
}

func jsonSchemaForInput(ctx context.Context, op *proto.Operation, input *proto.OperationInput, schema *proto.Schema) JSONSchema {
	if op.Type == proto.OperationType_OPERATION_TYPE_LIST {
		return jsonSchemaForQueryObject(ctx, op, input, schema)
	}

	prop := JSONSchema{}

	switch input.Type.Type {
	case proto.Type_TYPE_ID:
		prop.Type = "string"
	case proto.Type_TYPE_STRING:
		prop.Type = "string"
	case proto.Type_TYPE_DATE:
		prop.Type = "string"
		prop.Format = "date"
	case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		prop.Type = "string"
		prop.Format = "date-time"
	case proto.Type_TYPE_BOOL:
		prop.Type = "boolean"
	case proto.Type_TYPE_INT:
		prop.Type = "number"
	case proto.Type_TYPE_ENUM:
		prop.Type = "string"
		enum, _ := lo.Find(schema.Enums, func(e *proto.Enum) bool {
			return e.Name == input.Type.EnumName.Value
		})
		for _, v := range enum.Values {
			prop.Enum = append(prop.Enum, v.Name)
		}
	}

	isImplicit := input.Behaviour == proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT
	isWrite := input.Mode == proto.InputMode_INPUT_MODE_WRITE

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
			// TODO: when we want to support OpenAPI generation we'll need
			// to support both styles. For now we only support JSON Schema.
			prop.Type = []any{prop.Type, "null"}
		}
	}

	return prop
}

func jsonSchemaForQueryObject(ctx context.Context, op *proto.Operation, input *proto.OperationInput, schema *proto.Schema) JSONSchema {
	switch input.Type.Type {
	case proto.Type_TYPE_ID:
		t := JSONSchema{
			Type: "string",
		}
		return JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"equals": t,
				"oneOf":  {Type: "array", Items: &t},
			},
		}
	case proto.Type_TYPE_STRING:
		t := JSONSchema{
			Type: "string",
		}
		return JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"equals":     t,
				"startsWith": t,
				"endsWith":   t,
				"contains":   t,
				"oneOf":      {Type: "array", Items: &t},
			},
		}
	case proto.Type_TYPE_DATE:
		t := JSONSchema{
			Type:   "string",
			Format: "date",
		}
		return JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"equals":     t,
				"before":     t,
				"onOrBefore": t,
				"after":      t,
				"onOrAfter":  t,
			},
		}
	case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		t := JSONSchema{
			Type:   "string",
			Format: "date-time",
		}
		return JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"before": t,
				"after":  t,
			},
		}
	case proto.Type_TYPE_BOOL:
		return JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"equals": {Type: "boolean"},
			},
		}
	case proto.Type_TYPE_INT:
		t := JSONSchema{
			Type: "number",
		}
		return JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"equals":              t,
				"lessThan":            t,
				"lessThanOrEquals":    t,
				"greaterThan":         t,
				"greaterThanOrEquals": t,
			},
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
		return JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"equals": t,
				"oneOf": {
					Type:  "array",
					Items: &t,
				},
			},
		}
	}

	return JSONSchema{}
}
