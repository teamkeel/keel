package actions

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
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

	Contains
	NotContains
	Equals
	NotEquals
	StartsWith
	EndsWith
	GreaterThan
	GreaterThanEquals
	LessThan
	LessThanEquals
	OneOf
	NotOneOf
	After
	Before
	OnOrAfter
	OnOrBefore

	AllEquals
	AnyEquals
	AllNotEquals
	AnyNotEquals
	AllGreaterThan
	AnyGreaterThan
	AllGreaterThanEquals
	AnyGreaterThanEquals
	AllLessThan
	AnyLessThan
	AllLessThanEquals
	AnyLessThanEquals
	AllAfter
	AnyAfter
	AllBefore
	AnyBefore
	AllOnOrAfter
	AnyOnOrAfter
	AllOnOrBefore
	AnyOnOrBefore

	Addition
	Subtraction
	Multiplication
	Division

	Not

	Or
	And
)

// queryOperatorToActionOperator converts the conditional operators that are used
// in GraphQL request input structures (such as "lessThanOrEquals") to its symbolic constant,
// machine-readable, ActionOperator value.
func queryOperatorToActionOperator(in string) (out ActionOperator, err error) {
	switch in {
	case "equals":
		return Equals, nil
	case "notEquals":
		return NotEquals, nil
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

func anyQueryOperationToActionOperator(in string) (out ActionOperator, err error) {
	switch in {
	case "equals":
		return AnyEquals, nil
	case "notEquals":
		return AnyNotEquals, nil
	case "lessThan":
		return AnyLessThan, nil
	case "lessThanOrEquals":
		return AnyLessThanEquals, nil
	case "greaterThan":
		return AnyGreaterThan, nil
	case "greaterThanOrEquals":
		return AnyGreaterThanEquals, nil
	case "before":
		return AnyBefore, nil
	case "after":
		return AnyAfter, nil
	case "onOrBefore":
		return AnyOnOrBefore, nil
	case "onOrAfter":
		return AnyOnOrAfter, nil
	default:
		return out, fmt.Errorf("unrecognized operator for any query: %s", in)
	}
}

func allQueryOperatorToActionOperator(in string) (out ActionOperator, err error) {
	switch in {
	case "equals":
		return AllEquals, nil
	case "notEquals":
		return AllNotEquals, nil
	case "lessThan":
		return AllLessThan, nil
	case "lessThanOrEquals":
		return AllLessThanEquals, nil
	case "greaterThan":
		return AllGreaterThan, nil
	case "greaterThanOrEquals":
		return AllGreaterThanEquals, nil
	case "before":
		return AllBefore, nil
	case "after":
		return AllAfter, nil
	case "onOrBefore":
		return AllOnOrBefore, nil
	case "onOrAfter":
		return AllOnOrAfter, nil
	default:
		return out, fmt.Errorf("unrecognized operator for all query: %s", in)
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

func toSql(o proto.OrderDirection) (string, error) {
	switch o {
	case proto.OrderDirection_ORDER_DIRECTION_ASCENDING:
		return "ASC", nil
	case proto.OrderDirection_ORDER_DIRECTION_DECENDING:
		return "DESC", nil
	default:
		return "", errors.New("cannot parse ORDER_DIRECTION_UNKNOWN")
	}
}
