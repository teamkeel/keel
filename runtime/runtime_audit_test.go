package runtime_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/karlseguin/typed"
	"github.com/nsf/jsondiff"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/testhelpers"
	"go.opentelemetry.io/otel/trace"
)

const (
	traceId = "71f835dc7ac2750bed2135c7b30dc7fe"
	spanId  = "b4c9e2a6a0d84702"
)

func newContext(t *testing.T) (context.Context, db.Database, *proto.Schema) {
	dbConnInfo := &db.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Password: "postgres",
		Database: "keel",
	}

	var keelSchema = `
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

		@permission(expression: true, actions: [create, update, delete])
	}
	model WeddingInvitee {
		fields {
			wedding Wedding
			firstName Text
		}
	}
	`

	schema := protoSchema(t, keelSchema)

	ctx := context.Background()

	// Add private key to context
	pk, err := testhelpers.GetEmbeddedPrivateKey()
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, pk)

	// Add database to context
	database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, "audit_test")
	require.NoError(t, err)
	ctx = db.WithDatabase(ctx, database)

	return ctx, database, schema
}

func withIdentity(t *testing.T, ctx context.Context, schema *proto.Schema) (context.Context, *auth.Identity) {
	identity, err := actions.CreateIdentity(ctx, schema, "dave.new@keel.xyz", "1234")
	require.NoError(t, err)
	return auth.WithIdentity(ctx, identity), identity
}

func withTracing(t *testing.T, ctx context.Context) context.Context {
	traceIdBytes, err := hex.DecodeString(traceId)
	require.NoError(t, err)
	spanIdBytes, err := hex.DecodeString(spanId)
	require.NoError(t, err)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID(traceIdBytes),
		SpanID:     trace.SpanID(spanIdBytes),
		TraceFlags: trace.FlagsSampled,
	})
	require.True(t, spanContext.IsValid())
	return trace.ContextWithSpanContext(ctx, spanContext)
}

