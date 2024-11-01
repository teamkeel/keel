package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

// Visitor lets you define "enter" and "leave" functions for AST nodes.
// This struct may not have fields for all AST nodes, so if you need to
// visit a node that is not currently supported add the necessary fields
// to this struct. For your functions to get called you must name these
// fields correctly.
//
// For a node type called "SomethingNode" the hooks would be:
//
//	EnterSomething: func(n *parser.SomethingNode)
//	LeaveSomething: func(n *parser.SomethingNode)
type Visitor struct {
	EnterModel func(n *parser.ModelNode)
	LeaveModel func(n *parser.ModelNode)

	EnterModelSection func(n *parser.ModelSectionNode)
	LeaveModelSection func(n *parser.ModelSectionNode)

	EnterMessage func(n *parser.MessageNode)
	LeaveMessage func(n *parser.MessageNode)

	EnterField func(n *parser.FieldNode)
	LeaveField func(n *parser.FieldNode)

	EnterAction func(n *parser.ActionNode)
	LeaveAction func(n *parser.ActionNode)

	EnterWith func(n *parser.ActionNode)
	LeaveWith func(n *parser.ActionNode)

	EnterActionInput func(n *parser.ActionInputNode)
	LeaveActionInput func(n *parser.ActionInputNode)

	EnterEnum func(n *parser.EnumNode)
	LeaveEnum func(n *parser.EnumNode)

	EnterRole func(n *parser.RoleNode)
	LeaveRole func(n *parser.RoleNode)

	EnterAttribute func(n *parser.AttributeNode)
	LeaveAttribute func(n *parser.AttributeNode)

	EnterAttributeArgument func(n *parser.AttributeArgumentNode)
	LeaveAttributeArgument func(n *parser.AttributeArgumentNode)

	EnterAPI func(n *parser.APINode)
	LeaveAPI func(n *parser.APINode)

	EnterAPIModel func(n *parser.APIModelNode)
	LeaveAPIModel func(n *parser.APIModelNode)

	EnterAPIModelAction func(n *parser.APIModelActionNode)
	LeaveAPIModelAction func(n *parser.APIModelActionNode)

	EnterJob func(n *parser.JobNode)
	LeaveJob func(n *parser.JobNode)

	EnterJobInput func(n *parser.JobInputNode)
	LeaveJobInput func(n *parser.JobInputNode)

	EnterExpression func(e *parser.Expression)
	LeaveExpression func(e *parser.Expression)
}

type VisitorFunc func([]*parser.AST, *errorhandling.ValidationErrors) Visitor

func runVisitors(asts []*parser.AST, visitors []Visitor) {
	for _, ast := range asts {
		visit(reflect.ValueOf(ast), visitors)
	}
}

func visit(v reflect.Value, visitors []Visitor) {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			visit(v.Index(i), visitors)
		}
		return
	}

	if v.Kind() != reflect.Struct {
		return
	}

	callVisitHook("Enter", v, visitors)
	defer callVisitHook("Leave", v, visitors)

	for i := 0; i < v.NumField(); i++ {
		visit(v.FieldByIndex([]int{i}), visitors)
	}
}

func callVisitHook(action string, v reflect.Value, visitors []Visitor) {
	hookName := fmt.Sprintf("%s%s", action, strings.TrimSuffix(v.Type().Name(), "Node"))

	for _, visitor := range visitors {
		hookFunc := reflect.ValueOf(visitor).FieldByName(hookName)
		if hookFunc.Kind() == reflect.Func && !hookFunc.IsNil() {
			hookFunc.Call([]reflect.Value{
				v.Addr(),
			})
		}
	}
}
