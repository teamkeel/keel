package actions

// A ListInput defines the input to a LIST action. Specifically,
// A filter - in terms of Where clauses, and also a mandate about which
// page from the potential results is required.
type ListInput struct {
	Page Page
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

// An Operator represents one of the built-in operators you can use in a filter for
// implicit inputs.
// type Operator string

// const (
// 	OperatorUnknown = "unknown"
// 	OperatorEquals  = "equal"

// 	// String
// 	OperatorStartsWith = "startsWith"
// 	OperatorEndsWith   = "endsWith"
// 	OperatorContains   = "contains"
// 	OperatorOneOf      = "oneOf"

// 	// Numeric
// 	OperatorLessThan          = "lessThan"
// 	OperatorLessThanEquals    = "lessThanOrEqualTo"
// 	OperatorGreaterThan       = "greaterThan"
// 	OperatorGreaterThanEquals = "greaterThanOrEqualTo"

// 	// Date
// 	OperatorBefore     = "before"
// 	OperatorAfter      = "after"
// 	OperatorOnOrBefore = "onOrBefore"
// 	OperatorOnOrAfter  = "onOrAfter"
// )

// Contains a mapping from graphql operator names
// to SQL operators
// var operatorsMap = map[Operator]string{
// 	OperatorEquals:            "=",
// 	OperatorOneOf:             "IN",
// 	OperatorStartsWith:        "LIKE",
// 	OperatorEndsWith:          "LIKE",
// 	OperatorLessThan:          "<",
// 	OperatorLessThanEquals:    "<=",
// 	OperatorGreaterThan:       ">",
// 	OperatorGreaterThanEquals: ">=",
// 	OperatorBefore:            "<",
// 	OperatorOnOrBefore:        "<=",
// 	OperatorAfter:             ">",
// 	OperatorOnOrAfter:         ">=",
// }
