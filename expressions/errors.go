package expressions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/types"
)

// converters which translate cel errors into better keel validation messages
var messageConverters = []errorConverter{
	undefinedField,
	noFieldSelection,
	noOperatorOverload,
	undeclaredOperatorReference,
	undeclaredVariableReference,
	unrecognisedToken,
	mismatchedInput,
}

type errorConverter struct {
	Regex     string
	Construct func(expectedReturnType *types.Type, values []string) string
}

var undefinedField = errorConverter{
	Regex: `undefined field '(.+)'`,
	Construct: func(expectedReturnType *types.Type, values []string) string {
		return fmt.Sprintf("field '%s' does not exist", values[0])
	},
}

var noFieldSelection = errorConverter{
	Regex: `type '(.+)' does not support field selection`,
	Construct: func(expectedReturnType *types.Type, values []string) string {
		return fmt.Sprintf("type %s does not have any fields to select", mapType(values[0]))
	},
}

var noOperatorOverload = errorConverter{
	Regex: `found no matching overload for '(.+)' applied to '\((.+),\s*(.+)\)'`,
	Construct: func(expectedReturnType *types.Type, values []string) string {
		return fmt.Sprintf("cannot use operator '%s' with types %s and %s", mapOperator(values[0]), mapType(values[1]), mapType(values[2]))
	},
}

var undeclaredOperatorReference = errorConverter{
	Regex: `undeclared reference to '_(.+)_' \(in container ''\)`,
	Construct: func(expectedReturnType *types.Type, values []string) string {
		return fmt.Sprintf("operator '%s' not supported in this context", mapOperator(values[0]))
	},
}

var undeclaredVariableReference = errorConverter{
	Regex: `undeclared reference to '(.+)' \(in container ''\)`,
	Construct: func(expectedReturnType *types.Type, values []string) string {
		switch {
		case expectedReturnType.String() == "_Role" || expectedReturnType.String() == "_Role[]":
			return fmt.Sprintf("%s is not a role defined in your schema", values[0])
		}
		return fmt.Sprintf("unknown identifier '%s'", values[0])
	},
}

var unrecognisedToken = errorConverter{
	Regex: `Syntax error: token recognition error at: '(.+)'`,
	Construct: func(expectedReturnType *types.Type, values []string) string {
		if values[0] == "= " {
			return "assignment operator '=' not valid - did you mean to use the comparison operator '=='?"
		}
		return fmt.Sprintf("invalid character(s) '%s' in expression", strings.Trim(values[0], " "))
	},
}

var mismatchedInput = errorConverter{
	Regex: `Syntax error: mismatched input '(.+)' expecting (.+)`,
	Construct: func(expectedReturnType *types.Type, values []string) string {
		return fmt.Sprintf("unknown or unsupported identifier or operator '%s' in expression", values[0])
	},
}

func mapOperator(op string) string {
	switch op {
	case operators.In:
		return "in"
	default:
		return strings.Trim(op, "_")
	}
}

func mapType(t string) string {
	isArray := false

	pattern := regexp.MustCompile(`list\((.+)\)`)
	if matches := pattern.FindStringSubmatch(t); matches != nil {
		isArray = true
		t = matches[1]
	}

	var keelType string
	switch t {
	case "string":
		keelType = "Text"
	case "int":
		keelType = "Number"
	case "double":
		keelType = "Decimal"
	case "bool":
		keelType = "Boolean"
	case "timestamp":
		keelType = "Timestamp"
	default:
		// Enum or Model name
		keelType = strings.TrimPrefix(t, "_")
	}

	if isArray {
		keelType += "[]"
	}

	return keelType
}
