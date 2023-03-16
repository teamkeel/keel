package actions

import (
	"github.com/teamkeel/keel/runtime/common"
)

func Update(scope *Scope, input map[string]any) (map[string]any, error) {
	query := NewQuery(scope.model)

	values, ok := input["values"].(map[string]any)
	if !ok {
		values = map[string]any{}
	}

	where, ok := input["where"].(map[string]any)
	if !ok {
		where = map[string]any{}
	}

	// Generate the SQL statement
	statement, err := GenerateUpdateStatement(query, scope, input)
	if err != nil {
		return nil, err
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

	isAuthorised, err := query.isAuthorised(scope, permissionInputs)
	if err != nil {
		return nil, err
	}

	if !isAuthorised {
		return nil, common.NewPermissionError()
	}

	// Execute database request, expecting a single result
	result, err := statement.ExecuteToSingle(scope.context)

	// TODO: if error is multiple rows affected then rollback transaction
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, common.NewNotFoundError()
	}

	return result, nil
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
