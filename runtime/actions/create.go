package actions

import (
	"context"

	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

func Create(scope *Scope, input map[string]any) (res map[string]any, err error) {
	database, err := runtimectx.GetDatabase(scope.context)
	if err != nil {
		return nil, err
	}

	err = database.Transaction(scope.context, func(ctx context.Context) error {
		scope := scope.WithContext(ctx)
		query := NewQuery(scope.model)

		// Generate the SQL statement
		statement, err := GenerateCreateStatement(query, scope, input)
		if err != nil {
			return err
		}

		// Execute database request, expecting a single result
		res, err = statement.ExecuteToSingle(scope.context)
		if err != nil {
			return err
		}

		isAuthorised, err := AuthoriseSingle(scope, res)
		if err != nil {
			return err
		}

		if !isAuthorised {
			return common.NewPermissionError()
		}

		return nil
	})

	return res, err
}

func GenerateCreateStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, error) {
	err := query.captureWriteValues(scope, input)
	if err != nil {
		return nil, err
	}

	err = query.captureSetValues(scope, input)
	if err != nil {
		return nil, err
	}

	// Return the inserted row
	query.AppendReturning(AllFields())

	return query.InsertStatement(), nil
}
