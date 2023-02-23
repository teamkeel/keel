package db

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
)

type ExecuteQueryResult struct {
	Rows []map[string]any
}

type ExecuteStatementResult struct {
	RowsAffected int64
}

var (
	ErrNotNullConstraintViolation    = errors.New("null value violates not null column constraint")
	ErrForeignKeyConstraintViolation = errors.New("insert or update violates foreign key constraint")
	ErrUniqueConstraintViolation     = errors.New("duplicate key value violates unique constraint")
)

type DbError struct {
	Column string
	Err    error
}

func (err *DbError) Error() string {
	return err.Err.Error()
}

func (err *DbError) Unwrap() error {
	return err.Err
}

type Db interface {
	// Executes SQL query statement and returns rows.
	ExecuteQuery(ctx context.Context, sql string, values ...any) (*ExecuteQueryResult, error)
	// Executes SQL statement and returns number of rows affected.
	ExecuteStatement(ctx context.Context, sql string, values ...any) (*ExecuteStatementResult, error)
	// Begins a new transaction.
	BeginTransaction(ctx context.Context) error
	// Commits the current transaction.
	CommitTransaction(ctx context.Context) error
	// Rolls back the current transaction.
	RollbackTransaction(ctx context.Context) error
}

func New(ctx context.Context, dbConnInfo *ConnectionInfo) (Db, error) {
	sqlDb, err := sql.Open("postgres", dbConnInfo.String())
	if err != nil {
		return nil, err
	}
	return &postgres{conn: sqlDb}, nil
}

func NewFromConnection(ctx context.Context, sqlDb *sql.DB) (Db, error) {
	return &postgres{conn: sqlDb}, nil
}

var supportedValueTypes = []string{
	"<nil>",
	"bool",
	"string",
	"int",
	"int64",
	"float64",
	"time.Time",
}
