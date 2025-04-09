package migrations

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/auditing"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/schema/parser"
	"golang.org/x/exp/slices"
)

var PostgresFieldTypes map[proto.Type]string = map[proto.Type]string{
	proto.Type_TYPE_ID:        "TEXT",
	proto.Type_TYPE_STRING:    "TEXT",
	proto.Type_TYPE_INT:       "INTEGER",
	proto.Type_TYPE_DECIMAL:   "NUMERIC",
	proto.Type_TYPE_BOOL:      "BOOL",
	proto.Type_TYPE_TIMESTAMP: "TIMESTAMPTZ",
	proto.Type_TYPE_DATETIME:  "TIMESTAMPTZ",
	proto.Type_TYPE_DATE:      "DATE",
	proto.Type_TYPE_ENUM:      "TEXT",
	proto.Type_TYPE_SECRET:    "TEXT",
	proto.Type_TYPE_PASSWORD:  "TEXT",
	proto.Type_TYPE_MARKDOWN:  "TEXT",
	proto.Type_TYPE_VECTOR:    "VECTOR",
	proto.Type_TYPE_FILE:      "JSONB",
	proto.Type_TYPE_DURATION:  "INTERVAL",
}

// Matches the type cast on a Postgrs value eg. on "'foo'::text" matches "::text"
var typeCastRegex = regexp.MustCompile(`::([\w\s]+)(?:\[\])?$`)

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

		if field.IsBelongsTo() {
			name = fmt.Sprintf("%sId", name)
		}

		if field.IsHasMany() || field.IsHasOne() {
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
// ALTER TABLE "thisTable" ADD FOREIGN KEY ("thisColumn") REFERENCES "otherTable"("otherColumn") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;
func addForeignKeyConstraintStmt(thisTable string, thisColumn string, otherTable string, otherColumn string, onDelete string) string {
	return fmt.Sprintf("ALTER TABLE %s ADD FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s DEFERRABLE INITIALLY IMMEDIATE;",
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

		// Strip cast from default value e.g. 'Foo'::text -> 'Foo'
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
			// If computed, then we don't set the NOT NULL constraint yet
			// This is because we may still need to populate existing rows
			if field.ComputedExpression == nil {
				change = "SET NOT NULL"
			}

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
		if change != "" {
			stmts = append(stmts, fmt.Sprintf("%s %s;", alterColumnStmtPrefix, change))
		}
	}

	return strings.Join(stmts, "\n"), nil
}

func hashOfExpression(expression string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(expression)))[:8]
}

// computedFieldFuncName generates the name of the a computed field's function
func computedFieldFuncName(field *proto.Field) string {
	// shortened alphanumeric hash from an expression
	hash := hashOfExpression(field.ComputedExpression.Source)
	return fmt.Sprintf("%s__%s__%s__comp", strcase.ToSnake(field.ModelName), strcase.ToSnake(field.Name), hash)
}

// computedExecFuncName generates the name for the table function which executed all computed functions
func computedExecFuncName(model *proto.Model) string {
	return fmt.Sprintf("%s__exec_comp_fns", strcase.ToSnake(model.Name))
}

// computedTriggerName generates the name for the trigger which runs the function which executes computed functions
func computedTriggerName(model *proto.Model) string {
	return fmt.Sprintf("%s__comp", strcase.ToSnake(model.Name))
}

func computedDependencyFuncName(model *proto.Model, dependentModel *proto.Model, fragments []string) string {
	hash := hashOfExpression(strings.Join(fragments, "."))
	return fmt.Sprintf("%s__to__%s__%s__comp_dep", strcase.ToSnake(dependentModel.Name), strcase.ToSnake(model.Name), hash)
}

// fieldFromComputedFnName determines the field from computed function name
func fieldFromComputedFnName(schema *proto.Schema, fn string) *proto.Field {
	parts := strings.Split(fn, "__")
	model := schema.FindModel(strcase.ToCamel(parts[0]))
	for _, f := range model.Fields {
		if f.Name == strcase.ToLowerCamel(parts[1]) {
			return f
		}
	}
	return nil
}

