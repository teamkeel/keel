package actions

import (
	"context"
	"errors"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
)

func Delete(scope *Scope, input map[string]any) (res *string, err error) {
	database, err := db.GetDatabase(scope.Context)
	if err != nil {
		return nil, err
	}

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
		err = common.NewPermissionError()
	case canResolveEarly && authorised:
		// Execute database request without starting a transaction or performing any row-based authorization
		row, err = statement.ExecuteToSingle(scope.Context)
	case !canResolveEarly:
		err = database.Transaction(scope.Context, func(ctx context.Context) error {
			scope := scope.WithContext(ctx)

			query.AppendSelect(IdField())
			query.AppendDistinctOn(IdField())
			rows, err := query.SelectStatement().ExecuteToSingle(scope.Context)
			if err != nil {
				return err
			}

			rowsToAuthorise := []map[string]any{}
			if rows != nil {
				rowsToAuthorise = append(rowsToAuthorise, rows)
			}

			isAuthorised, err := AuthoriseAction(scope, input, rowsToAuthorise)
			if err != nil {
				return err
			}

			if !isAuthorised {
				return common.NewPermissionError()
			}

			// Execute database request
			row, err = statement.ExecuteToSingle(scope.Context)
			if err != nil {
				return err
			}

			return nil
		})
	}

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
