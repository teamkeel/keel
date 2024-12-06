package attributes

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/cel-go/common/operators"
)

type match struct {
	Regex     string
	Construct func([]string) string
}

func ConvertMessage(message string) (string, error) {
	var matches = []match{
		unexpectedResolvedType,
		noOperatorOverload,
		undeclaredOperatorReference,
		undeclaredVariableReference,
	}

	for _, match := range matches {
		pattern, err := regexp.Compile(match.Regex)
		if err != nil {
			return "", err
		}
		if matches := pattern.FindStringSubmatch(message); matches != nil {
			return match.Construct(matches[1:]), nil
		}
	}

	return message, nil
}

// "expression expected to resolve to type list(_Role) but it is _Role
var unexpectedResolvedType = match{
	Regex: `expression expected to resolve to type (.+) but it is (.+)`,
	Construct: func(values []string) string {
		return fmt.Sprintf("expression expected to resolve to type %s but it is %s", mapType(values[0]), mapType(values[1]))
	},
}

var noOperatorOverload = match{
	Regex: `found no matching overload for '(.+)' applied to '\((.+),\s*(.+)\)'`,
	Construct: func(values []string) string {
		return fmt.Sprintf("cannot use operator '%s' with types %s and %s", mapOperator(values[0]), mapType(values[1]), mapType(values[2]))
	},
}

var undeclaredOperatorReference = match{
	Regex: `undeclared reference to '_(.+)_' \(in container ''\)`,
	Construct: func(values []string) string {
		return fmt.Sprintf("operator '%s' not supported in this context", mapOperator(values[0]))
	},
}

var undeclaredVariableReference = match{
	Regex: `undeclared reference to '(.+)' \(in container ''\)`,
	Construct: func(values []string) string {
		return fmt.Sprintf("unknown identifier '%s'", values[0])
	},
}

func mapOperator(op string) string {
	switch op {
	case operators.In:
		return "in"
	default:
		v := strings.Trim(op, "_")

		if v == "&&" {
			v = "and"
		} else if v == "||" {
			v = "or"
		}

		return v
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
	default:
		// Enum or Model name
		keelType = strings.TrimPrefix(t, "_")
	}

	if isArray {
		keelType += "[]"
	}

	return keelType
}
