package actions

import "github.com/teamkeel/keel/runtime/common"

func Get(scope *Scope, input map[string]any) (map[string]any, error) {
	query := NewQuery(scope.Context, scope.Model)

	// Generate the SQL statement
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

	err = query.applyExplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	// Select all columns and distinct on id
	query.AppendSelect(AllFields())
	query.AppendDistinctOn(IdField())

	return query.SelectStatement(), nil
}
