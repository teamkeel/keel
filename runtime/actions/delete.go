package actions

import (
	"errors"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
)

func Delete(scope *Scope, input map[string]any) (res *string, err error) {
	// Attempt to resolve permissions early; i.e. before row-based database querying.
	permissions := proto.PermissionsForAction(scope.Schema, scope.Action)
	canResolveEarly, authorised, err := TryResolveAuthorisationEarly(scope, permissions)
	if err != nil {
		return nil, err
	}

	// Generate the SQL statement
	query := NewQuery(scope.Model)
	statement, err := GenerateDeleteStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	var row map[string]any

	switch {
	case canResolveEarly && !authorised:
		return nil, common.NewPermissionError()
	case !canResolveEarly:
		query.AppendSelect(IdField())
		query.AppendDistinctOn(IdField())
		rows, err := query.SelectStatement().ExecuteToSingle(scope.Context)
		if err != nil {
			return nil, err
		}

		rowsToAuthorise := []map[string]any{}
		if rows != nil {
			rowsToAuthorise = append(rowsToAuthorise, rows)
		}

		isAuthorised, err := AuthoriseAction(scope, input, rowsToAuthorise)
		if err != nil {
			return nil, err
		}

		if !isAuthorised {
			return nil, common.NewPermissionError()
		}
	}

	// Execute database request, expecting a single result
	row, err = statement.ExecuteToSingle(scope.Context)
	if err != nil {
		return nil, err
	}

	if row == nil {
		return nil, common.NewNotFoundError()
	}

	id, ok := row["id"].(string)
	if !ok {
		return nil, errors.New("could not parse id key")
	}

	return &id, err
}

func GenerateDeleteStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, error) {
	err := query.applyImplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	err = query.applyExplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	query.AppendReturning(Field("id"))

	return query.DeleteStatement(scope.Context), nil
}
