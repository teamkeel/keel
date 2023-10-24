package actions

import (
	"context"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/runtime/common"
)

func Update(scope *Scope, input map[string]any) (res map[string]any, err error) {
	database, err := db.GetDatabase(scope.Context)
	if err != nil {
		return nil, err
	}

	err = database.Transaction(scope.Context, func(ctx context.Context) error {
		scope := scope.WithContext(ctx)
		query := NewQuery(scope.Context, scope.Model)

		// Generate the SQL statement
		statement, err := GenerateUpdateStatement(query, scope, input)
		if err != nil {
			return err
		}

		query.AppendSelect(IdField())
		query.AppendDistinctOn(IdField())
		rowToAuthorise, err := query.SelectStatement().ExecuteToSingle(scope.Context)
		if err != nil {
			return err
		}

		rowsToAuthorise := []map[string]any{}
		if rowToAuthorise != nil {
			rowsToAuthorise = append(rowsToAuthorise, rowToAuthorise)
		}

		isAuthorised, err := AuthoriseAction(scope, input, rowsToAuthorise)
		if err != nil {
			return err
		}

		if !isAuthorised {
			return common.NewPermissionError()
		}

		// Execute database request, expecting a single result
		res, err = statement.ExecuteToSingle(scope.Context)
		if err != nil {
			return err
		}

		if res == nil {
			return common.NewNotFoundError()
		}

		return nil
	})

	return res, err
}

func GenerateUpdateStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, error) {
	values, ok := input["values"].(map[string]any)
	if !ok {
		values = map[string]any{}
	}

	err := query.captureWriteValues(scope, values)
	if err != nil {
		return nil, err
	}

	err = query.captureSetValues(scope, values)
	if err != nil {
		return nil, err
	}

	where, ok := input["where"].(map[string]any)
	if !ok {
		where = map[string]any{}
	}

	err = query.applyImplicitFilters(scope, where)
	if err != nil {
		return nil, err
	}

	err = query.applyExplicitFilters(scope, where)
	if err != nil {
		return nil, err
	}

	// Return the updated row
	query.AppendReturning(AllFields())

	return query.UpdateStatement(), nil
}
