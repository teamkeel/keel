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
	op, err := opFromActionType(proto.ActionType_ACTION_TYPE_CREATE)
	require.Equal(t, Insert, op)
	require.NoError(t, err)
}

func TestAuditOpFromUpdatedEvent(t *testing.T) {
	op, err := opFromActionType(proto.ActionType_ACTION_TYPE_UPDATE)
	require.Equal(t, Update, op)
	require.NoError(t, err)
}

func TestAuditOpFromDeletedEvent(t *testing.T) {
	op, err := opFromActionType(proto.ActionType_ACTION_TYPE_DELETE)
	require.Equal(t, Delete, op)
	require.NoError(t, err)
}

func TestProcessEventSqlSingleEvent(t *testing.T) {
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

	sql, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.NoError(t, err)

	expectedSql := `
		UPDATE keel_audit 
		SET event_processed_at = NOW() 
		WHERE 
			trace_id = '0ffe82e8dcfd9f9fbe4c639d5ef4f1ba' AND 
			event_processed_at IS NULL AND 
			(table_name = 'person' AND op = 'update')
		RETURNING *`

	require.Equal(t, clean(expectedSql), clean(sql))
}

func TestProcessEventSqlComplexTableName(t *testing.T) {
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

	sql, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.NoError(t, err)

	expectedSql := `
		UPDATE keel_audit 
		SET event_processed_at = NOW() 
		WHERE 
			trace_id = '0ffe82e8dcfd9f9fbe4c639d5ef4f1ba' AND 
			event_processed_at IS NULL AND 
			(table_name = 'employee_of_company_1' AND op = 'update')
		RETURNING *`

	require.Equal(t, clean(expectedSql), clean(sql))
}

func TestProcessEventSqlMultipleEventsOneAttribute(t *testing.T) {
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

	sql, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.NoError(t, err)

	expectedSql := `
		UPDATE keel_audit 
		SET event_processed_at = NOW() 
		WHERE 
			trace_id = '0ffe82e8dcfd9f9fbe4c639d5ef4f1ba' AND 
			event_processed_at IS NULL AND 
			((table_name = 'person' AND op = 'update') OR 
			(table_name = 'person' AND op = 'insert') OR 
			(table_name = 'person' AND op = 'delete')) 
		RETURNING *`

	require.Equal(t, clean(expectedSql), clean(sql))
}

func TestProcessEventSqlMultipleEventsManyAttribute(t *testing.T) {
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

	sql, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.NoError(t, err)

	expectedSql := `
		UPDATE keel_audit 
		SET event_processed_at = NOW() 
		WHERE 
			trace_id = '0ffe82e8dcfd9f9fbe4c639d5ef4f1ba' AND 
			event_processed_at IS NULL AND 
			((table_name = 'person' AND op = 'update') OR 
			(table_name = 'person' AND op = 'insert') OR 
			(table_name = 'person' AND op = 'delete')) 
		RETURNING *`

	require.Equal(t, clean(expectedSql), clean(sql))
}

func TestProcessEventSqlNoEvents(t *testing.T) {
	var keelSchema = `
		model Person {
			fields {
				name Text
			}
		}`

	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(keelSchema, config.Empty)
	require.NoError(t, err)

	sql, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.Error(t, err)
	require.Empty(t, sql)
}

func TestProcessEventSqlEmptyTraceId(t *testing.T) {
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

	sql, err := processEventsSql(schema, "")
	require.Error(t, err)
	require.Empty(t, sql)
}

func TestProcessEventSqlWithMultipleModels(t *testing.T) {
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

	sql, err := processEventsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.NoError(t, err)

	expectedSql := `
		UPDATE keel_audit 
		SET event_processed_at = NOW() 
		WHERE 
			trace_id = '0ffe82e8dcfd9f9fbe4c639d5ef4f1ba' AND 
			event_processed_at IS NULL AND 
			((table_name = 'wedding' AND op = 'insert') OR 
			(table_name = 'wedding' AND op = 'update') OR 
			(table_name = 'wedding' AND op = 'delete') OR 
			(table_name = 'wedding_invitee' AND op = 'insert') OR 
			(table_name = 'wedding_invitee' AND op = 'update') OR 
			(table_name = 'person' AND op = 'update')) 
		RETURNING *`

	require.Equal(t, clean(expectedSql), clean(sql))
}

// Trims and removes redundant spacing
func clean(sql string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(sql)), " ")
}
