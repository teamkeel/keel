package actions

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
)

func Update(scope *Scope, input map[string]any) (Row, error) {
	values, ok := input["values"].(map[string]any)
	if !ok {
		values = map[string]any{}
	}

	query := NewQuery(scope.schema, scope.operation)

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

	// TODO: update so that permissions can't access inputs
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
		return nil, errors.New("not authorized to access this operation")
	}

	op := scope.operation
	if op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		return ParseUpdateResponse(scope.context, op, input)
	}

	// Return the updated row
	query.AppendReturning("*")

	// Execute database request with results
	results, affected, err := query.UpdateStatement().ExecuteWithResults(scope)
	if err != nil {
		return nil, err
	}

	if affected == 0 {
		return nil, errors.New("no records found for Update() operation")
	} else if affected > 1 {
		return nil, fmt.Errorf("Update() operation should update only one record, it updated: %d", affected)
	}

	result := toLowerCamelMap(results[0])
	return result, nil
}
