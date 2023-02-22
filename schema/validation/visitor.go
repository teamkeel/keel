package validation

import (
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

type Visitor struct {
	EnterModel             func(n *parser.ModelNode)
	EnterField             func(n *parser.FieldNode)
	EnterAction            func(n *parser.ActionNode)
	ExitAction             func(n *parser.ActionNode)
	EnterOperation         func(n *parser.ActionNode)
	ExitOperation          func(n *parser.ActionNode)
	EnterFunction          func(n *parser.ActionNode)
	ExitFunction           func(n *parser.ActionNode)
	EnterInput             func(n *parser.ActionInputNode)
	EnterReadInput         func(n *parser.ActionInputNode)
	EnterWriteInput        func(n *parser.ActionInputNode)
	EnterEnum              func(n *parser.EnumNode)
	EnterRole              func(n *parser.RoleNode)
	EnterAttribute         func(n *parser.AttributeNode)
	EnterAttributeArgument func(n *parser.AttributeArgumentNode)
	ExitAttribute          func(n *parser.AttributeNode)
	EnterAPI               func(n *parser.APINode)
}

type VisitorFunc func([]*parser.AST, *errorhandling.ValidationErrors) Visitor

func runVisitors(asts []*parser.AST, visitors []Visitor) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			switch {
			case decl.Model != nil:
				visitModel(decl.Model, visitors)
			case decl.Role != nil:
				for _, v := range visitors {
					if v.EnterRole != nil {
						v.EnterRole(decl.Role)
					}
				}
			case decl.API != nil:
				for _, v := range visitors {
					if v.EnterAPI != nil {
						v.EnterAPI(decl.API)
					}
				}
			}
		}
	}
}

func visitModel(m *parser.ModelNode, visitors []Visitor) {
	for _, v := range visitors {
		if v.EnterModel != nil {
			v.EnterModel(m)
		}
	}
	for _, section := range m.Sections {
		switch {
		case len(section.Fields) > 0:
			for _, field := range section.Fields {
				visitField(field, visitors)
			}
		case len(section.Operations) > 0:
			for _, op := range section.Operations {
				visitAction(op, false, visitors)
			}
		case len(section.Functions) > 0:
			for _, op := range section.Functions {
				visitAction(op, true, visitors)
			}
		case section.Attribute != nil:
			visitAttribute(section.Attribute, visitors)
		}
	}
}

func visitField(f *parser.FieldNode, visitors []Visitor) {
	for _, v := range visitors {
		if v.EnterField != nil {
			v.EnterField(f)
		}
	}
	for _, attr := range f.Attributes {
		visitAttribute(attr, visitors)
	}
}

func visitAttribute(n *parser.AttributeNode, visitors []Visitor) {
	for _, v := range visitors {
		if v.EnterAttribute != nil {
			v.EnterAttribute(n)
		}
	}
	for _, arg := range n.Arguments {
		for _, v := range visitors {
			if v.EnterAttributeArgument != nil {
				v.EnterAttributeArgument(arg)
			}
		}
	}
	for _, v := range visitors {
		if v.ExitAttribute != nil {
			v.ExitAttribute(n)
		}
	}
}

func visitAction(n *parser.ActionNode, isFunction bool, visitors []Visitor) {
	for _, v := range visitors {
		if v.EnterAction != nil {
			v.EnterAction(n)
		}
		if isFunction {
			if v.EnterFunction != nil {
				v.EnterFunction(n)
			}
		} else {
			if v.EnterOperation != nil {
				v.EnterOperation(n)
			}
		}
	}

	for _, input := range n.Inputs {
		for _, v := range visitors {
			if v.EnterInput != nil {
				v.EnterInput(input)
			}
			if v.EnterReadInput != nil {
				v.EnterReadInput(input)
			}
		}
	}

	for _, input := range n.With {
		for _, v := range visitors {
			if v.EnterInput != nil {
				v.EnterInput(input)
			}
			if v.EnterWriteInput != nil {
				v.EnterWriteInput(input)
			}
		}
	}

	for _, attr := range n.Attributes {
		for _, v := range visitors {
			if v.EnterAttribute != nil {
				v.EnterAttribute(attr)
			}
		}
	}

	for _, v := range visitors {
		if v.ExitAction != nil {
			v.ExitAction(n)
		}
		if isFunction {
			if v.ExitFunction != nil {
				v.ExitFunction(n)
			}
		} else {
			if v.ExitOperation != nil {
				v.ExitOperation(n)
			}
		}
	}
}
