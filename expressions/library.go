package expressions

import (
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

// Our library of syntax support
type keelLibrary struct{}

var _ cel.Library = new(keelLibrary)

func KeelLib() cel.EnvOption {
	return cel.Lib(&keelLibrary{})
}

// LibraryName implements the SingletonLibrary interface method.
func (*keelLibrary) LibraryName() string {
	return "keel"
}

// CompileOptions returns options for the standard CEL function declarations and macros.
func (l *keelLibrary) CompileOptions() []cel.EnvOption {
	paramA := types.NewTypeParamType("A")
	paramB := types.NewTypeParamType("B")
	listOfA := types.NewListType(paramA)
	mapOfAB := types.NewMapType(paramA, paramB)

	return []cel.EnvOption{

		// Logical operators
		cel.Function(operators.Conditional,
			decls.Overload(overloads.Conditional, argTypes(types.BoolType, paramA, paramA), paramA, decls.OverloadIsNonStrict())),
		cel.Function(operators.LogicalAnd,
			decls.Overload(overloads.LogicalAnd, argTypes(types.BoolType, types.BoolType), types.BoolType, decls.OverloadIsNonStrict())),
		cel.Function(operators.LogicalOr,
			decls.Overload(overloads.LogicalOr, argTypes(types.BoolType, types.BoolType), types.BoolType, decls.OverloadIsNonStrict())),
		cel.Function(operators.LogicalNot,
			decls.Overload(overloads.LogicalNot, argTypes(types.BoolType), types.BoolType)),

		// Equals
		cel.Function(operators.Equals,
			decls.Overload(overloads.Equals, argTypes(paramA, paramA), types.BoolType),
			decls.SingletonBinaryBinding(noBinaryOverrides)),

		cel.Function(operators.NotEquals,
			decls.Overload(overloads.NotEquals, argTypes(paramA, paramA), types.BoolType)),

		// Addition operator
		cel.Function(operators.Add,
			decls.Overload(overloads.AddString,
				argTypes(types.StringType, types.StringType), types.StringType),
			decls.Overload(overloads.AddDouble,
				argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
			decls.Overload(overloads.AddInt64,
				argTypes(types.IntType, types.IntType), types.IntType),
			decls.Overload(overloads.AddUint64,
				argTypes(types.UintType, types.UintType), types.UintType),
			decls.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				return lhs.(traits.Adder).Add(rhs)
			}, traits.AdderType)),

		// Subtraction operator
		cel.Function(operators.Subtract,
			decls.Overload(overloads.SubtractDouble,
				argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
			decls.Overload(overloads.SubtractInt64,
				argTypes(types.IntType, types.IntType), types.IntType),
			decls.Overload(overloads.SubtractUint64,
				argTypes(types.UintType, types.UintType), types.UintType),
			decls.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				return lhs.(traits.Subtractor).Subtract(rhs)
			}, traits.SubtractorType)),

		// Multiplication
		cel.Function(operators.Multiply,
			decls.Overload(overloads.MultiplyDouble,
				argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
			decls.Overload(overloads.MultiplyInt64,
				argTypes(types.IntType, types.IntType), types.IntType),
			decls.Overload(overloads.MultiplyUint64,
				argTypes(types.UintType, types.UintType), types.UintType),
			decls.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				return lhs.(traits.Multiplier).Multiply(rhs)
			}, traits.MultiplierType)),

		// Division
		cel.Function(operators.Divide,
			decls.Overload(overloads.DivideDouble,
				argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
			decls.Overload(overloads.DivideInt64,
				argTypes(types.IntType, types.IntType), types.IntType),
			decls.Overload(overloads.DivideUint64,
				argTypes(types.UintType, types.UintType), types.UintType),
			decls.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				return lhs.(traits.Divider).Divide(rhs)
			}, traits.DividerType)),

		// Negation
		cel.Function(operators.Negate,
			decls.Overload(overloads.NegateDouble, argTypes(types.DoubleType), types.DoubleType),
			decls.Overload(overloads.NegateInt64, argTypes(types.IntType), types.IntType),
			decls.SingletonUnaryBinding(func(val ref.Val) ref.Val {
				if types.IsBool(val) {
					return types.MaybeNoSuchOverloadErr(val)
				}
				return val.(traits.Negater).Negate()
			}, traits.NegatorType)),

		// Greater
		cel.Function(operators.Greater,
			decls.Overload(overloads.GreaterInt64,
				argTypes(types.IntType, types.IntType), types.BoolType),
			decls.Overload(overloads.GreaterInt64Double,
				argTypes(types.IntType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.GreaterInt64Uint64,
				argTypes(types.IntType, types.UintType), types.BoolType),
			decls.Overload(overloads.GreaterUint64,
				argTypes(types.UintType, types.UintType), types.BoolType),
			decls.Overload(overloads.GreaterUint64Double,
				argTypes(types.UintType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.GreaterUint64Int64,
				argTypes(types.UintType, types.IntType), types.BoolType),
			decls.Overload(overloads.GreaterDouble,
				argTypes(types.DoubleType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.GreaterDoubleInt64,
				argTypes(types.DoubleType, types.IntType), types.BoolType),
			decls.Overload(overloads.GreaterDoubleUint64,
				argTypes(types.DoubleType, types.UintType), types.BoolType),
			decls.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				cmp := lhs.(traits.Comparer).Compare(rhs)
				if cmp == types.IntOne {
					return types.True
				}
				if cmp == types.IntNegOne || cmp == types.IntZero {
					return types.False
				}
				return cmp
			}, traits.ComparerType)),

		// Greater or equals
		cel.Function(operators.GreaterEquals,
			decls.Overload(overloads.GreaterEqualsInt64,
				argTypes(types.IntType, types.IntType), types.BoolType),
			decls.Overload(overloads.GreaterEqualsInt64Double,
				argTypes(types.IntType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.GreaterEqualsInt64Uint64,
				argTypes(types.IntType, types.UintType), types.BoolType),
			decls.Overload(overloads.GreaterEqualsUint64,
				argTypes(types.UintType, types.UintType), types.BoolType),
			decls.Overload(overloads.GreaterEqualsUint64Double,
				argTypes(types.UintType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.GreaterEqualsUint64Int64,
				argTypes(types.UintType, types.IntType), types.BoolType),
			decls.Overload(overloads.GreaterEqualsDouble,
				argTypes(types.DoubleType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.GreaterEqualsDoubleInt64,
				argTypes(types.DoubleType, types.IntType), types.BoolType),
			decls.Overload(overloads.GreaterEqualsDoubleUint64,
				argTypes(types.DoubleType, types.UintType), types.BoolType),
			decls.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				cmp := lhs.(traits.Comparer).Compare(rhs)
				if cmp == types.IntOne || cmp == types.IntZero {
					return types.True
				}
				if cmp == types.IntNegOne {
					return types.False
				}
				return cmp
			}, traits.ComparerType)),

		// Less
		cel.Function(operators.Less,
			decls.Overload(overloads.LessInt64,
				argTypes(types.IntType, types.IntType), types.BoolType),
			decls.Overload(overloads.LessInt64Double,
				argTypes(types.IntType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.LessInt64Uint64,
				argTypes(types.IntType, types.UintType), types.BoolType),
			decls.Overload(overloads.LessUint64,
				argTypes(types.UintType, types.UintType), types.BoolType),
			decls.Overload(overloads.LessUint64Double,
				argTypes(types.UintType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.LessUint64Int64,
				argTypes(types.UintType, types.IntType), types.BoolType),
			decls.Overload(overloads.LessDouble,
				argTypes(types.DoubleType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.LessDoubleInt64,
				argTypes(types.DoubleType, types.IntType), types.BoolType),
			decls.Overload(overloads.LessDoubleUint64,
				argTypes(types.DoubleType, types.UintType), types.BoolType),
			decls.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				cmp := lhs.(traits.Comparer).Compare(rhs)
				if cmp == types.IntNegOne {
					return types.True
				}
				if cmp == types.IntOne || cmp == types.IntZero {
					return types.False
				}
				return cmp
			}, traits.ComparerType)),

		// Less or equals
		cel.Function(operators.LessEquals,
			decls.Overload(overloads.LessEqualsInt64,
				argTypes(types.IntType, types.IntType), types.BoolType),
			decls.Overload(overloads.LessEqualsInt64Double,
				argTypes(types.IntType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.LessEqualsInt64Uint64,
				argTypes(types.IntType, types.UintType), types.BoolType),
			decls.Overload(overloads.LessEqualsUint64,
				argTypes(types.UintType, types.UintType), types.BoolType),
			decls.Overload(overloads.LessEqualsUint64Double,
				argTypes(types.UintType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.LessEqualsUint64Int64,
				argTypes(types.UintType, types.IntType), types.BoolType),
			decls.Overload(overloads.LessEqualsDouble,
				argTypes(types.DoubleType, types.DoubleType), types.BoolType),
			decls.Overload(overloads.LessEqualsDoubleInt64,
				argTypes(types.DoubleType, types.IntType), types.BoolType),
			decls.Overload(overloads.LessEqualsDoubleUint64,
				argTypes(types.DoubleType, types.UintType), types.BoolType),
			decls.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				cmp := lhs.(traits.Comparer).Compare(rhs)
				if cmp == types.IntNegOne || cmp == types.IntZero {
					return types.True
				}
				if cmp == types.IntOne {
					return types.False
				}
				return cmp
			}, traits.ComparerType)),

		// Indexing
		cel.Function(operators.Index,
			decls.Overload(overloads.IndexMap, argTypes(mapOfAB, paramA), paramB),
			decls.Overload(overloads.IndexList, argTypes(listOfA, types.IntType), paramA),
			decls.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
				return lhs.(traits.Indexer).Get(rhs)
			}, traits.IndexerType)),

		// UPPER(string) custom global function
		cel.Function("upper",
			cel.Overload("upper_string",
				[]*cel.Type{cel.StringType},
				cel.StringType,
				cel.UnaryBinding(func(lhs ref.Val) ref.Val {
					return types.String(strings.ToUpper(fmt.Sprintf("%s", lhs)))
				}),
			),
		),
	}
}

func (*keelLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func argTypes(args ...*types.Type) []*types.Type {
	return args
}

func noBinaryOverrides(rhs, lhs ref.Val) ref.Val {
	return types.NoSuchOverloadErr()
}

func noFunctionOverrides(args ...ref.Val) ref.Val {
	return types.NoSuchOverloadErr()
}
