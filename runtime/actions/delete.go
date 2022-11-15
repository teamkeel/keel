package actions

import (
	"errors"

	"github.com/teamkeel/keel/proto"
)

func Delete(scope *Scope, input map[string]any) (bool, error) {
	err := DefaultApplyImplicitFilters(scope, input)
	if err != nil {
		return false, err
	}

	err = DefaultApplyExplicitFilters(scope, input)
	if err != nil {
		return false, err
	}

	isAuthorised, err := DefaultIsAuthorised(scope, input)
	if err != nil {
		return false, err
	}

	if !isAuthorised {
		return false, errors.New("not authorized to access this operation")
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
