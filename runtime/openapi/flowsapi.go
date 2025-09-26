package openapi

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"maps"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/flows"
	"github.com/teamkeel/keel/runtime/jsonschema"
)

//go:embed uiConfig.json
var uiConfigRaw []byte

//go:embed flowConfig.json
var flowConfigRaw []byte

// GenerateFlows generates an openAPI schema for the Flows API for the given schema.
func GenerateFlows(ctx context.Context, schema *proto.Schema) OpenAPI {
	var flowConfigSchema jsonschema.JSONSchema
	_ = json.Unmarshal(flowConfigRaw, &flowConfigSchema)

	paginationParams := []ParameterObject{
		{
			Name:     "limit",
			In:       "query",
			Required: false,
			Schema:   jsonschema.JSONSchema{Type: "number"},
		},
		{
			Name:     "before",
			In:       "query",
			Required: false,
			Schema:   jsonschema.JSONSchema{Type: "string"},
		},
		{
			Name:     "after",
			In:       "query",
			Required: false,
			Schema:   jsonschema.JSONSchema{Type: "string"},
		},
	}

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
			"startTime": {Type: "string", Format: "date-time"},
			"createdAt": {Type: "string", Format: "date-time"},
			"updatedAt": {Type: "string", Format: "date-time"},
			"steps":     {Type: "array", Items: &jsonschema.JSONSchema{Ref: "#/components/schemas/Step"}},
			"config":    flowConfigSchema,
			"input":     {Type: []string{"object", "null"}, AdditionalProperties: BoolPointer(true)},
			"startedBy": {Type: []string{"string", "null"}},
			"data":      {Type: []string{"object", "null"}},
		},
		Required: []string{"id", "status", "name", "traceId", "createdAt", "updatedAt", "steps", "config", "input"},
	}

	statsResponseSchema := jsonschema.JSONSchema{
		Type: "object",
		Properties: map[string]jsonschema.JSONSchema{
			"name":           {Type: "string"},
			"lastRun":        {Type: "string", Format: "date-time"},
			"totalRuns":      {Type: "number"},
			"errorRate":      {Type: "number"},
			"activeRuns":     {Type: "number"},
			"completedToday": {Type: "number"},
			"timeSeries":     {Type: []string{"array", "null"}, Items: &jsonschema.JSONSchema{Ref: "#/components/schemas/StatsBucket"}},
		},
		Required: []string{"name", "lastRun", "totalRuns", "errorRate", "activeRuns", "completedToday"},
	}

	statsBucketResponseSchema := jsonschema.JSONSchema{
		Type: "object",
		Properties: map[string]jsonschema.JSONSchema{
			"time":       {Type: "string", Format: "date-time"},
			"totalRuns":  {Type: "number"},
			"failedRuns": {Type: "number"},
		},
		Required: []string{"time", "totalRuns", "failedRunts"},
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
					StringPointer(string(flows.StepStatusCancelled)),
				},
			},
			"type": {
				Type: "string",
				Enum: []*string{
					StringPointer(string(flows.StepTypeFunction)),
					StringPointer(string(flows.StepTypeUI)),
					StringPointer(string(flows.StepTypeComplete)),
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
						"name":     {Type: "string"},
						"schedule": {Type: []string{"string", "null"}},
					},
				},
			},
		},
	}

	// Remap the $ref paths in the uiConfigRaw to point to the correct location in the components
	uiConfigRaw = bytes.ReplaceAll(uiConfigRaw, []byte("#/$defs/"), []byte("#/components/schemas/__UiConfigSchemas/$defs/"))
	var uiConfigSchema jsonschema.JSONSchema
	_ = json.Unmarshal(uiConfigRaw, &uiConfigSchema)

	// Create a copy of uiConfigSchema without the definitions
	uiConfigSchemaWithoutDefs := uiConfigSchema
	uiConfigSchemaWithoutDefs.Definitions = nil

	spec := OpenAPI{
		OpenAPI: "3.1.0",
		Info: InfoObject{
			Title:   "FlowsAPI",
			Version: "1",
		},
		Paths: map[string]PathItemObject{},
		Components: &ComponentsObject{
			Schemas: map[string]jsonschema.JSONSchema{
				"Run":               runResponseSchema,
				"Step":              stepResponseSchema,
				"Stats":             statsResponseSchema,
				"StatsBucket":       statsBucketResponseSchema,
				"UiConfig":          uiConfigSchemaWithoutDefs,
				"__UiConfigSchemas": {Definitions: uiConfigSchema.Definitions},
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
	for _, flow := range schema.GetAllFlows() {
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

	spec.Paths["/flows/json/stats"] = PathItemObject{
		Parameters: []ParameterObject{
			{
				Name:     "before",
				In:       "query",
				Required: false,
				Schema:   jsonschema.JSONSchema{Type: "string", Format: "date-time"},
			},
			{
				Name:     "after",
				In:       "query",
				Required: false,
				Schema:   jsonschema.JSONSchema{Type: "string", Format: "date-time"},
			},
			{
				Name:        "interval",
				In:          "query",
				Required:    false,
				Description: "If supplied, the results will include a time series with buckets defined by this interval period.",
				Schema: jsonschema.JSONSchema{Type: "string", Enum: []*string{
					StringPointer(flows.StatsIntervalDaily),
					StringPointer(flows.StatsIntervalHourly),
				}},
			},
		},
		Get: &OperationObject{
			OperationID: StringPointer("getRunsStats"),
			Responses: map[string]ResponseObject{
				"200": {
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: jsonschema.JSONSchema{
								Type:  "array",
								Items: &jsonschema.JSONSchema{Ref: "#/components/schemas/Stats"},
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

	spec.Paths["/flows/json/myRuns"] = PathItemObject{
		Parameters: func() []ParameterObject {
			return append(paginationParams, ParameterObject{
				Name:     "status",
				In:       "query",
				Required: false,
				Schema: jsonschema.JSONSchema{
					Type: "array",
					Items: &jsonschema.JSONSchema{
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
				},
				Style:   "form",
				Explode: BoolPointer(false),
			})
		}(),
		Get: &OperationObject{
			OperationID: StringPointer("getMyRuns"),
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

	spec.Paths["/flows/json/{flow}"] = PathItemObject{
		Parameters: func() []ParameterObject {
			return append(paginationParams, ParameterObject{
				Name:     "flow",
				In:       "path",
				Required: true,
				Schema: jsonschema.JSONSchema{
					Type: "string",
				},
			})
		}(),
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

	spec.Paths["/flows/json/{flow}/{runId}/back"] = PathItemObject{
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
			OperationID: StringPointer("backFlowRun"),
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
			{
				Name:     "action",
				In:       "query",
				Required: false,
				Schema: jsonschema.JSONSchema{
					Type: "string",
				},
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

	spec.Paths["/flows/json/{flow}/{runId}/{stepId}/callback"] = PathItemObject{
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
			{
				Name:     "callback",
				In:       "query",
				Required: true,
				Schema: jsonschema.JSONSchema{
					Type: "string",
				},
			},
			{
				Name:     "element",
				In:       "query",
				Required: true,
				Schema: jsonschema.JSONSchema{
					Type: "string",
				},
			},
		},
		Post: &OperationObject{
			OperationID: StringPointer("callback"),
			RequestBody: &RequestBodyObject{
				Content: map[string]MediaTypeObject{
					"application/json": {
						Schema: jsonschema.JSONSchema{Type: []string{"object", "string", "number", "boolean", "array"}, AdditionalProperties: BoolPointer(true)},
					},
				},
			},
			Responses: map[string]ResponseObject{
				"200": {
					Description: "Callback Response",
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: jsonschema.JSONSchema{Type: []string{"object", "string", "number", "boolean", "array"}, AdditionalProperties: BoolPointer(true)},
						},
					},
				},
				"400": {
					Description: "Callback Response Errors",
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: responseErrorSchema,
						},
					},
				},
			},
		},
	}

	return spec
}
