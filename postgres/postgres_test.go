package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// This test is just a code exerciser - to see if the start/stop cycle
// runs without error (two cycles). And to provide a debugging entry point.
func TestBringUpShutDown(t *testing.T) {
	err := BringUpPostgresLocally()
	require.NoError(t, err)

	err = StopThePostgresContainer()
	require.NoError(t, err)

	// Another full up/down cycle...

	err = BringUpPostgresLocally()
	require.NoError(t, err)

	err = StopThePostgresContainer()
	require.NoError(t, err)
}

// todo - we need better tests than this, but to design them is hard because
// the behaviour very stateful, and the state I'm talking about is the
// state of the host's docker environment.
//
// To create test fixtures that managed that state - would need the very same
// code as we are trying to test.
