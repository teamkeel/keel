package actions

import (
	"fmt"
	"strconv"
)

// A Page describes which page you want from a list of records,
// in the style of this "Connection" pattern:
// https://relay.dev/graphql/connections.htm
//
// It can handle both cursor and offset pagination. The default is cursor pagination, but if Limit is set to > 0,
// then Offset pagination will be used
//
// Cursor pagination:
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
	Offset int
	Limit  int
}

// ParsePage extracts page mandate information from the given map and uses it to
// compose a Page.
func ParsePage(args map[string]any) (Page, error) {
	page := Page{}

	if arg, ok := extractIntArg(args, "first"); ok {
		page.First = arg
	}
	if arg, ok := extractIntArg(args, "last"); ok {
		page.Last = arg
	}
	if arg, ok := extractIntArg(args, "offset"); ok {
		if arg > 0 {
			page.Offset = arg
		}
	}
	if arg, ok := extractIntArg(args, "limit"); ok {
		page.Limit = arg
	}

	// If none specified - use a sensible default for cursor pagination
	if !page.OffsetPagination() && page.First == 0 && page.Last == 0 {
		page.First = 50
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

	return page, nil
}

// extractIntArg checks if the given map contains a int value at the key defined by the name param
func extractIntArg(args map[string]any, name string) (int, bool) {
	if arg, ok := args[name]; ok {
		switch v := arg.(type) {
		case int64:
			return int(v), true
		case int:
			return v, true
		case float64:
			return int(v), true
		case string:
			num, err := strconv.Atoi(v)

			if err == nil {
				return num, true
			}
		}
	}

	return 0, false
}

// OffsetPagination tells us if the page is to use offset pagination
func (p *Page) OffsetPagination() bool {
	return p.Limit > 0
}

// IsBackwards tells us if the page is backwards paginated (e.g. we're requesting elements before a cursor)
func (p *Page) IsBackwards() bool {
	return p.Before != "" && p.Last > 0
}

// Cursor returns the cursor used in the pagination based on the direction of pagination:
// - if backwards, it's `before`
// - if forward pagination, it's `after`
func (p *Page) Cursor() string {
	if p.IsBackwards() {
		return p.Before
	}

	return p.After
}

// GetLimit returns the page limit to be used in the SQL query
func (p *Page) GetLimit() int {
	if p.OffsetPagination() {
		return p.Limit
	}
	if p.IsBackwards() {
		return p.Last
	}

	return p.First
}

// PageNumber returns the number of the current; it is only applicable for OffsetPagination
func (p *Page) PageNumber() *int {
	if p == nil || !p.OffsetPagination() {
		return nil
	}

	pNum := 1
	if p.Offset > 0 {
		pNum += p.Offset / p.Limit
	}

	return &pNum
}
