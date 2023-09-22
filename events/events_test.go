package events

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/auditing"
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
