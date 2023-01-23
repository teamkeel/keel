package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata/types"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type dataApi struct {
	client               *rdsdata.Client
	dbResourceArn        string
	dbSecretArn          string
	dbName               string
	ongoingTransactionId *string
}

func replaceQuestionMarksWithColonParam(query string) string {
	output := ""
	nextInputId := 1
	for _, char := range query {
		charStr := string(char)
		if charStr == "?" {
			output += ":param" + strconv.Itoa(nextInputId)
			nextInputId += 1
		} else {
			output += charStr
		}
	}
	return output
}

func valuesToColonParamMap(values []any) ([]types.SqlParameter, error) {
	result := []types.SqlParameter{}
	for i, value := range values {
		paramName := "param" + strconv.Itoa(i+1)
		field, typeHint, err := convertValueToField(value)
		if err != nil {
			return nil, err
		}
		result = append(result, types.SqlParameter{
			Name:     &paramName,
			TypeHint: typeHint,
			Value:    field,
		})
	}
	return result, nil
}

// copied from "github.com/krotscheck/go-rds-driver" dialect.go
func isNil(i any) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

// adapted copy from "github.com/krotscheck/go-rds-driver" dialect.go
func convertValueToField(value any) (types.Field, types.TypeHint, error) {
	var noTypeHint types.TypeHint
	if isNil(value) {
		return &types.FieldMemberIsNull{Value: true}, noTypeHint, nil
	}
	switch t := value.(type) {
	case string:
		return &types.FieldMemberStringValue{Value: t}, noTypeHint, nil
	case []byte:
		return &types.FieldMemberBlobValue{Value: t}, noTypeHint, nil
	case bool:
		return &types.FieldMemberBooleanValue{Value: t}, noTypeHint, nil
	case float32:
		return &types.FieldMemberDoubleValue{Value: float64(t)}, noTypeHint, nil
	case float64:
		return &types.FieldMemberDoubleValue{Value: t}, noTypeHint, nil
	case int:
		return &types.FieldMemberLongValue{Value: int64(t)}, noTypeHint, nil
	case int8:
		return &types.FieldMemberLongValue{Value: int64(t)}, noTypeHint, nil
	case int16:
		return &types.FieldMemberLongValue{Value: int64(t)}, noTypeHint, nil
	case int32:
		return &types.FieldMemberLongValue{Value: int64(t)}, noTypeHint, nil
	case int64:
		return &types.FieldMemberLongValue{Value: t}, noTypeHint, nil
	case uint:
		return &types.FieldMemberLongValue{Value: int64(t)}, noTypeHint, nil
	case uint8:
		return &types.FieldMemberBlobValue{Value: []byte{t}}, noTypeHint, nil
	case uint16:
		return &types.FieldMemberLongValue{Value: int64(t)}, noTypeHint, nil
	case uint32:
		return &types.FieldMemberLongValue{Value: int64(t)}, noTypeHint, nil
	case uint64:
		return &types.FieldMemberLongValue{Value: int64(t)}, noTypeHint, nil
	case time.Time:
		return &types.FieldMemberStringValue{
			Value: t.Format("2006-01-02 15:04:05.999"),
		}, types.TypeHintTimestamp, nil
	case nil:
		return &types.FieldMemberIsNull{Value: true}, noTypeHint, nil
	default:
		return nil, noTypeHint, fmt.Errorf("unsupported type: %#v", value)
	}
}

// adapted copy from "github.com/krotscheck/go-rds-driver" dialect.go & dialect_postgres.go
func convertFieldToValue(field types.Field, columnType string) (any, error) {
	_, isNull := field.(*types.FieldMemberIsNull)
	if isNull {
		return nil, nil
	}
	switch strings.ToLower(columnType) {
	case "numeric":
		return strconv.ParseFloat(field.(*types.FieldMemberStringValue).Value, 64)
	case "date":
		t, err := time.Parse("2006-01-02", field.(*types.FieldMemberStringValue).Value)
		if err != nil {
			return nil, err
		}
		return t, nil
	case "time":
		timeStringVal := field.(*types.FieldMemberStringValue).Value
		return time.Parse("15:04:05", timeStringVal)
	case "timestamp":
		t, err := time.Parse("2006-01-02 15:04:05", field.(*types.FieldMemberStringValue).Value)
		if err != nil {
			return nil, err
		}
		return t, nil
	}

	switch v := field.(type) {
	case *types.FieldMemberArrayValue:
		return v.Value, nil
	case *types.FieldMemberBlobValue:
		return v.Value, nil
	case *types.FieldMemberBooleanValue:
		return v.Value, nil
	case *types.FieldMemberDoubleValue:
		return v.Value, nil
	case *types.FieldMemberLongValue:
		return v.Value, nil
	case *types.FieldMemberStringValue:
		return v.Value, nil
	default:
		return nil, fmt.Errorf("unrecognized RDS field type: %#v", field)
	}
}

