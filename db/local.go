package db

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
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

func (db *localDb) ExecuteQuery(ctx context.Context, sqlQuery string, values ...any) (*ExecuteQueryResult, error) {
	rows := []map[string]any{}

	var result *sql.Rows
	var err error
	if db.ongoingTransaction != nil {
		result, err = db.ongoingTransaction.QueryContext(ctx, replaceQuestionMarksWithNumberedInputs(sqlQuery), values...)
		if err != nil {
			return nil, err
		}
	} else {
		result, err = db.conn.QueryContext(ctx, replaceQuestionMarksWithNumberedInputs(sqlQuery), values...)
		if err != nil {
			return nil, err
		}
	}
	columns, err := result.Columns()
	if err != nil {
		return nil, err
	}
	for result.Next() {
		row := make([]any, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range pointers {
			pointers[i] = &row[i]
		}
		err = result.Scan(pointers...)
		if err != nil {
			return nil, err
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
	var result sql.Result
	var err error
	if db.ongoingTransaction != nil {
		result, err = db.ongoingTransaction.ExecContext(ctx, replaceQuestionMarksWithNumberedInputs(sqlQuery), values...)
		if err != nil {
			return nil, err
		}
	} else {
		result, err = db.conn.ExecContext(ctx, replaceQuestionMarksWithNumberedInputs(sqlQuery), values...)
		if err != nil {
			return nil, err
		}
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
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
