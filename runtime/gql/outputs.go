package gql

import (
	"time"
)

// Output is a type you can embed in a structure in order to make
// that structure implement graphql.Output
type Output struct{}

func (op Output) Name() string {
	return "Name() return value"
}

func (op Output) String() string {
	return "String() return value"
}

func (op Output) Error() error {
	return nil
}

func (op Output) Description() string {
	return "Description return value"
}

// DataTimeAsGQLOutput wraps a time.time such that it implement the graphql.Output interface.
type DateTimeAsGQLOutput struct {
	Output
	DateTime time.Time
}

// SimpleMapAsGQLOutput wraps a simple map such that it implement the graphql.Output interface.
type SimpleMapAsGQLOutput struct {
	Output
	Map map[string]any
}
