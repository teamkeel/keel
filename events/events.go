package events

import (
	"context"
	"fmt"
	"time"

	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel/trace"
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
func SendEvents(ctx context.Context) error {
	if !HasEventHandler(ctx) {
		return nil
	}

	handler, err := GetEventHandler(ctx)
	if err != nil {
		return err
	}

	spanContext := trace.SpanContextFromContext(ctx)
	traceparent := util.GetTraceparent(spanContext)

	identityId := ""
	if auth.IsAuthenticated(ctx) {
		identity, err := auth.GetIdentity(ctx)
		if err != nil {
			return err
		}

		identityId = identity.Id
	}

	// 1. Retrieve rows from the audit table by this ctx's trace_id
	// 2. Do we have any events in the schema matching these rows?
	// 3. If so, call handleEvent for each subscriber of that event with the payload.

	// PLACEHOLDER CODE
	testEvent := &Event{
		EventName:  "person.created",
		OccurredAt: time.Now().UTC(),
		IdentityId: identityId,
		Target: &EventTarget{
			Id:   "2342342",
			Type: "Person",
			Data: map[string]any{
				"id":    "2342342",
				"name":  "Dave",
				"email": "dave@hello.com",
			},
		},
	}

	return handler(ctx, "verifyEmail", testEvent, traceparent)
}
