package events

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/auditing"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/testhelpers"
)

func TestEventNameFromInsertAudit(t *testing.T) {
	eventName, err := eventNameFromAudit("company_employee", auditing.Insert)
	require.Equal(t, "company_employee.created", eventName)
	require.NoError(t, err)
}

func TestEventNameFromUpdateAudit(t *testing.T) {
	eventName, err := eventNameFromAudit("company_employee", auditing.Update)
	require.Equal(t, "company_employee.updated", eventName)
	require.NoError(t, err)
}

func TestEventNameFromDeleteAudit(t *testing.T) {
	eventName, err := eventNameFromAudit("company_employee", auditing.Delete)
	require.Equal(t, "company_employee.deleted", eventName)
	require.NoError(t, err)
}

func TestEventNameFromUnknown(t *testing.T) {
	eventName, err := eventNameFromAudit("company_employee", "unknown")
	require.Empty(t, eventName)
	require.Error(t, err)
}

func TestAuditOpFromCreatedEvent(t *testing.T) {
	op, err := auditOpFromAction(proto.ActionType_ACTION_TYPE_CREATE)
	require.Equal(t, auditing.Insert, op)
	require.NoError(t, err)
}

func TestAuditOpFromUpdatedEvent(t *testing.T) {
	op, err := auditOpFromAction(proto.ActionType_ACTION_TYPE_UPDATE)
	require.Equal(t, auditing.Update, op)
	require.NoError(t, err)
}

func TestAuditOpFromDeletedEvent(t *testing.T) {
	op, err := auditOpFromAction(proto.ActionType_ACTION_TYPE_DELETE)
	require.Equal(t, auditing.Delete, op)
	require.NoError(t, err)
}

func TestSingleEvent(t *testing.T) {
	var keelSchema = `
		model Person {
			fields {
				name Text
			}
			@on([update], verifyDetails)
		}`

	schema, err := testhelpers.MakeSchemaFromString(keelSchema)
	require.NoError(t, err)

	sql, err := processAuditLogsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
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

func TestComplexTableName(t *testing.T) {
	var keelSchema = `
		model EmployeeOfCompany1 {
			fields {
				name Text
			}
			@on([update], verifyDetails)
		}`

	schema, err := testhelpers.MakeSchemaFromString(keelSchema)
	require.NoError(t, err)

	sql, err := processAuditLogsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
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

func TestMultipleEventsOneAttribute(t *testing.T) {
	var keelSchema = `
		model Person {
			fields {
				name Text
			}
			@on([update, create, delete], verifyDetails)
		}`

	schema, err := testhelpers.MakeSchemaFromString(keelSchema)
	require.NoError(t, err)

	sql, err := processAuditLogsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
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

func TestMultipleEventsManyAttribute(t *testing.T) {
	var keelSchema = `
		model Person {
			fields {
				name Text
			}
			@on([update], verifyDetails)
			@on([create], verifyDetails)
			@on([delete], verifyDetails)
		}`

	schema, err := testhelpers.MakeSchemaFromString(keelSchema)
	require.NoError(t, err)

	sql, err := processAuditLogsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
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

func TestNoEvents(t *testing.T) {
	var keelSchema = `
		model Person {
			fields {
				name Text
			}
		}`

	schema, err := testhelpers.MakeSchemaFromString(keelSchema)
	require.NoError(t, err)

	sql, err := processAuditLogsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
	require.Error(t, err)
	require.Empty(t, sql)
}

func TestEmptyTraceId(t *testing.T) {
	var keelSchema = `
		model Person {
			fields {
				name Text
			}
			@on([update], verifyDetails)
		}`

	schema, err := testhelpers.MakeSchemaFromString(keelSchema)
	require.NoError(t, err)

	sql, err := processAuditLogsSql(schema, "")
	require.Error(t, err)
	require.Empty(t, sql)
}

func TestWithMultipleModels(t *testing.T) {
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

	schema, err := testhelpers.MakeSchemaFromString(keelSchema)
	require.NoError(t, err)

	sql, err := processAuditLogsSql(schema, "0ffe82e8dcfd9f9fbe4c639d5ef4f1ba")
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
