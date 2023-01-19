package actions

import (
	"github.com/teamkeel/keel/runtime/common"
)

func Create(scope *Scope, input map[string]any) (res map[string]any, err error) {
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

	// Return the inserted row
	query.AppendReturning(AllFields())

	// Begin a transaction and defer rollback which will run if a commit hasn't occurred.
	query.Begin(scope.context)
	// Defer ensures a rollback as the function may return early due to an error.
	defer func() {
		if err != nil {
			query.Rollback(scope.context)
		}
	}()

	// Execute database request, expecting a single result
	result, err := query.
		InsertStatement().
		ExecuteToSingle(scope.context)

	if err != nil {
		return nil, err
	}

	// Retrieve the newly created row so we can check permissions
	query.Where(IdField(), Equals, Value(result["id"]))

	// Check permissions and roles conditions
	isAuthorised, err := query.isAuthorised(scope, input)
	if err != nil {
		return nil, err
	}

	if !isAuthorised {
		return nil, common.RuntimeError{Code: common.ErrPermissionDenied, Message: "not authorized to access this operation"}
	}

	err = query.Commit(scope.context)
	if err != nil {
		return nil, err
	}

	return result, nil
}
