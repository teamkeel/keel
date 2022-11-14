package rpc

import (
	"context"

	"github.com/teamkeel/keel/proto"
	"github.com/xeipuuv/gojsonschema"
)

type JSONSchema struct {
	Type       string                `json:"type,omitempty"`
	Properties map[string]JSONSchema `json:"properties,omitempty"`
	Required   []string              `json:"required,omitempty"`
}

func ValidateRequest(ctx context.Context, schema *proto.Schema, op *proto.Operation, input map[string]any) (*gojsonschema.Result, error) {
	requestType := jsonSchemaForOperation(ctx, schema, op)
	return gojsonschema.Validate(gojsonschema.NewGoLoader(requestType), gojsonschema.NewGoLoader(input))
}

func jsonSchemaForOperation(ctx context.Context, schema *proto.Schema, op *proto.Operation) JSONSchema {
	s := JSONSchema{
		Properties: map[string]JSONSchema{},
	}

	for _, input := range op.Inputs {
		prop := JSONSchema{}

		switch input.Type.Type {
		case proto.Type_TYPE_ID:
			prop.Type = "string"
		case proto.Type_TYPE_STRING:
			prop.Type = "string"
		}

		s.Properties[input.Name] = prop

		if !input.Optional {
			s.Required = append(s.Required, input.Name)
		}
	}

	return s
}
