package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	_ "github.com/lib/pq"
	"os"
)

type ExecuteQueryResult struct {
	Rows []map[string]any
}

type ExecuteStatementResult struct {
	RowsAffected int64
}

type Db interface {
	ExecuteQuery(ctx context.Context, sql string, values ...any) (*ExecuteQueryResult, error)
	ExecuteStatement(ctx context.Context, sql string, values ...any) (*ExecuteStatementResult, error)
	BeginTransaction(ctx context.Context) error
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error
}

func ResolveFromEnv(ctx context.Context, localConnString string) (Db, error) {
	dbConnType := os.Getenv("DB_CONN_TYPE")
	switch dbConnType {
	case "", "pg":
		return Local(ctx, localConnString)
	case "dataapi":
		return DataAPI(ctx)
	default:
		return nil, fmt.Errorf("unexpected DB_CONN_TYPE: %s", dbConnType)
	}
}

func Local(ctx context.Context, connString string) (Db, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return &localDb{conn: conn}, nil
}

func DataAPI(ctx context.Context) (Db, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := rdsdata.NewFromConfig(cfg)
	dbResourceArn := os.Getenv("DB_RESOURCE_ARN")
	if dbResourceArn == "" {
		return nil, errors.New("the DB_RESOURCE_ARN env var is not set")
	}
	dbSecretArn := os.Getenv("DB_SECRET_ARN")
	if dbSecretArn == "" {
		return nil, errors.New("the DB_SECRET_ARN env var is not set")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return nil, errors.New("the DB_NAME env var is not set")
	}
	return &dataApi{
		client:               client,
		dbResourceArn:        dbResourceArn,
		dbSecretArn:          dbSecretArn,
		dbName:               dbName,
		ongoingTransactionId: nil,
	}, nil
}
