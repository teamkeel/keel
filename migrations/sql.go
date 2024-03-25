package migrations

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/auditing"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"golang.org/x/exp/slices"
)

var PostgresFieldTypes map[proto.Type]string = map[proto.Type]string{
	proto.Type_TYPE_ID:        "TEXT",
	proto.Type_TYPE_STRING:    "TEXT",
	proto.Type_TYPE_INT:       "INTEGER",
	proto.Type_TYPE_BOOL:      "BOOL",
	proto.Type_TYPE_TIMESTAMP: "TIMESTAMPTZ",
	proto.Type_TYPE_DATETIME:  "TIMESTAMPTZ",
	proto.Type_TYPE_DATE:      "DATE",
	proto.Type_TYPE_ENUM:      "TEXT",
	proto.Type_TYPE_SECRET:    "TEXT",
	proto.Type_TYPE_PASSWORD:  "TEXT",
}

// Matches the type cast on a Postgrs value eg. on "'foo'::text" matches "::text"
var typeCastRegex = regexp.MustCompile(`::(\w)+$`)

// Identifier converts v into an identifier that can be used
// for table or column names in Postgres. The value is converted
// to snake case and then quoted. The former is done to create
// a more idiomatic postgres schema and the latter is so you
// can have a table name called "select" that would otherwise
// not be allowed as it clashes with the keyword.
func Identifier(v string) string {
	return db.QuoteIdentifier(casing.ToSnake(v))
}

func UniqueConstraintName(modelName string, fieldNames []string) string {
	slices.Sort(fieldNames)
	snaked := lo.Map(fieldNames, func(s string, _ int) string {
		return casing.ToSnake(s)
	})
	return fmt.Sprintf("%s_%s_udx", casing.ToSnake(modelName), strings.Join(snaked, "_"))
}

func PrimaryKeyConstraintName(modelName string, fieldName string) string {
	return fmt.Sprintf("%s_%s_pkey", casing.ToSnake(modelName), casing.ToSnake(fieldName))
}

func createTableStmt(schema *proto.Schema, model *proto.Model) (string, error) {
	statements := []string{}
	output := fmt.Sprintf("CREATE TABLE %s (\n", Identifier(model.Name))

	// Exclude fields of type Model - these exists only in proto land - and has no corresponding
	// column in the database.
	fields := lo.Filter(model.Fields, func(field *proto.Field, _ int) bool {
		return field.Type.Type != proto.Type_TYPE_MODEL
	})

	for i, field := range fields {
		stmt, err := fieldDefinition(field)
		if err != nil {
			return "", err
		}
		output += stmt
		if i < len(fields)-1 {
			output += ","
		}
		output += "\n"
	}
	output += ");"
	statements = append(statements, output)

	for _, field := range fields {
		if field.PrimaryKey {
			statements = append(statements, fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s);",
				Identifier(model.Name),
				PrimaryKeyConstraintName(model.Name, field.Name),
				Identifier(field.Name)))
		}
		if field.Unique && !field.PrimaryKey {
			uniqueStmt, err := addUniqueConstraintStmt(schema, model.Name, []string{field.Name})
			if err != nil {
				return "", err
			}
			statements = append(statements, uniqueStmt)
		}
	}

	// Passing an empty slice of constraints here as this is a new table so no existing constraints
	stmts, err := compositeUniqueConstraints(schema, model, []*ConstraintRow{})
	if err != nil {
		return "", err
	}

	statements = append(statements, stmts...)

	return strings.Join(statements, "\n"), nil
}

func dropTableStmt(name string) string {
	return fmt.Sprintf("DROP TABLE %s CASCADE;", Identifier(name))
}

func addUniqueConstraintStmt(schema *proto.Schema, modelName string, fieldNames []string) (string, error) {
	slices.Sort(fieldNames)

	columnNames := []string{}
	for _, name := range fieldNames {
		field := proto.FindField(schema.Models, modelName, name)

		if proto.IsBelongsTo(field) {
			name = fmt.Sprintf("%sId", name)
		}

		if proto.IsHasMany(field) || proto.IsHasOne(field) {
			return "", fmt.Errorf("cannot create unique constraint on has-many or has-one model field '%s'", name)
		}

		columnNames = append(columnNames, Identifier(name))
	}

	return fmt.Sprintf(
		"ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s);",
		Identifier(modelName),
		UniqueConstraintName(modelName, fieldNames),
		strings.Join(columnNames, ", ")), nil
}

