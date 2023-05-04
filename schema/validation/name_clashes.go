package validation

import (
	"fmt"
	"strings"

	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
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
			err := checkName(m.Name.Value, m.Name.Node)

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
		if strings.HasSuffix(name, suffix) {
			return errorhandling.NewValidationErrorWithDetails(
				errorhandling.NamingError,
				errorhandling.ErrorDetails{
					Message: fmt.Sprintf("Model names cannot end with '%s'", suffix),
					Hint:    "Try using a different name",
				},
				node,
			)
		}
	}

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
					Message: fmt.Sprintf("Reserved model name '%s'", name),
					Hint:    "Try using a different name",
				},
				node,
			)
		}
	}

	return nil
}

func getReservedSuffixes() []string {
	return []string{"Input", "Connection", "Edge", "Values", "Where", "Request", "Response"}
}

func getReservedKeelNames() []string {
	return []string{"Any", "ID", "Identity", "Text", "Boolean", "Secret", "Image", "Float", "Image", "File", "Coordinate", "Location", "Email", "Phone", "PageInfo"}
}

func getReservedGraphQLNames() []string {
	return []string{"Mutation", "Query", "Subscription"}
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

func merge(slices ...[]string) (ret []string) {
	for _, slice := range slices {
		ret = append(ret, slice...)
	}

	return ret
}
