package permissions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/db"
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
	NumberValue int    // Set if Type is ValueNumber
	HeaderKey   string // Set if Type is ValueHeader
	SecretKey   string // Set if Type is ValueSecret
}

// ToSQL creates a single SQL query that can be run to determine if permission is granted for the
// given action and a set of records.
//
// The returned SQL uses "?" placeholders for values and the returned list of values indicates
// what values should be provided to the query at runtime.
func ToSQL(s *proto.Schema, m *proto.Model, action *proto.Action) (sql string, values []*Value, err error) {
	tableName := identifier(m.Name)
	pkField := identifier(m.PrimaryKeyFieldName())

	stmt := &statement{}
	permissions := proto.PermissionsForAction(s, action)

	for _, p := range permissions {
		if p.Expression == nil {
			continue
		}

		// Combine multiple permission rules with "or"
		if stmt.expression != "" {
			stmt.expression += " or "
		}

		expr, err := parser.ParseExpression(p.Expression.Source)
		if err != nil {
			return sql, values, err
		}

		err = handleExpression(s, m, expr, stmt)
		if err != nil {
			return sql, values, err
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

func handleExpression(s *proto.Schema, m *proto.Model, expr *parser.Expression, stmt *statement) (err error) {
	stmt.expression += "("
	for i, or := range expr.Or {
		if i > 0 {
			stmt.expression += " or "
		}
		for j, and := range or.And {
			if j > 0 {
				stmt.expression += " and "
			}

			if and.Expression != nil {
				err = handleExpression(s, m, and.Expression, stmt)
				if err != nil {
					return err
				}
				continue
			}

			cond := and.Condition
			err = handleOperand(s, m, cond.LHS, stmt)
			if err != nil {
				return err
			}

			// If no RHS then we're done
			if cond.Operator == nil {
				continue
			}

			op := operatorMap[cond.Operator.Symbol]
			opOpen := op[0]
			opClose := ""
			if len(op) == 2 {
				opClose = op[1]
			}

			stmt.expression += opOpen

			err = handleOperand(s, m, cond.RHS, stmt)
			if err != nil {
				return err
			}

			if opClose != "" {
				stmt.expression += opClose
			}
		}
	}

	stmt.expression += ")"
	return nil
}

func handleOperand(s *proto.Schema, model *proto.Model, o *parser.Operand, stmt *statement) (err error) {
	switch {
	case o.True:
		stmt.expression += "true"
		return nil
	case o.False:
		stmt.expression += "false"
		return nil
	case o.String != nil:
		stmt.expression += "?"
		stmt.values = append(stmt.values, &Value{Type: ValueString, StringValue: *o.String})
		return nil
	case o.Number != nil:
		stmt.expression += "?"
		stmt.values = append(stmt.values, &Value{Type: ValueNumber, NumberValue: int(*o.Number)})
		return nil
	case o.Null:
		stmt.expression += "null"
		return nil
	case o.Array != nil:
		return errors.New("arrays in permission rules not yet supported")
	case o.Ident != nil:
		switch o.Ident.Fragments[0].Fragment {
		case "ctx":
			return handleContext(s, o, stmt)
		case casing.ToLowerCamel(model.Name):
			return handleModel(s, model, o.Ident, stmt)
		default:
			// If not context of model must be enum, but still worth checking to be sure
			enum := proto.FindEnum(s.Enums, o.Ident.Fragments[0].Fragment)
			if enum == nil {
				return fmt.Errorf("unknown ident %s", o.Ident.Fragments[0].Fragment)
			}

			stmt.expression += "?"
			stmt.values = append(stmt.values, &Value{Type: ValueString, StringValue: o.Ident.LastFragment()})

			return nil
		}
	}

	return fmt.Errorf("unsupported operand: %s", o.ToString())
}

func handleContext(s *proto.Schema, o *parser.Operand, stmt *statement) error {
	if len(o.Ident.Fragments) < 2 {
		return errors.New("ctx used in expression with no properties")
	}

	switch o.Ident.Fragments[1].Fragment {
	case "identity":
		// ctx.identity is the same as ctx.identity.id
		if len(o.Ident.Fragments) == 2 {
			stmt.expression += "?"
			stmt.values = append(stmt.values, &Value{Type: ValueIdentityID})
			return nil
		}
		switch o.Ident.Fragments[2].Fragment {
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
				&parser.Ident{
					// We can drop the first fragments, which is "ctx"
					Fragments: o.Ident.Fragments[1:],
				},
				inner,
			)
			if err != nil {
				return err
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
		key := o.Ident.Fragments[2].Fragment
		stmt.values = append(stmt.values, &Value{Type: ValueHeader, HeaderKey: key})
		return nil
	case "secrets":
		stmt.expression += "?"
		key := o.Ident.Fragments[2].Fragment
		stmt.values = append(stmt.values, &Value{Type: ValueSecret, SecretKey: key})
		return nil
	default:
		return fmt.Errorf("unknown property %s of ctx", o.Ident.Fragments[1].Fragment)
	}
}

func handleModel(s *proto.Schema, model *proto.Model, ident *parser.Ident, stmt *statement) (err error) {
	fieldName := ""
	for i, f := range ident.Fragments {
		switch {
		// The first fragment
		case i == 0:
			fieldName += casing.ToSnake(f.Fragment)

		// Remaining fragments
		default:
			field := proto.FindField(s.Models, model.Name, f.Fragment)
			if field == nil {
				return fmt.Errorf("model %s has no field %s", model.Name, f.Fragment)
			}

			isLast := i == len(ident.Fragments)-1
			isModel := field.Type.Type == proto.Type_TYPE_MODEL
			hasFk := field.ForeignKeyFieldName != nil

			if isModel && (!isLast || !hasFk) {
				// Left alias is the source table
				leftAlias := fieldName

				// Append fragment to identifier
				fieldName += "$" + casing.ToSnake(f.Fragment)

				// Right alias is the join table
				rightAlias := fieldName

				field := proto.FindField(s.Models, model.Name, f.Fragment)
				if field == nil {
					return fmt.Errorf("model %s has no field %s", model.Name, f.Fragment)
				}

				joinModel := proto.FindModel(s.Models, field.Type.ModelName.Value)
				if joinModel == nil {
					return fmt.Errorf("model %s not found in schema", model.Name)
				}

				leftFieldName := proto.GetForeignKeyFieldName(s.Models, field)
				rightFieldName := joinModel.PrimaryKeyFieldName()

				// If not belongs to then swap foreign/primary key
				if !field.IsBelongsTo() {
					leftFieldName = model.PrimaryKeyFieldName()
					rightFieldName = proto.GetForeignKeyFieldName(s.Models, field)
				}

				stmt.joins = append(stmt.joins, fmt.Sprintf(
					"LEFT JOIN %s AS %s ON %s.%s = %s.%s",
					identifier(joinModel.Name),
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
				if field.Type.Type == proto.Type_TYPE_MODEL {
					if field.ForeignKeyFieldName != nil {
						fieldName += "." + identifier(field.ForeignKeyFieldName.Value)
					} else {
						fieldName += "." + identifier("id")
					}
				} else {
					fieldName += "." + identifier(f.Fragment)
				}
			}
		}
	}

	sql := fieldName
	stmt.expression += sql
	return nil
}

// identifier converts s to snake cases and wraps it in double quotes
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
// or as a pair of opening and closing values
var operatorMap = map[string][]string{
	"==": {" IS NOT DISTINCT FROM "},
	"!=": {" IS DISTINCT FROM "},
	"<":  {" < "},
	"<=": {" <= "},
	">":  {" > "},
	">=": {" >= "},
	"in": {" IS NOT DISTINCT FROM "},
}
