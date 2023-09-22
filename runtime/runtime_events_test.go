package runtime_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/karlseguin/typed"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
)

var eventsSchema = `
model Wedding {
	fields {
		name Text
		guests WeddingInvitee[]
	}
	actions {
		create createWedding() with (name) 
		create createWeddingWithGuests() with (name, guests.firstName) 
		update updateWedding(id) with (name) 
		delete deleteWedding(id)
	}

	@on([create], sendInvites)
	@on([update], sendUpdates)
	@on([delete], sendCancellations)

	@permission(expression: true, actions: [create, update, delete])
}
model WeddingInvitee  {
	fields {
		firstName Text
		wedding Wedding?
	}

	actions {
		create createInvitee() with (firstName) 
		update updateInvitee(id) with (firstName) 
	}

	@on([create], sendInvites)
	@on([create, update], verifyDetails)

	@permission(expression: true, actions: [create, update, delete])
}
model Person {
	actions {
		create createPerson()
	}
	@on([update], verifyDetails)
	@permission(expression: true, actions: [create])
}
`

func NewEventHandler() EventHandler {
	return EventHandler{
		subscribedEvents: map[string][]*events.Event{},
	}
}

type EventHandler struct {
	subscribedEvents map[string][]*events.Event
}

func (handler *EventHandler) HandleEvent(ctx context.Context, subscriber string, event *events.Event, traceparent string) error {
	handler.subscribedEvents[subscriber] = append(handler.subscribedEvents[subscriber], event)

	if subscriber == "" || event == nil {
		return errors.New("invalid params for event handler")
	}

	return nil
}

