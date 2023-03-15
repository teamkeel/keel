package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSuite(t *testing.T, createTestDb func(t *testing.T, ctx context.Context) Db) {
	suite := dbTestSuite{
		CreateTestDb: createTestDb,
	}
	t.Run("testDbTransactionError", suite.testDbTransactionError)
	t.Run("testDbTransactionCommit", suite.testDbTransactionCommit)
	t.Run("testDbTransactionRollback", suite.testDbTransactionRollback)
	t.Run("testDbStatements", suite.testDbStatements)
	t.Run("testErrUniqueConstraintViolation", suite.testErrUniqueConstraintViolation)
	t.Run("testErrForeignKeyConstraintViolation", suite.testErrForeignKeyConstraintViolation)
	t.Run("testErrNotNullConstraintViolation", suite.testErrNotNullConstraintViolation)
}

type dbTestSuite struct {
	CreateTestDb func(t *testing.T, ctx context.Context) Db
}

func (suite dbTestSuite) testDbTransactionError(t *testing.T) {
	ctx := context.Background()
	db := suite.CreateTestDb(t, ctx)

	err := db.CommitTransaction(ctx)
	assert.ErrorContains(t, err, "cannot commit transaction when there is no ongoing transaction")
	err = db.RollbackTransaction(ctx)
	assert.ErrorContains(t, err, "cannot rollback transaction when there is no ongoing transaction")
}

func (suite dbTestSuite) testDbTransactionCommit(t *testing.T) {
	ctx := context.Background()
	db := suite.CreateTestDb(t, ctx)
	otherDb := suite.CreateTestDb(t, ctx)

	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS test_local_transaction_commit_table")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS test_local_transaction_commit_table")
	})

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

func (suite dbTestSuite) testDbTransactionRollback(t *testing.T) {
	ctx := context.Background()
	db := suite.CreateTestDb(t, ctx)
	otherDb := suite.CreateTestDb(t, ctx)

	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS testapi_local_transaction_rollback_table")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS testapi_local_transaction_rollback_table")
	})

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

