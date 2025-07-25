package permissions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
)

type ValueType int

const (
	ValueIdentityID      ValueType = iota // Identity ID of caller
	ValueIdentityEmail                    // Identity email of caller
	ValueIsAuthenticated                  // Is authenticated flag
	ValueNow                              // Current timestamp
	ValueHeader                           // Header value
	ValueSecret                           // Secret value
	ValueString                           // A string literal
	ValueNumber                           // A number literal
	ValueRecordIDs                        // The ID's of the records to check permission for
)

type Value struct {
	Type        ValueType
	StringValue string // Set if Type is ValueString
	NumberValue int64  // Set if Type is ValueNumber
	HeaderKey   string // Set if Type is ValueHeader
	SecretKey   string // Set if Type is ValueSecret
}

type permissionGen struct {
	schema *proto.Schema
	model  *proto.Model
	stmt   *statement
}

func gen(schema *proto.Schema, model *proto.Model, stmt *statement) resolve.Visitor[*statement] {
	return &permissionGen{
		schema: schema,
		model:  model,
		stmt:   stmt,
	}
}

func (v *permissionGen) StartTerm(nested bool) error {
	if nested {
		v.stmt.expression += "("
	}
	return nil
}
func (v *permissionGen) EndTerm(nested bool) error {
	if nested {
		v.stmt.expression += ")"
	}

	return nil
}

func (v *permissionGen) StartFunction(name string) error {
	return nil
}

func (v *permissionGen) EndFunction() error {
	return nil
}

func (v *permissionGen) StartArgument(num int) error {
	return nil
}

func (v *permissionGen) EndArgument() error {
	return nil
}

func (v *permissionGen) VisitAnd() error {
	v.stmt.expression += " and "
	return nil
}
func (v *permissionGen) VisitOr() error {
	v.stmt.expression += " or "
	return nil
}
func (v *permissionGen) VisitNot() error {
	return nil
}
func (v *permissionGen) VisitOperator(op string) error {
	v.stmt.expression += operatorMap[op][0]
	return nil
}
func (v *permissionGen) VisitLiteral(value any) error {
	if value == nil {
		v.stmt.expression += "null"
		return nil
	}

	switch value := value.(type) {
	case int64:
		v.stmt.expression += "?"
		v.stmt.values = append(v.stmt.values, &Value{Type: ValueNumber, NumberValue: value})
	case uint64:
		v.stmt.expression += "?"
		v.stmt.values = append(v.stmt.values, &Value{Type: ValueNumber, NumberValue: int64(value)})
	case string:
		v.stmt.expression += "?"
		v.stmt.values = append(v.stmt.values, &Value{Type: ValueString, StringValue: fmt.Sprintf("\"%s\"", value)})
	case bool:
		if value {
			v.stmt.expression += "true"
		} else {
			v.stmt.expression += "false"
		}
	default:
		return fmt.Errorf("unsupported type in permission expression")
	}

	return nil
}
func (v *permissionGen) VisitIdent(ident *parser.ExpressionIdent) error {
	return handleOperand(v.schema, v.model, ident, v.stmt)
}
func (v *permissionGen) VisitIdentArray(idents []*parser.ExpressionIdent) error {
	return nil
}
func (v *permissionGen) Result() (*statement, error) {
	return v.stmt, nil
}

var _ resolve.Visitor[*statement] = new(permissionGen)