func (db *dataApi) ExecuteQuery(ctx context.Context, sql string, values ...any) (*ExecuteQueryResult, error) {
	query := replaceQuestionMarksWithColonParam(sql)
	params, err := valuesToColonParamMap(values)
	if err != nil {
		return nil, err
	}
	executeStatementOutput, err := db.client.ExecuteStatement(ctx, &rdsdata.ExecuteStatementInput{
		ResourceArn:           &db.dbResourceArn,
		SecretArn:             &db.dbSecretArn,
		Sql:                   &query,
		Database:              &db.dbName,
		Parameters:            params,
		TransactionId:         db.ongoingTransactionId,
		IncludeResultMetadata: true,
	})
	if err != nil {
		return nil, err
	}

	rows := []map[string]any{}

	columns := executeStatementOutput.ColumnMetadata
	for _, record := range executeStatementOutput.Records {
		row := map[string]any{}
		for i, cellField := range record {
			value, err := convertFieldToValue(cellField, *columns[i].TypeName)
			if err != nil {
				return nil, err
			}
			columnName := *columns[i].Name
			row[columnName] = value
		}
		rows = append(rows, row)
	}

	return &ExecuteQueryResult{
		Rows: rows,
	}, nil
}

func (db *dataApi) ExecuteStatement(ctx context.Context, sql string, values ...any) (*ExecuteStatementResult, error) {
	query := replaceQuestionMarksWithColonParam(sql)
	params, err := valuesToColonParamMap(values)
	if err != nil {
		return nil, err
	}
	executeStatementOutput, err := db.client.ExecuteStatement(ctx, &rdsdata.ExecuteStatementInput{
		ResourceArn:   &db.dbResourceArn,
		SecretArn:     &db.dbSecretArn,
		Sql:           &query,
		Database:      &db.dbName,
		Parameters:    params,
		TransactionId: db.ongoingTransactionId,
	})
	if err != nil {
		return nil, err
	}
	return &ExecuteStatementResult{RowsAffected: executeStatementOutput.NumberOfRecordsUpdated}, nil

}

func (db *dataApi) BeginTransaction(ctx context.Context) error {
	if db.ongoingTransactionId != nil {
		return errors.New("cannot begin transaction when there is an ongoing transaction")
	}
	beginTransactionOutput, err := db.client.BeginTransaction(ctx, &rdsdata.BeginTransactionInput{
		ResourceArn: &db.dbResourceArn,
		SecretArn:   &db.dbSecretArn,
		Database:    &db.dbName,
	})
	if err != nil {
		return err
	}
	db.ongoingTransactionId = beginTransactionOutput.TransactionId
	return nil
}

func (db *dataApi) CommitTransaction(ctx context.Context) error {
	if db.ongoingTransactionId == nil {
		return errors.New("cannot commit transaction when there is no ongoing transaction")
	}
	_, err := db.client.CommitTransaction(ctx, &rdsdata.CommitTransactionInput{
		ResourceArn:   &db.dbResourceArn,
		SecretArn:     &db.dbSecretArn,
		TransactionId: db.ongoingTransactionId,
	})
	if err != nil {
		return err
	}
	db.ongoingTransactionId = nil
	return nil
}

func (db *dataApi) RollbackTransaction(ctx context.Context) error {
	if db.ongoingTransactionId == nil {
		return errors.New("cannot rollback transaction when there is no ongoing transaction")
	}
	_, err := db.client.RollbackTransaction(ctx, &rdsdata.RollbackTransactionInput{
		ResourceArn:   &db.dbResourceArn,
		SecretArn:     &db.dbSecretArn,
		TransactionId: db.ongoingTransactionId,
	})
	if err != nil {
		return err
	}
	db.ongoingTransactionId = nil
	return nil
}
