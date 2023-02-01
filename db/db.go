package db

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
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
	Table  string
	Column string
	Value  string
	Err    error
}

func (err DbError) Error() string {
	return err.Err.Error()
}

func (err DbError) Unwrap() error {
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

// Local data operations for a local database connection.
func Local(ctx context.Context, dbConnInfo *ConnectionInfo) (Db, error) {
	sqlDb, err := sql.Open("postgres", dbConnInfo.String())
	if err != nil {
		return nil, err
	}
	return &localDb{conn: sqlDb}, nil
}

// LocalFromConnection using an existing sql.DB connection to provide data operations to a database.
// Typically used for testing where it makes sense to reuse an existing connection.
func LocalFromConnection(ctx context.Context, sqlDb *sql.DB) (Db, error) {
	return &localDb{conn: sqlDb}, nil
}

func WithDataApiCredentials(resourceArn, secretArn, dbName string) func(d *dataApi) {
	return func(d *dataApi) {
		d.dbResourceArn = resourceArn
		d.dbSecretArn = secretArn
		d.dbName = dbName
	}
}

// DataAPI provides data operations for an Amazon Aurora database.
func DataAPI(ctx context.Context, opts ...func(d *dataApi)) (Db, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := rdsdata.NewFromConfig(cfg)

	d := &dataApi{
		client:        client,
		dbResourceArn: os.Getenv("DB_RESOURCE_ARN"),
		dbSecretArn:   os.Getenv("DB_SECRET_ARN"),
		dbName:        os.Getenv("DB_NAME"),
	}

	for _, o := range opts {
		o(d)
	}

	if d.dbResourceArn == "" {
		return nil, errors.New("the resource ARN was not provided")
	}

	if d.dbSecretArn == "" {
		return nil, errors.New("the secret ARN was not provided")
	}

	if d.dbName == "" {
		return nil, errors.New("the db name was not provided")
	}

	return d, nil
}
