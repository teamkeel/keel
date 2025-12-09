package openapi

import (
	"context"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/jsonschema"
	"github.com/teamkeel/keel/runtime/tasks"
)

// GenerateTasks generates an openAPI schema for the Tasks API for the given schema.
func GenerateTasks(ctx context.Context, schema *proto.Schema) OpenAPI {
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

	taskSchema := jsonschema.JSONSchema{
		Type: "object",
		Properties: map[string]jsonschema.JSONSchema{
			"id":   {Type: "string"},
			"name": {Type: "string"},
			"status": {
				Type: "string",
				Enum: []*string{
					StringPointer(string(tasks.StatusNew)),
					StringPointer(string(tasks.StatusAssigned)),
					StringPointer(string(tasks.StatusCompleted)),
					StringPointer(string(tasks.StatusDeferred)),
					StringPointer(string(tasks.StatusCancelled)),
					StringPointer(string(tasks.StatusStarted)),
				},
			},
			"flowRunId":     {Type: []string{"string", "null"}},
			"createdAt":     {Type: "string", Format: "date-time"},
			"updatedAt":     {Type: "string", Format: "date-time"},
			"assignedTo":    {Type: []string{"string", "null"}},
			"assignedAt":    {Type: "string", Format: "date-time"},
			"resolvedAt":    {Type: "string", Format: "date-time"},
			"deferredUntil": {Type: "string", Format: "date-time"},
		},
		Required: []string{"id", "name", "createdAt", "updatedAt"},
	}

	topicSchema := jsonschema.JSONSchema{
		Type: "object",
		Properties: map[string]jsonschema.JSONSchema{
			"name": {
				Type: "string",
			},
			"metrics": {
				Type: []string{"object", "null"},
				Ref:  "#/components/schemas/Metrics",
			},
			"stats": {
				Type: []string{"object", "null"},
				Ref:  "#/components/schemas/Stats",
			},
		},
		Required: []string{"name"},
	}

	metricsSchema := jsonschema.JSONSchema{
		Type: "object",
		Properties: map[string]jsonschema.JSONSchema{
			"totalCount": {
				Type: "number",
			},
			"completedToday": {
				Type: "number",
			},
			"oldestUnresolved": {
				Type:   []string{"string", "null"},
				Format: "date-time",
			},
		},
		Required: []string{"totalCount", "completedToday"},
	}

	statsSchema := jsonschema.JSONSchema{
		Type: "object",
		Properties: map[string]jsonschema.JSONSchema{
			"openCount": {
				Type: "number",
			},
			"assignedCount": {
				Type: "number",
			},
			"deferredCount": {
				Type: "number",
			},
			"completionRate": {
				Type: "number",
			},
			"completionTimeMedian": {
				Type: []string{"number", "null"},
			},
			"completionTime90P": {
				Type: []string{"number", "null"},
			},
			"completionTime99P": {
				Type: []string{"number", "null"},
			},
		},
		Required: []string{"openCount", "assignedCount", "deferredCount", "completionRate"},
	}

	topicResponse := map[string]ResponseObject{
		"200": {
			Description: "Topic Response",
			Content: map[string]MediaTypeObject{
				"application/json": {
					Schema: jsonschema.JSONSchema{Ref: "#/components/schemas/Topic"},
				},
			},
		},
		"400": {
			Description: "Topic Response Errors",
			Content: map[string]MediaTypeObject{
				"application/json": {
					Schema: responseErrorSchema,
				},
			},
		},
	}

	spec := OpenAPI{
		OpenAPI: "3.1.0",
		Info: InfoObject{
			Title:   "TasksAPI",
			Version: "1",
		},
		Paths: map[string]PathItemObject{},
		Components: &ComponentsObject{
			Schemas: map[string]jsonschema.JSONSchema{
				"Task":    taskSchema,
				"Topic":   topicSchema,
				"Metrics": metricsSchema,
				"Stats":   statsSchema,
			},
		},
	}

	taskResponse := map[string]ResponseObject{
		"200": {
			Description: "Task Response",
			Content: map[string]MediaTypeObject{
				"application/json": {
					Schema: jsonschema.JSONSchema{Ref: "#/components/schemas/Task"},
				},
			},
		},
		"400": {
			Description: "Task Response Errors",
			Content: map[string]MediaTypeObject{
				"application/json": {
					Schema: responseErrorSchema,
				},
			},
		},
	}

	taskIdParam := ParameterObject{
		Name:     "taskId",
		In:       "path",
		Required: true,
		Schema: jsonschema.JSONSchema{
			Type: "string",
		},
	}
	topicParam := ParameterObject{
		Name:     "topic",
		In:       "path",
		Required: true,
		Schema: jsonschema.JSONSchema{
			Type: "string",
		},
	}

	spec.Paths = map[string]PathItemObject{}

	spec.Paths["/topics/json"] = PathItemObject{
		Get: &OperationObject{
			OperationID: StringPointer("listTopics"),
			Responses: map[string]ResponseObject{
				"200": {
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: jsonschema.JSONSchema{
								Type: "object",
								Properties: map[string]jsonschema.JSONSchema{
									"topics": {
										Type:  "array",
										Items: &jsonschema.JSONSchema{Ref: "#/components/schemas/Topic"},
									},
								},
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

	spec.Paths["/topics/json/{topic}"] = PathItemObject{
		Parameters: []ParameterObject{topicParam},
		Get: &OperationObject{
			OperationID: StringPointer("getTopic"),
			Responses:   topicResponse,
		},
	}
	spec.Paths["/topics/json/{topic}/stats"] = PathItemObject{
		Parameters: []ParameterObject{topicParam},
		Get: &OperationObject{
			OperationID: StringPointer("getTopicStats"),
			Responses:   topicResponse,
		},
	}

	spec.Paths["/topics/json/{topic}/tasks"] = PathItemObject{
		Parameters: func() []ParameterObject {
			return append(paginationParams, topicParam)
		}(),
		Get: &OperationObject{
			OperationID: StringPointer("getTasks"),
			Responses: map[string]ResponseObject{
				"200": {
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: jsonschema.JSONSchema{
								Type: "object",
								Properties: map[string]jsonschema.JSONSchema{
									"tasks": {
										Type:  "array",
										Items: &jsonschema.JSONSchema{Ref: "#/components/schemas/Task"},
									},
								},
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
		Post: &OperationObject{
			OperationID: StringPointer("createTask"),
			RequestBody: &RequestBodyObject{
				Content: map[string]MediaTypeObject{
					"application/json": {
						Schema: jsonschema.JSONSchema{
							Type:                 "object",
							AdditionalProperties: BoolPointer(false),
							Properties: map[string]jsonschema.JSONSchema{
								"defer_until": {Type: "string", Format: "date-time"},
							},
						},
					},
				},
			},
			Responses: map[string]ResponseObject{
				"200": {
					Content: map[string]MediaTypeObject{
						"application/json": {
							Schema: jsonschema.JSONSchema{Ref: "#/components/schemas/Task"},
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

	spec.Paths["/topics/json/{topic}/tasks/{taskId}/complete"] = PathItemObject{
		Parameters: []ParameterObject{topicParam, taskIdParam},
		Put: &OperationObject{
			OperationID: StringPointer("completeTask"),
			Responses:   taskResponse,
		},
	}
	spec.Paths["/topics/json/{topic}/tasks/{taskId}/defer"] = PathItemObject{
		Parameters: []ParameterObject{topicParam, taskIdParam},
		Put: &OperationObject{
			OperationID: StringPointer("deferTask"),
			Responses:   taskResponse,
			RequestBody: &RequestBodyObject{
				Content: map[string]MediaTypeObject{
					"application/json": {
						Schema: jsonschema.JSONSchema{
							Type:                 "object",
							AdditionalProperties: BoolPointer(false),
							Properties: map[string]jsonschema.JSONSchema{
								"defer_until": {Type: "string", Format: "date-time"},
							},
							Required: []string{"defer_until"},
						},
					},
				},
			},
		},
	}
	spec.Paths["/topics/json/{topic}/tasks/{taskId}/assign"] = PathItemObject{
		Parameters: []ParameterObject{topicParam, taskIdParam},
		Put: &OperationObject{
			OperationID: StringPointer("assignTask"),
			Responses:   taskResponse,
			RequestBody: &RequestBodyObject{
				Content: map[string]MediaTypeObject{
					"application/json": {
						Schema: jsonschema.JSONSchema{
							Type:                 "object",
							AdditionalProperties: BoolPointer(false),
							Properties: map[string]jsonschema.JSONSchema{
								"assigned_to": {Type: "string"},
							},
							Required: []string{"assigned_to"},
						},
					},
				},
			},
		},
	}

	spec.Paths["/topics/json/{topic}/tasks/next"] = PathItemObject{
		Parameters: []ParameterObject{topicParam},
		Post: &OperationObject{
			OperationID: StringPointer("nextTask"),
			Responses:   taskResponse,
		},
	}

	spec.Paths["/topics/json/{topic}/tasks/{taskId}/start"] = PathItemObject{
		Parameters: []ParameterObject{topicParam, taskIdParam},
		Put: &OperationObject{
			OperationID: StringPointer("startTask"),
			Responses:   taskResponse,
		},
	}

	spec.Paths["/topics/json/{topic}/tasks/{taskId}/cancel"] = PathItemObject{
		Parameters: []ParameterObject{topicParam, taskIdParam},
		Put: &OperationObject{
			OperationID: StringPointer("cancelTask"),
			Responses:   taskResponse,
		},
	}

	spec.Paths["/topics/json/{topic}/tasks/{taskId}/unassign"] = PathItemObject{
		Parameters: []ParameterObject{topicParam, taskIdParam},
		Put: &OperationObject{
			OperationID: StringPointer("unassignTask"),
			Responses:   taskResponse,
		},
	}

	return spec
}
