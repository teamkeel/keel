package auditing

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
)

func TestAuditOpFromCreatedEvent(t *testing.T) {
	t.Parallel()
	op, err := opFromActionType(proto.ActionType_ACTION_TYPE_CREATE)
	require.Equal(t, Insert, op)
	require.NoError(t, err)
}

func TestAuditOpFromUpdatedEvent(t *testing.T) {
	t.Parallel()
	op, err := opFromActionType(proto.ActionType_ACTION_TYPE_UPDATE)
	require.Equal(t, Update, op)
	require.NoError(t, err)
}

func TestAuditOpFromDeletedEvent(t *testing.T) {
	t.Parallel()
	op, err := opFromActionType(proto.ActionType_ACTION_TYPE_DELETE)
	require.Equal(t, Delete, op)
	require.NoError(t, err)
}

func TestProcessEventSqlSingleEvent(t *testing.T) {
	t.Parallel()
	var keelSchema = `
		model Person {
			fields {
				name Text
			}
			@on([update], verifyDetails)
		}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	sql, args, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.NoError(t, err)

	expectedSql := `
		UPDATE keel_audit 
		SET event_processed_at = now() 
		WHERE 
			trace_id = ? AND 
			event_processed_at IS NULL AND 
			(table_name = ? AND op = ?)
		RETURNING *`

	require.Equal(t, clean(expectedSql), clean(sql))
	require.Len(t, args, 3)
	require.Equal(t, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba", args[0])
	require.Equal(t, "person", args[1])
	require.Equal(t, "update", args[2])
}

func TestProcessEventSqlComplexTableName(t *testing.T) {
	t.Parallel()
	var keelSchema = `
		model EmployeeOfCompany1 {
			fields {
				name Text
			}
			@on([update], verifyDetails)
		}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	sql, args, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.NoError(t, err)

	expectedSql := `
		UPDATE keel_audit 
		SET event_processed_at = now() 
		WHERE 
			trace_id = ? AND 
			event_processed_at IS NULL AND 
			(table_name = ? AND op = ?)
		RETURNING *`

	require.Equal(t, clean(expectedSql), clean(sql))
	require.Len(t, args, 3)
	require.Equal(t, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba", args[0])
	require.Equal(t, "employee_of_company_1", args[1])
	require.Equal(t, "update", args[2])
}

func TestProcessEventSqlMultipleEventsOneAttribute(t *testing.T) {
	t.Parallel()
	var keelSchema = `
		model Person {
			fields {
				name Text
			}
			@on([update, create, delete], verifyDetails)
		}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	sql, args, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.NoError(t, err)

	expectedSql := `
		UPDATE keel_audit 
		SET event_processed_at = now() 
		WHERE 
			trace_id = ? AND 
			event_processed_at IS NULL AND 
			((table_name = ? AND op = ?) OR 
			(table_name = ? AND op = ?) OR 
			(table_name = ? AND op = ?)) 
		RETURNING *`

	require.Equal(t, clean(expectedSql), clean(sql))
	require.Len(t, args, 7)
	require.Equal(t, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba", args[0])
	require.Equal(t, "person", args[1])
	require.Equal(t, "update", args[2])
	require.Equal(t, "person", args[3])
	require.Equal(t, "insert", args[4])
	require.Equal(t, "person", args[5])
	require.Equal(t, "delete", args[6])
}

func TestProcessEventSqlMultipleEventsManyAttribute(t *testing.T) {
	t.Parallel()
	var keelSchema = `
		model Person {
			fields {
				name Text
			}
			@on([update], verifyDetails)
			@on([create], verifyDetails)
			@on([delete], verifyDetails)
		}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	sql, args, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.NoError(t, err)

	expectedSql := `
		UPDATE keel_audit 
		SET event_processed_at = now() 
		WHERE 
			trace_id = ? AND 
			event_processed_at IS NULL AND 
			((table_name = ? AND op = ?) OR 
			(table_name = ? AND op = ?) OR 
			(table_name = ? AND op = ?)) 
		RETURNING *`

	require.Equal(t, clean(expectedSql), clean(sql))
	require.Len(t, args, 7)
	require.Equal(t, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba", args[0])
	require.Equal(t, "person", args[1])
	require.Equal(t, "update", args[2])
	require.Equal(t, "person", args[3])
	require.Equal(t, "insert", args[4])
	require.Equal(t, "person", args[5])
	require.Equal(t, "delete", args[6])
}

func TestProcessEventSqlNoEvents(t *testing.T) {
	t.Parallel()
	var keelSchema = `
		model Person {
			fields {
				name Text
			}
		}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	sql, _, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.Error(t, err)
	require.Empty(t, sql)
}

func TestProcessEventSqlEmptyTraceId(t *testing.T) {
	t.Parallel()
	var keelSchema = `
		model Person {
			fields {
				name Text
			}
			@on([update], verifyDetails)
		}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	sql, _, err := processEventsSql(schema, "")
	require.Error(t, err)
	require.Empty(t, sql)
}

func TestProcessEventSqlWithMultipleModels(t *testing.T) {
	t.Parallel()
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
			@on([update], verifyDetails)
		}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	sql, args, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.NoError(t, err)

	expectedSql := `
		UPDATE keel_audit 
		SET event_processed_at = now() 
		WHERE 
			trace_id = ? AND 
			event_processed_at IS NULL AND 
			((table_name = ? AND op = ?) OR 
			(table_name = ? AND op = ?) OR 
			(table_name = ? AND op = ?) OR 
			(table_name = ? AND op = ?) OR 
			(table_name = ? AND op = ?) OR 
			(table_name = ? AND op = ?)) 
		RETURNING *`

	require.Equal(t, clean(expectedSql), clean(sql))
	require.Len(t, args, 13)
	require.Equal(t, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba", args[0])
	require.Equal(t, "wedding", args[1])
	require.Equal(t, "insert", args[2])
	require.Equal(t, "wedding", args[3])
	require.Equal(t, "update", args[4])
	require.Equal(t, "wedding", args[5])
	require.Equal(t, "delete", args[6])
	require.Equal(t, "wedding_invitee", args[7])
	require.Equal(t, "insert", args[8])
	require.Equal(t, "wedding_invitee", args[9])
	require.Equal(t, "update", args[10])
	require.Equal(t, "person", args[11])
	require.Equal(t, "update", args[12])
}

// Trims and removes redundant spacing
func clean(sql string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(sql)), " ")
}
