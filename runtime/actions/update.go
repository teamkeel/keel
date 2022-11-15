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

	err := DefaultCaptureImplicitWriteInputValues(scope.operation.Inputs, values, scope)
	if err != nil {
		return nil, err
	}

	err = DefaultCaptureSetValues(scope, values)
	if err != nil {
		return nil, err
	}

	where, ok := input["where"].(map[string]any)
	if !ok {
		where = map[string]any{}
	}

	err = DefaultApplyImplicitFilters(scope, where)
	if err != nil {
		return nil, err
	}

	err = DefaultApplyExplicitFilters(scope, where)
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

	isAuthorised, err := DefaultIsAuthorised(scope, permissionInputs)
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

	err = scope.query.Updates(scope.writeValues).Error
	if err != nil {
		return nil, err
	}

	// todo: Use RETURNING statement on UPDATE
	// https://linear.app/keel/issue/RUN-146/gorm-use-returning-on-insert-and-update-statements
	results := []map[string]any{}
	scope.query = scope.query.WithContext(scope.context).Find(&results)

	if scope.query.Error != nil {
		return nil, scope.query.Error
	}

	n := len(results)
	if n == 0 {
		return nil, errors.New("no records found for Update() operation")
	}
	if n > 1 {
		return nil, fmt.Errorf("Update() operation should find only one record, it found: %d", n)
	}

	result := toLowerCamelMap(results[0])
	return result, nil
}