func (suite dbTestSuite) testDbStatements(t *testing.T) {
	ctx := context.Background()
	db := suite.CreateTestDb(t, ctx)
	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	})
	_, err = db.ExecuteStatement(ctx, `CREATE TABLE person(
        id               text PRIMARY KEY,
        name             text,
        married          boolean,
        favourite_number integer,
        date             timestamptz
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
		{"date": time.Date(2013, time.March, 1, 9, 10, 59, 897000, time.UTC), "favourite_number": int64(10), "id": "id1", "married": true, "name": "Keel Keelson"},
		{"date": time.Date(2022, time.April, 3, 12, 1, 33, 567000, time.UTC), "favourite_number": int64(1), "id": "id2", "married": false, "name": "Agent Smith"},
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

func (suite dbTestSuite) testErrUniqueConstraintViolation(t *testing.T) {
	ctx := context.Background()
	db := suite.CreateTestDb(t, ctx)
	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	})

	_, err = db.ExecuteStatement(ctx, `CREATE TABLE person(
        id               text PRIMARY KEY,
        name             text
    );`)
	assert.NoError(t, err)

	_, err = db.ExecuteStatement(ctx, "ALTER TABLE person ADD CONSTRAINT name_unique_constraint UNIQUE(name);")
	assert.NoError(t, err)

	keelKeelsonValues := []any{"id1", "Keel Keelson"}
	keelKeelsonValuesNotUniqueName := []any{"id2", "Keel Keelson"}

	statementResult, err := db.ExecuteStatement(ctx, "INSERT INTO person (id, name) VALUES (?, ?)", keelKeelsonValues...)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), statementResult.RowsAffected)

	_, err = db.ExecuteStatement(ctx, "INSERT INTO person (id, name) VALUES (?, ?)", keelKeelsonValuesNotUniqueName...)
	assert.ErrorIs(t, err, ErrUniqueConstraintViolation)
	dbError1 := &DbError{}
	if assert.ErrorAs(t, err, &dbError1) {
		assert.Equal(t, dbError1.Column, "name")
	}

	_, err = db.ExecuteQuery(ctx, "INSERT INTO person (id, name) VALUES (?, ?)", keelKeelsonValuesNotUniqueName...)
	assert.ErrorIs(t, err, ErrUniqueConstraintViolation)
	dbError2 := &DbError{}
	if assert.ErrorAs(t, err, &dbError2) {
		assert.Equal(t, dbError2.Column, "name")
	}
}

func (suite dbTestSuite) testErrNotNullConstraintViolation(t *testing.T) {
	ctx := context.Background()
	db := suite.CreateTestDb(t, ctx)
	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	})

	_, err = db.ExecuteStatement(ctx, `CREATE TABLE person(
        id               text PRIMARY KEY,
        name             text NOT NULL
    );`)
	assert.NoError(t, err)

	_, err = db.ExecuteStatement(ctx, "ALTER TABLE person ADD CONSTRAINT name_unique_constraint UNIQUE(name);")
	assert.NoError(t, err)

	keelKeelsonValues := []any{"id1", "Keel Keelson"}
	notNameValues := []any{"id2", nil}

	statementResult, err := db.ExecuteStatement(ctx, "INSERT INTO person (id, name) VALUES (?, ?)", keelKeelsonValues...)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), statementResult.RowsAffected)

	_, err = db.ExecuteStatement(ctx, "INSERT INTO person (id, name) VALUES (?, ?)", notNameValues...)
	assert.ErrorIs(t, err, ErrNotNullConstraintViolation)
	dbError1 := &DbError{}
	if assert.ErrorAs(t, err, &dbError1) {
		assert.Equal(t, dbError1.Column, "name")
	}

	_, err = db.ExecuteQuery(ctx, "INSERT INTO person (id, name) VALUES (?, ?)", notNameValues...)
	assert.ErrorIs(t, err, ErrNotNullConstraintViolation)
	dbError2 := &DbError{}
	if assert.ErrorAs(t, err, &dbError2) {
		assert.Equal(t, dbError2.Column, "name")
	}
}

func (suite dbTestSuite) testErrForeignKeyConstraintViolation(t *testing.T) {
	ctx := context.Background()
	db := suite.CreateTestDb(t, ctx)

	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	})

	_, err = db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS company")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS company")
	})

	_, err = db.ExecuteStatement(ctx, `CREATE TABLE person(
        id               text PRIMARY KEY,
        name             text,
		company_id		 text
    );`)
	assert.NoError(t, err)

	_, err = db.ExecuteStatement(ctx, `CREATE TABLE company(
        id               text PRIMARY KEY
    );`)
	assert.NoError(t, err)

	_, err = db.ExecuteStatement(ctx, "ALTER TABLE person ADD FOREIGN KEY (company_id) REFERENCES company(id)")
	assert.NoError(t, err)

	keelKeelsonValues := []any{"id1", "Keel Keelson", "123"}
	noCompanyValues := []any{"id2", "No Company", "999"}

	statementResultCompany, err := db.ExecuteStatement(ctx, "INSERT INTO company (id) VALUES (?)", "123")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), statementResultCompany.RowsAffected)

	statementResult, err := db.ExecuteStatement(ctx, "INSERT INTO person (id, name, company_id) VALUES (?, ?, ?)", keelKeelsonValues...)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), statementResult.RowsAffected)

	_, err = db.ExecuteStatement(ctx, "INSERT INTO person (id, name, company_id) VALUES (?, ?, ?)", noCompanyValues...)
	assert.ErrorIs(t, err, ErrForeignKeyConstraintViolation)
	dbError1 := &DbError{}
	if assert.ErrorAs(t, err, &dbError1) {
		assert.Equal(t, dbError1.Column, "company_id")
	}

	_, err = db.ExecuteQuery(ctx, "INSERT INTO person (id, name, company_id) VALUES (?, ?, ?)", noCompanyValues...)
	assert.ErrorIs(t, err, ErrForeignKeyConstraintViolation)
	dbError2 := &DbError{}
	if assert.ErrorAs(t, err, &dbError2) {
		assert.Equal(t, dbError2.Column, "company_id")
	}
}