func TestAuditCreateAction(t *testing.T) {
	ctx, database, schema := newContext(t)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)
	ctx = withTracing(t, ctx)

	_, err := actions.Create(
		actions.NewScope(ctx, proto.FindAction(schema, "createWedding"), schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	var weddings []map[string]any
	db.Raw("SELECT * FROM wedding").Scan(&weddings)
	require.Len(t, weddings, 1)
	wedding := weddings[0]

	var audits []map[string]any
	db.Raw("SELECT * FROM keel_audit WHERE table_name='wedding'").Scan(&audits)
	require.Len(t, audits, 1)
	audit := audits[0]

	// This is due to https://linear.app/keel/issue/BLD-824/storing-dates-in-utc-and-not-with-timezone
	wedding["created_at"] = wedding["created_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")
	wedding["updated_at"] = wedding["updated_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")

	expectedData, err := json.Marshal(wedding)
	require.NoError(t, err)

	require.Equal(t, "wedding", audit["table_name"])
	require.Equal(t, "insert", audit["op"])
	require.NotNil(t, audit["id"])
	require.NotNil(t, audit["created_at"])
	require.Equal(t, identity.Id, audit["identity_id"])
	require.Equal(t, traceId, audit["trace_id"])

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expectedData, []byte(audit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}
}

func TestAuditNestedCreateAction(t *testing.T) {
	ctx, database, schema := newContext(t)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)
	ctx = withTracing(t, ctx)

	_, err := actions.Create(
		actions.NewScope(ctx, proto.FindAction(schema, "createWeddingWithGuests"), schema),
		map[string]any{
			"name": "Dave",
			"guests": []any{
				map[string]any{"firstName": "Pete"},
				map[string]any{"firstName": "Adam"},
			},
		})
	require.NoError(t, err)

	var weddings []map[string]any
	db.Raw("SELECT * FROM wedding").Scan(&weddings)
	require.Len(t, weddings, 1)
	wedding := weddings[0]

	var weddingAudits []map[string]any
	db.Raw("SELECT * FROM keel_audit WHERE table_name='wedding'").Scan(&weddingAudits)
	require.Len(t, weddingAudits, 1)
	weddingAudit := weddingAudits[0]

	// This is due to https://linear.app/keel/issue/BLD-824/storing-dates-in-utc-and-not-with-timezone
	wedding["created_at"] = wedding["created_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")
	wedding["updated_at"] = wedding["updated_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")

	expectedData, err := json.Marshal(wedding)
	require.NoError(t, err)

	require.Equal(t, "wedding", weddingAudit["table_name"])
	require.Equal(t, "insert", weddingAudit["op"])
	require.NotNil(t, weddingAudit["id"])
	require.NotNil(t, weddingAudit["created_at"])
	require.Equal(t, identity.Id, weddingAudit["identity_id"])
	require.Equal(t, traceId, weddingAudit["trace_id"])

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expectedData, []byte(weddingAudit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}

	var invitees []map[string]any
	db.Raw("SELECT * FROM wedding_invitee").Scan(&invitees)
	require.Len(t, invitees, 2)
	pete, found := lo.Find(invitees, func(i map[string]any) bool { return i["first_name"] == "Pete" })
	require.True(t, found)
	adam, found := lo.Find(invitees, func(i map[string]any) bool { return i["first_name"] == "Adam" })
	require.True(t, found)

	// This is due to https://linear.app/keel/issue/BLD-824/storing-dates-in-utc-and-not-with-timezone
	pete["created_at"] = pete["created_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")
	pete["updated_at"] = pete["updated_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")

	// This is due to https://linear.app/keel/issue/BLD-824/storing-dates-in-utc-and-not-with-timezone
	adam["created_at"] = adam["created_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")
	adam["updated_at"] = adam["updated_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")

	var peteAudits []map[string]any
	db.Raw(fmt.Sprintf("SELECT * FROM keel_audit WHERE table_name='wedding_invitee' AND data ->> 'id' = '%s'", pete["id"])).Scan(&peteAudits)
	require.Len(t, peteAudits, 1)
	peteAudit := peteAudits[0]

	expectedPeteData, err := json.Marshal(pete)
	require.NoError(t, err)

	require.Equal(t, "wedding_invitee", peteAudit["table_name"])
	require.Equal(t, "insert", peteAudit["op"])
	require.NotNil(t, peteAudit["id"])
	require.NotNil(t, peteAudit["created_at"])
	require.Equal(t, identity.Id, peteAudit["identity_id"])
	require.Equal(t, traceId, peteAudit["trace_id"])

	diff, explanation = jsondiff.Compare(expectedPeteData, []byte(peteAudit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}

	var adamAudits []map[string]any
	db.Raw(fmt.Sprintf("SELECT * FROM keel_audit WHERE table_name='wedding_invitee' AND data ->> 'id' = '%s'", adam["id"])).Scan(&adamAudits)
	require.Len(t, adamAudits, 1)
	adamAudit := adamAudits[0]

	expectedAdamData, err := json.Marshal(adam)
	require.NoError(t, err)

	require.Equal(t, "wedding_invitee", adamAudit["table_name"])
	require.Equal(t, "insert", adamAudit["op"])
	require.NotNil(t, adamAudit["id"])
	require.NotNil(t, peteAudit["created_at"])
	require.Equal(t, identity.Id, adamAudit["identity_id"])
	require.Equal(t, traceId, adamAudit["trace_id"])

	diff, explanation = jsondiff.Compare(expectedAdamData, []byte(adamAudit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}
}

func TestAuditUpdateAction(t *testing.T) {
	ctx, database, schema := newContext(t)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)
	ctx = withTracing(t, ctx)

	create := proto.FindAction(schema, "createWedding")
	createResult, _, err := actions.Execute(
		actions.NewScope(ctx, create, schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	update := proto.FindAction(schema, "updateWedding")
	_, _, err = actions.Execute(
		actions.NewScope(ctx, update, schema),
		map[string]any{
			"where":  map[string]any{"id": createResult.(map[string]any)["id"]},
			"values": map[string]any{"name": "Adam"},
		})
	require.NoError(t, err)

	var weddings []map[string]any
	db.Raw("SELECT * FROM wedding").Scan(&weddings)
	require.Len(t, weddings, 1)
	wedding := weddings[0]

	// This is due to https://linear.app/keel/issue/BLD-824/storing-dates-in-utc-and-not-with-timezone
	wedding["created_at"] = wedding["created_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")
	wedding["updated_at"] = wedding["updated_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")

	var audits []map[string]any
	db.Raw("SELECT * FROM keel_audit WHERE op='insert' and table_name='wedding'").Scan(&audits)
	require.Len(t, audits, 1)

	audits = nil
	db.Raw("SELECT * FROM keel_audit WHERE op='update' and table_name='wedding'").Scan(&audits)
	require.Len(t, audits, 1)
	audit := audits[0]

	expectedData, err := json.Marshal(wedding)
	require.NoError(t, err)

	require.Equal(t, "wedding", audit["table_name"])
	require.Equal(t, "update", audit["op"])
	require.NotNil(t, audit["id"])
	require.NotNil(t, audit["created_at"])
	require.Equal(t, identity.Id, audit["identity_id"])
	require.Equal(t, traceId, audit["trace_id"])

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expectedData, []byte(audit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}
}

func TestAuditDeleteAction(t *testing.T) {
	ctx, database, schema := newContext(t)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)
	ctx = withTracing(t, ctx)

	create := proto.FindAction(schema, "createWedding")
	createResult, _, err := actions.Execute(
		actions.NewScope(ctx, create, schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	var weddings []map[string]any
	db.Raw("SELECT * FROM wedding").Scan(&weddings)
	require.Len(t, weddings, 1)
	wedding := weddings[0]

	// This is due to https://linear.app/keel/issue/BLD-824/storing-dates-in-utc-and-not-with-timezone
	wedding["created_at"] = wedding["created_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")
	wedding["updated_at"] = wedding["updated_at"].(time.Time).UTC().Format("2006-01-02T15:04:05.999999999-07:00")

	delete := proto.FindAction(schema, "deleteWedding")
	_, _, err = actions.Execute(
		actions.NewScope(ctx, delete, schema),
		map[string]any{"id": createResult.(map[string]any)["id"]},
	)
	require.NoError(t, err)

	var audits []map[string]any
	db.Raw("SELECT * FROM keel_audit WHERE op='insert' and table_name='wedding'").Scan(&audits)
	require.Len(t, audits, 1)

	audits = nil
	db.Raw("SELECT * FROM keel_audit WHERE op='delete' and table_name='wedding'").Scan(&audits)
	require.Len(t, audits, 1)
	audit := audits[0]

	expectedData, err := json.Marshal(wedding)
	require.NoError(t, err)

	require.Equal(t, "wedding", audit["table_name"])
	require.Equal(t, "delete", audit["op"])
	require.NotNil(t, audit["id"])
	require.NotNil(t, audit["created_at"])
	require.Equal(t, identity.Id, audit["identity_id"])
	require.Equal(t, traceId, audit["trace_id"])

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expectedData, []byte(audit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}
}

func TestAuditTablesWithOnlyIdentity(t *testing.T) {
	ctx, database, schema := newContext(t)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)

	action := proto.FindAction(schema, "createWedding")
	input := map[string]any{"name": "Dave"}
	scope := actions.NewScope(ctx, action, schema)
	_, err := actions.Create(scope, input)
	require.NoError(t, err)

	var audits []map[string]any
	db.Raw("SELECT * FROM keel_audit WHERE table_name='wedding'").Scan(&audits)
	require.Len(t, audits, 1)
	audit := audits[0]

	require.Equal(t, "wedding", audit["table_name"])
	require.Equal(t, "insert", audit["op"])
	require.NotNil(t, audit["id"])
	require.NotNil(t, audit["created_at"])
	require.Equal(t, identity.Id, audit["identity_id"])
	require.Nil(t, audit["trace_id"])
}

func TestAuditTablesWithOnlyTracing(t *testing.T) {
	ctx, database, schema := newContext(t)
	defer database.Close()
	db := database.GetDB()

	ctx = withTracing(t, ctx)

	action := proto.FindAction(schema, "createWedding")
	input := map[string]any{"name": "Dave"}
	scope := actions.NewScope(ctx, action, schema)
	_, err := actions.Create(scope, input)
	require.NoError(t, err)

	var audits []map[string]any
	db.Raw("SELECT * FROM keel_audit WHERE table_name='wedding'").Scan(&audits)
	require.Len(t, audits, 1)
	audit := audits[0]

	require.Equal(t, "wedding", audit["table_name"])
	require.Equal(t, "insert", audit["op"])
	require.NotNil(t, audit["id"])
	require.NotNil(t, audit["created_at"])
	require.Nil(t, audit["identity_id"])
	require.Equal(t, traceId, audit["trace_id"])
}

func TestAuditOnStatementExecuteWithoutResult(t *testing.T) {
	ctx, database, schema := newContext(t)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)
	ctx = withTracing(t, ctx)

	result, err := actions.Create(
		actions.NewScope(ctx, proto.FindAction(schema, "createWedding"), schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	action := proto.FindAction(schema, "createWedding")

	scope := actions.NewScope(ctx, action, schema)
	query := actions.NewQuery(scope.Context, scope.Model)
	err = query.Where(actions.IdField(), actions.Equals, actions.Value(result["id"]))
	require.NoError(t, err)
	query.AddWriteValue(actions.Field("name"), "Devin")
	affected, err := query.UpdateStatement().Execute(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, affected)

	var audits []map[string]any
	db.Raw("SELECT * FROM keel_audit WHERE table_name='wedding' and op='update'").Scan(&audits)
	require.Len(t, audits, 1)
	audit := audits[0]

	require.Equal(t, "wedding", audit["table_name"])
	require.Equal(t, "update", audit["op"])
	require.NotNil(t, audit["id"])
	require.NotNil(t, audit["created_at"])
	require.Equal(t, identity.Id, audit["identity_id"])
	require.Equal(t, traceId, audit["trace_id"])

	data, err := typed.JsonString(audit["data"].(string))
	require.NoError(t, err)
	require.Equal(t, data["name"], "Devin")
}

func TestAuditFieldsAreDroppedOnCreate(t *testing.T) {
	ctx, database, schema := newContext(t)
	defer database.Close()

	ctx, _ = withIdentity(t, ctx, schema)
	ctx = withTracing(t, ctx)

	result, err := actions.Create(
		actions.NewScope(ctx, proto.FindAction(schema, "createWedding"), schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	require.Nil(t, result["keelIdentityId"])
	require.Nil(t, result["keelTraceId"])
	require.Equal(t, 4, len(result))
}
