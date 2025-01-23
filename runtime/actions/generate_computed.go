package actions

import (
	"errors"
	"fmt"

	"github.com/google/cel-go/common/operators"
	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"

	"github.com/teamkeel/keel/schema/parser"
)

// GenerateComputedFunction visits the expression and generates a SQL expression
func GenerateComputedFunction(schema *proto.Schema, model *proto.Model, field *proto.Field) resolve.Visitor[string] {
	return &computedQueryGen{
		schema: schema,
		model:  model,
		field:  field,
		sql:    "",
	}
}

var _ resolve.Visitor[string] = new(computedQueryGen)

type computedQueryGen struct {
	schema *proto.Schema
	model  *proto.Model
	field  *proto.Field
	sql    string
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

	if len(ident.Fragments) == 2 {
		v.sql += "r." + sqlQuote(strcase.ToSnake(field.Name))
	} else if len(ident.Fragments) > 2 {
		// Join together all the tables based on the ident fragments
		model = v.schema.FindModel(field.Type.ModelName.Value)
		query := NewQuery(model)
		err := query.AddJoinFromFragments(v.schema, ident.Fragments[1:])
		if err != nil {
			return err
		}

		// Select the column as specified in the last ident fragment
		fieldName := ident.Fragments[len(ident.Fragments)-1]
		fragments := ident.Fragments[1 : len(ident.Fragments)-1]
		query.Select(ExpressionField(fragments, fieldName, false))

		// Filter by this model's row's ID
		relatedModelField := proto.FindField(v.schema.Models, v.model.Name, ident.Fragments[1])
		foreignKeyField := proto.GetForeignKeyFieldName(v.schema.Models, relatedModelField)
		fk := fmt.Sprintf("r.\"%s\"", strcase.ToSnake(foreignKeyField))
		err = query.Where(IdField(), Equals, Raw(fk))
		if err != nil {
			return err
		}

		stmt := query.SelectStatement()
		v.sql += fmt.Sprintf("(%s)", stmt.SqlTemplate())
	}
	return nil
}

func (v *computedQueryGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return errors.New("ident arrays not supported in computed expressions")
}

func (v *computedQueryGen) Result() (string, error) {
	return cleanSql(v.sql), nil
}
