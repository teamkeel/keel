package actions

import (
	"fmt"
	"strconv"

	"github.com/teamkeel/keel/proto"
)

type ListAction struct {
	scope *Scope
}

type ListResult struct {
	Collection  []map[string]any `json:"collection"`
	HasNextPage bool             `json:"hasNextPage"`
}

func (action *ListAction) Initialise(scope *Scope) ActionBuilder[ListResult] {
	action.scope = scope
	return action
}

// Keep the no-op methods in a group together

func (action *ListAction) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder[ListResult] {
	return action // no-op
}

func (action *ListAction) CaptureSetValues(args RequestArguments) ActionBuilder[ListResult] {
	return action // no-op
}

func (action *ListAction) IsAuthorised(args RequestArguments) ActionBuilder[ListResult] {
	return action
}

// ----------------

func (action *ListAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[ListResult] {
	if action.scope.Error != nil {
		return action
	}

inputs:
	for _, input := range action.scope.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		fieldName := input.Target[0]
		value, ok := args[fieldName]

		// not found
		if !ok {
			if input.Optional {
				continue inputs
			}

			action.scope.Error = fmt.Errorf("did not find required '%s' input in where clause", fieldName)
		}

		valueMap, ok := value.(map[string]any)

		if !ok {
			if input.Optional {
				// do not do any further processing if the input is not a map
				// as it is likely nil
				continue inputs
			}

			action.scope.Error = fmt.Errorf("'%s' input value %v to not in correct format", fieldName, value)
			return action
		}

		for operatorStr, operand := range valueMap {
			operatorName, err := operator(operatorStr) // { "rating": { "greaterThanOrEquals": 1 } }
			if err != nil {
				action.scope.Error = err
				return action
			}

			addImplicitFilter(action.scope, input, operatorName, operand)
		}
	}

	return action
}

func (action *ListAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[ListResult] {
	if action.scope.Error != nil {
		return action
	}
	// We delegate to a function that may get used by other Actions later on, once we have
	// unified how we handle operators in both schema where clauses and in implicit inputs language.
	err := DefaultApplyExplicitFilters(action.scope, args)
	if err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *ListAction) Execute(args RequestArguments) (*ActionResult[ListResult], error) {
	if action.scope.Error != nil {
		return nil, action.scope.Error
	}

	// We update the query to implement the paging request

	page, err := parsePage(args)
	if err != nil {
		return nil, err
	}

	// Specify the ORDER BY - but also a "LEAD" extra column to harvest extra data
	// that helps to determine "hasNextPage".
	const by = "id"
	selectArgs := `
		 *,
			CASE WHEN lead("id") OVER ( order by ? ) is not null THEN true ELSE false
			END as hasNext
		`
	action.scope.query = action.scope.query.Select(selectArgs, by)
	action.scope.query = action.scope.query.Order(by)

	// A Where clause to implement the after/before paging request
	switch {
	case page.After != "":
		action.scope.query = action.scope.query.Where("ID > ?", page.After)
	case page.Before != "":
		action.scope.query = action.scope.query.Where("ID < ?", page.Before)
	}

	switch {
	case page.First != 0:
		action.scope.query = action.scope.query.Limit(page.First)
	case page.Last != 0:
		action.scope.query = action.scope.query.Limit(page.Last)
	}

	// Execute the query
	result := []map[string]any{}
	action.scope.query = action.scope.query.Find(&result)
	if action.scope.query.Error != nil {
		return nil, action.scope.query.Error
	}

	// Sort out the hasNextPage value, and clean up the response.
	hasNextPage := false
	if len(result) > 0 {
		last := result[len(result)-1]
		hasNextPage = last["hasnext"].(bool)
	}

	for _, row := range result {
		delete(row, "has_next")
	}
	collection := toLowerCamelMaps(result)

	return &ActionResult[ListResult]{
		Value: ListResult{
			Collection:  collection,
			HasNextPage: hasNextPage,
		},
	}, nil
}

// parsePage extracts page mandate information from the given map and uses it to
// compose a Page.
func parsePage(args map[string]any) (Page, error) {
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
