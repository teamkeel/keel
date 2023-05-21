package migrations

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
)

var PostgresFieldTypes map[proto.Type]string = map[proto.Type]string{
	proto.Type_TYPE_ID:        "TEXT",
	proto.Type_TYPE_STRING:    "TEXT",
	proto.Type_TYPE_INT:       "INTEGER",
	proto.Type_TYPE_BOOL:      "BOOL",
	proto.Type_TYPE_TIMESTAMP: "TIMESTAMPTZ",
	proto.Type_TYPE_DATETIME:  "TIMESTAMPTZ",
	proto.Type_TYPE_DATE:      "DATE",
	proto.Type_TYPE_MODEL:     "TEXT", // id of the target
	proto.Type_TYPE_ENUM:      "TEXT",
	proto.Type_TYPE_SECRET:    "TEXT",
	proto.Type_TYPE_PASSWORD:  "TEXT",
}

// Identifier converts v into an identifier that can be used
// for table or column names in Postgres. The value is converted
// to snake case and then quoted. The former is done to create
// a more idiomatic postgres schema and the latter is so you
// can have a table name called "select" that would otherwise
// not be allowed as it clashes with the keyword.
func Identifier(v string) string {
	return db.QuoteIdentifier(casing.ToSnake(v))
}

func UniqueConstraintName(modelName string, fieldName string) string {
	return fmt.Sprintf("%s_%s_udx", casing.ToSnake(modelName), casing.ToSnake(fieldName))
}

func PrimaryKeyConstraintName(modelName string, fieldName string) string {
	return fmt.Sprintf("%s_%s_pkey", casing.ToSnake(modelName), casing.ToSnake(fieldName))
}

func createTableStmt(model *proto.Model) string {
	statements := []string{}
	output := fmt.Sprintf("CREATE TABLE %s (\n", Identifier(model.Name))

	// This type of field exists only in proto land - and has no corresponding
	// column in the database.
	fields := lo.Filter(model.Fields, func(field *proto.Field, _ int) bool {
		return field.Type.Type != proto.Type_TYPE_MODEL
	})

	for i, field := range fields {
		output += fieldDefinition(field)
		if i < len(fields)-1 {
			output += ","
		}
		output += "\n"
	}
	output += ");"
	statements = append(statements, output)

	for _, field := range fields {
		if field.Unique {
			statements = append(statements, fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s);",
				Identifier(model.Name),
				UniqueConstraintName(model.Name, field.Name),
				Identifier(field.Name)))
		}
		if field.PrimaryKey {
			statements = append(statements, fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s);",
				Identifier(model.Name),
				PrimaryKeyConstraintName(model.Name, field.Name),
				Identifier(field.Name)))
		}
	}

	return strings.Join(statements, "\n")
}

func dropTableStmt(name string) string {
	return fmt.Sprintf("DROP TABLE %s CASCADE;", Identifier(name))
}

func addColumnStmt(modelName string, field *proto.Field) string {
	statements := []string{}

	statements = append(statements,
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", Identifier(modelName), fieldDefinition(field)),
	)

	if field.Unique {
		statements = append(statements,
			fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s);",
				Identifier(modelName),
				UniqueConstraintName(modelName, field.Name),
				Identifier(field.Name)),
		)
	}

	return strings.Join(statements, "\n")
}

// addForeignKeyConstraintStmt generates a string of this form:
// ALTER TABLE "thisTable" ADD FOREIGN KEY ("thisColumn") REFERENCES "otherTable"("otherColumn")
func addForeignKeyConstraintStmt(thisTable string, thisColumn string, otherTable string, otherColumn string, onDelete string) string {
	return fmt.Sprintf("ALTER TABLE %s ADD FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s;",
		thisTable,
		thisColumn,
		otherTable,
		otherColumn,
		onDelete,
	)
}

func alterColumnStmt(modelName string, newField *proto.Field, currField *proto.Field) string {
	stmts := []string{}

	if newField.Optional != currField.Optional {
		output := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s", Identifier(modelName), Identifier(currField.Name))

		if newField.Optional && !currField.Optional {
			output += " DROP NOT NULL"
		}
		if !newField.Optional && currField.Optional {
			output += " SET NOT NULL"
		}
		output += ";"
		stmts = append(stmts, output)
	}

	if newField.Unique != currField.Unique {
		constraintName := UniqueConstraintName(modelName, newField.Name)
		output := fmt.Sprintf("ALTER TABLE %s ", Identifier(modelName))
		if !newField.Unique {
			output += fmt.Sprintf("DROP CONSTRAINT %s", constraintName)
		} else {
			output += fmt.Sprintf("ADD CONSTRAINT %s UNIQUE (%s)", constraintName, Identifier(newField.Name))
		}
		output += ";"
		stmts = append(stmts, output)
	}

	return strings.Join(stmts, "\n")
}

func fieldDefinition(field *proto.Field) string {
	columnName := Identifier(field.Name)
	output := fmt.Sprintf("%s %s", columnName, PostgresFieldTypes[field.Type.Type])

	if !field.Optional {
		output += " NOT NULL"
	}

	return output
}

func dropColumnStmt(modelName string, field *proto.Field) string {
	output := fmt.Sprintf("ALTER TABLE %s ", Identifier(modelName))
	output += fmt.Sprintf("DROP COLUMN %s;", Identifier(field.Name))
	return output
}

func SelectSingleColumn(tableName string, columnName string) string {
	return fmt.Sprintf("SELECT \"%s\" FROM \"%s\";", columnName, tableName)
}

func InsertRowComprisingSingleString(tableName string, theString string) string {
	output := fmt.Sprintf("INSERT INTO \"%s\"\n", tableName)
	output += fmt.Sprintf("VALUES ('%s');", theString)
	return output
}

func UpdateSingleStringColumn(tableName string, column string, newValue string) string {
	output := fmt.Sprintf("UPDATE \"%s\" SET \"%s\"='%s';", tableName, column, newValue)
	return output
}
