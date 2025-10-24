package db

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestDb(t *testing.T, ctx context.Context) Database {
	dbConnInfo := &ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Password: "postgres",
		Database: "keel",
	}

	db, err := New(ctx, dbConnInfo.String())
	require.NoError(t, err)
	return db
}

func TestDbTransactionConcurrency(t *testing.T) {
	ctx := t.Context()
	db := createTestDb(t, ctx)

	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS testdbtransactionconcurrency")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, err = db.ExecuteStatement(context.Background(), "DROP TABLE IF EXISTS testdbtransactionconcurrency")
		assert.NoError(t, err)
	})

	_, err = db.ExecuteStatement(ctx, `CREATE TABLE testdbtransactionconcurrency(
        id               text PRIMARY KEY
    );`)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	expectedRows := 0

	for i := range 20 {
		wg.Add(1)

		rollback := i%2 == 0
		if !rollback {
			expectedRows++
		}

		go func() {
			defer wg.Done()

			err = db.Transaction(ctx, func(ctx context.Context) error {
				_, err = db.ExecuteStatement(ctx, `INSERT INTO testdbtransactionconcurrency (id) values (?)`, ksuid.New().String())
				assert.NoError(t, err)

				if rollback {
					return errors.New("rollback pls")
				}

				return nil
			})

			if rollback {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		}()
	}

	wg.Wait()

	r, err := db.ExecuteQuery(ctx, "select * from testdbtransactionconcurrency")
	assert.NoError(t, err)
	assert.Len(t, r.Rows, expectedRows)
}

func TestDbTransactionCommit(t *testing.T) {
	ctx := t.Context()
	db := createTestDb(t, ctx)

	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS test_local_transaction_commit_table")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(context.Background(), "DROP TABLE IF EXISTS test_local_transaction_commit_table")
	})

	_, err = db.ExecuteStatement(ctx, "CREATE TABLE test_local_transaction_commit_table (id text, foo boolean)")
	assert.NoError(t, err)

	err = db.Transaction(ctx, func(ctx context.Context) error {
		_, err = db.ExecuteQuery(ctx, "INSERT INTO test_local_transaction_commit_table (id, foo) values (?, ?)", "1", true)
		assert.NoError(t, err)

		// Querying table outside of the transaction should return no rows
		result, err := db.ExecuteQuery(t.Context(), "SELECT * FROM test_local_transaction_commit_table")
		assert.NoError(t, err)
		assert.Equal(t, []map[string]any{}, result.Rows)

		// Return no error and commit
		return nil
	})
	assert.NoError(t, err)

	// Transaction was committed, row should be returned
	result, err := db.ExecuteQuery(ctx, "SELECT * FROM test_local_transaction_commit_table")
	assert.NoError(t, err)
	assert.Len(t, result.Rows, 1)
}

func TestDbTransactionRollback(t *testing.T) {
	ctx := t.Context()
	db := createTestDb(t, ctx)

	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS test_local_transaction_rollback_table")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(context.Background(), "DROP TABLE IF EXISTS test_local_transaction_rollback_table")
	})

	_, err = db.ExecuteStatement(ctx, "CREATE TABLE test_local_transaction_rollback_table (id text, foo boolean)")
	assert.NoError(t, err)

	err = db.Transaction(ctx, func(ctx context.Context) error {
		_, err = db.ExecuteQuery(ctx, "INSERT INTO test_local_transaction_rollback_table (id, foo) values (?, ?)", "1", true)
		assert.NoError(t, err)

		// Querying table outside of the transaction should return no rows
		result, err := db.ExecuteQuery(t.Context(), "SELECT * FROM test_local_transaction_rollback_table")
		assert.NoError(t, err)
		assert.Equal(t, []map[string]any{}, result.Rows)
		assert.Equal(t, []string{"id", "foo"}, result.Columns)

		// Return an error and rollback
		return errors.New("my error message")
	})
	assert.Error(t, err)
	assert.Equal(t, "my error message", err.Error())

	// Transaction was rolled bad, no rows should be returned
	result, err := db.ExecuteQuery(ctx, "SELECT * FROM test_local_transaction_rollback_table")
	assert.NoError(t, err)
	assert.Len(t, result.Rows, 0)
}

func TestDbStatements(t *testing.T) {
	ctx := t.Context()
	db := createTestDb(t, ctx)
	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(context.Background(), "DROP TABLE IF EXISTS person")
	})
	_, err = db.ExecuteStatement(ctx, `CREATE TABLE person(
        id               text PRIMARY KEY,
        name             text,
        married          boolean,
        favourite_number integer,
        date             timestamptz
    );`)
	assert.NoError(t, err)

	keelKeelsonValues := []any{"id1", "Keel Keelson", true, 10, time.Date(2013, 3, 1, 9, 10, 59, 897000, time.Local)}
	agentSmithValues := []any{"id2", "Agent Smith", false, 1, time.Date(2022, 4, 3, 12, 1, 33, 567000, time.Local)}
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
	assert.Equal(t, []string{"id", "name", "married", "favourite_number", "date"}, result.Columns)
	expectedData := []map[string]interface{}{
		{"date": time.Date(2013, time.March, 1, 9, 10, 59, 897000, time.Local), "favourite_number": int32(10), "id": "id1", "married": true, "name": "Keel Keelson"},
		{"date": time.Date(2022, time.April, 3, 12, 1, 33, 567000, time.Local), "favourite_number": int32(1), "id": "id2", "married": false, "name": "Agent Smith"},
		{"date": interface{}(nil), "favourite_number": interface{}(nil), "id": "id3", "married": interface{}(nil), "name": interface{}(nil)},
	}
	assert.Equal(t, 3, len(result.Rows))
	for i, row := range result.Rows {
		// NOTE: this may fail on your local due to localisation issues
		assert.Equal(t, expectedData[i], row)
	}

	statementResult, err = db.ExecuteStatement(ctx, "UPDATE person SET name = 'named' WHERE name IS NOT DISTINCT FROM ?", nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), statementResult.RowsAffected)

	statementResult, err = db.ExecuteStatement(ctx, "DELETE FROM person")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), statementResult.RowsAffected)
}

