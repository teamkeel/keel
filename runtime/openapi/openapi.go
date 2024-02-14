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
	Post *OperationObject `json:"post,omitempty"`
	Get  *OperationObject `json:"get,omitempty"`
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
			Title:   api.Name,
			Version: "1",
		},
		Paths: map[string]PathItemObject{},
	}

	components := ComponentsObject{
		Schemas: map[string]jsonschema.JSONSchema{},
	}

	for _, actionName := range proto.GetActionNamesForApi(schema, api) {
		action := proto.FindAction(schema, actionName)

		inputSchema := jsonschema.JSONSchemaForActionInput(ctx, schema, action)
		endpoint := fmt.Sprintf("/%s/json/%s", strings.ToLower(api.Name), action.Name)

		// Merge components from this request schema into OpenAPI components
		if inputSchema.Components != nil {
			for name, comp := range inputSchema.Components.Schemas {
				components.Schemas[name] = comp
			}
			inputSchema.Components = nil
		}

		responseSchema := jsonschema.JSONSchemaForActionResponse(ctx, schema, action)

		if responseSchema.Components != nil {
			for name, comp := range responseSchema.Components.Schemas {
				components.Schemas[name] = comp
			}

			responseSchema.Components = nil
		}

		spec.Paths[endpoint] = PathItemObject{
			Post: &OperationObject{
				OperationID: &action.Name,
				RequestBody: &RequestBodyObject{
					Description: action.Name + " Request",
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: inputSchema,
						},
					},
				},
				Responses: map[string]ResponseObject{
					"200": {
						Description: action.Name + " Response",
						Content: map[string]MediaTypeObject{
							"application/json": {
								Schema: responseSchema,
							},
						},
					},
					"400": {
						Description: action.Name + " Response Errors",
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

	for _, job := range schema.Jobs {
		if job.Name == jobName {
			msg := proto.FindMessage(schema.Messages, job.InputMessageName)
			if msg == nil {
				continue
			}
			inputSchema := jsonschema.JSONSchemaForMessage(ctx, schema, nil, msg)
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
						Description: job.Name + " Request",
						Content: map[string]MediaTypeObject{
							"application/json": {
								Schema: inputSchema,
							},
						},
					},
					Responses: map[string]ResponseObject{
						"200": {
							Description: job.Name + " Response",
							Content: map[string]MediaTypeObject{
								"application/json": {
									Schema: responseSchema,
								},
							},
						},
						"400": {
							Description: job.Name + " Response Errors",
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
