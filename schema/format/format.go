package format

import (
	"regexp"
	"strings"
	"text/scanner"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/schema/node"
	"github.com/teamkeel/keel/schema/parser"
)

const (
	indentSize    = 4
	maxLineLength = 80
)

func HasComments(nodes []node.ParserNode) bool {
	for _, n := range nodes {
		for _, t := range n.GetTokens() {
			if t.Type == scanner.Comment {
				return true
			}
		}
	}
	return false
}

func Format(ast *parser.AST) string {
	writer := &Writer{
		commentStack: [][]lexer.Token{},
		commentCache: map[string]bool{},
	}

	for i, decl := range ast.Declarations {
		if i > 0 {
			writer.writeLine("")
		}
		writer.comments(decl, func() {
			switch {
			case decl.Model != nil:
				printModel(writer, decl.Model)
			case decl.Enum != nil:
				printEnum(writer, decl.Enum)
			case decl.Role != nil:
				printRole(writer, decl.Role)
			case decl.API != nil:
				printApi(writer, decl.API)
			case decl.Message != nil:
				printMessage(writer, decl.Message)
			case decl.Job != nil:
				printJob(writer, decl.Job)
			case decl.Flow != nil:
				printFlow(writer, decl.Flow)
			case decl.Routes != nil:
				printRoute(writer, decl.Routes)
			}
		})
	}

	return writer.string()
}

func printMessage(writer *Writer, message *parser.MessageNode) {
	writer.comments(message, func() {
		writer.write("message %s", camel(message.Name.Value))
		writer.block(func() {
			for _, field := range message.Fields {
				writer.comments(field, func() {
					writer.write(
						"%s %s",
						lowerCamel(field.Name.Value),
						field.Type.Value,
					)

					if field.Optional {
						writer.write("?")
					}

					if field.Repeated {
						writer.write("[]")
					}

					writer.writeLine("")
				})
			}
		})
	})
}

func printJob(writer *Writer, job *parser.JobNode) {
	writer.comments(job, func() {
		writer.write("job %s", camel(job.Name.Value))
		writer.block(func() {
			for _, section := range job.Sections {
				writer.comments(section, func() {
					switch {
					case len(section.Inputs) > 0:
						writer.write("inputs")
						writer.block(func() {
							for _, input := range section.Inputs {
								writer.comments(input, func() {
									// CamelCase input types except for ID
									inputType := input.Type.Value
									if inputType != parser.FieldTypeID {
										inputType = camel(inputType)
									}
									writer.write(
										"%s %s",
										lowerCamel(input.Name.Value),
										inputType,
									)
									if input.Optional {
										writer.write("?")
									}
									writer.writeLine("")
								})
							}
						},
						)
						if len(section.Inputs) > 0 {
							writer.writeLine("")
						}
					case section.Attribute != nil:
						printAttributes(writer, []*parser.AttributeNode{section.Attribute})
					}
				})
			}
		})
	})
}

func printFlow(writer *Writer, flow *parser.FlowNode) {
	writer.comments(flow, func() {
		writer.write("flow %s", camel(flow.Name.Value))
		writer.block(func() {
			for _, section := range flow.Sections {
				writer.comments(section, func() {
					switch {
					case len(section.Inputs) > 0:
						writer.write("inputs")
						writer.block(func() {
							for _, input := range section.Inputs {
								writer.comments(input, func() {
									// CamelCase input types except for ID
									inputType := input.Type.Value
									if inputType != parser.FieldTypeID {
										inputType = camel(inputType)
									}
									writer.write(
										"%s %s",
										lowerCamel(input.Name.Value),
										inputType,
									)
									if input.Optional {
										writer.write("?")
									}
									writer.writeLine("")
								})
							}
						},
						)
					case section.Attribute != nil:
						printAttributes(writer, []*parser.AttributeNode{section.Attribute})
					}
				})
			}
		})
	})
}

