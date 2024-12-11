package expressions

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
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
	return []cel.EnvOption{}
}

func (*standardKeelLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func argTypes(args ...*types.Type) []*types.Type {
	return args
}
