package validation

import (
	"fmt"
	"strings"

	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// AttributeArgumentsRules tests for the very basic rules around required arguments for attributes
func AttributeArgumentsRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var model *parser.ModelNode
	var field *parser.FieldNode
	var action *parser.ActionNode
	var job *parser.JobNode

	return Visitor{
		EnterModel: func(m *parser.ModelNode) {
			model = m
		},
		LeaveModel: func(*parser.ModelNode) {
			model = nil
		},
		EnterAction: func(a *parser.ActionNode) {
			action = a
		},
		LeaveAction: func(*parser.ActionNode) {
			action = nil
		},
		EnterField: func(f *parser.FieldNode) {
			field = f
		},
		LeaveField: func(*parser.FieldNode) {
			field = nil
		},
		EnterJob: func(j *parser.JobNode) {
			job = j
		},
		LeaveJob: func(*parser.JobNode) {
			job = nil
		},
		EnterAttribute: func(attribute *parser.AttributeNode) {
			var template map[string]bool
			validationErrors := []*errorhandling.ValidationError{}

			switch attribute.Name.Value {
			case parser.AttributeDefault:
				// A single optional argument
				template = map[string]bool{
					"": false,
				}
			case parser.AttributeFunction:
				// No arguments
				template = map[string]bool{}
			case parser.AttributeUnique:
				if field == nil {
					// A composite unique requires a single argument
					template = map[string]bool{
						"": true,
					}
				} else {
					// A field-level unique has no arguments
					template = map[string]bool{}
				}
			case parser.AttributeWhere:
				// Optional expression argument
				template = map[string]bool{
					"":           false,
					"expression": false,
				}
			case parser.AttributeRelation, parser.AttributeSet, parser.AttributeSchedule:
				// A single required argument without a label
				template = map[string]bool{
					"": true,
				}
			case parser.AttributePermission:
				if action != nil {
					template = map[string]bool{
						"expression": false,
						"roles":      false,
					}
				} else if job != nil {
					template = map[string]bool{
						"expression": false,
						"roles":      false,
					}
				} else {
					template = map[string]bool{
						"expression": false,
						"roles":      false,
						"actions":    false,
					}
				}
			case parser.AttributeEmbed:
				template = map[string]bool{}
				for _, f := range query.ModelFields(model) {
					if query.Model(asts, f.Type.Value) != nil {
						template[f.Name.Value] = false
					}
				}
			default:
				return
			}

			var expected []string
			for k := range template {
				expected = append(expected, k)
			}

			hint := fmt.Sprintf("the signature for this @%s attribute is (%s)", attribute.Name.Value, strings.Join(expected, ", "))

			// copy a map
			arguments := make(map[string]bool)
			for k, v := range template {
				arguments[k] = v
			}

			// Look for unexpected arguments
			for _, arg := range attribute.Arguments {
				label := ""
				if arg.Label != nil {
					label = arg.Label.Value
				}

				if _, has := arguments[label]; has {
					delete(arguments, label)
				} else {
					var message string
					if label == "" {
						message = fmt.Sprintf("unexpected argument for @%s", attribute.Name.Value)
					} else {
						message = fmt.Sprintf("unexpected argument '%s' for @%s", label, attribute.Name.Value)
					}

					if len(expected) == 1 && expected[0] == "" {
						message = message + " as only a single argument is expected"
					} else if len(expected) == 0 {
						message = message + " as no arguments are expected"
					}

					var node node.ParserNode
					if label != "" {
						node = arg.Label
					} else {
						node = arg
					}

					validationErrors = append(validationErrors, errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: message,
							//Hint:    hint,
						},
						node,
					))
				}
			}

			if len(validationErrors) > 0 {
				errs.AppendErrors(validationErrors)
				return
			}

			// Look for expected arguments
			for k, v := range arguments {
				if !v {
					// if the argument is optional, then continue
					continue
				}
				if k == "" {
					validationErrors = append(validationErrors, errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("expected an argument for @%s", attribute.Name.Value),
							Hint:    hint,
						},
						attribute,
					))
				} else {
					validationErrors = append(validationErrors, errorhandling.NewValidationErrorWithDetails(
						errorhandling.AttributeArgumentError,
						errorhandling.ErrorDetails{
							Message: fmt.Sprintf("expected argument '%s' for @%s", k, attribute.Name.Value),
							Hint:    hint,
						},
						attribute,
					))
				}
			}

			errs.AppendErrors(validationErrors)

		},
	}
}

// E024:
// message: "{{ .ActualArgsCount }} argument(s) provided to @{{ .AttributeName }} but expected {{ .ExpectedArgsCount }}"
// hint: '{{ if eq .Signature "()" }}@{{ .AttributeName }} doesn''t accept any arguments{{ else }}the signature of this attribute is @{{ .AttributeName }}{{ .Signature }}{{ end }}'

// func makeWhereExpressionError(t errorhandling.ErrorType, message string, hint string, node node.ParserNode) *errorhandling.ValidationError {
// 	return errorhandling.NewValidationErrorWithDetails(
// 		t,
// 		errorhandling.ErrorDetails{
// 			Message: message,
// 			Hint:    hint,
// 		},
// 		node,
// 	)
// }