func TestUniqueConstraintViolation(t *testing.T) {
	ctx := t.Context()
	db := createTestDb(t, ctx)
	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(context.Background(), "DROP TABLE IF EXISTS person")
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
	dbError1 := &DbError{}
	if assert.ErrorAs(t, err, &dbError1) {
		assert.Equal(t, dbError1.Columns[0], "name")
		assert.Equal(t, dbError1.Table, "person")
		assert.NotEmpty(t, dbError1.Message)
		assert.Equal(t, dbError1.PgErrCode, PgUniqueConstraintViolation)
		pgErr := &pgconn.PgError{}
		assert.ErrorAs(t, dbError1.Err, &pgErr)
		assert.Equal(t, PgUniqueConstraintViolation, pgErr.Code)
	}

	_, err = db.ExecuteQuery(ctx, "INSERT INTO person (id, name) VALUES (?, ?)", keelKeelsonValuesNotUniqueName...)
	dbError2 := &DbError{}
	if assert.ErrorAs(t, err, &dbError2) {
		assert.Equal(t, dbError2.Columns[0], "name")
		assert.Equal(t, dbError2.Table, "person")
		assert.NotEmpty(t, dbError2.Message)
		assert.Equal(t, dbError2.PgErrCode, PgUniqueConstraintViolation)
		pgErr := &pgconn.PgError{}
		assert.ErrorAs(t, dbError2.Err, &pgErr)
		assert.Equal(t, PgUniqueConstraintViolation, pgErr.Code)
	}
}

func TestNotNullConstraintViolation(t *testing.T) {
	ctx := t.Context()
	db := createTestDb(t, ctx)
	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(context.Background(), "DROP TABLE IF EXISTS person")
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
	dbError1 := &DbError{}
	if assert.ErrorAs(t, err, &dbError1) {
		assert.Equal(t, "name", dbError1.Columns[0])
		assert.Equal(t, "person", dbError1.Table)
		assert.NotEmpty(t, dbError1.Message)
		assert.Equal(t, PgNotNullConstraintViolation, dbError1.PgErrCode)
		pgErr := &pgconn.PgError{}
		assert.ErrorAs(t, dbError1.Err, &pgErr)
		assert.Equal(t, PgNotNullConstraintViolation, pgErr.Code)
	}

	_, err = db.ExecuteQuery(ctx, "INSERT INTO person (id, name) VALUES (?, ?)", notNameValues...)

	dbError2 := &DbError{}
	if assert.ErrorAs(t, err, &dbError2) {
		assert.Equal(t, dbError2.Columns[0], "name")
		assert.Equal(t, PgNotNullConstraintViolation, dbError2.PgErrCode)
	}
}

func TestForeignKeyConstraintViolation(t *testing.T) {
	ctx := t.Context()
	db := createTestDb(t, ctx)

	_, err := db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS person")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(context.Background(), "DROP TABLE IF EXISTS person")
	})

	_, err = db.ExecuteStatement(ctx, "DROP TABLE IF EXISTS company")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.ExecuteStatement(context.Background(), "DROP TABLE IF EXISTS company")
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
	dbError1 := &DbError{}
	if assert.ErrorAs(t, err, &dbError1) {
		assert.Equal(t, "company_id", dbError1.Columns[0])
		assert.Equal(t, "person", dbError1.Table)
		assert.NotEmpty(t, dbError1.Message)
		assert.Equal(t, PgForeignKeyConstraintViolation, dbError1.PgErrCode)
		pgErr := &pgconn.PgError{}
		assert.ErrorAs(t, dbError1.Err, &pgErr)
		assert.Equal(t, PgForeignKeyConstraintViolation, pgErr.Code)
	}

	_, err = db.ExecuteQuery(ctx, "INSERT INTO person (id, name, company_id) VALUES (?, ?, ?)", noCompanyValues...)
	dbError2 := &DbError{}
	if assert.ErrorAs(t, err, &dbError2) {
		assert.Equal(t, "company_id", dbError2.Columns[0])
		assert.Equal(t, "person", dbError2.Table)
		assert.NotEmpty(t, dbError2.Message)
		assert.Equal(t, PgForeignKeyConstraintViolation, dbError2.PgErrCode)
		pgErr := &pgconn.PgError{}
		assert.ErrorAs(t, dbError2.Err, &pgErr)
		assert.Equal(t, PgForeignKeyConstraintViolation, pgErr.Code)
	}
}
