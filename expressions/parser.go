package orderby_expression

import (
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"

	"github.com/google/cel-go/checker/decls"
	d "github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/google/cel-go/common/types"

	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

type ExpressionParser struct {
	env                *cel.Env
	ast                *cel.Ast
	provider           *typeProvider
	expectedReturnType *types.Type
}

type Option func(*ExpressionParser) error

func WithSchema(schema []*parser.AST) Option {
	return func(p *ExpressionParser) error {
		fields := map[string]*types.Type{}

		for _, m := range query.Models(schema) {
			model := query.Model(schema, m.Name.Value)

			for _, f := range query.ModelFields(model) {
				fields[f.Name.Value] = mapType(f)
			}

			p.provider.context[m.Name.Value] = fields
		}

		p.provider.asts = schema

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

		p.provider.context["Context"] = fields

		var err error
		p.env, err = p.env.Extend(cel.Declarations(decls.NewVar("ctx", decls.NewObjectType("Context"))))
		if err != nil {
			return err
		}

		return nil
	}
}

func WithVariable(identifier string, typeName string) Option {
	return func(p *ExpressionParser) error {
		if p.provider.context[typeName] == nil {
			return fmt.Errorf("the type '%s' does not exist in the providers context", typeName)
		}

		env, err := p.env.Extend(cel.Declarations(decls.NewVar(identifier, decls.NewObjectType(typeName))))
		if err != nil {
			return err
		}

		p.env = env

		return nil
	}
}

func WithLogicalOperators() Option {
	return func(p *ExpressionParser) error {
		var err error

		p.env, err = p.env.Extend(
			cel.Function(operators.LogicalAnd,
				d.Overload(overloads.LogicalAnd, argTypes(types.BoolType, types.BoolType), types.BoolType, d.OverloadIsNonStrict())),
			cel.Function(operators.LogicalOr,
				d.Overload(overloads.LogicalOr, argTypes(types.BoolType, types.BoolType), types.BoolType, d.OverloadIsNonStrict())),
			cel.Function(operators.LogicalNot,
				d.Overload(overloads.LogicalNot, argTypes(types.BoolType), types.BoolType)))
		if err != nil {
			return err
		}

		return nil
	}
}

func WithComparisonOperators() Option {
	return func(p *ExpressionParser) error {
		paramA := types.NewTypeParamType("A")
		var err error

		p.env, err = p.env.Extend(
			cel.Function(operators.Equals,
				d.Overload(overloads.Equals, argTypes(paramA, paramA), types.BoolType),
				d.SingletonBinaryBinding(noBinaryOverrides)),
			cel.Function(operators.NotEquals,
				d.Overload(overloads.NotEquals, argTypes(paramA, paramA), types.BoolType)),
			// Greater
			cel.Function(operators.Greater,
				d.Overload(overloads.GreaterInt64,
					argTypes(types.IntType, types.IntType), types.BoolType),
				d.Overload(overloads.GreaterInt64Double,
					argTypes(types.IntType, types.DoubleType), types.BoolType),
				d.Overload(overloads.GreaterInt64Uint64,
					argTypes(types.IntType, types.UintType), types.BoolType),
				d.Overload(overloads.GreaterUint64,
					argTypes(types.UintType, types.UintType), types.BoolType),
				d.Overload(overloads.GreaterUint64Double,
					argTypes(types.UintType, types.DoubleType), types.BoolType),
				d.Overload(overloads.GreaterUint64Int64,
					argTypes(types.UintType, types.IntType), types.BoolType),
				d.Overload(overloads.GreaterDouble,
					argTypes(types.DoubleType, types.DoubleType), types.BoolType),
				d.Overload(overloads.GreaterDoubleInt64,
					argTypes(types.DoubleType, types.IntType), types.BoolType),
				d.Overload(overloads.GreaterDoubleUint64,
					argTypes(types.DoubleType, types.UintType), types.BoolType),
				d.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
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
				d.Overload(overloads.GreaterEqualsInt64,
					argTypes(types.IntType, types.IntType), types.BoolType),
				d.Overload(overloads.GreaterEqualsInt64Double,
					argTypes(types.IntType, types.DoubleType), types.BoolType),
				d.Overload(overloads.GreaterEqualsInt64Uint64,
					argTypes(types.IntType, types.UintType), types.BoolType),
				d.Overload(overloads.GreaterEqualsUint64,
					argTypes(types.UintType, types.UintType), types.BoolType),
				d.Overload(overloads.GreaterEqualsUint64Double,
					argTypes(types.UintType, types.DoubleType), types.BoolType),
				d.Overload(overloads.GreaterEqualsUint64Int64,
					argTypes(types.UintType, types.IntType), types.BoolType),
				d.Overload(overloads.GreaterEqualsDouble,
					argTypes(types.DoubleType, types.DoubleType), types.BoolType),
				d.Overload(overloads.GreaterEqualsDoubleInt64,
					argTypes(types.DoubleType, types.IntType), types.BoolType),
				d.Overload(overloads.GreaterEqualsDoubleUint64,
					argTypes(types.DoubleType, types.UintType), types.BoolType),
				d.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
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
				d.Overload(overloads.LessInt64,
					argTypes(types.IntType, types.IntType), types.BoolType),
				d.Overload(overloads.LessInt64Double,
					argTypes(types.IntType, types.DoubleType), types.BoolType),
				d.Overload(overloads.LessInt64Uint64,
					argTypes(types.IntType, types.UintType), types.BoolType),
				d.Overload(overloads.LessUint64,
					argTypes(types.UintType, types.UintType), types.BoolType),
				d.Overload(overloads.LessUint64Double,
					argTypes(types.UintType, types.DoubleType), types.BoolType),
				d.Overload(overloads.LessUint64Int64,
					argTypes(types.UintType, types.IntType), types.BoolType),
				d.Overload(overloads.LessDouble,
					argTypes(types.DoubleType, types.DoubleType), types.BoolType),
				d.Overload(overloads.LessDoubleInt64,
					argTypes(types.DoubleType, types.IntType), types.BoolType),
				d.Overload(overloads.LessDoubleUint64,
					argTypes(types.DoubleType, types.UintType), types.BoolType),
				d.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
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
				d.Overload(overloads.LessEqualsInt64,
					argTypes(types.IntType, types.IntType), types.BoolType),
				d.Overload(overloads.LessEqualsInt64Double,
					argTypes(types.IntType, types.DoubleType), types.BoolType),
				d.Overload(overloads.LessEqualsInt64Uint64,
					argTypes(types.IntType, types.UintType), types.BoolType),
				d.Overload(overloads.LessEqualsUint64,
					argTypes(types.UintType, types.UintType), types.BoolType),
				d.Overload(overloads.LessEqualsUint64Double,
					argTypes(types.UintType, types.DoubleType), types.BoolType),
				d.Overload(overloads.LessEqualsUint64Int64,
					argTypes(types.UintType, types.IntType), types.BoolType),
				d.Overload(overloads.LessEqualsDouble,
					argTypes(types.DoubleType, types.DoubleType), types.BoolType),
				d.Overload(overloads.LessEqualsDoubleInt64,
					argTypes(types.DoubleType, types.IntType), types.BoolType),
				d.Overload(overloads.LessEqualsDoubleUint64,
					argTypes(types.DoubleType, types.UintType), types.BoolType),
				d.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
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

		p.env, err = p.env.Extend(
			// Addition operator
			cel.Function(operators.Add,
				d.Overload(overloads.AddString,
					argTypes(types.StringType, types.StringType), types.StringType),
				d.Overload(overloads.AddDouble,
					argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
				d.Overload(overloads.AddInt64,
					argTypes(types.IntType, types.IntType), types.IntType),
				d.Overload(overloads.AddUint64,
					argTypes(types.UintType, types.UintType), types.UintType),
				d.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					return lhs.(traits.Adder).Add(rhs)
				}, traits.AdderType)),

			// Subtraction operator
			cel.Function(operators.Subtract,
				d.Overload(overloads.SubtractDouble,
					argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
				d.Overload(overloads.SubtractInt64,
					argTypes(types.IntType, types.IntType), types.IntType),
				d.Overload(overloads.SubtractUint64,
					argTypes(types.UintType, types.UintType), types.UintType),
				d.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					return lhs.(traits.Subtractor).Subtract(rhs)
				}, traits.SubtractorType)),

			// Multiplication
			cel.Function(operators.Multiply,
				d.Overload(overloads.MultiplyDouble,
					argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
				d.Overload(overloads.MultiplyInt64,
					argTypes(types.IntType, types.IntType), types.IntType),
				d.Overload(overloads.MultiplyUint64,
					argTypes(types.UintType, types.UintType), types.UintType),
				d.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
					return lhs.(traits.Multiplier).Multiply(rhs)
				}, traits.MultiplierType)),

			// Division
			cel.Function(operators.Divide,
				d.Overload(overloads.DivideDouble,
					argTypes(types.DoubleType, types.DoubleType), types.DoubleType),
				d.Overload(overloads.DivideInt64,
					argTypes(types.IntType, types.IntType), types.IntType),
				d.Overload(overloads.DivideUint64,
					argTypes(types.UintType, types.UintType), types.UintType),
				d.SingletonBinaryBinding(func(lhs, rhs ref.Val) ref.Val {
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

		p.env, err = p.env.Extend(
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
		p.expectedReturnType = mapType(&parser.FieldNode{Type: parser.NameNode{Value: returnType}})

		return nil
	}
}

func NewParser(options ...Option) (*ExpressionParser, error) {
	typeProvider := NewTypeProvider()

	env, err := cel.NewCustomEnv(
		standardKeelLibrary(),
		cel.ClearMacros(),
		cel.CustomTypeProvider(typeProvider),
		cel.EagerlyValidateDeclarations(true),
	)
	if err != nil {
		return nil, fmt.Errorf("program setup err: %s", err)
	}

	parser := &ExpressionParser{
		env:      env,
		provider: typeProvider,
	}

	for _, opt := range options {
		if err := opt(parser); err != nil {
			return nil, err
		}
	}

	return parser, nil
}

// Validate parses and validates the expression
func (p *ExpressionParser) Validate(expression string) ([]string, error) {
	ast, issues := p.env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		validationErrors := []string{}
		for _, e := range issues.Errors() {
			validationErrors = append(validationErrors, e.Message)
		}
		return validationErrors, nil
	}

	if p.expectedReturnType != nil {
		if ast.OutputType() != p.expectedReturnType {
			return []string{fmt.Sprintf("expression expected to resolve to type '%s'", p.expectedReturnType.String())}, nil
		}
	}

	p.ast = ast

	// Valid expression
	return nil, nil
}
