package actions

import (
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
)

func Get(scope *Scope, input map[string]any) (map[string]any, error) {
	permissions := proto.PermissionsForAction(scope.Schema, scope.Action)

	// Attempt to resolve permissions early; i.e. before row-based database querying.
	canResolveEarly, authorised, err := TryResolveAuthorisationEarly(scope, permissions)
	if err != nil {
		return nil, err
	}
	if canResolveEarly && !authorised {
		return nil, common.NewPermissionError()
	}

	// Generate the SQL statement
	query := NewQuery(scope.Model)
	statement, err := GenerateGetStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	// Execute database request, expecting a single result.
	res, err := statement.ExecuteToSingle(scope.Context)
	if err != nil {
		return nil, err
	}

	rowsToAuthorise := []map[string]any{}
	if res != nil {
		rowsToAuthorise = append(rowsToAuthorise, res)
	}

	isAuthorised, err := AuthoriseAction(scope, input, rowsToAuthorise)
	if err != nil {
		return nil, err
	}

	if !isAuthorised {
		return nil, common.NewPermissionError()
	}

	return res, err
}

func GenerateGetStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, error) {
	err := query.applyImplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	err = query.applyExpressionFilters(scope, input)
	if err != nil {
		return nil, err
	}

	// Select all columns and distinct on id
	query.AppendSelect(AllFields())
	query.AppendDistinctOn(IdField())

	return query.SelectStatement(), nil
}
