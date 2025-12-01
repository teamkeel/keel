package runtime_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/karlseguin/typed"
	"github.com/nsf/jsondiff"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/oauth"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/testhelpers"
	keeltesting "github.com/teamkeel/keel/testing"
	"go.opentelemetry.io/otel/trace"
)

var auditSchema = `
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
}`

func withIdentity(t *testing.T, ctx context.Context, schema *proto.Schema) (context.Context, auth.Identity) {
	identity, err := actions.CreateIdentity(ctx, schema, "dave.new@keel.xyz", "1234", oauth.KeelIssuer)
	require.NoError(t, err)
	return auth.WithIdentity(ctx, identity), identity
}

func TestAuditCreateAction(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), auditSchema, true)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)

	_, err := actions.Create(
		actions.NewScope(ctx, schema.FindAction("createWedding"), schema),
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
	require.Equal(t, identity[parser.FieldNameId].(string), audit["identity_id"])
	require.Equal(t, testhelpers.TraceId, audit["trace_id"])
	require.Nil(t, audit["event_processed_at"])

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expectedData, []byte(audit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}
}

func TestAuditNestedCreateAction(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, t.Context(), auditSchema, true)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)

	_, err := actions.Create(
		actions.NewScope(ctx, schema.FindAction("createWeddingWithGuests"), schema),
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
	require.Equal(t, identity[parser.FieldNameId].(string), weddingAudit["identity_id"])
	require.Equal(t, testhelpers.TraceId, weddingAudit["trace_id"])
	require.Nil(t, weddingAudit["event_processed_at"])

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
	require.Equal(t, identity[parser.FieldNameId].(string), peteAudit["identity_id"])
	require.Equal(t, testhelpers.TraceId, peteAudit["trace_id"])
	require.Nil(t, peteAudit["event_processed_at"])

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
	require.Equal(t, identity[parser.FieldNameId].(string), adamAudit["identity_id"])
	require.Equal(t, testhelpers.TraceId, adamAudit["trace_id"])

	diff, explanation = jsondiff.Compare(expectedAdamData, []byte(adamAudit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}
}

func TestAuditUpdateAction(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), auditSchema, true)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)

	create := schema.FindAction("createWedding")
	createResult, _, err := actions.Execute(
		actions.NewScope(ctx, create, schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	update := schema.FindAction("updateWedding")
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
	require.Equal(t, identity[parser.FieldNameId].(string), audit["identity_id"])
	require.Equal(t, testhelpers.TraceId, audit["trace_id"])
	require.Nil(t, audit["event_processed_at"])

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expectedData, []byte(audit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}
}

func TestAuditDeleteAction(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), auditSchema, true)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)

	create := schema.FindAction("createWedding")
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

	del := schema.FindAction("deleteWedding")
	_, _, err = actions.Execute(
		actions.NewScope(ctx, del, schema),
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
	require.Equal(t, identity[parser.FieldNameId].(string), audit["identity_id"])
	require.Equal(t, testhelpers.TraceId, audit["trace_id"])
	require.Nil(t, audit["event_processed_at"])

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expectedData, []byte(audit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}
}

func TestAuditTablesWithOnlyIdentity(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), auditSchema, true)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)

	// Empty and invalid span context
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{})
	ctx = trace.ContextWithSpanContext(ctx, spanContext)

	action := schema.FindAction("createWedding")
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
	require.Equal(t, identity[parser.FieldNameId].(string), audit["identity_id"])
	require.Nil(t, audit["trace_id"])
	require.Nil(t, audit["event_processed_at"])
}

func TestAuditTablesWithOnlyTracing(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), auditSchema, true)
	defer database.Close()
	db := database.GetDB()

	action := schema.FindAction("createWedding")
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
	require.Equal(t, testhelpers.TraceId, audit["trace_id"])
	require.Nil(t, audit["event_processed_at"])
}