func printModel(writer *Writer, model *parser.ModelNode) {
	writer.comments(model, func() {
		writer.write("model %s", camel(model.Name.Value))
		writer.block(func() {
			fieldSections := []*parser.ModelSectionNode{}
			actionSections := []*parser.ModelSectionNode{}
			attributeSections := []*parser.ModelSectionNode{}

			for _, section := range model.Sections {
				if section.Fields != nil {
					fieldSections = append(fieldSections, section)
				}
				if section.Actions != nil {
					actionSections = append(actionSections, section)
				}
				if section.Attribute != nil {
					attributeSections = append(attributeSections, section)
				}
			}

			sections := 0

			for _, section := range fieldSections {
				fields := section.Fields
				writer.comments(section, func() {
					writer.write("fields")
					writer.block(func() {
						for _, field := range fields {
							if field.BuiltIn {
								continue
							}

							fieldType := ""
							switch field.Type.Value {
							case parser.FieldTypeID:
								// we dont want to camel case ID as it should be in all caps
								fieldType = field.Type.Value
							default:
								fieldType = camel(field.Type.Value)
							}

							if field.Repeated {
								fieldType += "[]"
							}
							if field.Optional {
								fieldType += "?"
							}

							writer.comments(field, func() {
								writer.write(
									"%s %s",
									lowerCamel(field.Name.Value),
									fieldType,
								)

								hasComments := false
								for _, attr := range field.Attributes {
									if attr.Tokens[0].Type == scanner.Comment {
										hasComments = true
									}
								}

								// TODO: this needs a lot more thought, but for now
								// we omit the curly braces if there is just one
								// attribute and no comments, otherwise the attributes
								// get wrapper in a block
								if len(field.Attributes) == 1 && !hasComments {
									writer.write(" ")
									printAttributes(writer, field.Attributes)
								} else {
									printAttributesBlock(writer, field.Attributes)
								}
							})
						}
					})
				})
				sections++
			}

			for _, section := range actionSections {
				if sections > 0 {
					writer.writeLine("")
				}
				printActionsBlock(writer, section)
				sections++
			}

			for _, section := range attributeSections {
				if sections > 0 {
					writer.writeLine("")
				}
				writer.comments(section, func() {
					printAttributes(writer, []*parser.AttributeNode{section.Attribute})
				})
				sections++
			}
		})
	})
}

func printActionsBlock(writer *Writer, section *parser.ModelSectionNode) {
	writer.comments(section, func() {
		actions := []*parser.ActionNode{}
		if section.Actions != nil {
			actions = section.Actions
			writer.write(parser.KeywordActions)
		}

		writer.block(func() {
			for _, op := range actions {
				writer.comments(op, func() {
					writer.write(
						"%s %s",
						lowerCamel(op.Type.Value),
						lowerCamel(op.Name.Value),
					)

					printActionInputs(writer, op.Inputs, op.IsArbitraryFunction())

					if len(op.With) > 0 {
						writer.write(" with ")
						printActionInputs(writer, op.With, op.IsArbitraryFunction())
					}

					if len(op.Returns) > 0 {
						writer.write(" returns ")
						printActionInputs(writer, op.Returns, op.IsArbitraryFunction())
					}

					printAttributesBlock(writer, op.Attributes)
				})
			}
		})
	})
}

func printActionInputs(writer *Writer, inputs []*parser.ActionInputNode, isArbitraryFunction bool) {
	writer.write("(")
	writer.indent()

	// If there any comments in the action inputs then we need to print
	// each input on it's own line, to allow space for the comments
	isMultiline := HasComments(lo.Map(inputs, func(i *parser.ActionInputNode, _ int) node.ParserNode {
		return i
	}))

	// TODO: find a more generic way to do line-wrapping
	if !isMultiline {
		length := writer.lineLength()
		for _, arg := range inputs {
			if arg.Label != nil {
				length += len(arg.Label.Value)
				length += 2 // account for ": "
				length += len(arg.Type.Fragments[0].Fragment)
			} else {
				for i, frag := range arg.Type.Fragments {
					if i > 0 {
						length += 1 // account for "."
					}
					length += len(frag.Fragment)
				}
			}
			if arg.Optional {
				length += 1 // account for "?"
			}
		}
		if length > maxLineLength {
			isMultiline = true
		}
	}

	for i, arg := range inputs {
		if isMultiline {
			writer.writeLine("")
		}

		writer.comments(arg, func() {
			if arg.Label != nil {
				// explicit input
				writer.write("%s: %s", arg.Label.Value, arg.Type.Fragments[0].Fragment)
			} else {
				// Note: not using arg.Type.ToString() here as we want to try
				// and fix any casing issues
				for i, fragment := range arg.Type.Fragments {
					if i > 0 {
						writer.write(".")
					}

					// if its an arbitrary function, then we dont want to automatically lowercase the input names
					if isArbitraryFunction {
						writer.write("%s", fragment.Fragment)
					} else {
						writer.write("%s", lowerCamel(fragment.Fragment))
					}
				}
			}

			if arg.Optional {
				writer.write("?")
			}
		})

		if len(inputs) > 1 {
			if isMultiline {
				writer.write(",")
			} else if i < len(inputs)-1 {
				writer.write(", ")
			}
		}
	}

	if isMultiline {
		writer.writeLine("")
	}

	writer.dedent()
	writer.write(")")
}

