package actions

import "fmt"

// A ListInput defines part of an input to a LIST action. Specifically,
// A filter - in terms of Where clauses, and also a mandate about which
// page from the potential results is required.
type ListInput struct {
	Page   Page
	Wheres []*Where
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

// A Where specifies a filter for one column in a database row.
// You specify the field of interest with Name.
// And then you populate just the relevant one of XXXQuery attributes - according
// to the type of the column.
type Where struct {
	Name     string
	Operator Operator
	Operand  any
}

type Operator string

const (
	OperatorUnknown = "unknown"
	OperatorEquals  = "equal"

	// String
	OperatorStartsWith = "startsWith"
	OperatorEndsWith   = "endsWith"
	OperatorContains   = "contains"
	OperatorOneOf      = "oneOf"

	// Numberic
	OperatorLessThan          = "lessThan"
	OperatorLessThanEquals    = "lessThanOrEqualTo"
	OperatorGreaterThan       = "greaterThan"
	OperatorGreaterThanEquals = "greaterThanOrEqualTo"

	// Date
	OperatorBefore     = "before"
	OperatorAfter      = "after"
	OperatorOnOrBefore = "onOrBefore"
	OperatorOnOrAfter  = "onOrAfter"
)

// operator converts the given string representation of an operator like
// "eq" into the corresponding Operator value.
func operator(operatorStr string) (op Operator, err error) {
	switch operatorStr {
	case "equals":
		return OperatorEquals, nil
	case "startsWith":
		return OperatorStartsWith, nil
	case "endsWith":
		return OperatorEndsWith, nil
	case "contains":
		return OperatorContains, nil
	case "oneOf":
		return OperatorOneOf, nil
	case "lessThan":
		return OperatorLessThan, nil
	case "lessThanOrEquals":
		return OperatorLessThanEquals, nil
	case "greaterThan":
		return OperatorGreaterThan, nil
	case "greaterThanOrEquals":
		return OperatorGreaterThanEquals, nil
	case "before":
		return OperatorBefore, nil
	case "after":
		return OperatorAfter, nil
	case "onOrBefore":
		return OperatorOnOrAfter, nil
	case "onOrAfter":
		return OperatorOnOrBefore, nil
	default:
		return op, fmt.Errorf("unrecognized operator: %s", operatorStr)
	}
}
