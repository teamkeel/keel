package actions

import (
	"errors"

	"github.com/teamkeel/keel/proto"
)

func Get(scope *Scope, input map[string]any) (map[string]any, error) {
	query := NewQuery(scope.model)

	err := query.applyImplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	err = query.applyExplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	isAuthorised, err := query.isAuthorised(scope, input)
	if err != nil {
		return nil, err
	}

	if !isAuthorised {
		return nil, errors.New("not authorized to access this operation")
	}

	if scope.operation.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		return ParseGetObjectResponse(scope.context, scope.operation, input)
	}

	// Select all columns and distinct on id
	query.AppendSelect(AllFields())
	query.AppendDistinctOn(IdField())

	// Execute database request, expecting a single result
	result, err := query.
		SelectStatement().
		ExecuteToSingle(scope.context)

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, errors.New("no records found for Get() operation")
	}

	return result, nil
}
