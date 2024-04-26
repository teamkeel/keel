package actions

import (
	"fmt"
	"strconv"
)

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

// ParsePage extracts page mandate information from the given map and uses it to
// compose a Page.
func ParsePage(args map[string]any) (Page, error) {
	page := Page{}

	if first, ok := args["first"]; ok {
		switch v := first.(type) {
		case int64:
			page.First = int(v)
		case int:
			page.First = v
		case float64:
			page.First = int(v)
		case string:
			num, err := strconv.Atoi(v)

			if err == nil {
				page.First = num
			}
		}
	}

	if last, ok := args["last"]; ok {
		switch v := last.(type) {
		case int64:
			page.Last = int(v)
		case float64:
			page.Last = int(v)
		case int:
			page.Last = v
		case string:
			num, err := strconv.Atoi(v)

			if err == nil {
				page.Last = num
			}
		}
	}

	// If none specified - use a sensible default
	if page.First == 0 && page.Last == 0 {
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
