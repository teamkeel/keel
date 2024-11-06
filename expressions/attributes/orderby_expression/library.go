package orderby_expression

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

type orderByExpressionLibrary struct{}

var _ cel.Library = new(orderByExpressionLibrary)

func orderByAttributeLib() cel.EnvOption {
	return cel.Lib(&orderByExpressionLibrary{})
}

func (*orderByExpressionLibrary) LibraryName() string {
	return "keel"
}

func (l *orderByExpressionLibrary) CompileOptions() []cel.EnvOption {

	return []cel.EnvOption{}
}

func (*orderByExpressionLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func argTypes(args ...*types.Type) []*types.Type {
	return args
}

func noBinaryOverrides(rhs, lhs ref.Val) ref.Val {
	return types.NoSuchOverloadErr()
}
