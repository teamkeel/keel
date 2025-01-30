package options

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/google/cel-go/common/types"
	"github.com/iancoleman/strcase"

	"github.com/teamkeel/keel/expressions"
	"github.com/teamkeel/keel/expressions/typing"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
)

// Defines which types are compatible with each other for each comparison operator
// This is used to generate all the necessary combinations of operator overloads
var typeCompatibilityMapping = map[string][][]*types.Type{
	operators.Equals: {
		{types.StringType, typing.Text, typing.ID, typing.Markdown},
		{types.IntType, types.DoubleType, typing.Number, typing.Decimal},
		{typing.Date, typing.Timestamp, types.TimestampType},
		{typing.Boolean, types.BoolType},
		{types.NewListType(types.StringType), typing.TextArray, typing.IDArray, typing.MarkdownArray},
		{types.NewListType(types.IntType), types.NewListType(types.DoubleType), typing.NumberArray, typing.DecimalArray},
	},
	operators.NotEquals: {
		{types.StringType, typing.Text, typing.ID, typing.Markdown},
		{types.IntType, types.DoubleType, typing.Number, typing.Decimal},
		{typing.Date, typing.Timestamp, types.TimestampType},
		{typing.Boolean, types.BoolType},
		{types.NewListType(types.StringType), typing.TextArray, typing.IDArray, typing.MarkdownArray},
		{types.NewListType(types.IntType), types.NewListType(types.DoubleType), typing.NumberArray, typing.DecimalArray},
	},
	operators.Greater: {
		{types.IntType, types.DoubleType, typing.Number, typing.Decimal},
		{typing.Date, typing.Timestamp, types.TimestampType},
	},
	operators.GreaterEquals: {
		{types.IntType, types.DoubleType, typing.Number, typing.Decimal},
		{typing.Date, typing.Timestamp, types.TimestampType},
	},
	operators.Less: {
		{types.IntType, types.DoubleType, typing.Number, typing.Decimal},
		{typing.Date, typing.Timestamp, types.TimestampType},
	},
	operators.LessEquals: {
		{types.IntType, types.DoubleType, typing.Number, typing.Decimal},
		{typing.Date, typing.Timestamp, types.TimestampType},
	},
	operators.Add: {
		{types.IntType, types.DoubleType, typing.Number, typing.Decimal},
	},
	operators.Subtract: {
		{types.IntType, types.DoubleType, typing.Number, typing.Decimal},
	},
	operators.Multiply: {
		{types.IntType, types.DoubleType, typing.Number, typing.Decimal},
	},
	operators.Divide: {
		{types.IntType, types.DoubleType, typing.Number, typing.Decimal},
	},
}

