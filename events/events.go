package events

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/auditing"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel"
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
	// The model data at the time of the event.
	Data map[string]any `json:"data"`
	// The previous state of the model data before the event.
	PreviousData map[string]any `json:"previousData"`
}

// The event handler function to be executed for each subscriber event generated.
type EventHandler func(ctx context.Context, subscriber string, event *Event, traceparent string) error

type handlerContextKey string

var contextKey handlerContextKey = "eventHandler"

func WithEventHandler(ctx context.Context, handler EventHandler) (context.Context, error) {
	// If no tracing provider is set up, then events will not work.
	// It is better to error than to let events silently malfunction.
	if otel.GetTracerProvider() == trace.NewNoopTracerProvider() {
		return nil, errors.New("cannot use events when there is no trace provider configured")
	}

	return context.WithValue(ctx, contextKey, handler), nil
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

	// If there are no events defined in the schema, then don't bother processing events.
	if len(schema.Events) == 0 {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	// If there is no valid trace, then no events can be sent.  This is because
	// events are produced from the auditing table by comparing the trace ids.
	// However, there should always be a valid trace at this point.
	if !spanContext.IsValid() {
		return errors.New("valid spanContext expected")
	}

	handler, err := GetEventHandler(ctx)
	if err != nil {
		return err
	}

	traceparent := util.GetTraceparent(spanContext)
	traceId := spanContext.TraceID().String()

	identityId := ""
	if auth.IsAuthenticated(ctx) {
		identity, err := auth.GetIdentity(ctx)
		if err != nil {
			return err
		}

		identityId = identity.Id
	}

	auditLogs, err := auditing.ProcessEventsFromAuditTrail(ctx, schema, traceId)
	if err != nil {
		return err
	}

	var handlerErrors error
	for _, log := range auditLogs {
		eventName, err := eventNameFromAudit(log.TableName, log.Op)
		if err != nil {
			return err
		}

		protoEvent := proto.FindEvent(schema.Events, eventName)
		if protoEvent == nil {
			return fmt.Errorf("event '%s' does not exist", eventName)
		}

		subscribers := proto.FindEventSubscriptions(schema, protoEvent)
		if len(subscribers) == 0 {
			return fmt.Errorf("event '%s' must have at least one subscriber", eventName)
		}

		var previous map[string]any
		if log.Op != Created {
			p, err := auditing.Previous(ctx, log)
			if err != nil {
				return err
			}

			if p != nil {
				previous = p.Data
			}
		}

		for _, subscriber := range subscribers {
			event := &Event{
				EventName:  eventName,
				OccurredAt: time.Now().UTC(),
				IdentityId: identityId,
				Target: &EventTarget{
					Id:           log.Data["id"].(string),
					Type:         strcase.ToCamel(log.TableName),
					Data:         toLowerCamelMap(log.Data),
					PreviousData: toLowerCamelMap(previous),
				},
			}

			err = handler(ctx, subscriber.Name, event, traceparent)
			if err != nil {
				// We do not error yet when the event handler fails
				handlerErrors = errors.Join(handlerErrors, err)
			} else {
				// For successfully fired events
				span.AddEvent(eventName)
			}
		}
	}

	return handlerErrors
}

// eventNameFromAudit generates an event name from audit table columns.
func eventNameFromAudit(tableName string, op string) (string, error) {
	action := ""

	switch op {
	case auditing.Insert:
		action = Created
	case auditing.Update:
		action = Updated
	case auditing.Delete:
		action = Deleted
	default:
		return "", fmt.Errorf("unknown op type '%s' when creating event", op)
	}

	return fmt.Sprintf("%s.%s", strcase.ToSnake(tableName), action), nil
}

func toLowerCamelMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}

	res := map[string]any{}
	for key, value := range m {
		res[casing.ToLowerCamel(key)] = value
	}
	return res
}
