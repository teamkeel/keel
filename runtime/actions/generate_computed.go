package actions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/google/cel-go/common/operators"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/expressions/typing"
	"github.com/teamkeel/keel/proto"

	"github.com/teamkeel/keel/schema/parser"
)

// GenerateComputedFunction visits the expression and generates a SQL expression.
func GenerateComputedFunction(ctx context.Context, schema *proto.Schema, model *proto.Model, field *proto.Field) resolve.Visitor[string] {
	return &computedQueryGen{
		ctx:       ctx,
		schema:    schema,
		model:     model,
		field:     field,
		sql:       "",
		functions: arraystack.New(),
		arguments: arraystack.New(),

		operators: arraystack.New(),
		operands:  arraystack.New(),
	}
}

var _ resolve.Visitor[string] = new(computedQueryGen)

type computedQueryGen struct {
	ctx       context.Context
	schema    *proto.Schema
	model     *proto.Model
	field     *proto.Field
	sql       string
	functions *arraystack.Stack
	arguments *arraystack.Stack

	query     *QueryBuilder
	operators *arraystack.Stack
	operands  *arraystack.Stack
}

func (v *computedQueryGen) StartTerm(nested bool) error {
	if v.functions.Size() > 0 {
		if op, ok := v.operators.Peek(); ok && op == Not {
			_, _ = v.operators.Pop()
			v.query.Not()
		}

		// Only add parenthesis if we're in a nested condition
		if nested {
			v.query.OpenParenthesis()
		}

		return nil
	}

	if nested {
		v.sql += "("
	}

	return nil
}

func (v *computedQueryGen) EndTerm(nested bool) error {
	if v.functions.Size() > 0 {
		if _, ok := v.operators.Peek(); ok && v.operands.Size() == 2 {
			operator, _ := v.operators.Pop()

			r, ok := v.operands.Pop()
			if !ok {
				return errors.New("expected rhs operand")
			}
			l, ok := v.operands.Pop()
			if !ok {
				return errors.New("expected lhs operand")
			}

			lhs := l.(*QueryOperand)
			rhs := r.(*QueryOperand)

			err := v.query.Where(lhs, operator.(ActionOperator), rhs)
			if err != nil {
				return err
			}
		} else if _, ok := v.operators.Peek(); !ok {
			l, hasOperand := v.operands.Pop()
			if hasOperand {
				lhs := l.(*QueryOperand)
				err := v.query.Where(lhs, Equals, Value(true))
				if err != nil {
					return err
				}
			}
		}

		// Only close parenthesis if we're nested
		if nested {
			v.query.CloseParenthesis()
		}

		return nil
	}

	if nested {
		v.sql += ")"
	}
	return nil
}

func (v *computedQueryGen) StartFunction(name string) error {
	v.functions.Push(name)
	v.arguments.Push(0)

	return nil
}

func (v *computedQueryGen) EndFunction() error {
	v.functions.Pop()
	v.arguments.Pop()

	stmt := v.query.SelectStatement()
	v.sql += fmt.Sprintf("(%s)", stmt.SqlTemplate())

	return nil
}

func (v *computedQueryGen) StartArgument(num int) error {
	arg, has := v.arguments.Pop()
	if !has {
		return errors.New("argument stack is empty")
	}

	v.arguments.Push(arg.(int) + 1)
	return nil
}

func (v *computedQueryGen) EndArgument() error {
	return nil
}

func (v *computedQueryGen) VisitAnd() error {
	if v.functions.Size() > 0 {
		v.query.And()
		return nil
	}

	v.sql += " AND "
	return nil
}

func (v *computedQueryGen) VisitOr() error {
	if v.functions.Size() > 0 {
		v.query.Or()
		return nil
	}

	v.sql += " OR "
	return nil
}

func (v *computedQueryGen) VisitNot() error {
	if v.functions.Size() > 0 {
		v.operators.Push(Not)
		return nil
	}

	v.sql += " NOT "
	return nil
}

