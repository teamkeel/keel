package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/db"
)

func CreateTestDb(t *testing.T, ctx context.Context) db.Db {
	dbConnInfo := &db.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Password: "postgres",
		Database: "keel",
	}

	db, err := db.New(ctx, dbConnInfo)
	require.NoError(t, err)
	return db
}

func TestLocalDb(t *testing.T) {
	db.TestSuite(t, CreateTestDb)
}
