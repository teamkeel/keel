package actions

import (
	"github.com/teamkeel/keel/runtime/common"
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
		return nil, common.RuntimeError{Code: common.ErrPermissionDenied, Message: "not authorized to access this operation"}
	}

	// Select all columns and distinct on id
	query.AppendSelect(AllFields())
	query.AppendDistinctOn(IdField())

	// Execute database request, expecting a single result
	return query.
		SelectStatement().
		ExecuteToSingle(scope.context)
}
