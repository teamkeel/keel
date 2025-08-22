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
// This is used to generate all the necessary combinations of operator overloads.
var typeCompatibilityMapping = map[string][][]*types.Type{
	operators.Equals: {
		{types.StringType, typing.TypeText, typing.TypeID, typing.TypeMarkdown},
		{types.IntType, types.DoubleType, typing.TypeNumber, typing.TypeDecimal},
		{typing.TypeDate, typing.TypeTimestamp, types.TimestampType},
		{typing.TypeBoolean, types.BoolType},
		{types.NewListType(types.StringType), typing.TypeTextArray, typing.TypeIDArray, typing.TypeMarkdownArray},
		{types.NewListType(types.IntType), types.NewListType(types.DoubleType), typing.TypeNumberArray, typing.TypeDecimalArray},
		{typing.TypeDuration},

		// The following are needed to support our special 1-M filtering expressions, such as invoice.items.isActive == true
		{typing.TypeDecimalArray, types.IntType},
		{typing.TypeDecimalArray, types.DoubleType},
		{typing.TypeDecimalArray, typing.TypeNumber},
		{typing.TypeDecimalArray, typing.TypeDecimal},
		{typing.TypeNumberArray, types.IntType},
		{typing.TypeNumberArray, types.DoubleType},
		{typing.TypeNumberArray, typing.TypeNumber},
		{typing.TypeNumberArray, typing.TypeDecimal},
		{typing.TypeTextArray, types.StringType},
		{typing.TypeTextArray, typing.TypeText},
		{typing.TypeIDArray, types.StringType},
		{typing.TypeIDArray, typing.TypeText},
		{typing.TypeBooleanArray, types.BoolType},
		{typing.TypeBooleanArray, typing.TypeBoolean},
		{typing.TypeDateArray, typing.TypeDate},
		{typing.TypeDateArray, typing.TypeTimestamp},
		{typing.TypeTimestampArray, typing.TypeDate},
		{typing.TypeTimestampArray, typing.TypeTimestamp},
		{typing.TypeDurationArray, typing.TypeDuration},
	},
	operators.NotEquals: {
		{types.StringType, typing.TypeText, typing.TypeID, typing.TypeMarkdown},
		{types.IntType, types.DoubleType, typing.TypeNumber, typing.TypeDecimal},
		{typing.TypeDate, typing.TypeTimestamp, types.TimestampType},
		{typing.TypeBoolean, types.BoolType},
		{types.NewListType(types.StringType), typing.TypeTextArray, typing.TypeIDArray, typing.TypeMarkdownArray},
		{types.NewListType(types.IntType), types.NewListType(types.DoubleType), typing.TypeNumberArray, typing.TypeDecimalArray},
		{typing.TypeDuration},

		// The following are needed to support our special 1-M filtering expressions, such as invoice.items.price != 0
		{typing.TypeDecimalArray, types.IntType},
		{typing.TypeDecimalArray, types.DoubleType},
		{typing.TypeDecimalArray, typing.TypeNumber},
		{typing.TypeDecimalArray, typing.TypeDecimal},
		{typing.TypeNumberArray, types.IntType},
		{typing.TypeNumberArray, types.DoubleType},
		{typing.TypeNumberArray, typing.TypeNumber},
		{typing.TypeNumberArray, typing.TypeDecimal},
		{typing.TypeTextArray, types.StringType},
		{typing.TypeTextArray, typing.TypeText},
		{typing.TypeIDArray, types.StringType},
		{typing.TypeIDArray, typing.TypeText},
		{typing.TypeBooleanArray, types.BoolType},
		{typing.TypeBooleanArray, typing.TypeBoolean},
		{typing.TypeDateArray, typing.TypeDate},
		{typing.TypeDateArray, typing.TypeTimestamp},
		{typing.TypeTimestampArray, typing.TypeDate},
		{typing.TypeTimestampArray, typing.TypeTimestamp},
		{typing.TypeDurationArray, typing.TypeDuration},
	},
	operators.Greater: {
		{types.IntType, types.DoubleType, typing.TypeNumber, typing.TypeDecimal},
		{typing.TypeDecimal, typing.TypeNumber},
		{typing.TypeDate, typing.TypeTimestamp, types.TimestampType},
		{typing.TypeDuration},

		// The following are needed to support our special 1-M filtering expressions, such as invoice.items.price > 0
		{typing.TypeDecimalArray, types.IntType},
		{typing.TypeDecimalArray, types.DoubleType},
		{typing.TypeDecimalArray, typing.TypeNumber},
		{typing.TypeDecimalArray, typing.TypeDecimal},
		{typing.TypeNumberArray, types.IntType},
		{typing.TypeNumberArray, types.DoubleType},
		{typing.TypeNumberArray, typing.TypeNumber},
		{typing.TypeNumberArray, typing.TypeDecimal},
		{typing.TypeDateArray, typing.TypeDate},
		{typing.TypeDateArray, typing.TypeTimestamp},
		{typing.TypeTimestampArray, typing.TypeDate},
		{typing.TypeTimestampArray, typing.TypeTimestamp},
		{typing.TypeDurationArray, typing.TypeDuration},
	},
	operators.GreaterEquals: {
		{types.IntType, types.DoubleType, typing.TypeNumber, typing.TypeDecimal},
		{typing.TypeDate, typing.TypeTimestamp, types.TimestampType},
		{typing.TypeDuration},

		// The following are needed to support our special 1-M filtering expressions, such as invoice.items.price >= 0
		{typing.TypeDecimalArray, types.IntType},
		{typing.TypeDecimalArray, types.DoubleType},
		{typing.TypeDecimalArray, typing.TypeNumber},
		{typing.TypeDecimalArray, typing.TypeDecimal},
		{typing.TypeNumberArray, types.IntType},
		{typing.TypeNumberArray, types.DoubleType},
		{typing.TypeNumberArray, typing.TypeNumber},
		{typing.TypeNumberArray, typing.TypeDecimal},
		{typing.TypeDateArray, typing.TypeDate},
		{typing.TypeDateArray, typing.TypeTimestamp},
		{typing.TypeTimestampArray, typing.TypeDate},
		{typing.TypeTimestampArray, typing.TypeTimestamp},
		{typing.TypeDurationArray, typing.TypeDuration},
	},
	operators.Less: {
		{types.IntType, types.DoubleType, typing.TypeNumber, typing.TypeDecimal},
		{typing.TypeDate, typing.TypeTimestamp, types.TimestampType},
		{typing.TypeDuration},

		// The following are needed to support our special 1-M filtering expressions, such as invoice.items.price < 0
		{typing.TypeDecimalArray, types.IntType},
		{typing.TypeDecimalArray, types.DoubleType},
		{typing.TypeDecimalArray, typing.TypeNumber},
		{typing.TypeDecimalArray, typing.TypeDecimal},
		{typing.TypeNumberArray, types.IntType},
		{typing.TypeNumberArray, types.DoubleType},
		{typing.TypeNumberArray, typing.TypeNumber},
		{typing.TypeNumberArray, typing.TypeDecimal},
		{typing.TypeDateArray, typing.TypeDate},
		{typing.TypeDateArray, typing.TypeTimestamp},
		{typing.TypeTimestampArray, typing.TypeDate},
		{typing.TypeTimestampArray, typing.TypeTimestamp},
		{typing.TypeDurationArray, typing.TypeDuration},
	},
	operators.LessEquals: {
		{types.IntType, types.DoubleType, typing.TypeNumber, typing.TypeDecimal},
		{typing.TypeDate, typing.TypeTimestamp, types.TimestampType},
		{typing.TypeDuration},

		// The following are needed to support our special 1-M filtering expressions, such as invoice.items.price <= 0
		{typing.TypeDecimalArray, types.IntType},
		{typing.TypeDecimalArray, types.DoubleType},
		{typing.TypeDecimalArray, typing.TypeNumber},
		{typing.TypeDecimalArray, typing.TypeDecimal},
		{typing.TypeNumberArray, types.IntType},
		{typing.TypeNumberArray, types.DoubleType},
		{typing.TypeNumberArray, typing.TypeNumber},
		{typing.TypeNumberArray, typing.TypeDecimal},
		{typing.TypeDateArray, typing.TypeDate},
		{typing.TypeDateArray, typing.TypeTimestamp},
		{typing.TypeTimestampArray, typing.TypeDate},
		{typing.TypeTimestampArray, typing.TypeTimestamp},
		{typing.TypeDurationArray, typing.TypeDuration},
	},
	operators.Add: {
		{types.IntType, types.DoubleType, typing.TypeNumber, typing.TypeDecimal},
		{typing.TypeDuration},
		{types.StringType, typing.TypeText},
	},
	operators.Subtract: {
		{types.IntType, types.DoubleType, typing.TypeNumber, typing.TypeDecimal},
		{typing.TypeDuration},
	},
	operators.Multiply: {
		{types.IntType, types.DoubleType, typing.TypeNumber, typing.TypeDecimal},
	},
	operators.Divide: {
		{types.IntType, types.DoubleType, typing.TypeNumber, typing.TypeDecimal},
	},
}

