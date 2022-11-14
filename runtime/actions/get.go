package actions

import (
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/proto"
)

type Row map[string]any

func Get(scope *Scope, input map[string]any) (Row, error) {
	err := DefaultApplyImplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	err = DefaultApplyExplicitFilters(scope, input)
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
		return ParseGetObjectResponse(scope.context, op, input)
	}

	results := []map[string]any{}
	scope.query = scope.query.
		WithContext(scope.context).
		Select(fmt.Sprintf("DISTINCT %s.*", strcase.ToSnake(scope.model.Name))). // TODO: expand to related models
		Find(&results)

	if scope.query.Error != nil {
		return nil, scope.query.Error
	}

	n := len(results)
	if n == 0 {
		return nil, errors.New("no records found for Get() operation")
	}

	if n > 1 {
		return nil, fmt.Errorf("Get() operation should find only one record, it found: %d", n)
	}

	return toLowerCamelMap(results[0]), nil
}
