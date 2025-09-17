package tasksapi

import (
	"errors"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/apis/httpjson"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/locale"
	"github.com/teamkeel/keel/runtime/openapi"
	"github.com/teamkeel/keel/runtime/tasks"
	"github.com/teamkeel/keel/schema/parser"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/apis/tasksapi")

func Handler(s *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "TasksAPI")
		defer span.End()

		identity, err := actions.HandleAuthorizationHeader(ctx, s, r.Header)
		if err != nil {
			return httpjson.NewErrorResponse(ctx, err, nil)
		}
		if identity != nil {
			ctx = auth.WithIdentity(ctx, identity)
		}

		identityID := identity[parser.FieldNameId].(string)
		if identityID == "" {
			return httpjson.NewErrorResponse(ctx, common.NewPermissionError(), nil)
		}

		path := path.Clean(r.URL.EscapedPath())
		pathParts := strings.Split(strings.TrimPrefix(path, "/topics/json/"), "/")

		topic := s.FindTask(pathParts[0])
		if topic == nil {
			return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
		}

		// authorise that the user is allowed to access this topic
		authorised, err := tasks.AuthoriseTopic(ctx, s, topic)
		if err != nil {
			return httpjson.NewErrorResponse(ctx, err, nil)
		}
		if !authorised {
			return httpjson.NewErrorResponse(ctx, common.NewPermissionError(), nil)
		}

		switch len(pathParts) {
		case 1:
			// GET topics/{name} - Retrieves specific topic with basic metric data
			if r.Method != http.MethodGet {
				return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP GET accepted"), nil)
			}
			tData, err := tasks.GetTopic(ctx, topic, false)
			if err != nil {
				return httpjson.NewErrorResponse(ctx, err, nil)
			}

			return common.NewJsonResponse(http.StatusOK, tData, nil)
		case 2:
			// GET topics/{name}/stats - Retrieves the topic’s detailed metric data
			// POST topics/{name}/tasks - Creates a new task for the queue
			// GET topics/{name}/tasks - Retrieves all tasks in a topic’s queue
			switch pathParts[1] {
			case "stats":
				// GET topics/{name}/stats - Retrieves the topic’s detailed metric data
				if r.Method != http.MethodGet {
					return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP GET accepted"), nil)
				}
				tData, err := tasks.GetTopic(ctx, topic, true)
				if err != nil {
					return httpjson.NewErrorResponse(ctx, err, nil)
				}

				return common.NewJsonResponse(http.StatusOK, tData, nil)
			case "tasks":
				switch r.Method {
				case http.MethodGet:
					// GET topics/{name}/tasks - Retrieves all tasks in a topic’s queue
					tasks, err := tasks.ListTasks(ctx, topic, common.ParseQueryParams(r))
					if err != nil {
						return httpjson.NewErrorResponse(ctx, err, nil)
					}

					return common.NewJsonResponse(http.StatusOK, tasks, nil)
				case http.MethodPost:
					// POST topics/{name}/tasks - Creates a new task for the queue
					// parse input
					inputs, err := common.ParseRequestData(r)
					if err != nil {
						return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("error parsing POST body"), nil)
					}

					inputsMap, ok := inputs.(map[string]any)
					if inputs == nil || !ok {
						return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("data not correctly formatted"), nil)
					}

					var deferUntil *time.Time
					strDate, ok := inputsMap["defer_until"].(string)
					if ok {
						date, err := time.Parse(time.RFC3339, strDate)
						if err != nil {
							return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("defer_until not correctly formatted"), nil)
						}
						deferUntil = &date
					}

					task, err := tasks.CreateTask(ctx, topic, identityID, deferUntil)
					if err != nil {
						if errors.Is(err, tasks.ErrTaskNotFound) {
							return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
						}

						return httpjson.NewErrorResponse(ctx, err, nil)
					}

					return common.NewJsonResponse(http.StatusOK, task, nil)
				}

				return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP GET or POST accepted"), nil)
			}
		case 3:
			// TODO: POST topics/{name}/tasks/next - Assigns next task to me (or returns my current task)
			return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
		case 4:
			// PUT topics/{name}/tasks/{id}/complete - Completes a task
			// PUT topics/{name}/tasks/{id}/defer - Defers a task until a later period
			// PUT topics/{name}/tasks/{id}/assign - Assigns the task to a new identity
			if r.Method != http.MethodPut {
				return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP PUT accepted"), nil)
			}

			switch pathParts[3] {
			case "complete":
				task, err := tasks.CompleteTask(ctx, topic, pathParts[2], identityID)
				if err != nil {
					if errors.Is(err, tasks.ErrTaskNotFound) {
						return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
					}

					return httpjson.NewErrorResponse(ctx, err, nil)
				}

				return common.NewJsonResponse(http.StatusOK, task, nil)
			case "defer":
				// parse input
				inputs, err := common.ParseRequestData(r)
				if err != nil {
					return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("error parsing POST body"), nil)
				}

				inputsMap, ok := inputs.(map[string]any)
				if inputs == nil || !ok {
					return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("data not correctly formatted"), nil)
				}

				strDate, ok := inputsMap["defer_until"].(string)
				if !ok {
					return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("data not correctly formatted"), nil)
				}

				deferUntil, err := time.Parse(time.RFC3339, strDate)
				if err != nil {
					return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("date not correctly formatted"), nil)
				}

				// defer task
				task, err := tasks.DeferTask(ctx, topic, pathParts[2], deferUntil, identityID)
				if err != nil {
					if errors.Is(err, tasks.ErrTaskNotFound) {
						return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
					}

					return httpjson.NewErrorResponse(ctx, err, nil)
				}

				return common.NewJsonResponse(http.StatusOK, task, nil)
			case "assign":
				// parse input
				inputs, err := common.ParseRequestData(r)
				if err != nil {
					return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("error parsing POST body"), nil)
				}

				inputsMap, ok := inputs.(map[string]any)
				if inputs == nil || !ok {
					return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("data not correctly formatted"), nil)
				}

				assignedTo, ok := inputsMap["assigned_to"].(string)
				if !ok {
					return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError("data not correctly formatted"), nil)
				}

				// assign task
				task, err := tasks.AssignTask(ctx, topic, pathParts[2], assignedTo, identityID)
				if err != nil {
					if errors.Is(err, tasks.ErrTaskNotFound) {
						return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
					}

					return httpjson.NewErrorResponse(ctx, err, nil)
				}

				return common.NewJsonResponse(http.StatusOK, task, nil)
			}
		}

		return httpjson.NewErrorResponse(ctx, common.NewNotFoundError("Not found"), nil)
	}
}

