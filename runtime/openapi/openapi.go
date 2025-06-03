package openapi

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/flows"
	"github.com/teamkeel/keel/runtime/jsonschema"
)

//go:embed uiConfig.json
var uiConfigRaw []byte

//go:embed flowConfig.json
var flowConfigRaw []byte

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

// GenerateFlows generates an openAPI schema for the Flows API for the given schema.
func GenerateFlows(ctx context.Context, schema *proto.Schema) OpenAPI {
	var flowConfigSchema jsonschema.JSONSchema
	_ = json.Unmarshal(flowConfigRaw, &flowConfigSchema)

	runResponseSchema := jsonschema.JSONSchema{
		Type: "object",
		Properties: map[string]jsonschema.JSONSchema{
			"id": {Type: "string"},
			"status": {
				Type: "string",
				Enum: []*string{
					StringPointer(string(flows.StatusNew)),
					StringPointer(string(flows.StatusRunning)),
					StringPointer(string(flows.StatusAwaitingInput)),
					StringPointer(string(flows.StatusFailed)),
					StringPointer(string(flows.StatusCompleted)),
					StringPointer(string(flows.StatusCancelled)),
				},
			},
			"name":      {Type: "string"},
			"traceId":   {Type: "string"},
			"createdAt": {Type: "string", Format: "date-time"},
			"updatedAt": {Type: "string", Format: "date-time"},
			"steps":     {Type: "array", Items: &jsonschema.JSONSchema{Ref: "#/components/schemas/Step"}},
			"config":    flowConfigSchema,
			"input":     {Type: []string{"object", "null"}, AdditionalProperties: BoolPointer(true)},
			"startedBy": {Type: []string{"string", "null"}},
		},
		Required: []string{"id", "status", "name", "traceId", "createdAt", "updatedAt", "steps", "config", "input"},
	}

	anyTypeSchema := jsonschema.JSONSchema{
		Type:                 []string{"string", "object", "array", "integer", "number", "boolean", "null"},
		AdditionalProperties: BoolPointer(true),
	}

	stepResponseSchema := jsonschema.JSONSchema{
		Type: "object",
		Properties: map[string]jsonschema.JSONSchema{
			"id":    {Type: "string"},
			"name":  {Type: "string"},
			"runId": {Type: "string"},
			"status": {
				Type: "string",
				Enum: []*string{
					StringPointer(string(flows.StepStatusPending)),
					StringPointer(string(flows.StepStatusFailed)),
					StringPointer(string(flows.StepStatusCompleted)),
				},
			},
			"type": {
				Type: "string",
				Enum: []*string{
					StringPointer(string(flows.StepTypeFunction)),
					StringPointer(string(flows.StepTypeUI)),
				},
			},
			"value":     anyTypeSchema,
			"startTime": {Type: []string{"string", "null"}, Format: "date-time"},
			"endTime":   {Type: []string{"string", "null"}, Format: "date-time"},
			"createdAt": {Type: "string", Format: "date-time"},
			"updatedAt": {Type: "string", Format: "date-time"},
			"ui":        {Ref: "#/components/schemas/UiConfig"},
			"error":     {Type: []string{"string", "null"}},
			"stage":     {Type: []string{"string", "null"}},
		},
		Required: []string{"id", "runId", "status", "name", "type", "createdAt", "updatedAt", "value", "ui", "startTime", "endTime", "error"},
	}

	listFlowsResponseSchema := jsonschema.JSONSchema{
		Type: "object",
		Properties: map[string]jsonschema.JSONSchema{
			"flows": {
				Type: "array",
				Items: &jsonschema.JSONSchema{
					Type: "object",
					Properties: map[string]jsonschema.JSONSchema{
						"name": {
							Type: "string",
						},
					},
				},
			},
		},
	}

	// Remap the $ref paths in the uiConfigRaw to point to the correct location in the components
	uiConfigRaw = bytes.ReplaceAll(uiConfigRaw, []byte("#/$defs/"), []byte("#/components/schemas/UiConfig/$defs/"))
	var uiConfigSchema jsonschema.JSONSchema
	_ = json.Unmarshal(uiConfigRaw, &uiConfigSchema)

	spec := OpenAPI{
		OpenAPI: "3.1.0",
		Info: InfoObject{
			Title:   "FlowsAPI",
			Version: "1",
		},
		Paths: map[string]PathItemObject{},
		Components: &ComponentsObject{
			Schemas: map[string]jsonschema.JSONSchema{
				"Run":      runResponseSchema,
				"Step":     stepResponseSchema,
				"UiConfig": uiConfigSchema,
			},
		},
	}

	flowRunResponse := map[string]ResponseObject{
		"200": {
			Description: "Flow Response",
			Content: map[string]MediaTypeObject{
				"application/json": {
					Schema: jsonschema.JSONSchema{Ref: "#/components/schemas/Run"},
				},
			},
		},
		"400": {
			Description: "Flow Response Errors",
			Content: map[string]MediaTypeObject{
				"application/json": {
					Schema: responseErrorSchema,
				},
			},
		},
	}

	spec.Paths = map[string]PathItemObject{}

	// Add specific flows endpoints with defined inputs
	for _, flow := range schema.GetFlows() {
		msg := schema.FindMessage(flow.GetInputMessageName())
		if msg == nil {
			continue
		}
		inputSchema := jsonschema.JSONSchemaForMessage(ctx, schema, nil, msg, true)
		endpoint := "/flows/json/" + flow.GetName()

		// Merge components from this request schema into OpenAPI components
		if inputSchema.Components != nil {
			maps.Copy(spec.Components.Schemas, inputSchema.Components.Schemas)
			inputSchema.Components = nil
		}

		spec.Paths[endpoint] = PathItemObject{
			Post: &OperationObject{
				OperationID: &flow.Name,
				RequestBody: &RequestBodyObject{
					Description: flow.GetName() + " Request",
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: inputSchema,
						},
					},
				},
				Responses: flowRunResponse,
			},
		}
	}

	spec.Paths["/flows/json"] = PathItemObject{
		Get: &OperationObject{
			OperationID: StringPointer("listFlows"),
			Responses: map[string]ResponseObject{
				"200": {
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: listFlowsResponseSchema,
						},
					},
				},
				"400": {
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: responseErrorSchema,
						},
					},
				},
			},
		},
	}

	spec.Paths["/flows/json/{flow}"] = PathItemObject{
		Parameters: []ParameterObject{
			{
				Name:     "flow",
				In:       "path",
				Required: true,
				Schema: jsonschema.JSONSchema{
					Type: "string",
				},
			},
		},
		Post: &OperationObject{
			OperationID: StringPointer("startFlow"),
			RequestBody: &RequestBodyObject{
				Content: map[string]MediaTypeObject{
					"application/json": {
						Schema: jsonschema.JSONSchema{Type: "object", AdditionalProperties: BoolPointer(true)},
					},
				},
			},
			Responses: flowRunResponse,
		},
		Get: &OperationObject{
			OperationID: StringPointer("getFlow"),
			Responses: map[string]ResponseObject{
				"200": {
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: jsonschema.JSONSchema{
								Type:  "array",
								Items: &jsonschema.JSONSchema{Ref: "#/components/schemas/Run"},
							},
						},
					},
				},
				"400": {
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: responseErrorSchema,
						},
					},
				},
			},
		},
	}

	spec.Paths["/flows/json/{flow}/{runId}"] = PathItemObject{
		Parameters: []ParameterObject{
			{
				Name:     "flow",
				In:       "path",
				Required: true,
				Schema: jsonschema.JSONSchema{
					Type: "string",
				},
			},
			{
				Name:     "runId",
				In:       "path",
				Required: true,
				Schema: jsonschema.JSONSchema{
					Type: "string",
				},
			},
		},
		Get: &OperationObject{
			OperationID: StringPointer("getFlowRun"),
			Responses:   flowRunResponse,
		},
	}

	spec.Paths["/flows/json/{flow}/{runId}/cancel"] = PathItemObject{
		Parameters: []ParameterObject{
			{
				Name:     "flow",
				In:       "path",
				Required: true,
				Schema:   jsonschema.JSONSchema{Type: "string"},
			},
			{
				Name:     "runId",
				In:       "path",
				Required: true,
				Schema:   jsonschema.JSONSchema{Type: "string"},
			},
		},
		Post: &OperationObject{
			OperationID: StringPointer("cancelFlowRun"),
			Responses:   flowRunResponse,
		},
	}

	spec.Paths["/flows/json/{flow}/{runId}/{stepId}"] = PathItemObject{
		Parameters: []ParameterObject{
			{
				Name:     "flow",
				In:       "path",
				Required: true,
				Schema: jsonschema.JSONSchema{
					Type: "string",
				},
			},
			{
				Name:     "runId",
				In:       "path",
				Required: true,
				Schema: jsonschema.JSONSchema{
					Type: "string",
				},
			},
			{
				Name:     "stepId",
				In:       "path",
				Required: true,
				Schema:   jsonschema.JSONSchema{Type: "string"},
			},
		},
		Put: &OperationObject{
			OperationID: StringPointer("putFlowStep"),
			RequestBody: &RequestBodyObject{
				Content: map[string]MediaTypeObject{
					"application/json": {
						Schema: jsonschema.JSONSchema{Type: "object", AdditionalProperties: BoolPointer(true)},
					},
				},
			},
			Responses: flowRunResponse,
		},
	}

	return spec
}
