package format

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

const (
	indentSize = 4
)

type Writer struct {
	b          strings.Builder
	currIndent int
}

func (w *Writer) WriteLine(s string, args ...any) {
	if w.isStartOfLine() && s != "" {
		w.b.WriteString(strings.Repeat(" ", w.currIndent))
	}
	w.b.WriteString(fmt.Sprintf(s+"\n", args...))
}

func (w *Writer) Write(s string, args ...any) {
	if w.isStartOfLine() && s != "" {
		w.b.WriteString(strings.Repeat(" ", w.currIndent))
	}
	w.b.WriteString(fmt.Sprintf(s, args...))
}

func (w *Writer) Indent() {
	w.currIndent += indentSize
}

func (w *Writer) Dedent() {
	w.currIndent -= indentSize
	if w.currIndent < 0 {
		w.currIndent = 0
	}
}

func (w *Writer) Block(fn func()) {
	w.WriteLine(" {")
	w.Indent()
	fn()
	w.Dedent()
	w.WriteLine("}")
}

func (w *Writer) String() string {
	return w.b.String()
}

func (w *Writer) isStartOfLine() bool {
	s := w.b.String()
	return len(s) > 0 && s[len(s)-1] == '\n'
}

func Format(ast *parser.AST) string {
	writer := &Writer{}

	for i, decl := range ast.Declarations {
		if i > 0 {
			writer.WriteLine("")
		}
		switch {
		case decl.Model != nil:
			printModel(writer, decl.Model)
		case decl.Enum != nil:
			printEnum(writer, decl.Enum)
		case decl.Role != nil:
			printRole(writer, decl.Role)
		case decl.API != nil:
			printApi(writer, decl.API)
		}
	}

	return writer.String()
}

func printModel(writer *Writer, model *parser.ModelNode) {
	writer.Write("model %s", camel(model.Name.Value))
	writer.Block(func() {

		fields := query.ModelFields(model, query.ExcludeBuiltInFields)
		sections := 0

		// fields
		if len(fields) > 0 {
			writer.Write("fields")
			writer.Block(func() {
				for _, field := range fields {
					fieldType := camel(field.Type)
					if field.Optional {
						fieldType += "?"
					}
					if field.Repeated {
						fieldType += "[]"
					}

					writer.Write(
						"%s %s",
						lowerCamel(field.Name.Value),
						fieldType,
					)

					switch len(field.Attributes) {
					case 1:
						writer.Write(" ")
						printAttributes(writer, field.Attributes)
					default:
						printAttributesBlock(writer, field.Attributes)
					}
				}
			})
			sections++
		}

		// operations
		ops := query.ModelOperations(model)
		if len(ops) > 0 {
			if sections > 0 {
				writer.WriteLine("")
			}
			sections++
			writer.Write("operations")
			printActionsBlock(writer, ops)
		}

		// functions
		funcs := query.ModelFunctions(model)
		if len(funcs) > 0 {
			if sections > 0 {
				writer.WriteLine("")
			}
			sections++
			writer.Write("functions")
			printActionsBlock(writer, ops)
		}

		// attributes
		attrs := query.ModelAttributes(model)
		if len(attrs) > 0 {
			if sections > 0 {
				writer.WriteLine("")
			}
			sections++
			printAttributes(writer, query.ModelAttributes(model))
		}
	})
}

func printActionsBlock(writer *Writer, actions []*parser.ActionNode) {
	writer.Block(func() {
		for _, op := range actions {
			writer.Write(
				"%s %s",
				lowerCamel(op.Type),
				lowerCamel(op.Name.Value),
			)

			writer.Write("(")
			for i, arg := range op.Arguments {
				if i > 0 {
					writer.Write(", ")
				}
				writer.Write(lowerCamel(arg.Name.Value))
			}
			writer.Write(")")
			printAttributesBlock(writer, op.Attributes)
		}

	})
}

func printRole(writer *Writer, role *parser.RoleNode) {
	writer.Write("role %s", camel(role.Name.Value))
	writer.Block(func() {
		sections := 0
		// domains
		for _, section := range role.Sections {
			if len(section.Domains) > 0 {
				sections++
				writer.Write("domains")
				writer.Block((func() {
					for _, domain := range section.Domains {
						writer.WriteLine(domain.Domain)
					}
				}))
			}
		}

		// emails
		for _, section := range role.Sections {
			if len(section.Emails) > 0 {
				if sections > 0 {
					writer.WriteLine("")
				}
				writer.Write("emails")
				writer.Block(func() {
					for _, email := range section.Emails {
						writer.WriteLine(email.Email)
					}
				})
			}
		}
	})
}

func printApi(writer *Writer, api *parser.APINode) {
	writer.Write("api %s", camel(api.Name.Value))
	writer.Block(func() {
		for i, section := range api.Sections {
			if i > 0 {
				writer.WriteLine("")
			}
			switch {
			case len(section.Models) > 0:
				writer.Write("models")
				writer.Block(func() {
					for _, model := range section.Models {
						writer.WriteLine(camel(model.Name.Value))
					}
				})
			case section.Attribute != nil:
				printAttributes(writer, []*parser.AttributeNode{section.Attribute})
			}
		}
	})
}

func printAttributesBlock(writer *Writer, attributes []*parser.AttributeNode) {
	if len(attributes) == 0 {
		writer.WriteLine("")
		return
	}

	writer.Block(func() {
		printAttributes(writer, attributes)
	})
}

func printAttributes(writer *Writer, attributes []*parser.AttributeNode) {
	for _, attr := range attributes {
		writer.Write("@%s", lowerCamel(attr.Name.Value))

		if len(attr.Arguments) > 0 {
			writer.Write("(")

			isMultiline := len(attr.Arguments) > 1
			if isMultiline {
				writer.WriteLine("")
				writer.Indent()
			}

			for i, arg := range attr.Arguments {
				if i > 0 {
					if isMultiline {
						writer.WriteLine(",")
					} else {
						writer.Write(", ")
					}
				}
				if arg.Name != nil {
					writer.Write("%s: ", lowerCamel(arg.Name.Value))
				}
				expr, _ := expressions.ToString(arg.Expression)
				writer.Write(expr)
			}

			if isMultiline {
				writer.WriteLine("")
				writer.Dedent()
			}

			writer.Write(")")
		}

		writer.WriteLine("")
	}
}

var allCapsRe = regexp.MustCompile("^[A-Z]+$")

func camel(s string) string {
	// Special case if the string is "FOOBAR" we want "Foobar" but
	// to get there we have to first lower case the string so
	// strcase.ToCamel does the right thing
	if allCapsRe.MatchString(s) {
		s = strings.ToLower(s)
	}

	return strcase.ToCamel(s)
}

func lowerCamel(s string) string {
	// Special case if the string is "FOOBAR" we want "foobar"
	if allCapsRe.MatchString(s) {
		return strings.ToLower(s)
	}

	return strcase.ToLowerCamel(s)
}

func printEnum(writer *Writer, enum *parser.EnumNode) {
	writer.Write("enum %s", camel(enum.Name.Value))
	writer.Block(func() {
		for _, v := range enum.Values {
			writer.WriteLine(camel(v.Name.Value))
		}
	})
}
