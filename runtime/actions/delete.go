package actions

import (
	"errors"

	"github.com/teamkeel/keel/proto"
)

func Delete(scope *Scope, input map[string]any) (bool, error) {
	err := DefaultApplyImplicitFilters(scope, input)
	if err != nil {
		scope.Error = err
		return false, scope.Error
	}

	err = DefaultApplyExplicitFilters(scope, input)
	if err != nil {
		scope.Error = err
		return false, scope.Error
	}

	isAuthorised, err := DefaultIsAuthorised(scope, input)
	if err != nil {
		scope.Error = err
		return false, scope.Error
	}

	if !isAuthorised {
		scope.Error = errors.New("not authorized to access this operation")
		return false, scope.Error
	}

	op := scope.operation

	if op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		return ParseDeleteResponse(scope.context, op, input)
	}

	records := []map[string]any{}
	err = scope.query.Delete(records).Error

	// TODO: handle this error properly
	return err == nil, nil
}
