package actions

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

type Operator int

const (
	OperatorEquals   = iota
	OperatorLessThan = iota
	OperatorContains = iota
)
