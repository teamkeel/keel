package actions

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
)

type Row map[string]any

func Get(scope *Scope, input map[string]any) (Row, error) {
	query := NewQuery(scope.schema, scope.operation)

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
	query.AppendSelect("*")
	query.AppendDistinctOn("id")

	// Execute database request with results
	results, affected, err := query.SelectStatement().ExecuteWithResults(scope)
	if err != nil {
		return nil, err
	}

	if affected == 0 {
		return nil, errors.New("no records found for Get() operation")
	} else if affected > 1 {
		return nil, fmt.Errorf("Get() operation should find only one record, it found: %d", affected)
	}

	return toLowerCamelMap(results[0]), nil
}