func (v *computedQueryGen) VisitOperator(op string) error {
	if v.functions.Size() > 0 {
		operator, err := toActionOperator(op)
		if err != nil {
			return err
		}

		v.operators.Push(operator)
		return nil
	}

	// Map CEL operators to SQL operators
	sqlOp := map[string]string{
		operators.Add:           "+",
		operators.Subtract:      "-",
		operators.Multiply:      "*",
		operators.Divide:        "/",
		operators.Equals:        "IS NOT DISTINCT FROM",
		operators.NotEquals:     "IS DISTINCT FROM",
		operators.Greater:       ">",
		operators.GreaterEquals: ">=",
		operators.Less:          "<",
		operators.LessEquals:    "<=",
	}[op]

	// Handle string concatenation
	if v.field.GetType().GetType() == proto.Type_TYPE_STRING && op == operators.Add {
		sqlOp = "||"
	}

	if sqlOp == "" {
		return fmt.Errorf("unsupported operator: %s", op)
	}

	v.sql += fmt.Sprintf(" %s ", sqlOp)

	return nil
}

func (v *computedQueryGen) VisitLiteral(value any) error {
	if v.functions.Size() > 0 {
		if value == nil {
			v.operands.Push(Null())
		} else {
			v.operands.Push(Value(value))
		}
		return nil
	}

	switch val := value.(type) {
	case int64:
		v.sql += fmt.Sprintf("%v", val)
	case float64:
		v.sql += fmt.Sprintf("%v", val)
	case string:
		v.sql += fmt.Sprintf("'%v'", val)
	case bool:
		v.sql += fmt.Sprintf("%t", val)
	case nil:
		v.sql += "NULL"
	default:
		return fmt.Errorf("unsupported literal type: %T", value)
	}
	return nil
}

