package validation

import (
	"fmt"

	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// AttributeArgumentsRules tests for the very basic rules around required arguments for attributes.
func AttributeArgumentsRules(asts []*parser.AST, errs *errorhandling.ValidationErrors) Visitor {
	var field *parser.FieldNode
	var action *parser.ActionNode
	var job *parser.JobNode
	var flow *parser.FlowNode

	return Visitor{
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
		EnterFlow: func(f *parser.FlowNode) {
			flow = f
		},
		LeaveFlow: func(*parser.FlowNode) {
			flow = nil
		},
		EnterAttribute: func(attribute *parser.AttributeNode) {
			var template map[string]bool
			var hint string
			validationErrors := []*errorhandling.ValidationError{}

			switch attribute.Name.Value {
			case parser.AttributeDefault:
				// A single optional argument
				template = map[string]bool{
					"": false,
				}

				hint = "the @default attribute takes a literal value - e.g. @default(1) - or no argument which will use the default value for this type"
			case parser.AttributeFunction:
				// No arguments
				template = map[string]bool{}

				hint = "the @function attribute does not accept any arguments"
			case parser.AttributeUnique:
				if field == nil {
					// A composite unique requires a single argument
					template = map[string]bool{
						"": true,
					}

					hint = "the @unique attribute at the model level is used to define composite uniques, for e.g. @unique(supplier, sku)"
				} else {
					// A field-level unique has no arguments
					template = map[string]bool{}

					hint = "the @unique attribute at the field level does not accept any arguments"
				}
			case parser.AttributeWhere:
				// Optional expression argument
				template = map[string]bool{
					"":           false,
					"expression": false,
				}

				hint = "the @where attribute accepts an expression as an argument, for e.g. @where(order.status == Status.Complete)"
			case parser.AttributeRelation:
				// A single required argument without a label
				template = map[string]bool{
					"": true,
				}

				hint = "the @relation attribute accepts the field name on the related model, for e.g. @relation(author)"
			case parser.AttributeSet:
				// A single required argument without a label
				template = map[string]bool{
					"": true,
				}

				hint = "the @set attribute sets a field on this model to some literal, for e.g. @set(order.status = Status.New)"
			case parser.AttributeSchedule:
				// A single required argument without a label
				template = map[string]bool{
					"": true,
				}

				hint = `the @schedule attribute accepts cron syntax as a string, for e.g. @schedule("every weekday at 9am")`
			case parser.AttributePermission:
				if action != nil {
					template = map[string]bool{
						"expression": false,
						"roles":      false,
					}

					hint = `the @permission attribute at the action level can accept either an expression or a roles argument, for e.g. @permission(expression: ctx.isAuthenticated)`
				} else if job != nil {
					template = map[string]bool{
						"expression": false,
						"roles":      false,
					}

					hint = `the @permission attribute in jobs can accept either an expression or a roles argument, for e.g. @permission(expression: ctx.isAuthenticated)`
				} else if flow != nil {
					template = map[string]bool{
						"expression": false,
						"roles":      false,
					}

					hint = `the @permission attribute in flows can accept either an expression or a roles argument, for e.g. @permission(roles: [Admin])`
				} else {
					template = map[string]bool{
						"expression": false,
						"roles":      false,
						"actions":    false,
					}

					hint = `the @permission attribute at the model level accepts either an expression or a roles argument, and an actions argument, for e.g. @permission(expression: ctx.isAuthenticated, actions: [get, list])`
				}
			default:
				return
			}

			var expected []string
			for k := range template {
				expected = append(expected, k)
			}

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
							Hint:    hint,
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
