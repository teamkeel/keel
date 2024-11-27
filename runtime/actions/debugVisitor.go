package actions

import (
	"fmt"
	"strings"
)

func DebugStringGenerator() expressionVisitor[string] {
	return &debugStringVisitor{
		output: strings.Builder{},
	}
}

var _ expressionVisitor[string] = new(debugStringVisitor)

type debugStringVisitor struct {
	indent string
	output strings.Builder
}

func (v *debugStringVisitor) result() string {
	return v.output.String()
}

func (v *debugStringVisitor) modelName() string {
	return ""
}

func (v *debugStringVisitor) startCondition(parenthesis bool) error {

	//v.output.WriteString(v.indent)
	v.output.WriteString("(")
	//v.indent += ""
	v.output.WriteString(v.indent)

	// if parenthesis {
	// 	err := v.startCondition(false)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (v *debugStringVisitor) endCondition(parenthesis bool) error {

	//v.indent = strings.TrimPrefix(v.indent, "-")
	v.output.WriteString(v.indent + ")")

	// if parenthesis {
	// 	err := v.endCondition(false)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (v *debugStringVisitor) visitAnd() error {
	v.output.WriteString(v.indent + "&&")
	v.output.WriteString(v.indent)

	return nil
}

func (v *debugStringVisitor) visitOr() error {
	v.output.WriteString(v.indent + "||")
	v.output.WriteString(v.indent)

	return nil
}

func (v *debugStringVisitor) visitOperator(op ActionOperator) error {
	v.output.WriteString(fmt.Sprintf("%s", op))
	return nil
}

func (v *debugStringVisitor) visitLiteral(value any) error {
	v.output.WriteString(fmt.Sprintf("%s", value))
	return nil
}

func (v *debugStringVisitor) visitInput(name string) error {
	v.output.WriteString(fmt.Sprintf("%s", name))
	return nil
}

func (v *debugStringVisitor) visitField(fragments []string) error {
	v.output.WriteString(fmt.Sprintf("%s", strings.Join(fragments, ".")))
	return nil
}