// addComputedFieldFuncStmt generates the function for a computed field
func addComputedFieldFuncStmt(schema *proto.Schema, model *proto.Model, field *proto.Field) (string, string, error) {
	var sqlType string
	switch field.Type.Type {
	case proto.Type_TYPE_DECIMAL, proto.Type_TYPE_INT, proto.Type_TYPE_BOOL, proto.Type_TYPE_STRING:
		sqlType = PostgresFieldTypes[field.Type.Type]
	case proto.Type_TYPE_MODEL:
		sqlType = "TEXT"
	default:
		return "", "", fmt.Errorf("type not supported for computed fields: %s", field.Type.Type)
	}

	expression, err := parser.ParseExpression(field.ComputedExpression.Source)
	if err != nil {
		return "", "", err
	}

	// Generate SQL from the computed attribute expression to set this field
	stmt, err := resolve.RunCelVisitor(expression, actions.GenerateComputedFunction(schema, model, field))
	if err != nil {
		return "", "", err
	}

	fn := computedFieldFuncName(field)
	sql := fmt.Sprintf("CREATE FUNCTION \"%s\"(r \"%s\") RETURNS %s AS $$ BEGIN\n\tRETURN %s;\nEND; $$ LANGUAGE plpgsql;",
		fn,
		strcase.ToSnake(model.Name),
		sqlType,
		stmt)

	return fn, sql, nil
}

func dropComputedExecFunctionStmt(model *proto.Model) string {
	return fmt.Sprintf("DROP FUNCTION \"%s__exec_comp_fns\";", strcase.ToSnake(model.Name))
}

