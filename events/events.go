package events

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/karlseguin/typed"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel/trace"
)

// Event names
const (
	Created = "created"
	Updated = "updated"
	Deleted = "deleted"
)

type Event struct {
	// The name of the event, e.g. member.created.
	EventName string `json:"eventName"`
	// The time at which the event was created.
	OccurredAt time.Time `json:"occurredAt"`
	// The identity that resulted in the triggered events.
	IdentityId string `json:"identityId,omitempty"`
	// The target impacted by this event.
	Target *EventTarget `json:"target"`
}

type EventTarget struct {
	// The id of the target, if applicable.
	Id string `json:"id"`
	// The type of event target, e.g. Employee
	Type string `json:"type"`
	// The data relevant to this target type.
	Data map[string]any `json:"data"`
}

// The event handler function to be executed for each subscriber event generated.
type EventHandler func(ctx context.Context, subscriber string, event *Event, traceparent string) error

type handlerContextKey string

var contextKey handlerContextKey = "eventHandler"

func WithEventHandler(ctx context.Context, handler EventHandler) context.Context {
	return context.WithValue(ctx, contextKey, handler)
}

func HasEventHandler(ctx context.Context) bool {
	return ctx.Value(contextKey) != nil
}

func GetEventHandler(ctx context.Context) (EventHandler, error) {
	v, ok := ctx.Value(contextKey).(EventHandler)
	if !ok {
		return nil, fmt.Errorf("context does not have key or is not EventHandler: %s", contextKey)
	}
	return v, nil
}

// SendEvents will gather, create and send events which have occurred within the scope of this context.
// It achieves this by inspecting the keel_audit table for rows which must be generated into events,
// updates the event_processed_at field on these rows, and then calls the event handler for each event.
func SendEvents(ctx context.Context, schema *proto.Schema) error {
	// If no event handler has been configured, then no events can be sent.
	if !HasEventHandler(ctx) {
		return nil
	}

	spanContext := trace.SpanContextFromContext(ctx)

	// If tracing is disabled, then no events can be sent.
	// This is because events are produced from the auditing table.
	if !spanContext.IsValid() {
		return nil
	}

	handler, err := GetEventHandler(ctx)
	if err != nil {
		return err
	}

	traceparent := util.GetTraceparent(spanContext)

	identityId := ""
	if auth.IsAuthenticated(ctx) {
		identity, err := auth.GetIdentity(ctx)
		if err != nil {
			return err
		}

		identityId = identity.Id
	}

	database, err := db.GetDatabase(ctx)
	if err != nil {
		return err
	}

	sql2 := fmt.Sprintf(
		`SELECT * FROM keel_audit`)
	result2, err := database.ExecuteQuery(ctx, sql2)
	fmt.Println(result2)
	sql, err := processAuditLogsSql(schema, spanContext.TraceID().String())
	if err != nil {
		return err
	}

	result, err := database.ExecuteQuery(ctx, sql)
	if err != nil {
		return err
	}

	for _, row := range result.Rows {
		audit := typed.New(row)

		tableName := audit.String("table_name")
		if tableName == "" {
			return errors.New("audit 'table' column cannot be parsed or is empty")
		}

		op := audit.String("op")
		if op == "" {
			return errors.New("audit 'op' column cannot be parsed or is empty")
		}

		eventName, err := eventNameFromAudit(tableName, op)
		if err != nil {
			return err
		}

		protoEvent := proto.FindEvent(schema.Events, eventName)
		if protoEvent == nil {
			continue
		}

		subscribers := proto.FindEventSubscriptions(schema, protoEvent)

		for _, subscriber := range subscribers {
			data, err := typed.JsonString(audit.String("data"))
			if err != nil {
				return err
			}

			id := data.String("id")
			if id == "" {
				return errors.New("error parsing audit table data")
			}

			event := &Event{
				EventName:  eventName,
				OccurredAt: time.Now().UTC(),
				IdentityId: identityId,
				Target: &EventTarget{
					Id:   id,
					Type: strcase.ToCamel(tableName),
					Data: toLowerCamelMap(data),
				},
			}

			err = handler(ctx, subscriber.Name, event, traceparent)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// processAuditLogsSql generates SQL which updates and returns the relevant audit log
// entries which are to be turned into events.
func processAuditLogsSql(schema *proto.Schema, traceId string) (string, error) {
	if traceId == "" {
		return "", errors.New("traceId cannot be empty")
	}

	if len(schema.Events) == 0 {
		return "", errors.New("there are no events defined in this schema")
	}

	conditions := []string{}
	for _, e := range schema.Events {
		table := strcase.ToSnake(e.ModelName)
		op, err := auditOpFromAction(e.ActionType)
		if err != nil {
			return "", err
		}

		conditions = append(conditions, fmt.Sprintf("(table_name = '%s' AND op = '%s')", table, op))
	}

	filter := strings.Join(conditions, " OR ")
	if len(conditions) > 1 {
		filter = fmt.Sprintf("(%s)", filter)
	}

	sql := fmt.Sprintf(
		`UPDATE keel_audit 
		SET event_processed_at = NOW()
		WHERE
			trace_id = '%s' AND 
			event_processed_at IS NULL AND
			%s
		RETURNING *`, traceId, filter)

	return sql, nil
}

// eventNameFromAudit generates an event name from audit table columns.
func eventNameFromAudit(tableName string, op string) (string, error) {
	action := ""

	switch op {
	case "insert":
		action = Created
	case "update":
		action = Updated
	case "delete":
		action = Deleted
	default:
		return "", fmt.Errorf("unknown op type '%s' when creating event", op)
	}

	return fmt.Sprintf("%s.%s", strcase.ToSnake(tableName), action), nil
}

// auditOpFromAction gets the audit operation for a specific action type.
func auditOpFromAction(action proto.ActionType) (string, error) {
	switch action {
	case proto.ActionType_ACTION_TYPE_CREATE:
		return "insert", nil
	case proto.ActionType_ACTION_TYPE_UPDATE:
		return "update", nil
	case proto.ActionType_ACTION_TYPE_DELETE:
		return "delete", nil
	default:
		return "", fmt.Errorf("unsupported action type '%s' when creating event", action)
	}
}

func toLowerCamelMap(m map[string]any) map[string]any {
	res := map[string]any{}
	for key, value := range m {
		res[casing.ToLowerCamel(key)] = value
	}
	return res
}
