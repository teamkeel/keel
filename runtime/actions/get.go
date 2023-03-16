package actions

import (
	"github.com/teamkeel/keel/runtime/common"
)

func Get(scope *Scope, input map[string]any) (map[string]any, error) {
	query := NewQuery(scope.model)

	// Generate the SQL statement
	statement, err := GenerateGetStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	isAuthorised, err := query.isAuthorised(scope, input)
	if err != nil {
		return nil, err
	}

	if !isAuthorised {
		return nil, common.NewPermissionError()
	}

	// Execute database request, expecting a single result.
	return statement.ExecuteToSingle(scope.context)
}

func GenerateGetStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, error) {
	err := query.applyImplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	err = query.applyExplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	// Select all columns and distinct on id
	query.AppendSelect(AllFields())
	query.AppendDistinctOn(IdField())

	return query.SelectStatement(), nil
}
