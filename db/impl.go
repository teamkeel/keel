package db

import (
	"context"
	"database/sql"
	"errors"
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
	conn               *sql.DB
	ongoingTransaction *sql.Tx
}

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
	ctx, span := tracer.Start(ctx, "ExecuteQuery")
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
	if db.ongoingTransaction != nil {
		result, err = db.ongoingTransaction.QueryContext(ctx, replaceQuestionMarksWithNumberedInputs(sqlQuery), values...)
	} else {
		result, err = db.conn.QueryContext(ctx, replaceQuestionMarksWithNumberedInputs(sqlQuery), values...)
	}
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
	ctx, span := tracer.Start(ctx, "ExecuteStatement")
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
	if db.ongoingTransaction != nil {
		result, err = db.ongoingTransaction.ExecContext(ctx, replaceQuestionMarksWithNumberedInputs(sqlQuery), values...)
	} else {
		result, err = db.conn.ExecContext(ctx, replaceQuestionMarksWithNumberedInputs(sqlQuery), values...)
	}
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

func (db *postgres) BeginTransaction(ctx context.Context) error {
	if db.ongoingTransaction != nil {
		return errors.New("cannot begin transaction when there is an ongoing transaction")
	}
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	db.ongoingTransaction = tx
	return nil
}

func (db *postgres) CommitTransaction(ctx context.Context) error {
	if db.ongoingTransaction == nil {
		return errors.New("cannot commit transaction when there is no ongoing transaction")
	}
	err := db.ongoingTransaction.Commit()
	if err != nil {
		return err
	}
	db.ongoingTransaction = nil
	return nil
}

func (db *postgres) RollbackTransaction(ctx context.Context) error {
	if db.ongoingTransaction == nil {
		return errors.New("cannot rollback transaction when there is no ongoing transaction")
	}
	err := db.ongoingTransaction.Rollback()
	if err != nil {
		return err
	}
	db.ongoingTransaction = nil
	return nil
}
