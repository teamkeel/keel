package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/cmd/database"
	"github.com/teamkeel/keel/db"
)

func CreateTestDb(t *testing.T, ctx context.Context) db.Db {
	dbConnInfo := &database.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Password: "postgres",
		Database: "keel",
	}

	db, err := db.Local(ctx, dbConnInfo)
	require.NoError(t, err)
	return db
}

func TestLocalDbTransactionError(t *testing.T) {
	ctx := context.Background()
	db := CreateTestDb(t, ctx)

	err := db.CommitTransaction(ctx)
	assert.ErrorContains(t, err, "cannot commit transaction when there is no ongoing transaction")
	err = db.RollbackTransaction(ctx)
	assert.ErrorContains(t, err, "cannot rollback transaction when there is no ongoing transaction")
}

func TestLocalDbTransactionCommit(t *testing.T) {
	ctx := context.Background()
	db := CreateTestDb(t, ctx)
	otherDb := CreateTestDb(t, ctx)

	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS test_local_transaction_commit_table")
	assert.NoError(t, err)

	err = db.BeginTransaction(ctx)
	assert.NoError(t, err)
	err = db.BeginTransaction(ctx)
	assert.ErrorContains(t, err, "cannot begin transaction when there is an ongoing transaction")

	_, err = db.ExecuteStatement(ctx, "CREATE TABLE test_local_transaction_commit_table (id text, foo boolean)")
	assert.NoError(t, err)

	_, err = db.ExecuteQuery(ctx, "SELECT * FROM test_local_transaction_commit_table")
	assert.NoError(t, err)

	_, err = otherDb.ExecuteQuery(ctx, "SELECT * FROM test_local_transaction_commit_table")
	assert.ErrorContains(t, err, "relation \"test_local_transaction_commit_table\" does not exist")

	err = db.CommitTransaction(ctx)
	assert.NoError(t, err)
	err = db.CommitTransaction(ctx)
	assert.ErrorContains(t, err, "cannot commit transaction when there is no ongoing transaction")

	result, err := otherDb.ExecuteQuery(ctx, "SELECT * FROM test_local_transaction_commit_table")
	assert.NoError(t, err)
	assert.Equal(t, []map[string]any{}, result.Rows)
}

func TestLocalDbTransactionRollback(t *testing.T) {
	ctx := context.Background()
	db := CreateTestDb(t, ctx)
	otherDb := CreateTestDb(t, ctx)

	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS testapi_local_transaction_rollback_table")
	assert.NoError(t, err)

	err = db.BeginTransaction(ctx)
	assert.NoError(t, err)

	_, err = db.ExecuteStatement(ctx, "CREATE TABLE testapi_local_transaction_rollback_table (id text, foo boolean)")
	assert.NoError(t, err)

	_, err = db.ExecuteQuery(ctx, "SELECT * FROM testapi_local_transaction_rollback_table")
	assert.NoError(t, err)

	_, err = otherDb.ExecuteQuery(ctx, "SELECT * FROM testapi_local_transaction_rollback_table")
	assert.ErrorContains(t, err, "relation \"testapi_local_transaction_rollback_table\" does not exist")

	err = db.RollbackTransaction(ctx)
	assert.NoError(t, err)
	err = db.RollbackTransaction(ctx)
	assert.ErrorContains(t, err, "cannot rollback transaction when there is no ongoing transaction")

	_, err = db.ExecuteQuery(ctx, "SELECT * FROM testapi_local_transaction_rollback_table")
	assert.ErrorContains(t, err, "relation \"testapi_local_transaction_rollback_table\" does not exist")

	_, err = otherDb.ExecuteQuery(ctx, "SELECT * FROM testapi_local_transaction_rollback_table")
	assert.ErrorContains(t, err, "relation \"testapi_local_transaction_rollback_table\" does not exist")
}

func TestLocalDbStatements(t *testing.T) {
	ctx := context.Background()
	db := CreateTestDb(t, ctx)
	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	assert.NoError(t, err)
	_, err = db.ExecuteStatement(ctx, `CREATE TABLE person(
        id               text PRIMARY KEY,
        name             text,
        married          boolean,
        favourite_number integer,
        date             timestamp
    );`)
	assert.NoError(t, err)

	keelKeelsonValues := []any{"id1", "Keel Keelson", true, 10, time.Date(2013, 3, 1, 9, 10, 59, 897000, time.UTC)}
	agentSmithValues := []any{"id2", "Agent Smith", false, 1, time.Date(2022, 4, 3, 12, 1, 33, 567000, time.UTC)}
	nullPersonValues := []any{"id3", nil, nil, nil, nil}

	statementResult, err := db.ExecuteStatement(ctx, "INSERT INTO person (id, name, married, favourite_number, date) VALUES (?, ?, ?, ?, ?)", keelKeelsonValues...)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), statementResult.RowsAffected)
	_, err = db.ExecuteStatement(ctx, "INSERT INTO person (id, name, married, favourite_number, date) VALUES (?, ?, ?, ?, ?)", agentSmithValues...)
	assert.NoError(t, err)
	_, err = db.ExecuteStatement(ctx, "INSERT INTO person (id, name, married, favourite_number, date) VALUES (?, ?, ?, ?, ?)", nullPersonValues...)
	assert.NoError(t, err)

	result, err := db.ExecuteQuery(ctx, "SELECT * FROM person ORDER BY id ASC")
	assert.NoError(t, err)
	expectedData := []map[string]interface{}{
		{"date": time.Date(2013, time.March, 1, 9, 10, 59, 0, time.UTC), "favourite_number": int64(10), "id": "id1", "married": true, "name": "Keel Keelson"},
		{"date": time.Date(2022, time.April, 3, 12, 1, 33, 0, time.UTC), "favourite_number": int64(1), "id": "id2", "married": false, "name": "Agent Smith"},
		{"date": interface{}(nil), "favourite_number": interface{}(nil), "id": "id3", "married": interface{}(nil), "name": interface{}(nil)},
	}
	assert.Equal(t, expectedData, result.Rows)

	statementResult, err = db.ExecuteStatement(ctx, "UPDATE person SET name = 'named' WHERE name IS NOT DISTINCT FROM ?", nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), statementResult.RowsAffected)

	statementResult, err = db.ExecuteStatement(ctx, "DELETE FROM person")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), statementResult.RowsAffected)
}
