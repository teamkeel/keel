package db

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/db")

type GormDB struct {
	db *gorm.DB
}

var _ Database = &GormDB{}

func (db *GormDB) ExecuteQuery(ctx context.Context, sqlQuery string, values ...any) (*ExecuteQueryResult, error) {
	ctx, span := tracer.Start(ctx, "Execute Query")
	defer span.End()

	span.SetAttributes(attribute.String("sql", sqlQuery))

	rows := []map[string]any{}
	conn := db.db.WithContext(ctx)

	// Check for a transaction
	if v, ok := ctx.Value(transactionCtxKey).(*gorm.DB); ok {
		conn = v
	}

	err := conn.Raw(sqlQuery, values...).Scan(&rows).Error
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, toDbError(err)
	}

	span.SetAttributes(attribute.Int("rows.count", len(rows)))
	return &ExecuteQueryResult{Rows: rows}, nil
}

func (db *GormDB) ExecuteStatement(ctx context.Context, sqlQuery string, values ...any) (*ExecuteStatementResult, error) {
	ctx, span := tracer.Start(ctx, "Execute Statement")
	defer span.End()

	span.SetAttributes(attribute.String("sql", sqlQuery))

	conn := db.db.WithContext(ctx)

	// Check for a transaction
	if v, ok := ctx.Value(transactionCtxKey).(*gorm.DB); ok {
		conn = v
	}

	result := conn.Exec(sqlQuery, values...)

	err := result.Error
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, toDbError(err)
	}

	span.SetAttributes(attribute.Int("rows.affected", int(result.RowsAffected)))
	return &ExecuteStatementResult{RowsAffected: result.RowsAffected}, nil
}

var transactionCtxKey struct{}

func (db *GormDB) Transaction(ctx context.Context, fn func(context.Context) error) error {
	ctx, span := tracer.Start(ctx, "Database Transaction")
	defer span.End()

	return db.db.Transaction(func(tx *gorm.DB) (err error) {
		ctx = context.WithValue(ctx, transactionCtxKey, tx)
		return fn(ctx)
	})
}

func (db *GormDB) Close() error {
	conn, err := db.db.DB()
	if err != nil {
		return err
	}

	return conn.Close()
}

func (db *GormDB) GetDB() *gorm.DB {
	return db.db
}

func toDbError(err error) error {
	var pgxErr *pgconn.PgError
	if !errors.As(err, &pgxErr) {
		return err
	}

	switch pgxErr.Code {
	case "23502":
		return &DbError{
			Columns: []string{pgxErr.ColumnName},
			Err:     ErrNotNullConstraintViolation,
			PgErr:   pgxErr,
		}
	case "23503":
		// Extract column and value from "Key (author_id)=(2L2ar5NCPvTTEdiDYqgcpF3f5QN1) is not present in table \"author\"."
		out := regexp.MustCompile(`\(([^)]+)\)`).FindAllStringSubmatch(pgxErr.Detail, -1)
		return &DbError{
			Columns: []string{out[0][1]},
			Err:     ErrForeignKeyConstraintViolation,
			PgErr:   pgxErr,
		}
	case "23505":
		// Extract column and value from "Key (code)=(1234) already exists."
		out := regexp.MustCompile(`\(([^)]+)\)`).FindAllStringSubmatch(pgxErr.Detail, -1)
		cols := strings.Split(out[0][1], ", ")
		return &DbError{
			Columns: cols,
			Err:     ErrUniqueConstraintViolation,
			PgErr:   pgxErr,
		}
	default:
		return err
	}
}
