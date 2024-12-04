package options

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

	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/typing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

func WithSchemaTypes(schema []*parser.AST) expressions.Option {
	return func(p *expressions.Parser) error {

		p.Provider.Schema = schema

		var options []cel.EnvOption
		for _, enum := range query.Enums(schema) {
			options = append(options, cel.Constant(enum.Name.Value, types.NewObjectType(fmt.Sprintf("%s_Enum", enum.Name.Value)), nil))
		}

		for _, role := range query.Roles(schema) {
			options = append(options, cel.Constant(role.Name.Value, types.NewOpaqueType("_Role"), nil))
		}

		if options != nil {
			var err error
			p.CelEnv, err = p.CelEnv.Extend(options...)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func WithConstant(value string, typeName string) expressions.Option {
	return func(p *expressions.Parser) error {
		var err error
		p.CelEnv, err = p.CelEnv.Extend(cel.Constant(value, types.NewOpaqueType(typeName), nil))
		if err != nil {
			return err
		}

		return nil
	}
}

func WithCtx() expressions.Option {
	return func(p *expressions.Parser) error {
		fields := map[string]*types.Type{
			"identity":        types.NewObjectType("Identity"),
			"isAuthenticated": types.BoolType,
			"now":             types.TimestampType,
			"secrets":         types.DynType,
			"env":             types.DynType,
			"headers":         types.DynType,
		}

		p.Provider.Objects["Context"] = fields

		var err error
		p.CelEnv, err = p.CelEnv.Extend(cel.Variable("ctx", types.NewObjectType("Context")))
		if err != nil {
			return err
		}

		return nil
	}
}

func WithVariable(identifier string, typeName string) expressions.Option {
	return func(p *expressions.Parser) error {
		t, err := typing.MapType(p.Provider.Schema, typeName)
		if err != nil {
			return err
		}

		env, err := p.CelEnv.Extend(cel.Variable(identifier, t))
		if err != nil {
			return err
		}

		p.CelEnv = env

		return nil
	}
}

func WithActionInputs(schema []*parser.AST, action *parser.ActionNode) expressions.Option {
	return func(p *expressions.Parser) error {
		model := query.ActionModel(schema, action.Name.Value)
		opts := []cel.EnvOption{}

		// Add filter inputs as variables
		for _, f := range action.Inputs {
			typeName := query.ResolveInputType(schema, f, model, action)

			t, err := typing.MapType(p.Provider.Schema, typeName)
			if err != nil {
				return err
			}

			opts = append(opts, cel.Variable(f.Name(), t))
		}

		// Add with inputs as variables
		for _, f := range action.With {
			typeName := query.ResolveInputType(schema, f, model, action)

			t, err := typing.MapType(p.Provider.Schema, typeName)
			if err != nil {
				return err
			}

			opts = append(opts, cel.Variable(f.Name(), t))
		}

		env, err := p.CelEnv.Extend(opts...)
		if err != nil {
			return err
		}

		p.CelEnv = env

		return nil
	}
}

func WithLogicalOperators() expressions.Option {
	return func(p *expressions.Parser) error {
		var err error

		p.CelEnv, err = p.CelEnv.Extend(
			cel.Function(operators.LogicalAnd,
				cel.Overload(overloads.LogicalAnd, []*types.Type{types.BoolType, types.BoolType}, types.BoolType, cel.OverloadIsNonStrict())),
			cel.Function(operators.LogicalOr,
				cel.Overload(overloads.LogicalOr, []*types.Type{types.BoolType, types.BoolType}, types.BoolType, cel.OverloadIsNonStrict())),
			cel.Function(operators.LogicalNot,
				cel.Overload(overloads.LogicalNot, []*types.Type{types.BoolType}, types.BoolType)))
		if err != nil {
			return err
		}

		return nil
	}
}

func WithComparisonOperators() expressions.Option {
	return func(p *expressions.Parser) error {
		paramA := types.NewTypeParamType("A")
		//structA := types.NewObjectType("B")
		//	listA := types.NewListType(paramA)
		//statusEnum := types.NewOpaqueType("Status")
		//s := decls.NewAbstractType("DATE")
		var err error

		if p.Provider.Schema != nil {

			// For each enum type, configure equals, not equals and 'in' operators
			for _, enum := range query.Enums(p.Provider.Schema) {
				enumType := types.NewOpaqueType(enum.Name.Value)
				enumTypeArr := cel.ObjectType(enum.Name.Value + "[]")

				p.CelEnv, err = p.CelEnv.Extend(
					cel.Function(operators.Equals,
						cel.Overload(fmt.Sprintf("equals_%s", strcase.ToLowerCamel(enum.Name.Value)), []*types.Type{enumType, enumType}, types.BoolType),
					),
					// cel.Function(operators.NotEquals,
					// 	cel.Overload(fmt.Sprintf("notequals_%s", strcase.ToLowerCamel(enum.Name.Value)), []*types.Type{enumType, enumType}, types.BoolType),
					// ),
					cel.Function(operators.In,
						cel.Overload(fmt.Sprintf("in_%s", strcase.ToLowerCamel(enum.Name.Value)), []*types.Type{enumType, enumTypeArr}, types.BoolType),
					))
				if err != nil {
					return err
				}
			}

			for _, model := range query.Models(p.Provider.Schema) {
				modelType := types.NewObjectType(model.Name.Value)
				modelTypeArr := cel.ObjectType(model.Name.Value + "[]")

				p.CelEnv, err = p.CelEnv.Extend(
					cel.Function(operators.Equals,
						cel.Overload(fmt.Sprintf("equals_%s", strcase.ToLowerCamel(model.Name.Value)), []*types.Type{modelType, modelType}, types.BoolType),
					),
					cel.Function(operators.In,
						cel.Overload(fmt.Sprintf("in_%s", strcase.ToLowerCamel(model.Name.Value)), []*types.Type{modelType, modelTypeArr}, types.BoolType),
					))
				if err != nil {
					return err
				}
			}
		}

		p.CelEnv, err = p.CelEnv.Extend(
			// Equals
			cel.Function(operators.Equals,
				cel.Overload("equals_string", []*types.Type{types.StringType, types.StringType}, types.BoolType),
				cel.Overload("equals_nullable_string", []*types.Type{types.NewNullableType(types.StringType), types.NullType}, types.BoolType),

				cel.Overload("equals_int", argTypes(types.IntType, types.IntType), types.BoolType),
				cel.Overload("equals_double", argTypes(types.DoubleType, types.DoubleType), types.BoolType),
				cel.Overload("equals_boolean", argTypes(types.BoolType, types.BoolType), types.BoolType),
				cel.Overload("equals_datetime", argTypes(types.TimestampType, types.TimestampType), types.BoolType),
				cel.Overload("equals_string_list", argTypes(types.NewListType(types.StringType), types.NewListType(types.StringType)), types.BoolType),
				cel.Overload("equals_int_list", argTypes(types.NewListType(types.IntType), types.NewListType(types.IntType)), types.BoolType),
				cel.Overload("equals_double_list", argTypes(types.NewListType(types.DoubleType), types.NewListType(types.DoubleType)), types.BoolType),
				cel.Overload("equals_bool_list", argTypes(types.NewListType(types.BoolType), types.NewListType(types.BoolType)), types.BoolType),
			),

			//TODO:the rest
			cel.Function(operators.NotEquals,
				cel.Overload(overloads.NotEquals, argTypes(paramA, paramA), types.BoolType)),

			// // In
			// cel.Function(operators.In,
			// 	cel.Overload("in_string", argTypes(types.StringType, types.NewListType(types.StringType)), types.BoolType),
			// 	cel.Overload("in_int", argTypes(types.IntType, types.NewListType(types.IntType)), types.BoolType),
			// 	cel.Overload("in_double", argTypes(types.DoubleType, types.NewListType(types.DoubleType)), types.BoolType),
			// 	cel.Overload("in_boolean", argTypes(types.BoolType, types.BoolType), types.BoolType),
			// ),

			// In
			cel.Function(operators.In,
				cel.Overload("in", argTypes(paramA, types.NewListType(paramA)), types.BoolType),

				//cel.Overload("in_string2", argTypes(structA, types.NewListType(structA)), types.BoolType),
				// cel.Overload("in_int", argTypes(types.IntType, types.NewListType(types.IntType)), types.BoolType),
				// cel.Overload("in_double", argTypes(types.DoubleType, types.NewListType(types.DoubleType)), types.BoolType),
				// cel.Overload("in_boolean", argTypes(types.BoolType, types.BoolType), types.BoolType),
			),

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

func WithArithmeticOperators() expressions.Option {
	return func(p *expressions.Parser) error {
		var err error

		p.CelEnv, err = p.CelEnv.Extend(
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

func WithFunctions() expressions.Option {
	return func(p *expressions.Parser) error {
		var err error

		p.CelEnv, err = p.CelEnv.Extend(
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

func WithReturnTypeAssertion(returnType string, asArray bool) expressions.Option {
	return func(p *expressions.Parser) error {
		var err error
		p.ExpectedReturnType, err = typing.MapType(p.Provider.Schema, returnType)

		if asArray {
			p.ExpectedReturnType = cel.ListType(p.ExpectedReturnType)
		}
		return err
	}
}

func argTypes(args ...*types.Type) []*types.Type {
	return args
}
