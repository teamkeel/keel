package expressions

import (
	"github.com/google/cel-go/cel"
)

type standardKeelLib struct{}

var _ cel.Library = new(standardKeelLib)

func standardKeelLibrary() cel.EnvOption {
	return cel.Lib(&standardKeelLib{})
}

// LibraryName returns our standard library for expressions.
func (*standardKeelLib) LibraryName() string {
	return "keel"
}

func (l *standardKeelLib) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		// Define any globally configured CEL options here
	}
}

func (*standardKeelLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
