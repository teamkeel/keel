package actions

import (
	"github.com/teamkeel/keel/runtime/common"
)

func Delete(scope *Scope, input map[string]any) (*string, error) {
	query := NewQuery(scope.model)

	// Generate the SQL statement
	statement, err := GenerateDeleteStatement(query, scope, input)
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

	// Execute database request
	row, err := statement.ExecuteToSingle(scope.context)

	// TODO: if the error is multiple rows affected then rollback transaction
	if err != nil {
		return nil, err
	}

	if row == nil {
		return nil, common.NewNotFoundError()
	}

	id, _ := row["id"].(string)
	return &id, nil
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