// ListTopicsHandler handles a request to /topics and returns data about all topics defined in the schema.
func ListTopicsHandler(p *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "TasksAPI")
		defer span.End()

		identity, err := actions.HandleAuthorizationHeader(ctx, p, r.Header)
		if err != nil {
			return httpjson.NewErrorResponse(ctx, err, nil)
		}
		if identity != nil {
			ctx = auth.WithIdentity(ctx, identity)
		}

		// handle any Time-Zone headers
		location, err := locale.HandleTimezoneHeader(ctx, r.Header)
		if err != nil {
			return httpjson.NewErrorResponse(ctx, common.NewInputMalformedError(err.Error()), nil)
		}
		ctx = locale.WithTimeLocation(ctx, location)

		if r.Method != http.MethodGet {
			return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP GET accepted"), nil)
		}

		authorisedTopics, err := tasks.AuthorisedTopics(ctx, p)
		if err != nil {
			return httpjson.NewErrorResponse(ctx, err, nil)
		}

		topics := []*tasks.Topic{}
		for _, t := range authorisedTopics {
			topic, err := tasks.GetTopic(ctx, t, false)
			if err != nil {
				return httpjson.NewErrorResponse(ctx, err, nil)
			}
			topics = append(topics, topic)
		}
		return common.NewJsonResponse(http.StatusOK, map[string]any{"topics": topics}, nil)
	}
}

func OpenAPISchemaHandler(p *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		ctx, span := tracer.Start(r.Context(), "TasksAPI")
		defer span.End()
		span.SetAttributes(
			attribute.String("api.protocol", "HTTP JSON"),
		)

		if r.Method != http.MethodGet {
			return httpjson.NewErrorResponse(ctx, common.NewHttpMethodNotAllowedError("only HTTP GET accepted"), nil)
		}

		sch := openapi.GenerateTasks(ctx, p)
		return common.NewJsonResponse(http.StatusOK, sch, nil)
	}
}