// ToSQL creates a single SQL query that can be run to determine if permission is granted for the
// given action and a set of records.
//
// The returned SQL uses "?" placeholders for values and the returned list of values indicates
// what values should be provided to the query at runtime.
func ToSQL(s *proto.Schema, m *proto.Model, permissions []*proto.PermissionRule) (sql string, values []*Value, err error) {
	tableName := identifier(m.GetName())
	pkField := identifier(m.PrimaryKeyFieldName())

	stmt := &statement{
		joins:  []string{},
		values: []*Value{},
	}

	for _, p := range permissions {
		if p.GetExpression() == nil {
			continue
		}

		// Combine multiple permission rules with "or"
		if stmt.expression != "" {
			stmt.expression += " or "
		}

		expr, err := parser.ParseExpression(p.GetExpression().GetSource())
		if err != nil {
			return sql, values, err
		}

		stmt, err = resolve.RunCelVisitor(expr, gen(s, m, stmt))
		if err != nil {
			return "", nil, err
		}
	}

	if stmt.expression == "" {
		return sql, values, nil
	}

	sql = fmt.Sprintf("SELECT DISTINCT %s.%s", tableName, pkField)
	sql += fmt.Sprintf(" FROM %s", tableName)
	if len(stmt.joins) > 0 {
		// lo.Unique to dedupe joins
		sql += " " + strings.Join(lo.Uniq(stmt.joins), " ")
	}

	expr := stmt.expression
	if len(permissions) > 1 {
		expr = fmt.Sprintf("(%s)", expr)
	}

	sql += fmt.Sprintf(" WHERE %s AND %s.%s IN (?)", expr, tableName, pkField)

	stmt.values = append(stmt.values, &Value{
		Type: ValueRecordIDs,
	})

	return sql, stmt.values, nil
}

func handleOperand(s *proto.Schema, model *proto.Model, ident *parser.ExpressionIdent, stmt *statement) (err error) {
	switch ident.Fragments[0] {
	case "ctx":
		return handleContext(s, ident, stmt)
	case casing.ToLowerCamel(model.GetName()):
		return handleModel(s, model, ident, stmt)
	default:
		// If not context of model must be enum, but still worth checking to be sure
		enum := proto.FindEnum(s.GetEnums(), ident.Fragments[0])
		if enum == nil {
			return fmt.Errorf("unknown ident %s", ident.Fragments[0])
		}

		stmt.expression += "?"
		stmt.values = append(stmt.values, &Value{Type: ValueString, StringValue: ident.Fragments[len(ident.Fragments)-1]})

		return nil
	}
}

func handleContext(s *proto.Schema, ident *parser.ExpressionIdent, stmt *statement) error {
	if len(ident.Fragments) < 2 {
		return errors.New("ctx used in expression with no properties")
	}

	switch ident.Fragments[1] {
	case "identity":
		// ctx.identity is the same as ctx.identity.id
		if len(ident.Fragments) == 2 {
			stmt.expression += "?"
			stmt.values = append(stmt.values, &Value{Type: ValueIdentityID})
			return nil
		}
		switch ident.Fragments[2] {
		case "id":
			stmt.expression += "?"
			stmt.values = append(stmt.values, &Value{Type: ValueIdentityID})
			return nil
		case "email":
			stmt.expression += "?"
			stmt.values = append(stmt.values, &Value{Type: ValueIdentityEmail})
			return nil
		default:
			inner := &statement{}
			err := handleModel(
				s,
				s.FindModel("Identity"),
				&parser.ExpressionIdent{
					// We can drop the first fragments, which is "ctx"
					Fragments: ident.Fragments[1:],
				},
				inner,
			)
			if err != nil {
				return err
			}

			// The RHS is a subquery so we need to replace the IS NOT DISTINCT FROM operator with IN
			if strings.HasSuffix(stmt.expression, " IS NOT DISTINCT FROM ") {
				stmt.expression = strings.TrimSuffix(stmt.expression, " IS NOT DISTINCT FROM ") + " IN "
			}

			stmt.expression += fmt.Sprintf(
				`(SELECT %s FROM "identity" %s WHERE "identity"."id" IS NOT DISTINCT FROM ?)`,
				inner.expression, strings.Join(inner.joins, " "),
			)
			stmt.values = append(stmt.values, &Value{Type: ValueIdentityID})
			return nil
		}
	case "isAuthenticated":
		// Explicit cast to boolean as Kysely seems to send value as string
		stmt.expression += "?::boolean"
		stmt.values = append(stmt.values, &Value{Type: ValueIsAuthenticated})
		return nil
	case "now":
		stmt.expression += "?"
		stmt.values = append(stmt.values, &Value{Type: ValueNow})
		return nil
	case "headers":
		stmt.expression += "?"
		key := ident.Fragments[2]
		stmt.values = append(stmt.values, &Value{Type: ValueHeader, HeaderKey: key})
		return nil
	case "secrets":
		stmt.expression += "?"
		key := ident.Fragments[2]
		stmt.values = append(stmt.values, &Value{Type: ValueSecret, SecretKey: key})
		return nil
	default:
		return fmt.Errorf("unknown property %s of ctx", ident.Fragments[1])
	}
}

