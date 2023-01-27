package actions

import (
	"fmt"

	"github.com/teamkeel/keel/schema/parser"
)

// An ActionOperator gives a symbolic, machine-readable name to each
// of the comparison operators that Keel Actions work with at a CONCEPTUAL
// level.
//
// By design, the ActionOperator has no knowledge (in of itself) of how these
// might be expressed in schema's or in request inputs, or in expressions for
// example.
type ActionOperator int

const (
	Unknown ActionOperator = iota

	After
	Before
	Contains
	Equals
	EndsWith
	GreaterThan
	GreaterThanEquals
	LessThan
	LessThanEquals
	NotContains
	NotEquals
	NotOneOf
	OneOf
	OnOrAfter
	OnOrBefore
	StartsWith
)

// graphQlOperatorToActionOperator converts the conditional operators that are used
// in GraphQL request input structures (such as "lessThanOrEquals") to its symbolic constant,
// machine-readable, ActionOperator value.
func graphQlOperatorToActionOperator(in string) (out ActionOperator, err error) {
	switch in {
	case "equals":
		return Equals, nil
	case "startsWith":
		return StartsWith, nil
	case "endsWith":
		return EndsWith, nil
	case "contains":
		return Contains, nil
	case "oneOf":
		return OneOf, nil
	case "lessThan":
		return LessThan, nil
	case "lessThanOrEquals":
		return LessThanEquals, nil
	case "greaterThan":
		return GreaterThan, nil
	case "greaterThanOrEquals":
		return GreaterThanEquals, nil
	case "before":
		return Before, nil
	case "after":
		return After, nil
	case "onOrBefore":
		return OnOrBefore, nil
	case "onOrAfter":
		return OnOrAfter, nil
	default:
		return out, fmt.Errorf("unrecognized operator: %s", in)
	}
}

// expressionOperatorToActionOperator converts the conditional operators that are used
// in Keel Expressions (such as ">=") to its symbolic constant,
// machine-readable, ActionOperator value.
func expressionOperatorToActionOperator(in string) (out ActionOperator, err error) {
	switch in {
	case parser.OperatorEquals:
		return Equals, nil
	case parser.OperatorNotEquals:
		return NotEquals, nil
	case parser.OperatorGreaterThanOrEqualTo:
		return GreaterThanEquals, nil
	case parser.OperatorLessThanOrEqualTo:
		return LessThanEquals, nil
	case parser.OperatorLessThan:
		return LessThan, nil
	case parser.OperatorGreaterThan:
		return GreaterThan, nil
	case parser.OperatorIn:
		return OneOf, nil
	case parser.OperatorNotIn:
		return NotOneOf, nil

	default:
		return Unknown, fmt.Errorf("this is not a recognized conditional operator: %s", in)
	}
}
