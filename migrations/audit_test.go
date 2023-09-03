package migrations_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/testhelpers"
)

var keelSchema = `
model Wedding {
	fields {
		name Text
		guests WeddingInvitee[]
	}
	actions {
		create createWedding() with (name) {
			@permission(expression: true)
		}
	}
}
model WeddingInvitee {
	fields {
		wedding Wedding
		name Text
		acceptedInvite Boolean
	}
}
`

func TestAuditTables(t *testing.T) {
	dbConnInfo := &db.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Password: "postgres",
		Database: "keel",
	}

	schema := protoSchema(t, keelSchema)
	ctx := context.Background()

	database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, "audit_test")
	require.NoError(t, err)
	defer database.Close()

	pk, err := testhelpers.GetEmbeddedPrivateKey()
	require.NoError(t, err)

	ctx = runtimectx.WithPrivateKey(ctx, pk)
	ctx = db.WithDatabase(ctx, database)

	action := proto.FindAction(schema, "createWedding")
	scope := actions.NewScope(ctx, action, schema)
	input := map[string]any{"name": "Dave"}

	_, err = actions.Create(scope, input)
	require.NoError(t, err)

	db := database.GetDB()

	var weddings []map[string]any
	db.Raw("SELECT * FROM wedding").Scan(&weddings)
	wedding := weddings[0]

	var audits []map[string]any
	db.Raw("SELECT * FROM keel_audit").Scan(&audits)
	audit := audits[0]

	expectedData, err := json.Marshal(wedding)
	require.NoError(t, err)

	require.Equal(t, "wedding", audit["table_name"])
	require.Equal(t, "insert", audit["op"])
	require.NotNil(t, audit["id"])
	require.NotNil(t, audit["created_at"])
	require.Nil(t, audit["identity_id"])
	require.Nil(t, audit["trace_id"])

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expectedData, []byte(audit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}
}

func TestAuditTablesWithIdentity(t *testing.T) {
	dbConnInfo := &db.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Password: "postgres",
		Database: "keel",
	}

	schema := protoSchema(t, keelSchema)
	ctx := context.Background()

	database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, "audit_test")
	require.NoError(t, err)
	defer database.Close()

	pk, err := testhelpers.GetEmbeddedPrivateKey()
	require.NoError(t, err)

	ctx = runtimectx.WithPrivateKey(ctx, pk)
	ctx = db.WithDatabase(ctx, database)

	identity, err := actions.CreateIdentity(ctx, schema, "dave.new@keel.xyz", "1234")
	require.NoError(t, err)

	ctx = auth.WithIdentity(ctx, identity)

	action := proto.FindAction(schema, "createWedding")
	input := map[string]any{"name": "Dave"}
	scope := actions.NewScope(ctx, action, schema)
	_, err = actions.Create(scope, input)
	require.NoError(t, err)

	db := database.GetDB()

	var weddings []map[string]any
	db.Raw("SELECT * FROM wedding").Scan(&weddings)
	wedding := weddings[0]

	var audits []map[string]any
	db.Raw("SELECT * FROM keel_audit WHERE table_name='wedding'").Scan(&audits)
	audit := audits[0]

	expectedData, err := json.Marshal(wedding)
	require.NoError(t, err)

	require.Equal(t, "wedding", audit["table_name"])
	require.Equal(t, "insert", audit["op"])
	require.NotNil(t, audit["id"])
	require.NotNil(t, audit["created_at"])
	require.Equal(t, identity.Id, audit["identity_id"])
	require.Nil(t, audit["trace_id"])

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expectedData, []byte(audit["data"].(string)), &opts)
	if diff != jsondiff.FullMatch {
		t.Errorf("data column does not match expected: %s", explanation)
	}
}

// func TestAuditTablesWithTracing(t *testing.T) {
// 	dbConnInfo := &db.ConnectionInfo{
// 		Host:     "localhost",
// 		Port:     "8001",
// 		Username: "postgres",
// 		Password: "postgres",
// 		Database: "keel",
// 	}

// 	schema := protoSchema(t, keelSchema)
// 	ctx := context.Background()

// 	database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, "audit_test")
// 	require.NoError(t, err)
// 	defer database.Close()