func TestCreateEvent(t *testing.T) {
	ctx, database, schema := newContext(t, eventsSchema)
	defer database.Close()

	ctx, identity := withIdentity(t, ctx, schema)
	ctx = withTracing(t, ctx)

	handler := NewEventHandler()
	ctx = events.WithEventHandler(ctx, handler.HandleEvent)

	result, _, err := actions.Execute(
		actions.NewScope(ctx, proto.FindAction(schema, "createWedding"), schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	wedding, ok := result.(map[string]any)
	require.True(t, ok)

	require.Len(t, handler.subscribedEvents, 1)

	events, ok := handler.subscribedEvents["sendInvites"]
	require.True(t, ok)
	require.Len(t, events, 1)

	require.Equal(t, "wedding.created", events[0].EventName)
	require.Equal(t, identity.Id, events[0].IdentityId)
	require.NotEmpty(t, events[0].OccurredAt)
	require.NotNil(t, events[0].Target)
	require.Equal(t, wedding["id"], events[0].Target.Id)
	require.Equal(t, "Wedding", events[0].Target.Type)

	data := typed.New(events[0].Target.Data)
	require.Equal(t, wedding["id"], data.String("id"))
	require.Equal(t, wedding["name"], data.String("name"))

	createdAt, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", data.String("createdAt"))
	require.NoError(t, err)
	require.Equal(t, wedding["createdAt"], createdAt)

	updatedAt, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", data.String("updatedAt"))
	require.NoError(t, err)
	require.Equal(t, wedding["updatedAt"], updatedAt)
}

func TestUpdateEvent(t *testing.T) {
	ctx, database, schema := newContext(t, eventsSchema)
	defer database.Close()

	result, _, err := actions.Execute(
		actions.NewScope(ctx, proto.FindAction(schema, "createWedding"), schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	wedding, ok := result.(map[string]any)
	require.True(t, ok)

	ctx, identity := withIdentity(t, ctx, schema)
	ctx = withTracing(t, ctx)

	handler := NewEventHandler()
	ctx = events.WithEventHandler(ctx, handler.HandleEvent)

	updated, _, err := actions.Execute(
		actions.NewScope(ctx, proto.FindAction(schema, "updateWedding"), schema),
		map[string]any{
			"where":  map[string]any{"id": wedding["id"]},
			"values": map[string]any{"name": "Adam"},
		})
	require.NoError(t, err)

	updatedWedding, ok := updated.(map[string]any)
	require.True(t, ok)
	require.Equal(t, wedding["id"], updatedWedding["id"])

	require.Len(t, handler.subscribedEvents, 1)

	events, ok := handler.subscribedEvents["sendUpdates"]
	require.True(t, ok)
	require.Len(t, events, 1)

	require.NotEmpty(t, events[0])
	require.Equal(t, "wedding.updated", events[0].EventName)
	require.Equal(t, identity.Id, events[0].IdentityId)
	require.NotEmpty(t, events[0].OccurredAt)
	require.NotNil(t, events[0].Target)
	require.Equal(t, wedding["id"], events[0].Target.Id)
	require.Equal(t, "Wedding", events[0].Target.Type)

	data := typed.New(events[0].Target.Data)
	require.Equal(t, wedding["id"], data.String("id"))
	require.Equal(t, updatedWedding["name"], data.String("name"))

	createdAt, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", data.String("createdAt"))
	require.NoError(t, err)
	require.Equal(t, wedding["createdAt"], createdAt)

	updatedAt, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", data.String("updatedAt"))
	require.NoError(t, err)
	require.Equal(t, wedding["updatedAt"], updatedAt)
}

func TestDeleteEvent(t *testing.T) {
	ctx, database, schema := newContext(t, eventsSchema)
	defer database.Close()

	result, _, err := actions.Execute(
		actions.NewScope(ctx, proto.FindAction(schema, "createWedding"), schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	wedding, ok := result.(map[string]any)
	require.True(t, ok)

	ctx, identity := withIdentity(t, ctx, schema)
	ctx = withTracing(t, ctx)

	handler := NewEventHandler()
	ctx = events.WithEventHandler(ctx, handler.HandleEvent)

	_, _, err = actions.Execute(
		actions.NewScope(ctx, proto.FindAction(schema, "deleteWedding"), schema),
		map[string]any{"id": wedding["id"]})
	require.NoError(t, err)

	require.Len(t, handler.subscribedEvents, 1)

	events, ok := handler.subscribedEvents["sendCancellations"]
	require.True(t, ok)
	require.Len(t, events, 1)

	require.Equal(t, "wedding.deleted", events[0].EventName)
	require.Equal(t, identity.Id, events[0].IdentityId)
	require.NotEmpty(t, events[0].OccurredAt)
	require.NotNil(t, events[0].Target)
	require.Equal(t, wedding["id"], events[0].Target.Id)
	require.Equal(t, "Wedding", events[0].Target.Type)

	data := typed.New(events[0].Target.Data)
	require.Equal(t, wedding["id"], data.String("id"))
	require.Equal(t, wedding["name"], data.String("name"))

	createdAt, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", data.String("createdAt"))
	require.NoError(t, err)
	require.Equal(t, wedding["createdAt"], createdAt)

	updatedAt, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", data.String("updatedAt"))
	require.NoError(t, err)
	require.Equal(t, wedding["updatedAt"], updatedAt)
}

func TestNoIdentityEvent(t *testing.T) {
	ctx, database, schema := newContext(t, eventsSchema)
	defer database.Close()

	ctx = withTracing(t, ctx)

	handler := NewEventHandler()
	ctx = events.WithEventHandler(ctx, handler.HandleEvent)

	_, _, err := actions.Execute(
		actions.NewScope(ctx, proto.FindAction(schema, "createWedding"), schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	require.Len(t, handler.subscribedEvents, 1)

	events, ok := handler.subscribedEvents["sendInvites"]
	require.True(t, ok)
	require.Len(t, events, 1)
	require.Empty(t, events[0].IdentityId)
}

func TestNestedCreateEvent(t *testing.T) {
	ctx, database, schema := newContext(t, eventsSchema)
	defer database.Close()

	ctx, _ = withIdentity(t, ctx, schema)
	ctx = withTracing(t, ctx)

	handler := NewEventHandler()
	ctx = events.WithEventHandler(ctx, handler.HandleEvent)

	result, _, err := actions.Execute(
		actions.NewScope(ctx, proto.FindAction(schema, "createWeddingWithGuests"), schema),
		map[string]any{
			"name": "Dave",
			"guests": []any{
				map[string]any{"firstName": "Pete"},
				map[string]any{"firstName": "Adam"},
			},
		})
	require.NoError(t, err)

	_, ok := result.(map[string]any)
	require.True(t, ok)

	require.Len(t, handler.subscribedEvents, 2)

	sendInvitesEvent, ok := handler.subscribedEvents["sendInvites"]
	require.True(t, ok)
	require.Len(t, sendInvitesEvent, 3)

	require.Equal(t, "wedding.created", sendInvitesEvent[0].EventName)
	require.Equal(t, "weddingInvitee.created", sendInvitesEvent[1].EventName)
	require.Equal(t, "weddingInvitee.created", sendInvitesEvent[2].EventName)

	verifyDetailsEvent, ok := handler.subscribedEvents["verifyDetails"]
	require.True(t, ok)
	require.Len(t, verifyDetailsEvent, 2)

	require.Equal(t, "weddingInvitee.created", verifyDetailsEvent[0].EventName)
	require.Equal(t, "weddingInvitee.created", verifyDetailsEvent[1].EventName)
}

func TestMultipleEvents(t *testing.T) {
	ctx, database, schema := newContext(t, eventsSchema)
	defer database.Close()

	ctx, _ = withIdentity(t, ctx, schema)
	ctx = withTracing(t, ctx)

	result, _, err := actions.Execute(
		actions.NewScope(ctx, proto.FindAction(schema, "createInvitee"), schema),
		map[string]any{"firstName": "Dave"})
	require.NoError(t, err)

	wedding, ok := result.(map[string]any)
	require.True(t, ok)

	handler := NewEventHandler()
	ctx = events.WithEventHandler(ctx, handler.HandleEvent)

	updated, _, err := actions.Execute(
		actions.NewScope(ctx, proto.FindAction(schema, "updateInvitee"), schema),
		map[string]any{
			"where":  map[string]any{"id": wedding["id"]},
			"values": map[string]any{"firstName": "Adam"},
		})
	require.NoError(t, err)

	updatedWedding, ok := updated.(map[string]any)
	require.True(t, ok)
	require.Equal(t, wedding["id"], updatedWedding["id"])

	require.Len(t, handler.subscribedEvents, 2)

	sendInvitesEvent, ok := handler.subscribedEvents["sendInvites"]
	require.True(t, ok)
	require.Len(t, sendInvitesEvent, 1)

	require.Equal(t, "weddingInvitee.created", sendInvitesEvent[0].EventName)

	verifyDetailsEvent, ok := handler.subscribedEvents["verifyDetails"]
	require.True(t, ok)
	require.Len(t, verifyDetailsEvent, 2)

	require.Equal(t, "weddingInvitee.created", verifyDetailsEvent[0].EventName)
	require.Equal(t, "weddingInvitee.updated", verifyDetailsEvent[1].EventName)
}

func TestAuditTableEventCreatedAtUpdated(t *testing.T) {
	ctx, database, schema := newContext(t, eventsSchema)
	defer database.Close()

	ctx = withTracing(t, ctx)

	handler := NewEventHandler()
	ctx = events.WithEventHandler(ctx, handler.HandleEvent)

	result, _, err := actions.Execute(
		actions.NewScope(ctx, proto.FindAction(schema, "createWedding"), schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)
	_, ok := result.(map[string]any)
	require.True(t, ok)

	result2, _, err := actions.Execute(
		actions.NewScope(ctx, proto.FindAction(schema, "createPerson"), schema),
		map[string]any{})
	require.NoError(t, err)
	_, ok2 := result2.(map[string]any)
	require.True(t, ok2)

	var audits []map[string]any
	database.GetDB().Raw("SELECT * FROM keel_audit").Scan(&audits)
	require.Len(t, audits, 2)

	auditWedding := typed.New(audits[0])
	eventCreatedAt, isDate := auditWedding.TimeIf("event_processed_at")
	require.NotEmpty(t, eventCreatedAt)
	require.True(t, isDate)
	require.GreaterOrEqual(t, time.Now().UTC(), eventCreatedAt)

	eventCreatedAtPerson, ok := audits[1]["event_processed_at"]
	require.Nil(t, eventCreatedAtPerson)
	require.True(t, ok)

	require.Len(t, handler.subscribedEvents, 1)
}
