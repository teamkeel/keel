package actions

import (
	"context"

	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

func Update(scope *Scope, input map[string]any) (res map[string]any, err error) {
	database, err := runtimectx.GetDatabase(scope.context)
	if err != nil {
		return nil, err
	}

	values, ok := input["values"].(map[string]any)
	if !ok {
		values = map[string]any{}
	}

	where, ok := input["where"].(map[string]any)
	if !ok {
		where = map[string]any{}
	}

	err = database.Transaction(scope.context, func(ctx context.Context) error {
		scope := scope.WithContext(ctx)
		query := NewQuery(scope.model)

		// Generate the SQL statement
		statement, err := GenerateUpdateStatement(query, scope, input)
		if err != nil {
			return err
		}

		// TODO: update so that permissions can't access inputs
		// https://linear.app/keel/issue/RUN-183/permission-expressions-barred-from-using-inputs
		permissionInputs := map[string]any{}
		for k, v := range where {
			permissionInputs[k] = v
		}
		for k, v := range values {
			permissionInputs[k] = v
		}

		// Execute database request, expecting a single result
		res, err = statement.ExecuteToSingle(scope.context)

		// TODO: if error is multiple rows affected then rollback transaction
		if err != nil {
			return err
		}

		if res == nil {
			return common.NewNotFoundError()
		}

		isAuthorised, err := AuthoriseSingle(scope, permissionInputs, res)
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