func (v *computedQueryGen) VisitIdent(ident *parser.ExpressionIdent) error {
	model := v.schema.FindModel(strcase.ToCamel(ident.Fragments[0]))

	if model == nil {
		enum := v.schema.FindEnum(ident.Fragments[0])
		if enum == nil {
			return fmt.Errorf("model or enum not found: %s", ident.Fragments[0])
		}

		var value string
		for _, v := range enum.GetValues() {
			if v.GetName() == ident.Fragments[1] {
				value = v.GetName()
				break
			}
		}

		if value == "" {
			return fmt.Errorf("enum value not found: %s", ident.Fragments[1])
		}

		v.sql += fmt.Sprintf("'%v'", value)
		return nil
	}

	field := proto.FindField(v.schema.GetModels(), model.GetName(), ident.Fragments[1])

	normalised, err := NormalisedFragments(v.schema, ident.Fragments)
	if err != nil {
		return err
	}

	if len(normalised) == 2 {
		v.sql += "r." + sqlQuote(strcase.ToSnake(field.GetName()))
	} else if len(normalised) > 2 {
		isToMany, err := v.isToManyLookup(ident)
		if err != nil {
			return err
		}

		if isToMany {

			arg, has := v.arguments.Peek()
			if !has {
				return errors.New("argument stack is empty")
			}

			model = v.schema.FindModel(field.GetType().GetModelName().GetValue())

			switch arg.(int) {
			case 1: //the first arg sets the SELECT
				model = v.schema.FindModel(field.GetType().GetModelName().GetValue())
				v.query = NewQuery(model, EmbedLiterals())

				relatedModelField := proto.FindField(v.schema.GetModels(), v.model.GetName(), normalised[1])
				foreignKeyField := proto.GetForeignKeyFieldName(v.schema.GetModels(), relatedModelField)

				r := proto.FindField(v.schema.GetModels(), v.model.GetName(), normalised[1])
				subFragments := normalised[1:]
				subFragments[0] = strcase.ToLowerCamel(r.GetType().GetModelName().GetValue())

				err := v.query.AddJoinFromFragments(v.schema, subFragments)
				if err != nil {
					return err
				}

				funcBegin, has := v.functions.Peek()
				if !has {
					return errors.New("no function found for 1:M lookup")
				}

				fieldName := normalised[len(normalised)-1]
				fragments := normalised[1 : len(normalised)-1]

				raw := ""
				selectField := sqlQuote(casing.ToSnake(strings.Join(fragments, "$"))) + "." + sqlQuote(casing.ToSnake(fieldName))
				switch funcBegin {
				case typing.FunctionSum, typing.FunctionSumIf:
					raw += fmt.Sprintf("COALESCE(SUM(%s), 0)", selectField)
				case typing.FunctionCount, typing.FunctionCountIf:
					raw += fmt.Sprintf("COALESCE(COUNT(%s), 0)", selectField)
				case typing.FunctionAvg, typing.FunctionAvgIf:
					raw += fmt.Sprintf("COALESCE(AVG(%s), 0)", selectField)
				case typing.FunctionMedian, typing.FunctionMedianIf:
					raw += fmt.Sprintf("COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY %s), 0)", selectField)
				case typing.FunctionMin, typing.FunctionMinIf:
					raw += fmt.Sprintf("COALESCE(MIN(%s), 0)", selectField)
				case typing.FunctionMax, typing.FunctionMaxIf:
					raw += fmt.Sprintf("COALESCE(MAX(%s), 0)", selectField)
				}

				v.query.Select(Raw(raw))

				// Filter by this model's row's ID
				fk := fmt.Sprintf("r.\"%s\"", parser.FieldNameId)
				err = v.query.Where(Field(foreignKeyField), Equals, Raw(fk))
				if err != nil {
					return err
				}
				v.query.And()

			case 2: // the second arg sets the WHERE
				model = v.schema.FindModel(field.GetType().GetModelName().GetValue())

				r := proto.FindField(v.schema.GetModels(), v.model.GetName(), normalised[1])
				subFragments := normalised[1:]
				subFragments[0] = strcase.ToLowerCamel(r.GetType().GetModelName().GetValue())

				operand, err := generateOperand(v.ctx, v.schema, model, nil, map[string]any{}, subFragments)
				if err != nil {
					return err
				}

				err = v.query.AddJoinFromFragments(v.schema, subFragments)
				if err != nil {
					return err
				}

				v.operands.Push(operand)
			}

		} else {
			// Join together all the tables based on the ident fragments
			model = v.schema.FindModel(field.GetType().GetModelName().GetValue())
			query := NewQuery(model)

			relatedModelField := proto.FindField(v.schema.GetModels(), v.model.GetName(), normalised[1])
			subFragments := normalised[1:]
			subFragments[0] = strcase.ToLowerCamel(relatedModelField.GetType().GetModelName().GetValue())

			err := query.AddJoinFromFragments(v.schema, subFragments)
			if err != nil {
				return err
			}

			// Select the column as specified in the last ident fragment
			fieldName := normalised[len(normalised)-1]
			fragments := subFragments[:len(subFragments)-1]
			query.Select(ExpressionField(fragments, fieldName, false))

			// Filter by this model's row's ID
			foreignKeyField := proto.GetForeignKeyFieldName(v.schema.GetModels(), relatedModelField)

			fk := fmt.Sprintf("r.\"%s\"", strcase.ToSnake(foreignKeyField))
			err = query.Where(IdField(), Equals, Raw(fk))
			if err != nil {
				return err
			}

			stmt := query.SelectStatement()
			v.sql += fmt.Sprintf("(%s)", stmt.SqlTemplate())
		}
	}

	return nil
}

func (v *computedQueryGen) isToManyLookup(idents *parser.ExpressionIdent) (bool, error) {
	model := v.schema.FindModel(strcase.ToCamel(idents.Fragments[0]))

	fragments, err := NormalisedFragments(v.schema, idents.Fragments)
	if err != nil {
		return false, err
	}

	for i := 1; i < len(fragments)-1; i++ {
		currentFragment := fragments[i]
		field := proto.FindField(v.schema.GetModels(), model.GetName(), currentFragment)
		if field.GetType().GetType() == proto.Type_TYPE_MODEL && field.GetType().GetRepeated() {
			return true, nil
		}
		model = v.schema.FindModel(field.GetType().GetModelName().GetValue())
	}

	return false, nil
}

func (v *computedQueryGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return errors.New("ident arrays not supported in computed expressions")
}

func (v *computedQueryGen) Result() (string, error) {
	return cleanSql(v.sql), nil
}
