package actions

import (
	"errors"

	"github.com/teamkeel/keel/proto"
)

func Delete(scope *Scope, input map[string]any) (bool, error) {
	query := NewQuery(scope.model)

	err := query.applyImplicitFilters(scope, input)
	if err != nil {
		return false, err
	}

	err = query.applyExplicitFilters(scope, input)
	if err != nil {
		return false, err
	}

	isAuthorised, err := query.isAuthorised(scope, input)
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

	// Execute database request
	affected, err := query.
		DeleteStatement().
		Execute(scope.context)

	if err != nil {
		return false, err
	}

	if affected == 0 {
		return false, errors.New("no records found for Delete() operation")
	}

	return true, nil
}
