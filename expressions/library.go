package expressions

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
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
	paramA := types.NewTypeParamType("A")
	paramB := types.NewTypeParamType("B")
	listOfA := types.NewListType(paramA)
	mapOfAB := types.NewMapType(paramA, paramB)

	return []cel.EnvOption{
		// Indexing
		cel.Function(operators.Index,
			decls.Overload(overloads.IndexMap, argTypes(mapOfAB, paramA), paramB),
			decls.Overload(overloads.IndexList, argTypes(listOfA, types.IntType), paramA),
			decls.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				return lhs.(traits.Indexer).Get(rhs)
			}, traits.IndexerType)),
	}
}

func (*standardKeelLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func argTypes(args ...*types.Type) []*types.Type {
	return args
}
