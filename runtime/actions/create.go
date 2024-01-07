package actions

import (
	"context"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
)

func Create(scope *Scope, input map[string]any) (res map[string]any, err error) {
	database, err := db.GetDatabase(scope.Context)
	if err != nil {
		return nil, err
	}

	permissions := proto.PermissionsForAction(scope.Schema, scope.Action)

	// Attempt to resolve permissions early; i.e. before row-based database querying.
	canResolveEarly, authorised, err := TryResolveAuthorisationEarly(scope, permissions)
	if err != nil {
		return nil, err
	}

	// Generate the SQL statement
	query := NewQuery(scope.Model)
	statement, err := GenerateCreateStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	switch {
	case canResolveEarly && !authorised:
		err = common.NewPermissionError()
	case canResolveEarly && authorised:
		// Execute database request without starting a transaction or performing any row-based authorization
		res, err = statement.ExecuteToSingle(scope.Context)
	case !canResolveEarly:
		err = database.Transaction(scope.Context, func(ctx context.Context) error {
			scope := scope.WithContext(ctx)

			// Execute database request, expecting a single result
			res, err = statement.ExecuteToSingle(scope.Context)
			if err != nil {
				return err
			}

			isAuthorised, err := AuthoriseAction(scope, input, []map[string]any{res})
			if err != nil {
				return err
			}

			if !isAuthorised {
				return common.NewPermissionError()
			}

			return nil
		})
	}

	return res, err
}

func GenerateCreateStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, error) {
	err := query.captureWriteValues(scope, input)
	if err != nil {
		return nil, err
	}

	err = query.captureSetValues(scope, input)
	if err != nil {
		return nil, err
	}

	// Return the inserted row
	query.AppendReturning(AllFields())

	return query.InsertStatement(scope.Context), nil
}
