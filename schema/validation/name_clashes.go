package validation

import (
	"fmt"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// NameClashes checks that the names of entities defined in a schema do not clash with
// built-in types or reserved keywords
func NameClashesRule(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			// skip built in models like Identity
			if m.BuiltIn {
				return
			}

			err := checkName(m.Name.Value, m.Name.Node)

			if err != nil {
				errs.AppendError(err)
			}
		},
		EnterMessage: func(m *parser.MessageNode) {
			err := checkMessageName(asts, m)

			if err != nil {
				errs.AppendError(err)
			}
		},
		EnterEnum: func(e *parser.EnumNode) {
			err := checkName(e.Name.Value, e.Name.Node)

			if err != nil {
				errs.AppendError(err)
			}
		},
	}
}

func checkName(name string, node node.Node) *errorhandling.ValidationError {
	reservedSuffixes := getReservedSuffixes()

	for _, suffix := range reservedSuffixes {
		entries := camelcase.Split(name)

		// if the prospective name split into word fragments (e.g CamelCase will become ["Camel", "Case"])
		// is just one item and that item matches one of the reserved suffixes, this is okay because it won't
		// clash with any other actions reserved entity names
		// e.g it is perfectly acceptable to have a model called "Edge" or "Connection"
		if strings.HasSuffix(name, suffix) && len(entries) > 1 {
			return errorhandling.NewValidationErrorWithDetails(
				errorhandling.NamingError,
				errorhandling.ErrorDetails{
					Message: fmt.Sprintf("Names cannot end with '%s'", suffix),
				},
				node,
			)
		}
	}

	err := checkReservedExactMatches(name, node)

	if err != nil {
		return err
	}

	return nil
}

func getReservedSuffixes() []string {
	return []string{"Input", "Connection", "Edge", "Values", "Where", "Request", "Response"}
}

func getReservedKeelNames() []string {
	return []string{"Any", "ID", "Text", "Boolean", "Secret", "Image", "Float", "Image", "File", "Coordinate", "Location", "Email", "Phone", "PageInfo"}
}

func getReservedGraphQLNames() []string {
	return []string{"Mutation", "Query", "Subscription"}
}

func checkReservedExactMatches(name string, node node.Node) *errorhandling.ValidationError {
	reservedNames := merge(
		getReservedKeelNames(),
		getReservedJavaScriptGlobals(),
		getReservedGraphQLNames(),
	)

	for _, reserved := range reservedNames {
		if name == reserved {
			return errorhandling.NewValidationErrorWithDetails(
				errorhandling.NamingError,
				errorhandling.ErrorDetails{
					Message: fmt.Sprintf("Reserved name '%s'", name),
				},
				node,
			)
		}
	}

	return nil
}

func getReservedJavaScriptGlobals() []string {
	// todo: other potential things to restrict:
	// 	ret = append(ret, "Global")
	// 	ret = append(ret, "Blob")
	// 	ret = append(ret, "Array")
	// 	ret = append(ret, "Buffer")
	// 	ret = append(ret, "Crypto")
	// 	ret = append(ret, "File")
	// 	ret = append(ret, "Event")
	// 	ret = append(ret, "EventTarget")
	// 	ret = append(ret, "Headers")
	// 	ret = append(ret, "FormData")
	// 	ret = append(ret, "TextDecoder")
	// 	ret = append(ret, "TextEncoder")
	// 	ret = append(ret, "Error")
	// 	ret = append(ret, "URL")

	return []string{"String", "Number", "Boolean", "Object", "Array", "Error"}
}

func checkMessageName(asts []*parser.AST, message *parser.MessageNode) *errorhandling.ValidationError {
	candidateActions := []parser.ActionNode{}

	// collect all of the built-in actions in the schema as these will be the clash candidates
	for _, model := range query.Models(asts) {
		for _, action := range query.ModelActions(model) {
			if action.Type.Value != parser.ActionTypeRead && action.Type.Value != parser.ActionTypeWrite {
				candidateActions = append(candidateActions, *action)
			}
		}
	}

	for _, a := range candidateActions {
		for _, suffix := range getReservedSuffixes() {
			// for every reserved suffix we want to check the combination of the action name + suffix
			// against the prospective messsage name.
			// e.g You shouldn't be allowed to have a message name called FooInput if there is already a built in
			// action called foo
			if !message.BuiltIn && message.Name.Value == fmt.Sprintf("%s%s", casing.ToCamel(a.Name.Value), suffix) {
				return errorhandling.NewValidationErrorWithDetails(
					errorhandling.NamingError,
					errorhandling.ErrorDetails{
						Message: fmt.Sprintf("Reserved message name '%s'", message.Name.Value),
					},
					message.Name,
				)
			}
		}
	}

	err := checkReservedExactMatches(message.Name.Value, message.Name.Node)

	if err != nil {
		return err
	}

	return nil
}

func merge(slices ...[]string) (ret []string) {
	for _, slice := range slices {
		ret = append(ret, slice...)
	}

	return ret
}