func dropComputedTriggerStmt(model *proto.Model) string {
	return fmt.Sprintf("DROP TRIGGER \"%s__comp\" ON \"%s\";", strcase.ToSnake(model.Name), strcase.ToSnake(model.Name))
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

	if field.Type.Repeated {
		fieldType = fmt.Sprintf("%s[]", fieldType)
	}

	output := fmt.Sprintf("%s %s", columnName, fieldType)

	// If computed, then we don't set the NOT NULL constraint yet
	// This is because we may still need to populate existing rows
	if !field.Optional && field.ComputedExpression == nil {
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
	// Handle zero values first
	if field.DefaultValue.UseZeroValue {
		return getZeroValue(field)
	}

	// Handle specific types
	switch {
	case field.Type.Type == proto.Type_TYPE_ENUM:
		return getEnumDefault(field)
	case field.Type.Repeated:
		return getRepeatedDefault(field)
	default:
		expression, err := parser.ParseExpression(field.DefaultValue.Expression.Source)
		if err != nil {
			return "", err
		}

		v, isNull, err := resolve.ToValue[any](expression)
		if err != nil {
			return "", err
		}

		if isNull {
			return "NULL", nil
		}

		return toSqlLiteral(v, field)
	}
}

// Helper functions to break down the logic
func getZeroValue(field *proto.Field) (string, error) {
	if field.Type.Repeated {
		return "{}", nil
	}

	zeroValues := map[proto.Type]string{
		proto.Type_TYPE_STRING:    db.QuoteLiteral(""),
		proto.Type_TYPE_MARKDOWN:  db.QuoteLiteral(""),
		proto.Type_TYPE_INT:       "0",
		proto.Type_TYPE_DECIMAL:   "0",
		proto.Type_TYPE_BOOL:      "false",
		proto.Type_TYPE_DATE:      "now()",
		proto.Type_TYPE_DATETIME:  "now()",
		proto.Type_TYPE_TIMESTAMP: "now()",
		proto.Type_TYPE_ID:        "ksuid()",
	}

	if value, ok := zeroValues[field.Type.Type]; ok {
		return value, nil
	}
	return "", fmt.Errorf("no zero value defined for type %v", field.Type.Type)
}

func getEnumDefault(field *proto.Field) (string, error) {
	expression, err := parser.ParseExpression(field.DefaultValue.Expression.Source)
	if err != nil {
		return "", err
	}

	if field.Type.Repeated {
		enums, err := resolve.AsIdentArray(expression)
		if err != nil {
			return "", err
		}

		if len(enums) == 0 {
			return "'{}'", nil
		}

		values := []string{}
		for _, el := range enums {
			values = append(values, db.QuoteLiteral(el.Fragments[1]))
		}

		return fmt.Sprintf("ARRAY[%s]::TEXT[]", strings.Join(values, ",")), nil
	}

	enum, err := resolve.AsIdent(expression)
	if err != nil {
		return "", err
	}

	return db.QuoteLiteral(enum.Fragments[1]), nil
}

func getRepeatedDefault(field *proto.Field) (string, error) {
	var (
		values []string
		err    error
	)

	switch field.Type.Type {
	case proto.Type_TYPE_INT:
		values, err = getArrayValues[int64](field)
	case proto.Type_TYPE_DECIMAL:
		values, err = getArrayValues[float64](field)
	case proto.Type_TYPE_BOOL:
		values, err = getArrayValues[bool](field)
	default:
		values, err = getArrayValues[string](field)
	}
	if err != nil {
		return "", err
	}

	if len(values) == 0 {
		return "'{}'", nil
	}

	typeCasts := map[proto.Type]string{
		proto.Type_TYPE_INT:     "INTEGER[]",
		proto.Type_TYPE_DECIMAL: "NUMERIC[]",
		proto.Type_TYPE_BOOL:    "BOOL[]",
	}
	cast := typeCasts[field.Type.Type]
	if cast == "" {
		cast = "TEXT[]"
	}

	return fmt.Sprintf("ARRAY[%s]::%s", strings.Join(values, ","), cast), nil
}

// Generic helper for array values
func getArrayValues[T any](field *proto.Field) ([]string, error) {
	expression, err := parser.ParseExpression(field.DefaultValue.Expression.Source)
	if err != nil {
		return nil, err
	}

	v, err := resolve.ToValueArray[T](expression)
	if err != nil {
		return nil, err
	}

	values := make([]string, len(v))
	for i, el := range v {
		val, err := toSqlLiteral(el, field)
		if err != nil {
			return nil, err
		}
		values[i] = val
	}
	return values, nil
}

func toSqlLiteral(value any, field *proto.Field) (string, error) {
	switch {
	case field.Type.Type == proto.Type_TYPE_STRING:
		return db.QuoteLiteral(fmt.Sprintf("%s", value)), nil
	case field.Type.Type == proto.Type_TYPE_DECIMAL:
		return fmt.Sprintf("%f", value), nil
	case field.Type.Type == proto.Type_TYPE_INT:
		return fmt.Sprintf("%d", value), nil
	case field.Type.Type == proto.Type_TYPE_BOOL:
		return fmt.Sprintf("%v", value), nil
	case field.Type.Type == proto.Type_TYPE_DURATION:
		return db.QuoteLiteral(fmt.Sprintf("%s", value)), nil
	default:
		return "", fmt.Errorf("field %s has unexpected default value %s", field.Name, value)
	}
}

func dropColumnStmt(modelName string, fieldName string) string {
	output := fmt.Sprintf("ALTER TABLE %s ", Identifier(modelName))
	output += fmt.Sprintf("DROP COLUMN %s CASCADE;", Identifier(fieldName))
	return output
}

// createAuditTriggerStmts generates the CREATE TRIGGER statements for auditing.
// Only creates a trigger if the trigger does not already exist in the database.
func createAuditTriggerStmts(triggers []*TriggerRow, model *proto.Model) string {
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

	return strings.Join(statements, "\n")
}

// createUpdatedAtTriggerStmts generates the CREATE TRIGGER statements for automatically updating each model's updatedAt column.
// Only creates a trigger if the trigger does not already exist in the database.
func createUpdatedAtTriggerStmts(triggers []*TriggerRow, model *proto.Model) string {
	modelLower := casing.ToSnake(model.Name)
	statements := []string{}

	updatedAt := fmt.Sprintf("%s_updated_at", modelLower)
	if _, found := lo.Find(triggers, func(t *TriggerRow) bool { return t.TriggerName == updatedAt && t.TableName == modelLower }); !found {
		statements = append(statements, fmt.Sprintf(
			`CREATE TRIGGER %s BEFORE UPDATE ON %s FOR EACH ROW EXECUTE PROCEDURE set_updated_at();`, updatedAt, Identifier(model.Name)))
	}

	return strings.Join(statements, "\n")
}

// createIndexStmts generates index changes for input fields and faceted fields
func createIndexStmts(schema *proto.Schema, existingIndexes []*IndexRow) []string {
	indexedFields := []*proto.Field{}
	for _, model := range schema.Models {
		for _, action := range model.Actions {
			if action.Type != proto.ActionType_ACTION_TYPE_LIST {
				continue
			}

			message := proto.FindWhereInputMessage(schema, action.Name)
			if message == nil {
				continue
			}

			// Find fields used as required inputs
			fieldsToIndex := findIndexableInputFields(schema, model, message)
			for _, field := range fieldsToIndex {
				// They could have been added already if used as another action's input
				if !lo.Contains(indexedFields, field) {
					indexedFields = append(indexedFields, field)
				}
			}

			// Find fields used as facets
			for _, facet := range action.Facets {
				field := model.FindField(facet)
				// They could have been added already as an input
				if !lo.Contains(indexedFields, field) {
					indexedFields = append(indexedFields, field)
				}
			}
		}
	}

	statements := []string{}

	// Add indexes which don't exist yet
	for _, field := range indexedFields {
		// Skip fields which are unique as these will already have an index
		if field.Unique {
			continue
		}

		// Skip fields which already have an index
		if lo.ContainsBy(existingIndexes, func(i *IndexRow) bool {
			return indexName(field.ModelName, field.Name) == i.IndexName
		}) {
			continue
		}

		stmt := fmt.Sprintf("CREATE INDEX \"%s\" ON %s (%s);", indexName(field.ModelName, field.Name), Identifier(field.ModelName), Identifier(field.Name))
		statements = append(statements, stmt)
	}

	// Drop existing indexes which don't exist anymore
	for _, index := range existingIndexes {
		if lo.ContainsBy(indexedFields, func(f *proto.Field) bool {
			return indexName(f.ModelName, f.Name) == index.IndexName
		}) {
			continue
		}

		// Skip dropping primary key and unique indexes as we are not concerned with these here
		if index.IsPrimary || index.IsUnique {
			continue
		}

		stmt := fmt.Sprintf("DROP INDEX IF EXISTS \"%s\";", index.IndexName)
		statements = append(statements, stmt)
	}

	return statements
}

func indexName(modelName string, fieldName string) string {
	return fmt.Sprintf("%s__%s__idx", casing.ToSnake(modelName), casing.ToSnake(fieldName))
}

func findIndexableInputFields(schema *proto.Schema, model *proto.Model, message *proto.Message) []*proto.Field {
	indexedFields := []*proto.Field{}

	for _, input := range message.Fields {
		field := proto.FindField(schema.Models, model.Name, input.Name)

		// Skip optional inputs
		if input.Optional {
			continue
		}

		if !input.IsModelField() && input.Type.Type == proto.Type_TYPE_MESSAGE {
			messageModel := schema.FindModel(field.Type.ModelName.Value)
			nestedMsg := schema.FindMessage(input.Type.MessageName.Value)
			indexedFields = append(indexedFields, findIndexableInputFields(schema, messageModel, nestedMsg)...)
		}

		// Skip indexing on optional fields
		if input.Optional {
			continue
		}

		// Skip inputs which don't correlate to a model field
		if len(input.Target) == 0 {
			continue
		}

		f := model.FindField(input.Target[len(input.Target)-1])

		indexedFields = append(indexedFields, f)
	}

	return indexedFields
}
