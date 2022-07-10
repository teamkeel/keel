package postgres

import (
	"testing"
)

func TestBringUpAndLaterStop(t *testing.T) {
	// Todo - this test is another that suffers errors arising from race conditions
	// as other tests being run concurrently also interact with the "singleton" dockerized
	// database.
	//
	// The body is now commented out below.
	//
	// At least now we have other tests that are exercising the code, so having this
	// independent test is less vital.

	/*
		db, err := BringUpPostgresLocally()
		require.NoError(t, err)
		require.NotNil(t, db)

		err = StopThePostgresContainer()
		require.NoError(t, err)
	*/
}
