package db

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"

	"github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slices"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/db")

type postgres struct {
	conn *sql.DB
}

var _ Db = &postgres{}

func replaceQuestionMarksWithNumberedInputs(query string) string {
	output := ""
	nextInputId := 1
	for _, char := range query {
		charStr := string(char)
		if charStr == "?" {
			output += "$" + strconv.Itoa(nextInputId)
			nextInputId += 1
		} else {
			output += charStr
		}
	}
	return output
}

func toDbError(err error) error {
	if value, ok := err.(*pq.Error); ok {
		switch value.Code {
		case "23502":
			return &DbError{Column: value.Column, Err: ErrNotNullConstraintViolation}
		case "23503":
			// Extract column and value from "Key (author_id)=(2L2ar5NCPvTTEdiDYqgcpF3f5QN1) is not present in table \"author\"."
			out := regexp.MustCompile(`\(([^)]+)\)`).FindAllStringSubmatch(value.Detail, -1)
			return &DbError{Column: out[0][1], Err: ErrForeignKeyConstraintViolation}
		case "23505":
			// Extract column and value from "Key (code)=(1234) already exists."
			out := regexp.MustCompile(`\(([^)]+)\)`).FindAllStringSubmatch(value.Detail, -1)
			return &DbError{Column: out[0][1], Err: ErrUniqueConstraintViolation}
		default:
			return err
		}
	}
	return err
}

func validateSupportedType(value any) error {
	if slices.Contains(supportedValueTypes, fmt.Sprintf("%T", value)) {
		return nil
	}
	return fmt.Errorf("unsupported %T", value)
}

func (db *postgres) ExecuteQuery(ctx context.Context, sqlQuery string, values ...any) (*ExecuteQueryResult, error) {
	ctx, span := tracer.Start(ctx, "Execute Query")
	defer span.End()

	span.SetAttributes(attribute.String("sql", sqlQuery))

	for _, value := range values {
		err := validateSupportedType(value)
		if err != nil {
			return nil, err
		}
	}
	rows := []map[string]any{}

	var result *sql.Rows
	var err error
	var conn interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}

	conn = db.conn

	// Check for transaction
	if v, ok := ctx.Value(transactionCtxKey).(*sql.Tx); ok {
		conn = v
	}

	result, err = conn.QueryContext(ctx, replaceQuestionMarksWithNumberedInputs(sqlQuery), values...)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, toDbError(err)
	}

	columns, err := result.Columns()
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, toDbError(err)
	}
	for result.Next() {
		row := make([]any, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range pointers {
			pointers[i] = &row[i]
		}
		err = result.Scan(pointers...)
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
			return nil, toDbError(err)
		}
		rowMap := map[string]any{}
		for i, cell := range row {
			rowMap[columns[i]] = cell
		}
		rows = append(rows, rowMap)
	}

	span.SetAttributes(attribute.Int("rows.count", len(rows)))
	return &ExecuteQueryResult{Rows: rows}, nil
}

func (db *postgres) ExecuteStatement(ctx context.Context, sqlQuery string, values ...any) (*ExecuteStatementResult, error) {
	ctx, span := tracer.Start(ctx, "Execute Statement")
	defer span.End()

	span.SetAttributes(attribute.String("sql", sqlQuery))

	for _, value := range values {
		err := validateSupportedType(value)
		if err != nil {
			return nil, err
		}
	}
	var result sql.Result
	var err error
	var conn interface {
		ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	}

	conn = db.conn

	// Check for a transaction
	if v, ok := ctx.Value(transactionCtxKey).(*sql.Tx); ok {
		conn = v
	}

	result, err = conn.ExecContext(ctx, replaceQuestionMarksWithNumberedInputs(sqlQuery), values...)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, toDbError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, toDbError(err)
	}

	span.SetAttributes(attribute.Int("rows.affected", int(rowsAffected)))
	return &ExecuteStatementResult{RowsAffected: rowsAffected}, nil
}

var transactionCtxKey struct{}

func (db *postgres) Transaction(ctx context.Context, fn func(context.Context) error) error {
	ctx, span := tracer.Start(ctx, "Database Transaction")
	defer span.End()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Make sure we rollback even if there is a panic
	defer func() {
		if r := recover(); r != nil {
			span.SetAttributes(attribute.Bool("panic", true))
			span.SetAttributes(attribute.Bool("rollback", true))
			if err, ok := r.(error); ok {
				span.RecordError(err, trace.WithStackTrace(true))
				span.SetStatus(codes.Error, err.Error())
			}
			_ = tx.Rollback()
			panic(err)
		}
	}()

	ctx = context.WithValue(ctx, transactionCtxKey, tx)

	err = fn(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("rollback", true))
		e := tx.Rollback()
		if e == nil {
			return err
		}
		return fmt.Errorf("error rolling back transaction: %s (original error: %w)", e.Error(), err)
	}

	span.SetAttributes(attribute.Bool("commit", true))
	return tx.Commit()
}
