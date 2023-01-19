package actions

import (
	"github.com/teamkeel/keel/runtime/common"
)

func Delete(scope *Scope, input map[string]any) (*string, error) {
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
		return nil, common.RuntimeError{Code: common.ErrPermissionDenied, Message: "not authorized to access this operation"}
	}

	query.AppendReturning(Field("id"))

	// Execute database request
	row, err := query.
		DeleteStatement().
		ExecuteToSingle(scope.context)

	// TODO: if the error is multiple rows affected then rollback transaction
	if err != nil {
		return nil, err
	}

	if row == nil {
		return nil, common.RuntimeError{Code: common.ErrRecordNotFound, Message: "record not found"}
	}

	id, _ := row["id"].(string)
	return &id, nil
}