func printRole(writer *Writer, role *parser.RoleNode) {
	writer.comments(role, func() {
		writer.write("role %s", camel(role.Name.Value))
		writer.block(func() {
			sections := 0
			// domains
			for _, section := range role.Sections {
				if len(section.Domains) > 0 {
					sections++
					writer.comments(section, func() {
						writer.write("domains")
						writer.block((func() {
							for _, domain := range section.Domains {
								writer.comments(domain, func() {
									writer.writeLine(domain.Domain)
								})
							}
						}))
					})
				}
			}

			// emails
			for _, section := range role.Sections {
				if len(section.Emails) > 0 {
					if sections > 0 {
						writer.writeLine("")
					}
					writer.comments(section, func() {
						writer.write("emails")
						writer.block(func() {
							for _, email := range section.Emails {
								writer.comments(email, func() {
									writer.writeLine(email.Email)
								})
							}
						})
					})
				}
			}
		})
	})
}

func printApi(writer *Writer, api *parser.APINode) {
	writer.comments(api, func() {
		writer.write("api %s", camel(api.Name.Value))
		writer.block(func() {
			for i, section := range api.Sections {
				if i > 0 {
					writer.writeLine("")
				}
				writer.comments(section, func() {
					switch {
					case len(section.Models) > 0:
						writer.write("models")
						writer.block(func() {
							for _, model := range section.Models {
								writer.comments(model, func() {
									writer.write("%s", camel(model.Name.Value))
									if len(model.Sections) == 1 {
										writer.block(func() {
											writer.write("actions")
											writer.block(func() {
												for j, action := range model.Sections[0].Actions {
													if j > 0 {
														writer.writeLine("")
													}
													writer.comments(action, func() {
														writer.write(action.Name.Value)
													})
												}
												writer.writeLine("")
											})
										})
									} else {
										writer.writeLine("")
									}
								})
							}
						})

					case section.Attribute != nil:
						printAttributes(writer, []*parser.AttributeNode{section.Attribute})
					}
				})
			}
		})
	})
}

func printAttributesBlock(writer *Writer, attributes []*parser.AttributeNode) {
	if len(attributes) == 0 {
		writer.writeLine("")
		return
	}

	if len(attributes) == 1 && attributes[0].Name.Value == parser.AttributeFunction {
		writer.write(" ")
		printAttributes(writer, attributes)

		return
	}

	writer.block(func() {
		printAttributes(writer, attributes)
	})
}

func printAttributes(writer *Writer, attributes []*parser.AttributeNode) {
	for _, attr := range attributes {
		writer.comments(attr, func() {
			writer.write("@%s", lowerCamel(attr.Name.Value))

			if len(attr.Arguments) > 0 {
				writer.write("(")

				isMultiline := len(attr.Arguments) > 1
				if isMultiline {
					writer.writeLine("")
					writer.indent()
				}

				for i, arg := range attr.Arguments {
					if i > 0 {
						if isMultiline {
							writer.writeLine(",")
						} else {
							writer.write(", ")
						}
					}
					writer.comments(arg, func() {
						if arg.Label != nil {
							writer.write("%s: ", lowerCamel(arg.Label.Value))
						}
						expr := arg.Expression.CleanString()
						writer.write(expr)
					})
				}

				if isMultiline {
					writer.writeLine("")
					writer.dedent()
				}

				writer.write(")")
			}

			writer.writeLine("")
		})
	}
}

var allCapsRe = regexp.MustCompile("^[A-Z]+$")

func camel(s string) string {
	// Special case if the string is "FOOBAR" we want "Foobar" but
	// to get there we have to first lower case the string so
	// casing.ToCamel does the right thing
	if allCapsRe.MatchString(s) {
		s = strings.ToLower(s)
	}

	return casing.ToCamel(s)
}

func lowerCamel(s string) string {
	// Special case if the string is "FOOBAR" we want "foobar"
	if allCapsRe.MatchString(s) {
		return strings.ToLower(s)
	}

	return casing.ToLowerCamel(s)
}

func printEnum(writer *Writer, enum *parser.EnumNode) {
	writer.comments(enum, func() {
		writer.write("enum %s", camel(enum.Name.Value))
		writer.block(func() {
			for _, v := range enum.Values {
				writer.comments(v, func() {
					writer.writeLine(v.Name.Value)
				})
			}
		})
	})
}

func printRoute(writer *Writer, routes *parser.RoutesNode) {
	writer.write("routes")
	writer.block(func() {
		writer.comments(routes, func() {
			for _, route := range routes.Routes {
				writer.comments(route, func() {
					writer.write(strings.ToLower(route.Method.Value))
					writer.write("(")
					writer.write(route.Pattern.Value)
					writer.write(", ")
					writer.write(route.Handler.Value)
					writer.writeLine(")")
				})
			}
		})
	})
}