// WithSchemaTypes declares schema models, enums and roles as types in the CEL environment.
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
				if p.Provider.Objects[typing.TypeNameEnvvars] == nil {
					p.Provider.Objects[typing.TypeNameEnvvars] = map[string]*types.Type{}
				}

				p.Provider.Objects[typing.TypeNameEnvvars][env] = types.StringType
			}
		}

		for _, ast := range schema {
			for _, env := range ast.Secrets {
				if p.Provider.Objects[typing.TypeNameSecrets] == nil {
					p.Provider.Objects[typing.TypeNameSecrets] = map[string]*types.Type{}
				}

				p.Provider.Objects[typing.TypeNameSecrets][env] = types.StringType
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

// WithVariable declares a new variable in the CEL environment.
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

// WithConstant declares a new constant in the CEL environment.
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

// WithCtx defines the ctx variable in the CEL environment.
func WithCtx() expressions.Option {
	return func(p *expressions.Parser) error {
		p.Provider.Objects[typing.TypeNameContext] = map[string]*types.Type{
			"identity":        types.NewObjectType(parser.IdentityModelName),
			"isAuthenticated": types.BoolType,
			"now":             typing.TypeTimestamp,
			"secrets":         typing.TypeSecrets,
			"env":             typing.TypeEnvvars,
			"headers":         typing.TypeHeaders,
		}

		if p.Provider.Objects[typing.TypeNameSecrets] == nil {
			p.Provider.Objects[typing.TypeNameSecrets] = map[string]*types.Type{}
		}

		if p.Provider.Objects[typing.TypeNameEnvvars] == nil {
			p.Provider.Objects[typing.TypeNameEnvvars] = map[string]*types.Type{}
		}

		if p.Provider.Objects[typing.TypeNameHeaders] == nil {
			p.Provider.Objects[typing.TypeNameHeaders] = map[string]*types.Type{}
		}

		var err error
		p.CelEnv, err = p.CelEnv.Extend(cel.Variable("ctx", typing.TypeContext))
		if err != nil {
			return err
		}

		return nil
	}
}

// WithActionInputs declares variables in the CEL environment for each action input.
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

// WithLogicalOperators enables support for the equals '==' and not equals '!=' operators for all types.
func WithLogicalOperators() expressions.Option {
	return func(p *expressions.Parser) error {
		var err error

		p.CelEnv, err = p.CelEnv.Extend(
			cel.Function(operators.LogicalAnd,
				cel.Overload(overloads.LogicalAnd, []*types.Type{types.BoolType, types.BoolType}, types.BoolType)),
			cel.Function(operators.LogicalOr,
				cel.Overload(overloads.LogicalOr, []*types.Type{types.BoolType, types.BoolType}, types.BoolType)),
			cel.Function(operators.LogicalNot,
				cel.Overload(overloads.LogicalNot, []*types.Type{types.BoolType}, types.BoolType),
				cel.Overload("logical_not_boolean", []*types.Type{typing.TypeBoolean}, types.BoolType),
				cel.Overload("logical_not_boolean_array", []*types.Type{typing.TypeBooleanArray}, types.BoolType)))
		if err != nil {
			return err
		}

		return nil
	}
}

// WithComparisonOperators enables support for comparison operators for all types.
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

			// For each entity, configure equals, not equals and 'in' operators
			for _, entity := range query.Entities(p.Provider.Schema) {
				entityType := types.NewObjectType(entity.GetName())
				entityTypeArr := types.NewObjectType(entity.GetName() + "[]")

				mapping[operators.Equals] = append(mapping[operators.Equals],
					[]*types.Type{entityType},
					[]*types.Type{entityTypeArr},
				)

				mapping[operators.NotEquals] = append(mapping[operators.NotEquals],
					[]*types.Type{entityType},
					[]*types.Type{entityTypeArr},
				)

				p.CelEnv, err = p.CelEnv.Extend(
					cel.Function(operators.In,
						cel.Overload(fmt.Sprintf("in_%s", strcase.ToLowerCamel(entity.GetName())), []*types.Type{entityType, entityTypeArr}, types.BoolType),
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

		// Subtracting two date/time variants will produce a duration. We define these separately because the return type is different to the operand types which doesnt fit with the generic mapping.
		options = append(options,
			cel.Function(operators.Subtract, cel.Overload(overloadName(operators.Subtract, typing.TypeDate, typing.TypeTimestamp), argTypes(typing.TypeDate, typing.TypeTimestamp), typing.TypeDuration)),
			cel.Function(operators.Subtract, cel.Overload(overloadName(operators.Subtract, typing.TypeDate, typing.TypeDate), argTypes(typing.TypeDate, typing.TypeDate), typing.TypeDuration)),
			cel.Function(operators.Subtract, cel.Overload(overloadName(operators.Subtract, typing.TypeTimestamp, typing.TypeTimestamp), argTypes(typing.TypeTimestamp, typing.TypeTimestamp), typing.TypeDuration)),
			cel.Function(operators.Subtract, cel.Overload(overloadName(operators.Subtract, typing.TypeTimestamp, typing.TypeDate), argTypes(typing.TypeTimestamp, typing.TypeDate), typing.TypeDuration)),
		)

		p.CelEnv, err = p.CelEnv.Extend(options...)
		if err != nil {
			return err
		}

		// Explicitly defining the 'in' operator overloads
		p.CelEnv, err = p.CelEnv.Extend(
			cel.Function(operators.In,
				cel.Overload("in_string_list(string)", argTypes(types.StringType, types.NewListType(types.StringType)), types.BoolType),
				cel.Overload("in_string_Text[]", argTypes(types.StringType, typing.TypeTextArray), types.BoolType),
				cel.Overload("in_Text_Text[]", argTypes(typing.TypeText, typing.TypeTextArray), types.BoolType),
				cel.Overload("in_Text_list(string)", argTypes(typing.TypeText, types.NewListType(types.StringType)), types.BoolType),

				cel.Overload("in_string_ID[]", argTypes(types.StringType, typing.TypeIDArray), types.BoolType),
				cel.Overload("in_ID_ID[]", argTypes(typing.TypeID, typing.TypeIDArray), types.BoolType),
				cel.Overload("in_ID_list(string)", argTypes(typing.TypeID, types.NewListType(types.StringType)), types.BoolType),

				cel.Overload("in_int_list(int)", argTypes(types.IntType, types.NewListType(types.IntType)), types.BoolType),
				cel.Overload("in_int_Number[]", argTypes(types.IntType, typing.TypeNumberArray), types.BoolType),
				cel.Overload("in_Number_Number[]", argTypes(typing.TypeNumber, typing.TypeNumberArray), types.BoolType),
				cel.Overload("in_Number_list(int)", argTypes(typing.TypeNumber, types.NewListType(types.IntType)), types.BoolType),

				cel.Overload("in_double_list(double)", argTypes(types.DoubleType, types.NewListType(types.DoubleType)), types.BoolType),
				cel.Overload("in_double_Decimal[]", argTypes(types.DoubleType, typing.TypeDecimalArray), types.BoolType),
				cel.Overload("in_Decimal_Decimal[]", argTypes(typing.TypeText, typing.TypeDecimalArray), types.BoolType),
				cel.Overload("in_Decimal_list(double)", argTypes(typing.TypeDecimal, types.NewListType(types.DoubleType)), types.BoolType),

				cel.Overload("in_bool_list(bool)", argTypes(types.BoolType, types.NewListType(types.DoubleType)), types.BoolType),
				cel.Overload("in_bool_Boolean[]", argTypes(types.BoolType, typing.TypeBooleanArray), types.BoolType),
				cel.Overload("in_Boolean_Boolean[]", argTypes(typing.TypeBoolean, typing.TypeBooleanArray), types.BoolType),
				cel.Overload("in_Boolean_list(bool)", argTypes(typing.TypeBoolean, types.NewListType(types.BoolType)), types.BoolType),

				cel.Overload("in_Timestamp_Timestamp[]", argTypes(typing.TypeTimestamp, typing.TypeTimestampArray), types.BoolType),
				cel.Overload("in_Date_Date[]", argTypes(typing.TypeDate, typing.TypeDateArray), types.BoolType),
			),
		)
		if err != nil {
			return err
		}

		return nil
	}
}

// WithArithmeticOperators enables support for arithmetic operators.
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
							opt := cel.Function(k, cel.Overload(overloadName(k, arg1, arg2), argTypes(arg1, arg2), arg1))
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

func WithAggregateFunctions() expressions.Option {
	return func(p *expressions.Parser) error {
		typeParamA := cel.TypeParamType("A")
		var err error
		p.CelEnv, err = p.CelEnv.Extend(
			cel.Function(typing.FunctionCount, cel.Overload("count", []*types.Type{typeParamA}, typing.TypeNumber)),
			cel.Function(typing.FunctionSum, cel.Overload("sum_decimal", []*types.Type{typing.TypeDecimalArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionSum, cel.Overload("sum_number", []*types.Type{typing.TypeNumberArray}, typing.TypeNumber)),
			cel.Function(typing.FunctionAvg, cel.Overload("avg_decimal", []*types.Type{typing.TypeDecimalArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionAvg, cel.Overload("avg_number", []*types.Type{typing.TypeNumberArray}, typing.TypeNumber)),
			cel.Function(typing.FunctionMin, cel.Overload("min_decimal", []*types.Type{typing.TypeDecimalArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMin, cel.Overload("min_number", []*types.Type{typing.TypeNumberArray}, typing.TypeNumber)),
			cel.Function(typing.FunctionMax, cel.Overload("max_decimal", []*types.Type{typing.TypeDecimalArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMax, cel.Overload("max_number", []*types.Type{typing.TypeNumberArray}, typing.TypeNumber)),
			cel.Function(typing.FunctionMedian, cel.Overload("median_decimal", []*types.Type{typing.TypeDecimalArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMedian, cel.Overload("median_number", []*types.Type{typing.TypeNumberArray}, typing.TypeNumber)),

			// These are necessary for expressions like: SUMIF(class.enrollments.grade, class.enrollments.student.isActive == true)
			cel.Function(typing.FunctionSumIf, cel.Overload("sumif_decimal", []*types.Type{typing.TypeDecimalArray, types.BoolType}, typing.TypeDecimal)),
			cel.Function(typing.FunctionSumIf, cel.Overload("sumif_number", []*types.Type{typing.TypeNumberArray, types.BoolType}, typing.TypeNumber)),
			cel.Function(typing.FunctionCountIf, cel.Overload("countif_decimal", []*types.Type{types.AnyType, types.BoolType}, typing.TypeNumber)),
			cel.Function(typing.FunctionAvgIf, cel.Overload("avgif_decimal", []*types.Type{typing.TypeDecimalArray, types.BoolType}, typing.TypeDecimal)),
			cel.Function(typing.FunctionAvgIf, cel.Overload("avgif_number", []*types.Type{typing.TypeNumberArray, types.BoolType}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMedianIf, cel.Overload("medianif_decimal", []*types.Type{typing.TypeDecimalArray, types.BoolType}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMedianIf, cel.Overload("medianif_number", []*types.Type{typing.TypeNumberArray, types.BoolType}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMinIf, cel.Overload("minif_decimal", []*types.Type{typing.TypeDecimalArray, types.BoolType}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMinIf, cel.Overload("minif_number", []*types.Type{typing.TypeNumberArray, types.BoolType}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMaxIf, cel.Overload("maxif_decimal", []*types.Type{typing.TypeDecimalArray, types.BoolType}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMaxIf, cel.Overload("maxif_number", []*types.Type{typing.TypeNumberArray, types.BoolType}, typing.TypeDecimal)),

			// These are necessary for expressions like: AVGIF(class.enrollments.grade, class.enrollments.student.isActive)
			cel.Function(typing.FunctionSumIf, cel.Overload("sumif_decimal_booleanarray", []*types.Type{typing.TypeDecimalArray, typing.TypeBooleanArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionSumIf, cel.Overload("sumif_number_booleanarray", []*types.Type{typing.TypeNumberArray, typing.TypeBooleanArray}, typing.TypeNumber)),
			cel.Function(typing.FunctionCountIf, cel.Overload("countif_booleanarray", []*types.Type{types.AnyType, typing.TypeBooleanArray}, typing.TypeNumber)),
			cel.Function(typing.FunctionAvgIf, cel.Overload("avgif_decimal_booleanarray", []*types.Type{typing.TypeDecimalArray, typing.TypeBooleanArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionAvgIf, cel.Overload("avgif_number_booleanarray", []*types.Type{typing.TypeNumberArray, typing.TypeBooleanArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMedianIf, cel.Overload("medianif_decimal_booleanarray", []*types.Type{typing.TypeDecimalArray, typing.TypeBooleanArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMedianIf, cel.Overload("medianif_number_booleanarray", []*types.Type{typing.TypeNumberArray, typing.TypeBooleanArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMinIf, cel.Overload("minif_decimal_booleanarray", []*types.Type{typing.TypeDecimalArray, typing.TypeBooleanArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMinIf, cel.Overload("minif_number_booleanarray", []*types.Type{typing.TypeNumberArray, typing.TypeBooleanArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMaxIf, cel.Overload("maxif_decimal_booleanarray", []*types.Type{typing.TypeDecimalArray, typing.TypeBooleanArray}, typing.TypeDecimal)),
			cel.Function(typing.FunctionMaxIf, cel.Overload("maxif_number_booleanarray", []*types.Type{typing.TypeNumberArray, typing.TypeBooleanArray}, typing.TypeDecimal)),
		)
		if err != nil {
			return err
		}

		return err
	}
}

// WithReturnTypeAssertion will check that the expression evaluates to a specific type.
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