func handleModel(s *proto.Schema, model *proto.Model, ident *parser.ExpressionIdent, stmt *statement) (err error) {
	fieldName := ""
	for i, f := range ident.Fragments {
		switch {
		// The first fragment
		case i == 0:
			fieldName += casing.ToSnake(f)

		// Remaining fragments
		default:
			field := proto.FindField(s.GetModels(), model.GetName(), f)
			if field == nil {
				return fmt.Errorf("model %s has no field %s", model.GetName(), f)
			}

			isLast := i == len(ident.Fragments)-1
			isModel := field.GetType().GetType() == proto.Type_TYPE_MODEL
			hasFk := field.GetForeignKeyFieldName() != nil

			if isModel && (!isLast || !hasFk) {
				// Left alias is the source table
				leftAlias := fieldName

				// Append fragment to identifier
				fieldName += "$" + casing.ToSnake(f)

				// Right alias is the join table
				rightAlias := fieldName

				field := proto.FindField(s.GetModels(), model.GetName(), f)
				if field == nil {
					return fmt.Errorf("model %s has no field %s", model.GetName(), f)
				}

				joinModel := s.FindModel(field.GetType().GetModelName().GetValue())
				if joinModel == nil {
					return fmt.Errorf("model %s not found in schema", model.GetName())
				}

				leftFieldName := proto.GetForeignKeyFieldName(s.GetModels(), field)
				rightFieldName := joinModel.PrimaryKeyFieldName()

				// If not belongs to then swap foreign/primary key
				if !field.IsBelongsTo() {
					leftFieldName = model.PrimaryKeyFieldName()
					rightFieldName = proto.GetForeignKeyFieldName(s.GetModels(), field)
				}

				stmt.joins = append(stmt.joins, fmt.Sprintf(
					"LEFT JOIN %s AS %s ON %s.%s = %s.%s",
					identifier(joinModel.GetName()),
					identifier(rightAlias),
					identifier(leftAlias),
					identifier(leftFieldName),
					identifier(rightAlias),
					identifier(rightFieldName),
				))

				model = joinModel
			}

			if isLast {
				// Turn the table into a quoted identifier
				fieldName = identifier(fieldName)

				// Then append the field name as a quoted identifier
				if field.GetType().GetType() == proto.Type_TYPE_MODEL {
					if field.GetForeignKeyFieldName() != nil {
						fieldName += "." + identifier(field.GetForeignKeyFieldName().GetValue())
					} else {
						fieldName += "." + identifier("id")
					}
				} else {
					fieldName += "." + identifier(f)
				}
			}
		}
	}

	sql := fieldName
	stmt.expression += sql
	return nil
}

// identifier converts s to snake cases and wraps it in double quotes.
func identifier(s string) string {
	return db.QuoteIdentifier(casing.ToSnake(s))
}

type statement struct {
	expression string
	joins      []string
	values     []*Value
}

// Map of Keel expression operators to SQL operators.
// SQL operators can be provided as just a simple value
// or as a pair of opening and closing values.
var operatorMap = map[string][]string{
	"_==_": {" IS NOT DISTINCT FROM "},
	"_!=_": {" IS DISTINCT FROM "},
	"_<_":  {" < "},
	"_<=_": {" <= "},
	"_>_":  {" > "},
	"_>=_": {" >= "},
	"@in":  {" IS NOT DISTINCT FROM "},
}
