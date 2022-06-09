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

// todo - we need better tests than this, but to design them is hard because
// the behaviour very stateful, and the state I'm talking about is the
// state of the host's docker environment.
//
// To create test fixtures that managed that state - would need the very same
// code as we are trying to test.
