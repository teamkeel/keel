package actions

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
)

type ListResult struct {
	Results     []map[string]any `json:"results"`
	HasNextPage bool             `json:"hasNextPage"`
}

func (query *QueryBuilder) applyImplicitFiltersForList(scope *Scope, args WhereArgs) error {
inputs:
	for _, input := range scope.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
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

			return fmt.Errorf("'%s' input value %v to not in correct format", fieldName, value)
		}

		for operatorStr, operand := range valueMap {
			operator, err := graphQlOperatorToActionOperator(operatorStr)
			if err != nil {
				return err
			}

			// Resolve the database statement for this expression
			err = query.whereByImplicitFilter(scope, input, fieldName, operator, operand)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func List(scope *Scope, input map[string]any) (*ListResult, error) {
	where, ok := input["where"].(map[string]any)
	if !ok {
		where = map[string]any{}
	}

	query := NewQuery(scope.model)

	err := query.applyImplicitFiltersForList(scope, where)
	if err != nil {
		return nil, err
	}

	err = query.applyExplicitFilters(scope, where)
	if err != nil {
		return nil, err
	}

	isAuthorised, err := query.isAuthorised(scope, where)
	if err != nil {
		return nil, err
	}

	if !isAuthorised {
		return nil, errors.New("not authorized to access this operation")
	}

	op := scope.operation
	if scope.operation.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		// TODO: the custom function should receive the whole input, not just the
		// where's
		return ParseListResponse(scope.context, op, where)
	}

	page, err := ParsePage(input)
	if err != nil {
		return nil, err
	}

	// Select all columns from this table and distinct on id
	query.AppendDistinctOn(IdField())
	query.AppendSelect(AllFields())
	query.ApplyPaging(page)

	// Execute database request with results
	results, _, hasNextPage, err := query.
		SelectStatement().
		ExecuteToMany(scope.context)

	if err != nil {
		return nil, err
	}

	return &ListResult{
		Results:     results,
		HasNextPage: hasNextPage,
	}, nil
}
