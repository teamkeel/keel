package actions

import (
	"errors"

	"github.com/teamkeel/keel/proto"
)

func Create(scope *Scope, input map[string]any) (Row, error) {
	var err error
	scope.writeValues, err = initialValueForModel(scope.model, scope.schema)
	if err != nil {
		return nil, err
	}

	err = DefaultCaptureImplicitWriteInputValues(scope.operation.Inputs, input, scope)
	if err != nil {
		return nil, err
	}

	err = DefaultCaptureSetValues(scope, input)
	if err != nil {
		return nil, err
	}

	isAuthorised, err := DefaultIsAuthorised(scope, input)
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

	err = scope.query.WithContext(scope.context).Create(scope.writeValues).Error
	if err != nil {
		return nil, err
	}

	// todo: Use RETURNING statement on INSERT
	// https://linear.app/keel/issue/RUN-146/gorm-use-returning-on-insert-and-update-statements
	return toLowerCamelMap(scope.writeValues), nil
}
