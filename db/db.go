package db

import (
	"context"
	"errors"

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
	ErrNotNullConstraintViolation    = errors.New("null value violates not null column constraint")
	ErrForeignKeyConstraintViolation = errors.New("insert or update violates foreign key constraint")
	ErrUniqueConstraintViolation     = errors.New("duplicate key value violates unique constraint")
)

type DbError struct {
	Columns []string
	Err     error
}

func (err *DbError) Error() string {
	return err.Err.Error()
}

func (err *DbError) Unwrap() error {
	return err.Err
}

type Database interface {
	// Executes SQL query statement and returns rows.
	ExecuteQuery(ctx context.Context, sql string, values ...any) (*ExecuteQueryResult, error)
	// Executes SQL statement and returns number of rows affected.
	ExecuteStatement(ctx context.Context, sql string, values ...any) (*ExecuteStatementResult, error)
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
