package actions

import (
	"fmt"
	"strconv"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

type ListAction struct {
	*Action[ListResult]
}

type ListResult struct {
	Collection  []map[string]any `json:"collection"`
	HasNextPage bool             `json:"hasNextPage"`
}

func (action *ListAction) Initialise(scope *Scope) ActionBuilder[ListResult] {
	action.Action = &Action[ListResult]{
		Scope: scope,
	}
	return action
}

func (action *ListAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[ListResult] {
	if action.HasError() {
		return action
	}

	allOptional := lo.EveryBy(action.operation.Inputs, func(input *proto.OperationInput) bool {
		return input.Optional
	})

inputs:
	for _, input := range action.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		fieldName := input.Target[0]

		whereInputs, ok := args["where"]
		if !ok {
			// We have some required inputs but there is no where key
			if !allOptional {
				return action.WithError(fmt.Errorf("arguments map does not contain a where key: %v", args))
			}
		} else {
			whereInputsAsMap, ok := whereInputs.(map[string]any)
			if !ok {
				return action.WithError(fmt.Errorf("cannot cast this: %v to a map[string]any", whereInputs))
			}

			value, ok := whereInputsAsMap[fieldName]

			if !ok {
				if input.Optional {
					// do not do any further processing if the input is not a map
					// as it is likely nil
					continue inputs
				}

				return action.WithError(fmt.Errorf("cannot cast this: %v to a map[string]any", value))
			}

			valueMap, ok := value.(map[string]any)

			if !ok {
				if input.Optional {
					// do not do any further processing if the input is not a map
					// as it is likely nil
					continue inputs
				}

				return action.WithError(fmt.Errorf("cannot cast this: %v to a map[string]any", value))
			}

			for operatorStr, operand := range valueMap {
				operatorName, err := operator(operatorStr) // { "rating": { "greaterThanOrEquals": 1 } }
				if err != nil {
					return action.WithError(err)
				}

				action.addImplicitFilter(input, operatorName, operand)
			}
		}
	}

	return action
}

func (action *ListAction) Execute(args RequestArguments) (*ActionResult[ListResult], error) {
	// how do we access original args?
	// simple: add

	if action.HasError() {
		return nil, action.curError
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
	action.query = action.query.Select(selectArgs, by)
	action.query = action.query.Order(by)

	// A Where clause to implement the after/before paging request
	switch {
	case page.After != "":
		action.query = action.query.Where("ID > ?", page.After)
	case page.Before != "":
		action.query = action.query.Where("ID < ?", page.Before)
	}

	switch {
	case page.First != 0:
		action.query = action.query.Limit(page.First)
	case page.Last != 0:
		action.query = action.query.Limit(page.Last)
	}

	// Execute the query
	result := []map[string]any{}
	action.query = action.query.Find(&result)
	if action.query.Error != nil {
		return nil, action.query.Error
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
	actionResult := ActionResult[ListResult]{
		Value: ListResult{
			Collection:  collection,
			HasNextPage: hasNextPage,
		},
	}

	return &actionResult, nil
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
