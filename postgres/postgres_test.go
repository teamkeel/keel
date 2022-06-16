package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBringUpAndLaterStop(t *testing.T) {
	db, err := BringUpPostgresLocally()
	require.NoError(t, err)
	require.NotNil(t, db)

	err = StopThePostgresContainer()
	require.NoError(t, err)
}
