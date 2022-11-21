package actions

import (
	"errors"
	"fmt"
	"strconv"

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
	query.AppendDistinctOn(Field("id"))
	query.AppendSelect(Field("*"))
	query.ApplyPaging(page)

	// Execute database request with results
	results, _, hasNextPage, err := query.
		SelectStatement().
		ExecuteWithResults(scope.context)

	if err != nil {
		return nil, err
	}

	return &ListResult{
		Results:     results,
		HasNextPage: hasNextPage,
	}, nil
}

// ParsePage extracts page mandate information from the given map and uses it to
// compose a Page.
func ParsePage(args map[string]any) (Page, error) {
	page := Page{}

	if first, ok := args["first"]; ok {
		asInt, ok := first.(int)
		if !ok {
			var err error
			asInt, err = strconv.Atoi(first.(string))
			if err != nil {
				return page, fmt.Errorf("cannot cast this: %v to an int", first)
			}
		}
		page.First = asInt
	}

	if last, ok := args["last"]; ok {
		asInt, ok := last.(int)
		if !ok {
			var err error
			asInt, err = strconv.Atoi(last.(string))
			if err != nil {
				return page, fmt.Errorf("cannot cast this: %v to an int", last)
			}
		}
		page.Last = asInt
	}

	if after, ok := args["after"]; ok {
		asString, ok := after.(string)
		if !ok {
			return page, fmt.Errorf("cannot cast this: %v to a string", after)
		}
		page.After = asString
	}

	if before, ok := args["before"]; ok {
		asString, ok := before.(string)
		if !ok {
			return page, fmt.Errorf("cannot cast this: %v to a string", before)
		}
		page.Before = asString
	}

	// If none specified - use a sensible default
	if page.First == 0 && page.Last == 0 {
		page = Page{First: 50}
	}

	return page, nil
}
