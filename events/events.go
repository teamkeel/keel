package events

import (
	"context"
	"errors"
	"fmt"
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

// Gather, create and send events which have occurred within the scope of this context.
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

	sql := fmt.Sprintf(
		`UPDATE keel_audit SET event_created_at = NOW()
		WHERE
			trace_id = '%s' AND 
			event_created_at IS NULL AND
			op IN ('insert', 'update', 'delete') RETURNING *`, spanContext.TraceID().String())

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

		eventName, err := toEventName(tableName, op)
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

func toEventName(tableName string, op string) (string, error) {
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

	return fmt.Sprintf("%s.%s", strcase.ToLowerCamel(tableName), action), nil
}

func toLowerCamelMap(m map[string]any) map[string]any {
	res := map[string]any{}
	for key, value := range m {
		res[casing.ToLowerCamel(key)] = value
	}
	return res
}