func dropConstraintStmt(tableName string, constraintName string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", Identifier(tableName), constraintName)
}

func addColumnStmt(schema *proto.Schema, modelName string, field *proto.Field) (string, error) {
	statements := []string{}

	stmt, err := fieldDefinition(field)
	if err != nil {
		return "", err
	}

	statements = append(statements,
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", Identifier(modelName), stmt),
	)

	if field.Unique && !field.PrimaryKey {
		stmt, err := addUniqueConstraintStmt(schema, modelName, []string{field.Name})
		if err != nil {
			return "", err
		}
		statements = append(statements, stmt)
	}

	return strings.Join(statements, "\n"), nil
}

// addForeignKeyConstraintStmt generates a string of this form:
// ALTER TABLE "thisTable" ADD FOREIGN KEY ("thisColumn") REFERENCES "otherTable"("otherColumn") ON DELETE CASCADE;
func addForeignKeyConstraintStmt(thisTable string, thisColumn string, otherTable string, otherColumn string, onDelete string) string {
	return fmt.Sprintf("ALTER TABLE %s ADD FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s;",
		thisTable,
		thisColumn,
		otherTable,
		otherColumn,
		onDelete,
	)
}

func alterColumnStmt(modelName string, field *proto.Field, column *ColumnRow) (string, error) {
	stmts := []string{}

	alterColumnStmtPrefix := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s", Identifier(modelName), Identifier(column.ColumnName))

	if field.DefaultValue == nil && column.HasDefault {
		output := fmt.Sprintf("%s DROP DEFAULT;", alterColumnStmtPrefix)
		stmts = append(stmts, output)
	}

	if field.DefaultValue != nil {
		value, err := getDefaultValue(field)
		if err != nil {
			return "", err
		}

		// Strip cast from default value e.g. 'Foo'::text -> 'Foo
		currentDefault := typeCastRegex.ReplaceAllString(column.DefaultValue, "")

		if !column.HasDefault || currentDefault != value {
			output := fmt.Sprintf("%s SET DEFAULT %s;", alterColumnStmtPrefix, value)
			stmts = append(stmts, output)
		}
	}

	// these two flags are opposites of each other, so if they are both true
	// or both false then there is a change to be applied
	if field.Optional == column.NotNull {
		var change string
		if field.Optional && column.NotNull {
			change = "DROP NOT NULL"
		} else {
			change = "SET NOT NULL"

			// Update all existing rows to the default value if they are null
			if field.DefaultValue != nil {
				value, err := getDefaultValue(field)
				if err != nil {
					return "", err
				}
				update := fmt.Sprintf("UPDATE %s SET %s = %s WHERE %s IS NULL;", Identifier(modelName), Identifier(column.ColumnName), value, Identifier(column.ColumnName))
				stmts = append(stmts, update)
			}

		}
		stmts = append(stmts, fmt.Sprintf("%s %s;", alterColumnStmtPrefix, change))
	}

	return strings.Join(stmts, "\n"), nil
}

func fieldDefinition(field *proto.Field) (string, error) {
	columnName := Identifier(field.Name)

	// We don't yet support Postgres JSON field types in Keel schemas.
	// But we need one for the special case of the keel_audit table.
	// So we hard code the JSON field type for now, for that special case.

	isAuditDataColumn := (field.ModelName == strcase.ToCamel(auditing.TableName)) && (field.Name == auditing.ColumnData)

	fieldType := lo.Ternary(
		isAuditDataColumn,
		"jsonb",
		PostgresFieldTypes[field.Type.Type])

	output := fmt.Sprintf("%s %s", columnName, fieldType)

	if !field.Optional {
		output += " NOT NULL"
	}

	if field.DefaultValue != nil {
		value, err := getDefaultValue(field)
		if err != nil {
			return "", err
		}

		output += " DEFAULT " + value
	}

	return output, nil
}