// 	pk, err := testhelpers.GetEmbeddedPrivateKey()
// 	require.NoError(t, err)

// 	ctx = runtimectx.WithPrivateKey(ctx, pk)
// 	ctx = db.WithDatabase(ctx, database)

// 	// provider := trace.NewTracerProvider()
// 	// otel.SetTracerProvider(provider)
// 	// otel.SetTextMapPropagator(propagation.TraceContext{})

// 	identity, err := actions.CreateIdentity(ctx, schema, "dave.new@keel.xyz", "1234")
// 	require.NoError(t, err)

// 	ctx = auth.WithIdentity(ctx, identity)
// 	var tracer = otel.Tracer("github.com/teamkeel/keel/db")
// 	ctx, span := tracer.Start(ctx, "Execute Statement")
// 	trace.ContextWithSpan()
// 	//spanContext := .SpanContextFromContext(ctx)
// 	action := proto.FindAction(schema, "createWedding")
// 	input := map[string]any{"name": "Dave"}
// 	scope := actions.NewScope(ctx, action, schema)
// 	_, err = actions.Create(scope, input)
// 	require.NoError(t, err)

// 	traceId := span.SpanContext().TraceID().String()
// 	span.End()

// 	db := database.GetDB()

// 	var weddings []map[string]any
// 	db.Raw("SELECT * FROM wedding").Scan(&weddings)
// 	wedding := weddings[0]

// 	var audits []map[string]any
// 	db.Raw("SELECT * FROM keel_audit WHERE table_name='wedding'").Scan(&audits)
// 	audit := audits[0]

// 	expectedData, err := json.Marshal(wedding)
// 	require.NoError(t, err)

// 	require.Equal(t, "wedding", audit["table_name"])
// 	require.Equal(t, "insert", audit["op"])
// 	require.NotNil(t, audit["id"])
// 	require.NotNil(t, audit["created_at"])
// 	require.Nil(t, identity.Id, audit["identity_id"])
// 	require.Equal(t, traceId, audit["trace_id"])

// 	opts := jsondiff.DefaultConsoleOptions()
// 	diff, explanation := jsondiff.Compare(expectedData, []byte(audit["data"].(string)), &opts)
// 	if diff != jsondiff.FullMatch {
// 		t.Errorf("data column does not match expected: %s", explanation)
// 	}
// }

// var rows []map[string]any
// db.Raw(`
// 	SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns
// 	WHERE table_schema = 'public' AND table_name = 'keel_audit';`).Scan(&rows)

// var expected = []map[string]any{
// 	{
// 		"column_name":    "id",
// 		"data_type":      "text",
// 		"is_nullable":    "NO",
// 		"column_default": "ksuid()",
// 	},
// 	{
// 		"column_name":    "table_name",
// 		"data_type":      "text",
// 		"is_nullable":    "NO",
// 		"column_default": nil,
// 	},
// 	{
// 		"column_name":    "op",
// 		"data_type":      "text",
// 		"is_nullable":    "NO",
// 		"column_default": nil,
// 	},
// 	{
// 		"column_name":    "data",
// 		"data_type":      "jsonb",
// 		"is_nullable":    "NO",
// 		"column_default": nil,
// 	},
// 	{
// 		"column_name":    "created_at",
// 		"data_type":      "timestamp with time zone",
// 		"is_nullable":    "NO",
// 		"column_default": "now()",
// 	},
// 	{
// 		"column_name":    "identity_id",
// 		"data_type":      "text",
// 		"is_nullable":    "YES",
// 		"column_default": nil,
// 	},
// 	{
// 		"column_name":    "trace_id",
// 		"data_type":      "text",
// 		"is_nullable":    "YES",
// 		"column_default": nil,
// 	},
// }

// require.Equal(t, len(expected), len(rows), "number of columns not matching")

// for _, e := range expected {
// 	row, found := lo.Find(rows, func(r map[string]any) bool {
// 		return r["column_name"] == e["column_name"]
// 	})

// 	require.True(t, found, fmt.Sprintf("Column '%s' not found", e["column_name"]))
// 	require.True(t, reflect.DeepEqual(e, row), fmt.Sprintf("Column '%s' doesnt match. Expected: {%s}, Found: {%s}", e["column_name"], e, row))
// }
