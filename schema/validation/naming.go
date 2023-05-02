package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// NamingRule checks that entities in the schema conform to our naming
// conventions.
//
// Models, enums, enum values, roles, and API's must written in UpperCamelCase.
// Fields, actions, and inputs must be written in lowerCamelCase.
func NamingRule(_ []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterModel: func(n *parser.ModelNode) {
			errs.AppendError(checkNamingError(n.Name, "model"))
		},
		EnterField: func(n *parser.FieldNode) {
			errs.AppendError(checkNamingError(n.Name, "field"))
		},
		EnterAction: func(n *parser.ActionNode) {
			errs.AppendError(checkNamingError(n.Name, "action"))
		},
		EnterActionInput: func(n *parser.ActionInputNode) {
			if n.Label == nil {
				return
			}
			errs.AppendError(checkNamingError(*n.Label, "input"))
		},
		EnterEnum: func(n *parser.EnumNode) {
			errs.AppendError(checkNamingError(n.Name, "enum"))
			for _, v := range n.Values {
				errs.AppendError(checkNamingError(v.Name, "enum value"))
			}
		},
		EnterMessage: func(n *parser.MessageNode) {
			errs.AppendError(checkNamingError(n.Name, "message"))
		},
		EnterRole: func(n *parser.RoleNode) {
			errs.AppendError(checkNamingError(n.Name, "role"))
		},
		EnterAPI: func(n *parser.APINode) {
			errs.AppendError(checkNamingError(n.Name, "api"))
		},
	}
}

func checkNamingError(node parser.NameNode, entity string) *errorhandling.ValidationError {
	expected := toCamelCase(node.Value)
	casing := "UpperCamelCase"

	switch entity {
	case "field", "action", "input":
		expected = toLowerCamelCase(node.Value)
		casing = "lowerCamelCase"
	}

	if entity == "enum value" {
		entity = "enum values"
	} else {
		entity = fmt.Sprintf("%s names", entity)
	}

	if node.Value == expected {
		return nil
	}

	return errorhandling.NewValidationErrorWithDetails(
		errorhandling.NamingError,
		errorhandling.ErrorDetails{
			Message: fmt.Sprintf("%s must use %s", entity, casing),
			Hint:    fmt.Sprintf("e.g. %s", expected),
		},
		node,
	)
}

var (
	allCapsRe = regexp.MustCompile("^[A-Z_]+$")
)

func toCamelCase(s string) string {
	// Special cases:
	//  - For "FOOBAR" we want "Foobar"
	//  - For "FOO_BAR" we want "FooBar"
	// To get there we have to first lower case the string so
	// casing.ToCamel does the right thing
	if allCapsRe.MatchString(s) {
		s = strings.ToLower(s)
	}

	return casing.ToCamel(s)
}

func toLowerCamelCase(s string) string {
	// Special cases:
	//  - For "FOOBAR" we want "foobar"
	//  - For "FOO_BAR" we want "fooBar"
	// To get there we have to first lower case the string so
	// casing.ToLowerCamel does the right thing
	if allCapsRe.MatchString(s) {
		s = strings.ToLower(s)
	}

	return casing.ToLowerCamel(s)
}
