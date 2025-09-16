package openapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/jsonschema"
)

const OpenApiSpecificationVersion = "3.1.0"

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
	Post       *OperationObject  `json:"post,omitempty"`
	Get        *OperationObject  `json:"get,omitempty"`
	Put        *OperationObject  `json:"put,omitempty"`
	Parameters []ParameterObject `json:"parameters,omitempty"`
}

type ParameterObject struct {
	Name        string                `json:"name"`
	In          string                `json:"in"`
	Required    bool                  `json:"required"`
	Description string                `json:"description"`
	Schema      jsonschema.JSONSchema `json:"schema"`
	Style       string                `json:"style,omitempty"`
	Explode     *bool                 `json:"explode,omitempty"`
}

type OperationObject struct {
	OperationID *string                   `json:"operationId,omitempty"`
	RequestBody *RequestBodyObject        `json:"requestBody,omitempty"`
	Responses   map[string]ResponseObject `json:"responses,omitempty"`
}

type RequestBodyObject struct {
	Description string                     `json:"description"`
	Content     map[string]MediaTypeObject `json:"content,omitempty"`
	Required    *bool                      `json:"required,omitempty"`
}

type ResponseObject struct {
	Description string                     `json:"description"`
	Content     map[string]MediaTypeObject `json:"content,omitempty"`
}

type MediaTypeObject struct {
	Schema jsonschema.JSONSchema `json:"schema,omitempty"`
}

func StringPointer(v string) *string {
	return &v
}

func BoolPointer(v bool) *bool {
	return &v
}

var (
	responseErrorSchema = jsonschema.JSONSchema{
		Properties: map[string]jsonschema.JSONSchema{
			"code": {
				Type: "string",
			},
			"message": {
				Type: "string",
			},
			"data": {
				Type: []string{"object", "null"},
				Properties: map[string]jsonschema.JSONSchema{
					"errors": {
						Type: "array",
						Properties: map[string]jsonschema.JSONSchema{
							"error": {
								Type: "string",
							},
							"field": {
								Type: "string",
							},
						},
					},
				},
			},
		},
	}
)

// Generate creates an OpenAPI 3.1 spec for the passed api.
func Generate(ctx context.Context, schema *proto.Schema, api *proto.Api) OpenAPI {
	spec := OpenAPI{
		OpenAPI: OpenApiSpecificationVersion,
		Info: InfoObject{
			Title:   api.GetName(),
			Version: "1",
		},
		Paths: map[string]PathItemObject{},
	}

	components := ComponentsObject{
		Schemas: map[string]jsonschema.JSONSchema{},
	}

	for _, actionName := range proto.GetActionNamesForApi(schema, api) {
		action := schema.FindAction(actionName)

		var requestBody *RequestBodyObject
		if action.GetInputMessageName() != "" {
			inputSchema := jsonschema.JSONSchemaForActionInput(ctx, schema, action)

			// Merge components from this request schema into OpenAPI components
			if inputSchema.Components != nil {
				for name, comp := range inputSchema.Components.Schemas {
					components.Schemas[name] = comp
				}
				inputSchema.Components = nil
			}

			requestBody = &RequestBodyObject{
				Description: action.GetName() + " Request",
				Content: map[string]MediaTypeObject{
					"application/json": {
						Schema: inputSchema,
					},
				},
			}
		}

		responseSchema := jsonschema.JSONSchemaForActionResponse(ctx, schema, action)

		if responseSchema.Components != nil {
			// Merge components from this response schema into OpenAPI components
			for name, comp := range responseSchema.Components.Schemas {
				components.Schemas[name] = comp
			}
			responseSchema.Components = nil
		}

		endpoint := fmt.Sprintf("/%s/json/%s", strings.ToLower(api.GetName()), action.GetName())

		spec.Paths[endpoint] = PathItemObject{
			Post: &OperationObject{
				OperationID: &action.Name,
				RequestBody: requestBody,
				Responses: map[string]ResponseObject{
					"200": {
						Description: action.GetName() + " Response",
						Content: map[string]MediaTypeObject{
							"application/json": {
								Schema: responseSchema,
							},
						},
					},
					"400": {
						Description: action.GetName() + " Response Errors",
						Content: map[string]MediaTypeObject{
							"application/json": {
								Schema: responseErrorSchema,
							},
						},
					},
				},
			},
		}
	}

	if len(components.Schemas) > 0 {
		spec.Components = &components
	}

	return spec
}

func GenerateJob(ctx context.Context, schema *proto.Schema, jobName string) OpenAPI {
	// loop over jobs in schema and find the one with the name and create openapi spec for it

	spec := OpenAPI{
		OpenAPI: "3.1.0",
		Info: InfoObject{
			Title:   jobName,
			Version: "1",
		},
		Paths: map[string]PathItemObject{},
	}

	for _, job := range schema.GetJobs() {
		if job.GetName() == jobName {
			msg := schema.FindMessage(job.GetInputMessageName())
			if msg == nil {
				continue
			}
			inputSchema := jsonschema.JSONSchemaForMessage(ctx, schema, nil, msg, true)
			endpoint := "/"

			// Merge components from this request schema into OpenAPI components
			if inputSchema.Components != nil {
				for name, comp := range inputSchema.Components.Schemas {
					spec.Components.Schemas[name] = comp
				}
				inputSchema.Components = nil
			}

			responseSchema := jsonschema.JSONSchema{
				Type: "object",
				Properties: map[string]jsonschema.JSONSchema{
					"status": {
						Type: "string",
					},
				},
			}
			spec.Paths = map[string]PathItemObject{}

			spec.Paths[endpoint] = PathItemObject{
				Post: &OperationObject{
					OperationID: &job.Name,
					RequestBody: &RequestBodyObject{
						Description: job.GetName() + " Request",
						Content: map[string]MediaTypeObject{
							"application/json": {
								Schema: inputSchema,
							},
						},
					},
					Responses: map[string]ResponseObject{
						"200": {
							Description: job.GetName() + " Response",
							Content: map[string]MediaTypeObject{
								"application/json": {
									Schema: responseSchema,
								},
							},
						},
						"400": {
							Description: job.GetName() + " Response Errors",
							Content: map[string]MediaTypeObject{
								"application/json": {
									Schema: responseErrorSchema,
								},
							},
						},
					},
				},
			}
		}
	}

	return spec
}
