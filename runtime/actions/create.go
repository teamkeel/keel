package actions

import (
	"errors"

	"github.com/teamkeel/keel/proto"
)

func Create(scope *Scope, input map[string]any) (map[string]any, error) {
	var err error

	query := NewQuery(scope.model)

	defaultValues, err := initialValueForModel(scope.model, scope.schema)
	if err != nil {
		return nil, err
	}

	query.AddWriteValues(defaultValues)

	err = query.captureWriteValues(scope, input)
	if err != nil {
		return nil, err
	}

	err = query.captureSetValues(scope, input)
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

	op := scope.operation
	if op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		return ParseCreateObjectResponse(scope.context, op, input)
	}

	// Return the inserted row
	query.AppendReturning(AllFields())

	// Execute database request, expecting a single result
	result, err := query.
		InsertStatement().
		ExecuteToSingle(scope.context)

	if err != nil {
		return nil, err
	}

	return result, nil
}
