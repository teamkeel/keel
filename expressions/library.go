package expressions

import (
	"github.com/google/cel-go/cel"
)

type standardKeelLib struct{}

var _ cel.Library = new(standardKeelLib)

func standardKeelLibrary() cel.EnvOption {
	return cel.Lib(&standardKeelLib{})
}

func (*standardKeelLib) LibraryName() string {
	return "keel"
}

func (l *standardKeelLib) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		// cel.Function(operators.LogicalNot,
		// 	decls.Overload(overloads.LogicalNot, []*types.Type{types.BoolType, types.BoolType}, types.BoolType)),
	}
}

func (*standardKeelLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