func TestAuditOnStatementExecuteWithoutResult(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), auditSchema, true)
	defer database.Close()
	db := database.GetDB()

	ctx, identity := withIdentity(t, ctx, schema)

	result, err := actions.Create(
		actions.NewScope(ctx, schema.FindAction("createWedding"), schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	action := schema.FindAction("createWedding")

	scope := actions.NewScope(ctx, action, schema)
	query := actions.NewQuery(scope.Model)
	err = query.Where(actions.IdField(), actions.Equals, actions.Value(result["id"]))
	require.NoError(t, err)
	query.AddWriteValue(actions.Field("name"), actions.Value("Devin"))
	affected, err := query.UpdateStatement(scope.Context).Execute(ctx)
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
	require.Equal(t, identity[parser.FieldNameId].(string), audit["identity_id"])
	require.Equal(t, testhelpers.TraceId, audit["trace_id"])
	require.Nil(t, audit["event_processed_at"])

	data, err := typed.JsonString(audit["data"].(string))
	require.NoError(t, err)
	require.Equal(t, data["name"], "Devin")
}

func TestAuditFieldsAreDroppedOnCreate(t *testing.T) {
	ctx, database, schema := keeltesting.MakeContext(t, context.TODO(), auditSchema, true)
	defer database.Close()

	ctx, _ = withIdentity(t, ctx, schema)

	result, err := actions.Create(
		actions.NewScope(ctx, schema.FindAction("createWedding"), schema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	require.Nil(t, result["keelIdentityId"])
	require.Nil(t, result["keelTraceId"])
	require.Equal(t, 4, len(result))
}

func TestAuditDatabaseMigration(t *testing.T) {
	var keelSchema = `
		model Person {
			fields {
				name Text
				age Number?
				isActive Boolean?
			}
			actions {
				create createPerson() with (name)
			}
			@permission(expression: true, actions: [create, update, delete])
		}`

	ctx, database, pSchema := keeltesting.MakeContext(t, context.TODO(), keelSchema, true)

	create := pSchema.FindAction("createPerson")
	_, _, err := actions.Execute(
		actions.NewScope(ctx, create, pSchema),
		map[string]any{"name": "Dave"})
	require.NoError(t, err)

	var updatedSchema = `
		model Person {
			fields {
				name Text
				age Number @default(0)
				isActive Boolean @default(true)
			}
			actions {
				create createPerson() with (name)
			}
			@permission(expression: true, actions: [create, update, delete])
		}`

	database.Close()
	ctx, database, pSchema = keeltesting.MakeContext(t, context.TODO(), updatedSchema, false)
	db := database.GetDB()
	defer database.Close()

	// Migrate the database to the new schema.
	m, err := migrations.New(ctx, pSchema, database)
	require.NoError(t, err)

	err = m.Apply(ctx, false)
	require.NoError(t, err)

	var audits []map[string]any
	db.Raw("SELECT * FROM keel_audit WHERE op='insert' and table_name='person'").Scan(&audits)
	require.Len(t, audits, 1)

	audits = nil
	db.Raw("SELECT * FROM keel_audit WHERE op='update' and table_name='person'").Scan(&audits)
	require.Len(t, audits, 2)
	ageUpdateAudit := audits[0]

	data, err := typed.JsonString(ageUpdateAudit["data"].(string))
	require.NoError(t, err)

	require.Equal(t, "person", ageUpdateAudit["table_name"])
	require.Equal(t, "update", ageUpdateAudit["op"])
	require.NotNil(t, ageUpdateAudit["id"])
	require.NotNil(t, ageUpdateAudit["created_at"])
	require.Nil(t, ageUpdateAudit["identity_id"])
	require.Equal(t, testhelpers.TraceId, ageUpdateAudit["trace_id"])
	require.Nil(t, ageUpdateAudit["event_processed_at"])
	require.Equal(t, 0, data.IntMust("age"))
	require.Empty(t, data["is_active"])

	isActiveUpdateAudit := audits[1]

	data, err = typed.JsonString(isActiveUpdateAudit["data"].(string))
	require.NoError(t, err)

	require.Equal(t, "person", ageUpdateAudit["table_name"])
	require.Equal(t, "update", ageUpdateAudit["op"])
	require.NotNil(t, ageUpdateAudit["id"])
	require.NotNil(t, ageUpdateAudit["created_at"])
	require.Nil(t, ageUpdateAudit["identity_id"])
	require.Equal(t, testhelpers.TraceId, ageUpdateAudit["trace_id"])
	require.Nil(t, ageUpdateAudit["event_processed_at"])
	require.Equal(t, 0, data.IntMust("age"))
	require.Equal(t, true, data.BoolMust("is_active"))
}