// WithSchemaTypes declares schema models, enums and roles as types in the CEL environment
func WithSchemaTypes(schema []*parser.AST) expressions.Option {
	return func(p *expressions.Parser) error {
		p.Provider.Schema = schema

		var options []cel.EnvOption
		for _, enum := range query.Enums(schema) {
			options = append(options, cel.Constant(enum.Name.Value, types.NewObjectType(fmt.Sprintf("%s_Enum", enum.Name.Value)), nil))
		}

		for _, role := range query.Roles(schema) {
			options = append(options, cel.Constant(role.Name.Value, typing.Role, nil))
		}

		for _, ast := range schema {
			for _, env := range ast.EnvironmentVariables {
				if p.Provider.Objects["_EnvironmentVariables"] == nil {
					p.Provider.Objects["_EnvironmentVariables"] = map[string]*types.Type{}
				}

				p.Provider.Objects["_EnvironmentVariables"][env] = types.StringType
			}
		}

		for _, ast := range schema {
			for _, env := range ast.Secrets {
				if p.Provider.Objects["_Secrets"] == nil {
					p.Provider.Objects["_Secrets"] = map[string]*types.Type{}
				}

				p.Provider.Objects["_Secrets"][env] = types.StringType
			}
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

// WithVariable declares a new variable in the CEL environment
func WithVariable(identifier string, typeName string, isRepeated bool) expressions.Option {
	return func(p *expressions.Parser) error {
		t, err := typing.MapType(p.Provider.Schema, typeName, isRepeated)
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

// WithConstant declares a new constant in the CEL environment
func WithConstant(identifier string, typeName string) expressions.Option {
	return func(p *expressions.Parser) error {
		t, err := typing.MapType(p.Provider.Schema, typeName, false)
		if err != nil {
			return err
		}

		p.CelEnv, err = p.CelEnv.Extend(cel.Constant(identifier, t, nil))
		if err != nil {
			return err
		}

		return nil
	}
}

// WithCtx defines the ctx variable in the CEL environment
func WithCtx() expressions.Option {
	return func(p *expressions.Parser) error {
		p.Provider.Objects["_Context"] = map[string]*types.Type{
			"identity":        types.NewObjectType(parser.IdentityModelName),
			"isAuthenticated": types.BoolType,
			"now":             typing.Timestamp,
			"secrets":         types.NewObjectType("_Secrets"),
			"env":             types.NewObjectType("_EnvironmentVariables"),
			"headers":         types.NewObjectType("_Headers"),
		}

		if p.Provider.Objects["_Secrets"] == nil {
			p.Provider.Objects["_Secrets"] = map[string]*types.Type{}
		}

		if p.Provider.Objects["_EnvironmentVariables"] == nil {
			p.Provider.Objects["_EnvironmentVariables"] = map[string]*types.Type{}
		}

		if p.Provider.Objects["_Headers"] == nil {
			p.Provider.Objects["_Headers"] = map[string]*types.Type{}
		}

		var err error
		p.CelEnv, err = p.CelEnv.Extend(cel.Variable("ctx", types.NewObjectType("_Context")))
		if err != nil {
			return err
		}

		return nil
	}
}

// WithActionInputs declares variables in the CEL environment for each action input
func WithActionInputs(schema []*parser.AST, action *parser.ActionNode) expressions.Option {
	return func(p *expressions.Parser) error {
		model := query.ActionModel(schema, action.Name.Value)
		opts := []cel.EnvOption{}

		// Add filter inputs as variables
		for _, f := range action.Inputs {
			if f.Type.Fragments[0].Fragment == parser.MessageFieldTypeAny {
				continue
			}

			if query.IsMessage(schema, f.Type.ToString()) {
				continue
			}

			typeName := query.ResolveInputType(schema, f, model, action)

			isRepeated := false
			if field := query.ResolveInputField(schema, f, model); field != nil {
				isRepeated = field.Repeated
			}

			t, err := typing.MapType(p.Provider.Schema, typeName, isRepeated)
			if err != nil {
				return err
			}

			opts = append(opts, cel.Variable(f.Name(), t))
		}

		// Add with inputs as variables
		for _, f := range action.With {
			typeName := query.ResolveInputType(schema, f, model, action)

			isRepeated := false
			if field := query.ResolveInputField(schema, f, model); field != nil {
				isRepeated = field.Repeated
			}

			t, err := typing.MapType(p.Provider.Schema, typeName, isRepeated)
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

// WithLogicalOperators enables support for the equals '==' and not equals '!=' operators for all types
func WithLogicalOperators() expressions.Option {
	return func(p *expressions.Parser) error {
		var err error

		p.CelEnv, err = p.CelEnv.Extend(
			cel.Function(operators.LogicalAnd,
				cel.Overload(overloads.LogicalAnd, []*types.Type{types.BoolType, types.BoolType}, types.BoolType)),
			cel.Function(operators.LogicalOr,
				cel.Overload(overloads.LogicalOr, []*types.Type{types.BoolType, types.BoolType}, types.BoolType)),
			cel.Function(operators.LogicalNot,
				cel.Overload(overloads.LogicalNot, []*types.Type{types.BoolType}, types.BoolType)))
		if err != nil {
			return err
		}

		return nil
	}
}

// WithComparisonOperators enables support for comparison operators for all types
func WithComparisonOperators() expressions.Option {
	return func(p *expressions.Parser) error {
		mapping := map[string][][]*types.Type{}

		var err error
		if p.Provider.Schema != nil {
			// For each enum type, configure equals, not equals and 'in' operators
			for _, enum := range query.Enums(p.Provider.Schema) {
				enumType := types.NewOpaqueType(enum.Name.Value)
				enumTypeArr := types.NewOpaqueType(enum.Name.Value + "[]")

				mapping[operators.Equals] = append(mapping[operators.Equals],
					[]*types.Type{enumType},
					[]*types.Type{enumTypeArr, types.NewListType(enumType)},
				)

				mapping[operators.NotEquals] = append(mapping[operators.NotEquals],
					[]*types.Type{enumType},
					[]*types.Type{enumTypeArr, types.NewListType(enumType)},
				)

				p.CelEnv, err = p.CelEnv.Extend(
					cel.Function(operators.In,
						cel.Overload(fmt.Sprintf("in_%s", strcase.ToLowerCamel(enum.Name.Value)), []*types.Type{enumType, enumTypeArr}, types.BoolType),
						cel.Overload(fmt.Sprintf("in_%s_literal", strcase.ToLowerCamel(enum.Name.Value)), []*types.Type{enumType, types.NewListType(enumType)}, types.BoolType),
					),
					cel.Function(operators.Equals,
						cel.Overload(fmt.Sprintf("equals_%s[]_%s", strcase.ToLowerCamel(enum.Name.Value), strcase.ToLowerCamel(enum.Name.Value)), argTypes(enumTypeArr, enumType), types.BoolType),
						cel.Overload(fmt.Sprintf("equals_%s_%s[]", strcase.ToLowerCamel(enum.Name.Value), strcase.ToLowerCamel(enum.Name.Value)), argTypes(enumType, enumTypeArr), types.BoolType),
					))
				if err != nil {
					return err
				}
			}

			// For each models, configure equals, not equals and 'in' operators
			for _, model := range query.Models(p.Provider.Schema) {
				modelType := types.NewObjectType(model.Name.Value)
				modelTypeArr := types.NewObjectType(model.Name.Value + "[]")

				mapping[operators.Equals] = append(mapping[operators.Equals],
					[]*types.Type{modelType},
					[]*types.Type{modelTypeArr},
				)

				mapping[operators.NotEquals] = append(mapping[operators.NotEquals],
					[]*types.Type{modelType},
					[]*types.Type{modelTypeArr},
				)

				p.CelEnv, err = p.CelEnv.Extend(
					cel.Function(operators.In,
						cel.Overload(fmt.Sprintf("in_%s", strcase.ToLowerCamel(model.Name.Value)), []*types.Type{modelType, modelTypeArr}, types.BoolType),
					))
				if err != nil {
					return err
				}
			}
		}

		for k, v := range typeCompatibilityMapping {
			mapping[k] = append(mapping[k], v...)
		}

		// Add operator overloads for each compatible type combination
		options := []cel.EnvOption{}
		for k, v := range mapping {
			switch k {
			case operators.Equals, operators.NotEquals, operators.Greater, operators.GreaterEquals, operators.Less, operators.LessEquals:
				for _, t := range v {
					for _, arg1 := range t {
						for _, arg2 := range t {
							opt := cel.Function(k, cel.Overload(overloadName(k, arg1, arg2), argTypes(arg1, arg2), types.BoolType))
							options = append(options, opt)
						}
					}
				}
			}
		}

		p.CelEnv, err = p.CelEnv.Extend(options...)
		if err != nil {
			return err
		}

		// Explicitly defining the 'in' operator overloads
		p.CelEnv, err = p.CelEnv.Extend(
			cel.Function(operators.In,
				cel.Overload("in_string_list(string)", argTypes(types.StringType, types.NewListType(types.StringType)), types.BoolType),
				cel.Overload("in_string_Text[]", argTypes(types.StringType, typing.TextArray), types.BoolType),
				cel.Overload("in_Text_Text[]", argTypes(typing.Text, typing.TextArray), types.BoolType),
				cel.Overload("in_Text_list(string)", argTypes(typing.Text, types.NewListType(types.StringType)), types.BoolType),

				cel.Overload("in_string_ID[]", argTypes(types.StringType, typing.IDArray), types.BoolType),
				cel.Overload("in_ID_ID[]", argTypes(typing.ID, typing.IDArray), types.BoolType),
				cel.Overload("in_ID_list(string)", argTypes(typing.ID, types.NewListType(types.StringType)), types.BoolType),

				cel.Overload("in_int_list(int)", argTypes(types.IntType, types.NewListType(types.IntType)), types.BoolType),
				cel.Overload("in_int_Number[]", argTypes(types.IntType, typing.NumberArray), types.BoolType),
				cel.Overload("in_Number_Number[]", argTypes(typing.Number, typing.NumberArray), types.BoolType),
				cel.Overload("in_Number_list(int)", argTypes(typing.Number, types.NewListType(types.IntType)), types.BoolType),

				cel.Overload("in_double_list(double)", argTypes(types.DoubleType, types.NewListType(types.DoubleType)), types.BoolType),
				cel.Overload("in_double_Decimal[]", argTypes(types.DoubleType, typing.DecimalArray), types.BoolType),
				cel.Overload("in_Decimal_Decimal[]", argTypes(typing.Text, typing.DecimalArray), types.BoolType),
				cel.Overload("in_Decimal_list(double)", argTypes(typing.Decimal, types.NewListType(types.DoubleType)), types.BoolType),

				cel.Overload("in_bool_list(bool)", argTypes(types.BoolType, types.NewListType(types.DoubleType)), types.BoolType),
				cel.Overload("in_bool_Boolean[]", argTypes(types.BoolType, typing.BooleanArray), types.BoolType),
				cel.Overload("in_Boolean_Boolean[]", argTypes(typing.Boolean, typing.BooleanArray), types.BoolType),
				cel.Overload("in_Boolean_list(bool)", argTypes(typing.Boolean, types.NewListType(types.BoolType)), types.BoolType),

				cel.Overload("in_Timestamp_Timestamp[]", argTypes(typing.Timestamp, typing.TimestampArray), types.BoolType),
				cel.Overload("in_Date_Date[]", argTypes(typing.Date, typing.DateArray), types.BoolType),
			),
		)
		if err != nil {
			return err
		}

		// Backwards compatibility for relationships expressions like `organisation.people.name == "Keel"` which is actually performing an "ANY" query
		// To be deprecated in favour of functions
		p.CelEnv, err = p.CelEnv.Extend(
			cel.Function(operators.Equals,
				cel.Overload("equals_Text[]_Text", argTypes(typing.TextArray, typing.Text), types.BoolType),
				cel.Overload("equals_Text[]_string", argTypes(typing.TextArray, types.StringType), types.BoolType),
				cel.Overload("equals_Text_Text[]", argTypes(typing.Text, typing.TextArray), types.BoolType),

				cel.Overload("equals_ID[]_ID", argTypes(typing.IDArray, typing.ID), types.BoolType),
				cel.Overload("equals_ID[]_string", argTypes(typing.IDArray, types.StringType), types.BoolType),
				cel.Overload("equals_ID_ID[]", argTypes(typing.ID, typing.IDArray), types.BoolType),

				cel.Overload("equals_Number[]_Number", argTypes(typing.NumberArray, typing.Number), types.BoolType),
				cel.Overload("equals_Number[]_int", argTypes(typing.NumberArray, types.IntType), types.BoolType),
				cel.Overload("equals_Number_Number[]", argTypes(typing.Number, typing.NumberArray), types.BoolType),
			),
		)
		if err != nil {
			return err
		}

		return nil
	}
}

// WithArithmeticOperators enables support for arithmetic operators
func WithArithmeticOperators() expressions.Option {
	return func(p *expressions.Parser) error {
		// Add operator overloads for each compatible type combination
		options := []cel.EnvOption{}
		for k, v := range typeCompatibilityMapping {
			switch k {
			case operators.Add, operators.Subtract, operators.Multiply, operators.Divide:
				for _, t := range v {
					for _, arg1 := range t {
						for _, arg2 := range t {
							opt := cel.Function(k, cel.Overload(overloadName(k, arg1, arg2), argTypes(arg1, arg2), typing.Decimal))
							options = append(options, opt)
						}
					}
				}
			}
		}

		var err error
		p.CelEnv, err = p.CelEnv.Extend(options...)
		if err != nil {
			return err
		}

		return nil
	}
}

func WithFunctions() expressions.Option {
	return func(p *expressions.Parser) error {
		typeParamA := cel.TypeParamType("A")
		var err error
		p.CelEnv, err = p.CelEnv.Extend(
			cel.Function(typing.FunctionCount, cel.Overload("count", []*types.Type{typeParamA}, typing.Number)),
			cel.Function(typing.FunctionSum, cel.Overload("sum_decimal", []*types.Type{typing.DecimalArray}, typing.Decimal)),
			cel.Function(typing.FunctionSum, cel.Overload("sum_number", []*types.Type{typing.NumberArray}, typing.Number)),
			cel.Function(typing.FunctionAvg, cel.Overload("avg_decimal", []*types.Type{typing.DecimalArray}, typing.Decimal)),
			cel.Function(typing.FunctionAvg, cel.Overload("avg_number", []*types.Type{typing.NumberArray}, typing.Number)),
			cel.Function(typing.FunctionMin, cel.Overload("min_decimal", []*types.Type{typing.DecimalArray}, typing.Decimal)),
			cel.Function(typing.FunctionMin, cel.Overload("min_number", []*types.Type{typing.NumberArray}, typing.Number)),
			cel.Function(typing.FunctionMax, cel.Overload("max_decimal", []*types.Type{typing.DecimalArray}, typing.Decimal)),
			cel.Function(typing.FunctionMax, cel.Overload("max_number", []*types.Type{typing.NumberArray}, typing.Number)),
			cel.Function(typing.FunctionMedian, cel.Overload("median_decimal", []*types.Type{typing.DecimalArray}, typing.Decimal)),
			cel.Function(typing.FunctionMedian, cel.Overload("median_number", []*types.Type{typing.NumberArray}, typing.Number)))
		if err != nil {
			return err
		}

		return err
	}
}

// WithReturnTypeAssertion will check that the expression evaluates to a specific type
func WithReturnTypeAssertion(returnType string, asArray bool) expressions.Option {
	return func(p *expressions.Parser) error {
		var err error
		p.ExpectedReturnType, err = typing.MapType(p.Provider.Schema, returnType, asArray)
		return err
	}
}

func argTypes(args ...*types.Type) []*types.Type {
	return args
}

func overloadName(op string, t1 *types.Type, t2 *types.Type) string {
	return fmt.Sprintf("%s_%s_%s", op, t1.String(), t2.String())
}