func getDefaultValue(field *proto.Field) (string, error) {
	if field.DefaultValue.UseZeroValue {
		switch field.Type.Type {
		case proto.Type_TYPE_STRING:
			return db.QuoteLiteral(""), nil
		case proto.Type_TYPE_INT:
			return "0", nil
		case proto.Type_TYPE_BOOL:
			return "false", nil
		case proto.Type_TYPE_DATE, proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
			return "now()", nil
		case proto.Type_TYPE_ID:
			return "ksuid()", nil
		}
	}

	expr, err := parser.ParseExpression(field.DefaultValue.Expression.Source)
	if err != nil {
		return "", err
	}

	value, err := expr.ToValue()
	if err != nil {
		return "", err
	}

	switch {
	case value.String != nil:
		s := *value.String
		// Remove wrapping quotes
		s = strings.TrimPrefix(s, `"`)
		s = strings.TrimSuffix(s, `"`)
		return db.QuoteLiteral(s), nil
	case value.Number != nil:
		return fmt.Sprintf("%d", *value.Number), nil
	case value.True:
		return "true", nil
	case value.False:
		return "false", nil
	case field.Type.Type == proto.Type_TYPE_ENUM && value.Ident != nil:
		if len(value.Ident.Fragments) != 2 {
			return "", fmt.Errorf("invalid default value %s for enum field %s", value.Ident.ToString(), field.Name)
		}

		return db.QuoteLiteral(value.Ident.Fragments[1].Fragment), nil
	default:
		return "", fmt.Errorf("field %s has unexpected default value %s", field.Name, value.ToString())
	}
}

func dropColumnStmt(modelName string, fieldName string) string {
	output := fmt.Sprintf("ALTER TABLE %s ", Identifier(modelName))
	output += fmt.Sprintf("DROP COLUMN %s;", Identifier(fieldName))
	return output
}

// createAuditTriggerStmts generates the CREATE TRIGGER statements for auditing.
// Only creates a trigger if the trigger does not already exist in the database.
func createAuditTriggerStmts(triggers []*TriggerRow, model *proto.Model) (string, error) {
	modelLower := casing.ToSnake(model.Name)
	statements := []string{}

	create := fmt.Sprintf("%s_create", modelLower)
	if _, found := lo.Find(triggers, func(t *TriggerRow) bool { return t.TriggerName == create && t.TableName == modelLower }); !found {
		statements = append(statements, fmt.Sprintf(
			`CREATE TRIGGER %s AFTER INSERT ON %s REFERENCING NEW TABLE AS new_table FOR EACH STATEMENT EXECUTE PROCEDURE process_audit();`, create, Identifier(model.Name)))
	}

	update := fmt.Sprintf("%s_update", modelLower)
	if _, found := lo.Find(triggers, func(t *TriggerRow) bool { return t.TriggerName == update && t.TableName == modelLower }); !found {
		statements = append(statements, fmt.Sprintf(
			`CREATE TRIGGER %s AFTER UPDATE ON %s REFERENCING NEW TABLE AS new_table OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE PROCEDURE process_audit();`, update, Identifier(model.Name)))
	}

	delete := fmt.Sprintf("%s_delete", modelLower)
	if _, found := lo.Find(triggers, func(t *TriggerRow) bool { return t.TriggerName == delete && t.TableName == modelLower }); !found {
		statements = append(statements, fmt.Sprintf(
			`CREATE TRIGGER %s AFTER DELETE ON %s REFERENCING OLD TABLE AS old_table FOR EACH STATEMENT EXECUTE PROCEDURE process_audit();`, delete, Identifier(model.Name)))
	}

	return strings.Join(statements, "\n"), nil
}

// createUpdatedAtTriggerStmts generates the CREATE TRIGGER statements for automatically updating each model's updatedAt column.
// Only creates a trigger if the trigger does not already exist in the database.
func createUpdatedAtTriggerStmts(triggers []*TriggerRow, model *proto.Model) (string, error) {
	modelLower := casing.ToSnake(model.Name)
	statements := []string{}

	updatedAt := fmt.Sprintf("%s_updated_at", modelLower)
	if _, found := lo.Find(triggers, func(t *TriggerRow) bool { return t.TriggerName == updatedAt && t.TableName == modelLower }); !found {
		statements = append(statements, fmt.Sprintf(
			`CREATE TRIGGER %s BEFORE UPDATE ON %s FOR EACH ROW EXECUTE PROCEDURE set_updated_at();`, updatedAt, Identifier(model.Name)))
	}

	return strings.Join(statements, "\n"), nil
}
