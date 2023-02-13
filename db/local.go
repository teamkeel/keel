package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"regexp"
	"strconv"
	"time"

	"github.com/lib/pq"
)

type localDb struct {
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

func convertTime(value time.Time) time.Time {
	// data api doesn't return nanos
	nanos := 0
	return time.Date(value.Year(), value.Month(), value.Day(), value.Hour(), value.Minute(), value.Second(), nanos, time.UTC)
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
	if slices.Contains(SupportedValueTypes, fmt.Sprintf("%T", value)) {
		return nil
	}
	return fmt.Errorf("unsupported %T", value)
}

func (db *localDb) ExecuteQuery(ctx context.Context, sqlQuery string, values ...any) (*ExecuteQueryResult, error) {
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
		return nil, toDbError(err)
	}

	columns, err := result.Columns()
	if err != nil {
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
			return nil, toDbError(err)
		}
		rowMap := map[string]any{}
		for i, cell := range row {
			timeCell, isTime := cell.(time.Time)
			if isTime {
				cell = convertTime(timeCell)
			}
			rowMap[columns[i]] = cell
		}
		rows = append(rows, rowMap)
	}
	return &ExecuteQueryResult{Rows: rows}, nil
}

func (db *localDb) ExecuteStatement(ctx context.Context, sqlQuery string, values ...any) (*ExecuteStatementResult, error) {
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
		return nil, toDbError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, toDbError(err)
	}
	return &ExecuteStatementResult{RowsAffected: rowsAffected}, nil
}

func (db *localDb) BeginTransaction(ctx context.Context) error {
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

func (db *localDb) CommitTransaction(ctx context.Context) error {
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

func (db *localDb) RollbackTransaction(ctx context.Context) error {
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
