package actions

import (
	"github.com/teamkeel/keel/runtime/common"
)

func Delete(scope *Scope, input map[string]any) (*string, error) {
	query := NewQuery(scope.Model)

	// Generate the SQL statement
	statement, err := GenerateDeleteStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	query.AppendSelect(IdField())
	query.AppendDistinctOn(IdField())
	rowToAuthorise, err := query.SelectStatement().ExecuteToSingle(scope.Context)
	if err != nil {
		return nil, err
	}

	isAuthorised, err := AuthoriseAction(scope, []map[string]any{rowToAuthorise})
	if err != nil {
		return nil, err
	}

	if !isAuthorised {
		return nil, common.NewPermissionError()
	}

	// Execute database request
	row, err := statement.ExecuteToSingle(scope.Context)

	// TODO: if the error is multiple rows affected then rollback transaction
	if err != nil {
		return nil, err
	}

	if row == nil {
		return nil, common.NewNotFoundError()
	}

	id, _ := row["id"].(string)
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

	return query.DeleteStatement(), nil
}