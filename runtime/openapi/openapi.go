package openapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/jsonschema"
)

// OpenAPI spec object - https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md
type OpenAPI struct {
	OpenAPI    string                    `json:"openapi"`
	Info       InfoObject                `json:"info"`
	Paths      map[string]PathItemObject `json:"paths,omitempty"`
	Components *ComponentsObject         `json:"components,omitempty"`
}

type InfoObject struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type ComponentsObject struct {
	Schemas map[string]jsonschema.JSONSchema `json:"schemas,omitempty"`
}

type PathItemObject struct {
	Post OperationObject `json:"post,omitempty"`
}

type OperationObject struct {
	RequestBody ResponseObject            `json:"requestBody"`
	Responses   map[string]ResponseObject `json:"responses,omitempty"`
}

type RequestBodyObject struct {
	Description string                     `json:"description"`
	Content     map[string]MediaTypeObject `json:"content,omitempty"`
	Required    bool                       `json:"required"`
}

type ResponseObject struct {
	Description string                     `json:"description"`
	Content     map[string]MediaTypeObject `json:"content,omitempty"`
}

type MediaTypeObject struct {
	Schema jsonschema.JSONSchema `json:"schema,omitempty"`
}

// Generate creates an OpenAPI 3.1 spec for the passed api.
func Generate(ctx context.Context, schema *proto.Schema, api *proto.Api) OpenAPI {
	spec := OpenAPI{
		OpenAPI: "3.1.0",
		Info: InfoObject{
			Title:   api.Name,
			Version: "1",
		},
		Paths: map[string]PathItemObject{},
	}
	components := ComponentsObject{
		Schemas: map[string]jsonschema.JSONSchema{},
	}

	for _, model := range schema.Models {
		if !lo.ContainsBy(api.ApiModels, func(m *proto.ApiModel) bool {
			return model.Name == m.ModelName
		}) {
			continue
		}

		for _, op := range model.Operations {
			opSchema := jsonschema.JSONSchemaForOperation(ctx, schema, op)
			endpoint := fmt.Sprintf("/%s/json/%s", strings.ToLower(api.Name), op.Name)

			// Merge components from this request schema into OpenAPI components
			if opSchema.Components != nil {
				for name, comp := range opSchema.Components.Schemas {
					components.Schemas[name] = comp
				}
				opSchema.Components = nil
			}

			spec.Paths[endpoint] = PathItemObject{
				Post: OperationObject{
					RequestBody: ResponseObject{
						Description: op.Name + " Request",
						Content: map[string]MediaTypeObject{
							"application/json": {
								Schema: opSchema,
							},
						},
					},
					Responses: map[string]ResponseObject{
						"200": {
							Description: op.Name + " Response",
							Content:     map[string]MediaTypeObject{
								// TODO add response schema
							},
						},
						// TODO: add possible errors
					},
				},
			}
		}
	}

	if len(components.Schemas) > 0 {
		spec.Components = &components
	}

	return spec
}
