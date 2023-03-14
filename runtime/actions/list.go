package actions

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
)

func (query *QueryBuilder) applyImplicitFiltersForList(scope *Scope, args map[string]any) error {
	message := proto.FindWhereInputMessage(scope.schema, scope.operation.Name)
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
	query := NewQuery(scope.model)

	// Generate the SQL statement.
	statement, page, err := GenerateListStatement(query, scope, input)
	if err != nil {
		return nil, err
	}

	// Execute database request with results
	results, _, hasNextPage, err := statement.ExecuteToMany(scope.context)

	if err != nil {
		return nil, err
	}

	if page.Last != 0 {
		results = lo.Reverse(results)
	}

	var startCursor string
	var endCursor string

	for i, record := range results {
		if i == 0 {
			startCursor, _ = record["id"].(string)
		} else if i == len(results)-1 {
			endCursor, _ = record["id"].(string)
		}
	}

	return map[string]any{
		"results":     results,
		"hasNextPage": hasNextPage,
		"startCursor": startCursor,
		"endCursor":   endCursor,
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

	isAuthorised, err := query.isAuthorised(scope, where)
	if err != nil {
		return nil, nil, err
	}

	if !isAuthorised {
		return nil, nil, common.RuntimeError{Code: common.ErrPermissionDenied, Message: "not authorized to access this operation"}
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
