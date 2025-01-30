package actions

import (
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

// GenerateComputedFunction visits the expression and generates a SQL expression
func GenerateComputedFunction(schema *proto.Schema, model *proto.Model, field *proto.Field) resolve.Visitor[string] {
	return &computedQueryGen{
		schema:    schema,
		model:     model,
		field:     field,
		sql:       "",
		functions: arraystack.New(),
	}
}

var _ resolve.Visitor[string] = new(computedQueryGen)

type computedQueryGen struct {
	schema    *proto.Schema
	model     *proto.Model
	field     *proto.Field
	sql       string
	functions *arraystack.Stack
}

func (v *computedQueryGen) StartTerm(nested bool) error {
	if nested {
		v.sql += "("
	}
	return nil
}

func (v *computedQueryGen) EndTerm(nested bool) error {
	if nested {
		v.sql += ")"
	}
	return nil
}

func (v *computedQueryGen) StartFunction(name string) error {
	v.functions.Push(name)
	return nil
}

func (v *computedQueryGen) EndFunction() error {
	return nil
}

func (v *computedQueryGen) VisitAnd() error {
	v.sql += " AND "
	return nil
}

func (v *computedQueryGen) VisitOr() error {
	v.sql += " OR "
	return nil
}

func (v *computedQueryGen) VisitNot() error {
	v.sql += " NOT "
	return nil
}

func (v *computedQueryGen) VisitOperator(op string) error {
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

	if sqlOp == "" {
		return fmt.Errorf("unsupported operator: %s", op)
	}

	v.sql += " " + sqlOp + " "
	return nil
}

func (v *computedQueryGen) VisitLiteral(value any) error {
	switch val := value.(type) {
	case int64:
		v.sql += fmt.Sprintf("%v", val)
	case float64:
		v.sql += fmt.Sprintf("%v", val)
	case string:
		v.sql += fmt.Sprintf("\"%v\"", val)
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
	field := proto.FindField(v.schema.Models, model.Name, ident.Fragments[1])

	normalised, err := NormalisedFragments(v.schema, ident.Fragments)
	if err != nil {
		return err
	}

	if len(normalised) == 2 {
		v.sql += "r." + sqlQuote(strcase.ToSnake(field.Name))
	} else if len(normalised) > 2 {
		isToMany, err := v.isToManyLookup(ident)
		if err != nil {
			return err
		}

		if isToMany {
			model = v.schema.FindModel(field.Type.ModelName.Value)
			query := NewQuery(model)

			relatedModelField := proto.FindField(v.schema.Models, v.model.Name, normalised[1])
			foreignKeyField := proto.GetForeignKeyFieldName(v.schema.Models, relatedModelField)

			r := proto.FindField(v.schema.Models, v.model.Name, normalised[1])
			subFragments := normalised[1:]
			subFragments[0] = strcase.ToLowerCamel(r.Type.ModelName.Value)

			err := query.AddJoinFromFragments(v.schema, subFragments)
			if err != nil {
				return err
			}

			funcBegin, has := v.functions.Pop()
			if !has {
				return errors.New("no function found for 1:M lookup")
			}

			fieldName := normalised[len(normalised)-1]
			fragments := normalised[1 : len(normalised)-1]

			raw := ""
			selectField := sqlQuote(casing.ToSnake(strings.Join(fragments, "$"))) + "." + sqlQuote(casing.ToSnake(fieldName))
			switch funcBegin {
			case typing.FunctionSum:
				raw += fmt.Sprintf("COALESCE(SUM(%s), 0)", selectField)
			case typing.FunctionCount:
				raw += fmt.Sprintf("COALESCE(COUNT(%s), 0)", selectField)
			case typing.FunctionAvg:
				raw += fmt.Sprintf("COALESCE(AVG(%s), 0)", selectField)
			case typing.FunctionMedian:
				raw += fmt.Sprintf("COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY %s), 0)", selectField)
			case typing.FunctionMin:
				raw += fmt.Sprintf("COALESCE(MIN(%s), 0)", selectField)
			case typing.FunctionMax:
				raw += fmt.Sprintf("COALESCE(MAX(%s), 0)", selectField)
			}

			query.Select(Raw(raw))

			// Filter by this model's row's ID
			fk := fmt.Sprintf("r.\"%s\"", parser.FieldNameId)
			err = query.Where(Field(foreignKeyField), Equals, Raw(fk))
			if err != nil {
				return err
			}

			stmt := query.SelectStatement()
			v.sql += fmt.Sprintf("(%s)", stmt.SqlTemplate())
		} else {
			// Join together all the tables based on the ident fragments
			model = v.schema.FindModel(field.Type.ModelName.Value)
			query := NewQuery(model)
			err := query.AddJoinFromFragments(v.schema, normalised[1:])
			if err != nil {
				return err
			}

			// Select the column as specified in the last ident fragment
			fieldName := normalised[len(normalised)-1]
			fragments := normalised[1 : len(normalised)-1]
			query.Select(ExpressionField(fragments, fieldName, false))

			// Filter by this model's row's ID
			relatedModelField := proto.FindField(v.schema.Models, v.model.Name, normalised[1])
			foreignKeyField := proto.GetForeignKeyFieldName(v.schema.Models, relatedModelField)

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
		field := proto.FindField(v.schema.Models, model.Name, currentFragment)
		if field.Type.Type == proto.Type_TYPE_MODEL && field.Type.Repeated {
			return true, nil
		}
		model = v.schema.FindModel(field.Type.ModelName.Value)
	}

	return false, nil
}

func (v *computedQueryGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return errors.New("ident arrays not supported in computed expressions")
}

func (v *computedQueryGen) Result() (string, error) {
	return cleanSql(v.sql), nil
}
