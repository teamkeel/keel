package expressions

import (
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

type Option func(*ExpressionParser) error

func WithSchema(schema []*parser.AST) Option {
	return func(p *ExpressionParser) error {
		p.provider.schema = schema

		var options []cel.EnvOption
		for _, enum := range query.Enums(schema) {
			options = append(options, cel.Constant(enum.Name.Value, types.NewObjectType(fmt.Sprintf("%s_EnumDefinition", enum.Name.Value)), nil))
		}

		if options != nil {
			var err error
			p.celEnv, err = p.celEnv.Extend(options...)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func WithCtx() Option {
	return func(p *ExpressionParser) error {
		fields := map[string]*types.Type{
			"identity":        types.NewObjectType("Identity"),
			"isAuthenticated": types.BoolType,
			"headers":         types.NewMapType(types.StringType, types.StringType),
			"now":             types.TimestampType,
			"secrets":         types.DynType,
		}

		p.provider.objects["Context"] = fields

		var err error
		p.celEnv, err = p.celEnv.Extend(cel.Variable("ctx", types.NewObjectType("Context")))
		if err != nil {
			return err
		}

		return nil
	}
}

func WithVariable(identifier string, typeName string) Option {
	return func(p *ExpressionParser) error {
		t, err := mapType(p.provider.schema, typeName)
		if err != nil {
			return err
		}

		env, err := p.celEnv.Extend(cel.Variable(identifier, t))
		if err != nil {
			return err
		}

		p.celEnv = env

		return nil
	}
}

func WithLogicalOperators() Option {
	return func(p *ExpressionParser) error {
		var err error

		p.celEnv, err = p.celEnv.Extend(
			cel.Function(operators.LogicalAnd,
				cel.Overload(overloads.LogicalAnd, argTypes(types.BoolType, types.BoolType), types.BoolType, cel.OverloadIsNonStrict())),
			cel.Function(operators.LogicalOr,
				cel.Overload(overloads.LogicalOr, argTypes(types.BoolType, types.BoolType), types.BoolType, cel.OverloadIsNonStrict())),
			cel.Function(operators.LogicalNot,
				cel.Overload(overloads.LogicalNot, argTypes(types.BoolType), types.BoolType)))
		if err != nil {
			return err
		}

		return nil
	}
}

func WithComparisonOperators() Option {
	return func(p *ExpressionParser) error {
		paramA := types.NewTypeParamType("A")
		//statusEnum := types.NewOpaqueType("Status")
		//s := decls.NewAbstractType("DATE")
		var err error

		if p.provider.schema != nil {
			for _, enum := range query.Enums(p.provider.schema) {
				enumType := types.NewOpaqueType(enum.Name.Value)
				p.celEnv, err = p.celEnv.Extend(
					cel.Function(operators.Equals,
						cel.Overload(fmt.Sprintf("equals_%s", strcase.ToLowerCamel(enum.Name.Value)), argTypes(enumType, enumType), types.BoolType),
					))
			}

			for _, model := range query.Models(p.provider.schema) {
				modelType := types.NewObjectType(model.Name.Value)
				p.celEnv, err = p.celEnv.Extend(
					cel.Function(operators.Equals,
						cel.Overload(fmt.Sprintf("equals_%s", strcase.ToLowerCamel(model.Name.Value)), argTypes(modelType, modelType), types.BoolType),
					))
			}

		}

		p.celEnv, err = p.celEnv.Extend(
			cel.Function(operators.Equals,
				cel.Overload("equals_string", argTypes(types.StringType, types.StringType), types.BoolType),
				cel.Overload("equals_int", argTypes(types.IntType, types.IntType), types.BoolType),
				cel.Overload("equals_uint", argTypes(types.UintType, types.UintType), types.BoolType),
				cel.Overload("equals_double", argTypes(types.DoubleType, types.DoubleType), types.BoolType),
				cel.Overload("equals_boolean", argTypes(types.BoolType, types.BoolType), types.BoolType),
			),
			cel.Function(operators.NotEquals,
				cel.Overload(overloads.NotEquals, argTypes(paramA, paramA), types.BoolType)),
			// Greater
			cel.Function(operators.Greater,
				cel.Overload(overloads.GreaterInt64, argTypes(types.IntType, types.IntType), types.BoolType),
				cel.Overload(overloads.GreaterInt64Double, argTypes(types.IntType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.GreaterInt64Uint64, argTypes(types.IntType, types.UintType), types.BoolType),
				cel.Overload(overloads.GreaterUint64, argTypes(types.UintType, types.UintType), types.BoolType),
				cel.Overload(overloads.GreaterUint64Double, argTypes(types.UintType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.GreaterUint64Int64, argTypes(types.UintType, types.IntType), types.BoolType),
				cel.Overload(overloads.GreaterDouble, argTypes(types.DoubleType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.GreaterDoubleInt64, argTypes(types.DoubleType, types.IntType), types.BoolType),
				cel.Overload(overloads.GreaterDoubleUint64, argTypes(types.DoubleType, types.UintType), types.BoolType),
				cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
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
				cel.Overload(overloads.GreaterEqualsInt64, argTypes(types.IntType, types.IntType), types.BoolType),
				cel.Overload(overloads.GreaterEqualsInt64Double, argTypes(types.IntType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.GreaterEqualsInt64Uint64, argTypes(types.IntType, types.UintType), types.BoolType),
				cel.Overload(overloads.GreaterEqualsUint64, argTypes(types.UintType, types.UintType), types.BoolType),
				cel.Overload(overloads.GreaterEqualsUint64Double, argTypes(types.UintType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.GreaterEqualsUint64Int64, argTypes(types.UintType, types.IntType), types.BoolType),
				cel.Overload(overloads.GreaterEqualsDouble,
					argTypes(types.DoubleType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.GreaterEqualsDoubleInt64,
					argTypes(types.DoubleType, types.IntType), types.BoolType),
				cel.Overload(overloads.GreaterEqualsDoubleUint64,
					argTypes(types.DoubleType, types.UintType), types.BoolType),
				cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
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
				cel.Overload(overloads.LessInt64,
					argTypes(types.IntType, types.IntType), types.BoolType),
				cel.Overload(overloads.LessInt64Double,
					argTypes(types.IntType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.LessInt64Uint64,
					argTypes(types.IntType, types.UintType), types.BoolType),
				cel.Overload(overloads.LessUint64,
					argTypes(types.UintType, types.UintType), types.BoolType),
				cel.Overload(overloads.LessUint64Double,
					argTypes(types.UintType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.LessUint64Int64,
					argTypes(types.UintType, types.IntType), types.BoolType),
				cel.Overload(overloads.LessDouble,
					argTypes(types.DoubleType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.LessDoubleInt64,
					argTypes(types.DoubleType, types.IntType), types.BoolType),
				cel.Overload(overloads.LessDoubleUint64,
					argTypes(types.DoubleType, types.UintType), types.BoolType),
				cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
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
				cel.Overload(overloads.LessEqualsInt64,
					argTypes(types.IntType, types.IntType), types.BoolType),
				cel.Overload(overloads.LessEqualsInt64Double,
					argTypes(types.IntType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.LessEqualsInt64Uint64,
					argTypes(types.IntType, types.UintType), types.BoolType),
				cel.Overload(overloads.LessEqualsUint64,
					argTypes(types.UintType, types.UintType), types.BoolType),
				cel.Overload(overloads.LessEqualsUint64Double,
					argTypes(types.UintType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.LessEqualsUint64Int64,
					argTypes(types.UintType, types.IntType), types.BoolType),
				cel.Overload(overloads.LessEqualsDouble,
					argTypes(types.DoubleType, types.DoubleType), types.BoolType),
				cel.Overload(overloads.LessEqualsDoubleInt64,
					argTypes(types.DoubleType, types.IntType), types.BoolType),
				cel.Overload(overloads.LessEqualsDoubleUint64,
					argTypes(types.DoubleType, types.UintType), types.BoolType),
				cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					cmp := lhs.(traits.Comparer).Compare(rhs)
					if cmp == types.IntNegOne || cmp == types.IntZero {
						return types.True
					}
					if cmp == types.IntOne {
						return types.False
					}
					return cmp
				}, traits.ComparerType)))
		if err != nil {
			return err
		}

		return nil
	}
}

func WithArithmeticOperators() Option {
	return func(p *ExpressionParser) error {
		var err error

		p.celEnv, err = p.celEnv.Extend(
			// Addition operator
			cel.Function(operators.Add,
				cel.Overload(overloads.AddString,
					argTypes(types.StringType, types.StringType), types.StringType),
				cel.Overload(overloads.AddDouble,
					argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
				cel.Overload(overloads.AddInt64,
					argTypes(types.IntType, types.IntType), types.IntType),
				cel.Overload(overloads.AddUint64,
					argTypes(types.UintType, types.UintType), types.UintType),
				cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					return lhs.(traits.Adder).Add(rhs)
				}, traits.AdderType)),

			// Subtraction operator
			cel.Function(operators.Subtract,
				cel.Overload(overloads.SubtractDouble,
					argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
				cel.Overload(overloads.SubtractInt64,
					argTypes(types.IntType, types.IntType), types.IntType),
				cel.Overload(overloads.SubtractUint64,
					argTypes(types.UintType, types.UintType), types.UintType),
				cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					return lhs.(traits.Subtractor).Subtract(rhs)
				}, traits.SubtractorType)),

			// Multiplication
			cel.Function(operators.Multiply,
				cel.Overload(overloads.MultiplyDouble,
					argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
				cel.Overload(overloads.MultiplyInt64,
					argTypes(types.IntType, types.IntType), types.IntType),
				cel.Overload(overloads.MultiplyUint64,
					argTypes(types.UintType, types.UintType), types.UintType),
				cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					return lhs.(traits.Multiplier).Multiply(rhs)
				}, traits.MultiplierType)),

			// Division
			cel.Function(operators.Divide,
				cel.Overload(overloads.DivideDouble,
					argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
				cel.Overload(overloads.DivideInt64,
					argTypes(types.IntType, types.IntType), types.IntType),
				cel.Overload(overloads.DivideUint64,
					argTypes(types.UintType, types.UintType), types.UintType),
				cel.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					return lhs.(traits.Divider).Divide(rhs)
				}, traits.DividerType)),
		)
		if err != nil {
			return err
		}

		return nil
	}
}

func WithFunctions() Option {
	return func(p *ExpressionParser) error {
		var err error

		p.celEnv, err = p.celEnv.Extend(
			// UPPER(string) custom global function
			cel.Function("UPPER",
				cel.Overload("upper_string",
					[]*cel.Type{cel.StringType},
					cel.StringType,
					cel.UnaryBinding(func(lhs ref.Val) ref.Val {
						return types.String(strings.ToUpper(fmt.Sprintf("%s", lhs)))
					}),
				),
			),
		)
		if err != nil {
			return err
		}

		return nil
	}
}

func WithReturnTypeAssertion(returnType string) Option {
	return func(p *ExpressionParser) error {
		var err error
		p.expectedReturnType, err = mapType(p.provider.schema, returnType)
		return err
	}
}
