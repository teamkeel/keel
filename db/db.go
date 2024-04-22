package db

import (
	"context"

	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ExecuteQueryResult struct {
	Rows []map[string]any
}

type ExecuteStatementResult struct {
	RowsAffected int64
}

var (
	PgNotNullConstraintViolation    = "23502"
	PgForeignKeyConstraintViolation = "23503"
	PgUniqueConstraintViolation     = "23505"
)

type DbError struct {
	// if the error was associated with a specific table, the name of the table
	Table string
	// if the error was associated with specific table columns, the names of these columns
	Columns []string
	// the primary human-readable error message. This should be accurate but terse (typically one line). Always present
	Message string
	// the SQLSTATE code for the error - https://www.postgresql.org/docs/current/errcodes-appendix.html. Always present
	PgErrCode string
	// the underlying error
	Err error
}

func (err *DbError) Error() string {
	return err.Err.Error()
}

func (err *DbError) Unwrap() error {
	return err.Err
}

type Database interface {
	// Executes SQL query statement and returns rows.
	ExecuteQuery(ctx context.Context, sql string, args ...any) (*ExecuteQueryResult, error)
	// Executes SQL statement and returns number of rows affected.
	ExecuteStatement(ctx context.Context, sql string, args ...any) (*ExecuteStatementResult, error)
	// Runs fn inside a transaction which is commited if fn returns a nil error
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	Close() error
	GetDB() *gorm.DB
}

func New(ctx context.Context, connString string) (Database, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  connString,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger:                 logger.Discard,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	return &GormDB{db: db}, nil
}

func QuoteIdentifier(name string) string {
	return pq.QuoteIdentifier(name)
}

func QuoteLiteral(literal string) string {
	return pq.QuoteLiteral(literal)
}
