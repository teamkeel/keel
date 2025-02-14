package actions

import (
	"errors"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/locale"
)

func Delete(scope *Scope, input map[string]any) (res *string, err error) {
	// Attempt to resolve permissions early; i.e. before row-based database querying.
	permissions := proto.PermissionsForAction(scope.Schema, scope.Action)
	canResolveEarly, authorised, err := TryResolveAuthorisationEarly(scope, input, permissions)
	if err != nil {
		return nil, err
	}

	// Generate the SQL statement
	opts := []QueryBuilderOption{}
	if location, err := locale.GetTimeLocation(scope.Context); err == nil {
		opts = append(opts, WithTimezone(location.String()))
	}
	query := NewQuery(scope.Model, opts...)

	statement, err := GenerateDeleteStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	var row map[string]any
	switch {
	case canResolveEarly && !authorised:
		return nil, common.NewPermissionError()
	case !canResolveEarly:
		authQuery := NewQuery(scope.Model)
		err := authQuery.ApplyImplicitFilters(scope, input)
		if err != nil {
			return nil, err
		}

		err = authQuery.applyExpressionFilters(scope, input)
		if err != nil {
			return nil, err
		}
		authQuery.Select(IdField())
		authQuery.DistinctOn(IdField())
		rows, err := authQuery.SelectStatement().ExecuteToSingle(scope.Context)
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
		return nil, common.NewNotFoundError("")
	}

	id, ok := row["id"].(string)
	if !ok {
		return nil, errors.New("could not parse id key")
	}

	return &id, err
}

func GenerateDeleteStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, error) {
	err := query.ApplyImplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	err = query.applyExpressionFilters(scope, input)
	if err != nil {
		return nil, err
	}

	query.AppendReturning(IdField())

	return query.DeleteStatement(scope.Context), nil
}
