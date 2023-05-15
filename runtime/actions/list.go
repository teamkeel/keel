package actions

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
)

func (query *QueryBuilder) applyImplicitFiltersForList(scope *Scope, args map[string]any) error {
	message := proto.FindWhereInputMessage(scope.Schema, scope.Operation.Name)
	if message == nil {
		return nil
	}

inputs:
	for _, input := range message.Fields {
		if !input.IsModelField() {
			continue
		}

		fieldName := input.Name
		value, ok := args[fieldName]

		// not found
		if !ok {
			if input.Optional {
				continue inputs
			}
			return fmt.Errorf("did not find required '%s' input in where clause", fieldName)
		}

		valueMap, ok := value.(map[string]any)

		if !ok {
			if input.Optional {
				// do not do any further processing if the input is not a map
				// as it is likely nil
				continue inputs
			}
			return fmt.Errorf("'%s' input value %v is not in correct format", fieldName, value)
		}

		for operatorStr, operand := range valueMap {
			operator, err := graphQlOperatorToActionOperator(operatorStr)
			if err != nil {
				return err
			}

			// Resolve the database statement for this expression
			err = query.whereByImplicitFilter(scope, input.Target, fieldName, operator, operand)
			if err != nil {
				return err
			}

			// Implicit input conditions are ANDed together
			query.And()
		}

	}

	return nil
}

func List(scope *Scope, input map[string]any) (map[string]any, error) {
	query := NewQuery(scope.Model)

	// Generate the SQL statement.
	statement, page, err := GenerateListStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	// Execute database request with results
	results, pageInfo, err := statement.ExecuteToMany(scope.Context, page)
	if err != nil {
		return nil, err
	}

	isAuthorised, err := AuthoriseAction(scope, results)
	if err != nil {
		return nil, err
	}

	if !isAuthorised {
		return nil, common.NewPermissionError()
	}

	return map[string]any{
		"results":  results,
		"pageInfo": pageInfo.ToMap(),
	}, nil
}

func GenerateListStatement(query *QueryBuilder, scope *Scope, input map[string]any) (*Statement, *Page, error) {
	where, ok := input["where"].(map[string]any)
	if !ok {
		where = map[string]any{}
	}

	err := query.applyImplicitFiltersForList(scope, where)
	if err != nil {
		return nil, nil, err
	}

	err = query.applyExplicitFilters(scope, where)
	if err != nil {
		return nil, nil, err
	}

	page, err := ParsePage(input)
	if err != nil {
		return nil, nil, err
	}

	// Select all columns from this table and distinct on id
	query.AppendDistinctOn(IdField())
	query.AppendSelect(AllFields())
	err = query.ApplyPaging(page)
	if err != nil {
		return nil, &page, err
	}

	return query.SelectStatement(), &page, nil
}
