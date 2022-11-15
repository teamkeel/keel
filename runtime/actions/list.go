package actions

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

type ListAction struct {
	scope *Scope
}

type ListResult struct {
	Results     []map[string]any `json:"results"`
	HasNextPage bool             `json:"hasNextPage"`
}

func (action *ListAction) Initialise(scope *Scope) ActionBuilder[ListResult] {
	action.scope = scope
	return action
}

func (action *ListAction) CaptureImplicitWriteInputValues(args ValueArgs) ActionBuilder[ListResult] {
	return action // no-op
}

func (action *ListAction) CaptureSetValues(args ValueArgs) ActionBuilder[ListResult] {
	return action // no-op
}

func (action *ListAction) IsAuthorised(args WhereArgs) ActionBuilder[ListResult] {
	if action.scope.Error != nil {
		return action
	}

	isAuthorised, err := DefaultIsAuthorised(action.scope, args)

	if err != nil {
		action.scope.Error = err
		return action
	}

	if !isAuthorised {
		action.scope.Error = errors.New("not authorized to access this operation")
	}

	return action
}

func (action *ListAction) ApplyImplicitFilters(args WhereArgs) ActionBuilder[ListResult] {
	if action.scope.Error != nil {
		return action
	}

	allJoins := []string{}

inputs:
	for _, input := range action.scope.operation.Inputs {
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
			operator, err := graphQlOperatorToActionOperator(operatorStr)
			if err != nil {
				action.scope.Error = err
				return action
			}

			// New filter resolver to generate a database query statement
			resolver := NewImplicitFilterResolverResolver(action.scope)

			// Resolve the database statement for this expression
			statement, joins, err := resolver.ResolveQueryStatement(input, fieldName, operand, operator)
			if err != nil {
				action.scope.Error = err
				return action
			}

			allJoins = append(allJoins, joins...)

			action.scope.query = action.scope.query.
				WithContext(action.scope.context).
				Where(statement)
		}
	}

	allJoins = lo.Uniq(allJoins)
	action.scope.query = action.scope.query.Joins(strings.Join(allJoins, " "))

	return action
}

func (action *ListAction) ApplyExplicitFilters(args WhereArgs) ActionBuilder[ListResult] {
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

func (action *ListAction) Execute(args WhereArgs) (*ActionResult[ListResult], error) {
	if action.scope.Error != nil {
		return nil, action.scope.Error
	}
	op := action.scope.operation

	if op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		return ParseListResponse(action.scope.context, op, args)
	}
	// We update the query to implement the paging request

	page, err := parsePage(args)
	if err != nil {
		return nil, err
	}

	// Specify the ORDER BY - but also a "LEAD" extra column to harvest extra data
	// that helps to determine "hasNextPage".
	by := fmt.Sprintf("%s.id", strcase.ToSnake(action.scope.model.Name))

	selectArgs := `DISTINCT ON (%[1]s.id) 
		%[1]s.*,
		CASE WHEN lead(%[1]s.id) OVER ( order by %[1]s.id ) is not null THEN true ELSE false END as hasNext
		`
	selectArgs = fmt.Sprintf(selectArgs, strcase.ToSnake(action.scope.model.Name))

	action.scope.query = action.scope.query.WithContext(action.scope.context).Select(selectArgs, by)
	action.scope.query = action.scope.query.WithContext(action.scope.context).Order(by)

	// A Where clause to implement the after/before paging request
	switch {
	case page.After != "":
		action.scope.query = action.scope.query.WithContext(action.scope.context).Where("ID > ?", page.After)
	case page.Before != "":
		action.scope.query = action.scope.query.WithContext(action.scope.context).Where("ID < ?", page.Before)
	}

	switch {
	case page.First != 0:
		action.scope.query = action.scope.query.WithContext(action.scope.context).Limit(page.First)
	case page.Last != 0:
		action.scope.query = action.scope.query.WithContext(action.scope.context).Limit(page.Last)
	}

	// Execute the query
	result := []map[string]any{}
	action.scope.query = action.scope.query.WithContext(action.scope.context).Find(&result)
	if action.scope.query.WithContext(action.scope.context).Error != nil {
		return nil, action.scope.query.WithContext(action.scope.context).Error
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
			Results:     collection,
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

// A Page describes which page you want from a list of records,
// in the style of this "Connection" pattern:
// https://relay.dev/graphql/connections.htm
//
// Consider for example, that you previously fetched a page of 10 records
// and from that previous response you also knew that the last of those 10 records
// could be referred to with the opaque cursor "abc123". Armed with that information you can
// ask for the next page of 10 records by setting First to 10, and After to "abc123".
//
// To move backwards, you'd set the Last and Before fields instead.
//
// When you have no prior positional context you should specify First but leave Before and After to
// the empty string. This gives you the first N records.
type Page struct {
	First  int
	Last   int
	After  string
	Before string
}
