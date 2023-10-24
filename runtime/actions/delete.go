package actions

import (
	"context"
	"errors"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/runtime/common"
)

func Delete(scope *Scope, input map[string]any) (res *string, err error) {
	database, err := db.GetDatabase(scope.Context)
	if err != nil {
		return nil, err
	}

	err = database.Transaction(scope.Context, func(ctx context.Context) error {
		query := NewQuery(scope.Context, scope.Model)

		// Generate the SQL statement
		statement, err := GenerateDeleteStatement(query, scope, input)
		if err != nil {
			return err
		}

		query.AppendSelect(IdField())
		query.AppendDistinctOn(IdField())
		rows, err := query.SelectStatement().ExecuteToSingle(scope.Context)
		if err != nil {
			return err
		}

		rowsToAuthorise := []map[string]any{}
		if rows != nil {
			rowsToAuthorise = append(rowsToAuthorise, rows)
		}

		isAuthorised, err := AuthoriseAction(scope, input, rowsToAuthorise)
		if err != nil {
			return err
		}

		if !isAuthorised {
			return common.NewPermissionError()
		}

		// Execute database request
		row, err := statement.ExecuteToSingle(scope.Context)
		if err != nil {
			return err
		}

		if row == nil {
			return common.NewNotFoundError()
		}

		id, ok := row["id"].(string)
		if !ok {
			return errors.New("could not parse id key")
		}

		res = &id
		return nil
	})

	return res, err
}

func GenerateDeleteStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, error) {
	err := query.applyImplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	err = query.applyExplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	query.AppendReturning(Field("id"))

	return query.DeleteStatement(), nil
}
