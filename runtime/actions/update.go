package actions

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
)

func Update(scope *Scope, input map[string]any) (res map[string]any, err error) {
	// Attempt to resolve permissions early; i.e. before row-based database querying.
	permissions := proto.PermissionsForAction(scope.Schema, scope.Action)
	canResolveEarly, authorised, err := TryResolveAuthorisationEarly(scope, permissions)
	if err != nil {
		return nil, err
	}

	if scope.Model.HasFiles() {
		// handle file uploads and change input values to file data if applicable
		if values, ok := input["values"].(map[string]any); ok {
			in, err := handleFileUploads(scope, values)
			if err != nil {
				return nil, fmt.Errorf("handling file uploads: %w", err)
			}
			input["values"] = in
		}
	}

	// Generate SQL statement
	query := NewQuery(scope.Model)
	statement, err := GenerateUpdateStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	switch {
	case canResolveEarly && !authorised:
		return nil, common.NewPermissionError()
	case !canResolveEarly:
		query.Select(IdField())
		query.DistinctOn(IdField())
		rowToAuthorise, err := query.SelectStatement().ExecuteToSingle(scope.Context)
		if err != nil {
			return nil, err
		}

		rowsToAuthorise := []map[string]any{}
		if rowToAuthorise != nil {
			rowsToAuthorise = append(rowsToAuthorise, rowToAuthorise)
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
	res, err = statement.ExecuteToSingle(scope.Context)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, common.NewNotFoundError("")
	}

	// if we have any files in our results we need to transform them to the object structure required
	if scope.Model.HasFiles() {
		res, err = transformModelFileResponses(scope.Context, scope.Model, res)
	}

	return res, err
}

func GenerateUpdateStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, error) {
	values, ok := input["values"].(map[string]any)
	if !ok {
		values = map[string]any{}
	}

	err := query.captureWriteValues(scope, values)
	if err != nil {
		return nil, err
	}

	err = query.captureSetValues(scope, values)
	if err != nil {
		return nil, err
	}

	where, ok := input["where"].(map[string]any)
	if !ok {
		where = map[string]any{}
	}

	err = query.applyImplicitFilters(scope, where)
	if err != nil {
		return nil, err
	}

	err = query.applyExpressionFilters(scope, where)
	if err != nil {
		return nil, err
	}

	// Return the updated row
	query.AppendReturning(AllFields())

	return query.UpdateStatement(scope.Context), nil
}
