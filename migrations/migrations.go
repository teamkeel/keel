package migrations

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/auditing"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/encoding/protojson"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/db")

const (
	ChangeTypeAdded    = "ADDED"
	ChangeTypeRemoved  = "REMOVED"
	ChangeTypeModified = "MODIFIED"
)

var ErrNoStoredSchema = errors.New("no schema stored in keel_schema table")
var ErrMultipleStoredSchemas = errors.New("more than one schema found in keel_schema table")

var (
	//go:embed ksuid.sql
	ksuidFunction string

	//go:embed process_audit.sql
	processAuditFunction string

	//go:embed set_identity_id.sql
	setIdentityId string

	//go:embed set_trace_id.sql
	setTraceId string

	//go:embed set_updated_at.sql
	setUpdatedAt string
)

type DatabaseChange struct {
	// The model this change applies to
	Model string

	// The field this change applies to (might be empty)
	Field string

	// The type of change
	Type string
}

func (c DatabaseChange) String() string {
	return fmt.Sprintf("Model: %s, Field: %s, Type: %s", c.Model, c.Field, c.Type)
}

type Migrations struct {
	database db.Database

	Schema *proto.Schema

	// Describes the changes that will be applied to the database
	// if SQL is run
	Changes []*DatabaseChange

	// The SQL to run to execute the database schema changes
	SQL string
}

// HasModelFieldChanges returns true if the migrations contain model field changes to be applied
func (m *Migrations) HasModelFieldChanges() bool {
	return m.SQL != ""
}

// Apply executes the migrations against the database
// If dryRun is true, then the changes are rolled back
func (m *Migrations) Apply(ctx context.Context, dryRun bool) error {
	ctx, span := tracer.Start(ctx, "Apply Migrations")
	defer span.End()

	span.SetAttributes(attribute.Bool("dryRun", dryRun))

	sql := strings.Builder{}

	if dryRun {
		sql.WriteString("BEGIN TRANSACTION;\n")
	}

	// Enable extensions
	sql.WriteString("CREATE EXTENSION IF NOT EXISTS pg_stat_statements;\n")
	sql.WriteString("CREATE EXTENSION IF NOT EXISTS vector;\n")

	// Functions
	sql.WriteString(ksuidFunction)
	sql.WriteString("\n")
	sql.WriteString(processAuditFunction)
	sql.WriteString("\n")
	sql.WriteString(setIdentityId)
	sql.WriteString("\n")
	sql.WriteString(setTraceId)
	sql.WriteString("\n")
	sql.WriteString(setUpdatedAt)
	sql.WriteString("\n")

	sql.WriteString("CREATE TABLE IF NOT EXISTS keel_schema (schema TEXT NOT NULL);\n")
	sql.WriteString("DELETE FROM keel_schema;\n")

	b, err := protojson.Marshal(m.Schema)
	if err != nil {
		return err
	}

	escapedJSON := db.QuoteLiteral(string(b))
	sql.WriteString(fmt.Sprintf("INSERT INTO keel_schema (schema) VALUES (%s);", escapedJSON))
	sql.WriteString("\n")

	sql.WriteString("CREATE TABLE IF NOT EXISTS keel_refresh_token (token TEXT NOT NULL PRIMARY KEY, identity_id TEXT NOT NULL, created_at TIMESTAMP, expires_at TIMESTAMP);\n")
	sql.WriteString("\n")

	sql.WriteString("CREATE TABLE IF NOT EXISTS keel_auth_code (code TEXT NOT NULL PRIMARY KEY, identity_id TEXT NOT NULL, created_at TIMESTAMP, expires_at TIMESTAMP);\n")
	sql.WriteString("\n")

	sql.WriteString(fmt.Sprintf("SELECT set_trace_id('%s');\n", span.SpanContext().TraceID().String()))

	sql.WriteString(m.SQL)
	sql.WriteString("\n")

	// For now, we do this here but this could belong in our proto once we start on the database indexing work.
	sql.WriteString("CREATE INDEX IF NOT EXISTS idx_keel_audit_trace_id ON keel_audit USING HASH(trace_id);\n")
	sql.WriteString("CREATE INDEX IF NOT EXISTS idx_keel_audit_table_name_data_id_created_at ON keel_audit (table_name, (data->>'id'), created_at);\n")

	// Data migration when migrating to new authentication methods.
	sql.WriteString("UPDATE identity SET issuer = 'https://keel.so' WHERE issuer = 'keel';\n")

	if dryRun {
		sql.WriteString("ROLLBACK TRANSACTION;\n")
	}

	_, err = m.database.ExecuteStatement(ctx, sql.String())
	if err != nil {
		// Rollback the transaction if we're doing a dry run. This needs to be a separate exec
		// because when a SQL migration error happens, then the rollback command won't be executed and
		// then the transaction will be left open.
		if dryRun {
			_, _ = m.database.ExecuteStatement(ctx, "ROLLBACK TRANSACTION;")
		}
		return err
	}

	return nil
}

// New creates a new Migrations instance for the given schema and database.
// Introspection is performed on the database to work out what schema changes
// need to be applied to result in the database schema matching the Keel schema
func New(ctx context.Context, schema *proto.Schema, database db.Database) (*Migrations, error) {
	_, span := tracer.Start(ctx, "Generate Migrations")
	defer span.End()

	columns, err := getColumns(database)
	if err != nil {
		return nil, err
	}

	constraints, err := getConstraints(database)
	if err != nil {
		return nil, err
	}

	existingTriggers, err := getTriggers(database)
	if err != nil {
		return nil, err
	}

	statements := []string{}
	changes := []*DatabaseChange{}
	modelsAdded := []*proto.Model{}
	existingModels := []*proto.Model{}

	// We're going to analyse the database changes required using a temporarily mutated schema.
	// Specifically we're going to inject a fake, hard-coded KeelAudit model into it.
	//
	// This mutated copy only lives for the lifespan of this New() function, and does
	// not interfere with the Migrations.Schema field in the Migration object returned.
	//
	// The reasons for this are:
	// 1) The audit table will get created if it doesn't exist.
	// 2) If we add a column to our hard-coded definition of the audit table, that will
	//    trigger a corresponding migration also.
	//
	pushAuditModel(schema)
	defer popAuditModel(schema)

	modelNames := schema.ModelNames()

	// Add any new models
	for _, modelName := range modelNames {
		model := schema.FindModel(modelName)
		_, exists := lo.Find(columns, func(c *ColumnRow) bool {
			return c.TableName == casing.ToSnake(model.Name)
		})
		if !exists {
			stmt, err := createTableStmt(schema, model)
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
			changes = append(changes, &DatabaseChange{
				Model: model.Name,
				Type:  ChangeTypeAdded,
			})
			modelsAdded = append(modelsAdded, model)
			continue
		}

		existingModels = append(existingModels, model)
	}

	// Foreign key constraints for new models (done after all tables have been created)
	for _, model := range modelsAdded {
		statements = append(statements, fkConstraintsForModel(model)...)
	}

	// Drop tables if models removed from schema
	tablesDeleted := map[string]bool{}
	for _, column := range columns {
		if _, ok := tablesDeleted[column.TableName]; ok {
			continue
		}

		modelName := casing.ToCamel(column.TableName)

		m := schema.FindModel(modelName)
		if m == nil {
			tablesDeleted[column.TableName] = true
			statements = append(statements, dropTableStmt(modelName))
			changes = append(changes, &DatabaseChange{
				Model: modelName,
				Type:  ChangeTypeRemoved,
			})
		}
	}

	// Add audit log triggers all model tables excluding the audit table itself.
	for _, model := range schema.Models {
		if model.Name != strcase.ToCamel(auditing.TableName) {
			stmt := createAuditTriggerStmts(existingTriggers, model)
			statements = append(statements, stmt)

			stmt = createUpdatedAtTriggerStmts(existingTriggers, model)
			statements = append(statements, stmt)
		}
	}

	// Updating columns for tables that already exist
	for _, model := range existingModels {
		tableName := casing.ToSnake(model.Name)

		tableColumns := lo.Filter(columns, func(c *ColumnRow, _ int) bool {
			return c.TableName == tableName
		})

		for _, field := range model.Fields {
			if field.Type.Type == proto.Type_TYPE_MODEL {
				continue
			}

			column, _ := lo.Find(tableColumns, func(c *ColumnRow) bool {
				return c.ColumnName == casing.ToSnake(field.Name)
			})
			if column == nil {
				// Add new column
				stmt, err := addColumnStmt(schema, model.Name, field)
				if err != nil {
					return nil, err
				}
				statements = append(statements, stmt)
				changes = append(changes, &DatabaseChange{
					Model: model.Name,
					Field: field.Name,
					Type:  ChangeTypeAdded,
				})

				// When the field added is a foreign key field, we add a corresponding foreign key constraint.
				if field.ForeignKeyInfo != nil {
					statements = append(statements, fkConstraint(field, model))
				}
				continue
			}

			// Column already exists - see if any changes need to be applied
			hasChanged := false

			alterSQL, err := alterColumnStmt(model.Name, field, column)
			if err != nil {
				return nil, err
			}
			if alterSQL != "" {
				statements = append(statements, alterSQL)
				hasChanged = true
			}

			uniqueConstraint, hasUniqueConstraint := lo.Find(constraints, func(c *ConstraintRow) bool {
				return c.TableName == tableName && c.ConstraintType == "u" && len(c.ConstrainedColumns) == 1 && c.ConstrainedColumns[0] == int64(column.ColumnNum)
			})

			if field.Unique && !field.PrimaryKey && !hasUniqueConstraint {
				uniqueStmt, err := addUniqueConstraintStmt(schema, model.Name, []string{field.Name})
				if err != nil {
					return nil, err
				}

				statements = append(statements, uniqueStmt)
				hasChanged = true
			}
			if !field.Unique && hasUniqueConstraint {
				statements = append(statements, dropConstraintStmt(uniqueConstraint.TableName, uniqueConstraint.ConstraintName))
				hasChanged = true
			}

			if hasChanged {
				changes = append(changes, &DatabaseChange{
					Model: model.Name,
					Field: field.Name,
					Type:  ChangeTypeModified,
				})
			}
		}

		// Drop columns if fields removed from model
		for _, column := range tableColumns {
			field := proto.FindField(schema.Models, model.Name, casing.ToLowerCamel(column.ColumnName))
			if field == nil {
				statements = append(statements, dropColumnStmt(model.Name, column.ColumnName))
				changes = append(changes, &DatabaseChange{
					Model: model.Name,
					Field: casing.ToLowerCamel(column.ColumnName),
					Type:  ChangeTypeRemoved,
				})
			}
		}

		stmts, err := compositeUniqueConstraints(schema, model, constraints)
		if err != nil {
			return nil, err
		}

		if len(stmts) > 0 {
			statements = append(statements, stmts...)
			changes = append(changes, &DatabaseChange{
				Model: model.Name,
				Type:  ChangeTypeModified,
			})
		}
	}

	// Fetch all computed functions in the database
	existingComputedFns, err := getComputedFunctions(database)
	if err != nil {
		return nil, err
	}

	// Computed fields functions and triggers
	computedChanges, stmts, err := computedFieldsStmts(schema, existingComputedFns)
	if err != nil {
		return nil, err
	}

	for _, change := range computedChanges {
		// Dont add the db change if the field was already modified elsewhere
		if lo.ContainsBy(changes, func(c *DatabaseChange) bool {
			return c.Model == change.Model && c.Field == change.Field
		}) {
			continue
		}

		// Dont add the db change if the model is new
		if lo.ContainsBy(changes, func(c *DatabaseChange) bool {
			return c.Model == change.Model && c.Field == "" && c.Type == ChangeTypeAdded
		}) {
			continue
		}

		changes = append(changes, computedChanges...)
	}

	statements = append(statements, stmts...)

	stringChanges := lo.Map(changes, func(c *DatabaseChange, _ int) string { return c.String() })
	span.SetAttributes(attribute.StringSlice("migration", stringChanges))

	return &Migrations{
		database: database,
		Schema:   schema,
		Changes:  changes,
		SQL:      strings.TrimSpace(strings.Join(statements, "\n")),
	}, nil
}

// compositeUniqueConstraintsForModel finds all composite unique constraints in model and
// returns a map where the keys are constraint names and the keys are the field names in
// that constraint
func compositeUniqueConstraintsForModel(model *proto.Model) map[string][]string {
	uniqueConstraints := map[string][]string{}
	for _, field := range model.Fields {
		if len(field.UniqueWith) > 0 {
			fieldNames := append([]string{field.Name}, field.UniqueWith...)
			constraintName := UniqueConstraintName(model.Name, fieldNames)
			uniqueConstraints[constraintName] = fieldNames
		}
	}
	return uniqueConstraints
}

// compositeUniqueConstraints generates SQL statements for dropping or creating composite
// unique constraints for model
func compositeUniqueConstraints(schema *proto.Schema, model *proto.Model, constraints []*ConstraintRow) (statements []string, err error) {
	uniqueConstraints := compositeUniqueConstraintsForModel(model)

	for _, c := range constraints {
		if c.TableName != casing.ToSnake(model.Name) || c.ConstraintType != "u" || len(c.ConstrainedColumns) == 1 {
			continue
		}

		if _, ok := uniqueConstraints[c.ConstraintName]; ok {
			delete(uniqueConstraints, c.ConstraintName)
			continue
		}

		stmt := dropConstraintStmt(c.TableName, c.ConstraintName)
		statements = append(statements, stmt)
	}

	for _, fieldNames := range uniqueConstraints {
		stmt, err := addUniqueConstraintStmt(schema, model.Name, fieldNames)
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}

	return statements, nil
}

// computedFieldDependencies returns a map of computed fields and every field it depends on
func computedFieldDependencies(schema *proto.Schema) (map[*proto.Field][]*proto.Field, error) {
	dependencies := map[*proto.Field][]*proto.Field{}

	for _, model := range schema.Models {
		for _, field := range model.Fields {
			if field.ComputedExpression == nil {
				continue
			}

			expr, err := parser.ParseExpression(field.ComputedExpression.Source)
			if err != nil {
				return nil, err
			}

			idents, err := resolve.IdentOperands(expr)
			if err != nil {
				return nil, err
			}

			for _, ident := range idents {
				for _, f := range schema.FindModel(strcase.ToCamel(ident.Fragments[0])).Fields {
					if f.Name == ident.Fragments[1] {
						dependencies[field] = append(dependencies[field], f)
						break
					}
				}
			}
		}
	}

	return dependencies, nil
}

// computedFieldsStmts generates SQL statements for dropping or creating functions and triggers for computed fields
func computedFieldsStmts(schema *proto.Schema, existingComputedFns []*FunctionRow) (changes []*DatabaseChange, statements []string, err error) {
	existingComputedFnNames := lo.Map(existingComputedFns, func(f *FunctionRow, _ int) string {
		return f.RoutineName
	})

	fns := map[string]string{}
	fieldsFns := map[*proto.Field]string{}
	changedFields := map[*proto.Field]bool{}

	// Adding computed field triggers and functions
	for _, model := range schema.Models {
		modelFns := map[string]string{}

		for _, field := range model.GetComputedFields() {
			changedFields[field] = false
			fnName, computedFuncStmt, err := addComputedFieldFuncStmt(schema, model, field)
			if err != nil {
				return nil, nil, err
			}

			fieldsFns[field] = fnName
			modelFns[fnName] = computedFuncStmt
		}

		// Get all the preexisting computed functions for computed fields on this model
		existingComputedFnNamesForModel := lo.Filter(existingComputedFnNames, func(f string, _ int) bool {
			return strings.HasPrefix(f, fmt.Sprintf("%s__", strcase.ToSnake(model.Name))) &&
				strings.HasSuffix(f, "_computed")
		})

		newFns, retiredFns := lo.Difference(lo.Keys(modelFns), existingComputedFnNamesForModel)
		slices.Sort(newFns)
		slices.Sort(retiredFns)

		// Functions to be created
		for _, fn := range newFns {
			statements = append(statements, modelFns[fn])

			f := fieldFromComputedFnName(schema, fn)
			changes = append(changes, &DatabaseChange{
				Model: f.ModelName,
				Field: f.Name,
				Type:  ChangeTypeModified,
			})
			changedFields[f] = true
		}

		// Functions to be dropped
		for _, fn := range retiredFns {
			statements = append(statements, fmt.Sprintf("DROP FUNCTION %s;", fn))

			f := fieldFromComputedFnName(schema, fn)
			if f != nil {
				change := &DatabaseChange{
					Model: f.ModelName,
					Field: f.Name,
					Type:  ChangeTypeModified,
				}
				if !lo.ContainsBy(changes, func(c *DatabaseChange) bool {
					return c.Model == change.Model && c.Field == change.Field
				}) {
					changes = append(changes, change)
				}
				changedFields[f] = true
			}
		}

		// When there all computed fields have been removed
		if len(modelFns) == 0 && len(retiredFns) > 0 {
			dropExecFn := dropComputedExecFunctionStmt(model)
			dropTrigger := dropComputedTriggerStmt(model)
			statements = append(statements, dropTrigger, dropExecFn)
		}

		for k, v := range modelFns {
			fns[k] = v
		}
	}

	dependencies, err := computedFieldDependencies(schema)
	if err != nil {
		return nil, nil, err
	}

	for _, model := range schema.Models {
		modelhasChanged := false
		for k, v := range changedFields {
			if k.ModelName == model.Name && v {
				modelhasChanged = true
			}
		}
		if !modelhasChanged {
			continue
		}

		computedFields := model.GetComputedFields()
		if len(computedFields) == 0 {
			continue
		}

		// Sort fields based on dependencies
		sorted := []*proto.Field{}
		visited := map[*proto.Field]bool{}
		var visit func(*proto.Field)
		visit = func(field *proto.Field) {
			if visited[field] || field.ComputedExpression == nil {
				return
			}
			visited[field] = true

			// Process dependencies first
			for _, dep := range dependencies[field] {
				if dep.ModelName == field.ModelName {
					visit(dep)
				}
			}
			sorted = append(sorted, field)
		}

		// Visit all fields to build sorted order
		for _, field := range computedFields {
			visit(field)
		}

		// Generate SQL statements in dependency order
		stmts := []string{}
		for _, field := range sorted {
			s := fmt.Sprintf("\tNEW.%s := %s(%s);\n", strcase.ToSnake(field.Name), fieldsFns[field], "NEW")
			stmts = append(stmts, s)
		}

		execFnName := computedExecFuncName(model)
		triggerName := computedTriggerName(model)

		// Generate the trigger function which executes all the computed field functions for the model.
		sql := fmt.Sprintf("CREATE OR REPLACE FUNCTION %s() RETURNS TRIGGER AS $$ BEGIN\n%s\tRETURN NEW;\nEND; $$ LANGUAGE plpgsql;", execFnName, strings.Join(stmts, ""))

		// Genrate the table trigger which executed the trigger function.
		trigger := fmt.Sprintf("CREATE OR REPLACE TRIGGER %s BEFORE INSERT OR UPDATE ON \"%s\" FOR EACH ROW EXECUTE PROCEDURE %s();", triggerName, strcase.ToSnake(model.Name), execFnName)

		statements = append(statements, sql, trigger)
	}

	return
}

func keelSchemaTableExists(ctx context.Context, database db.Database) (bool, error) {
	// to_regclass docs - https://www.postgresql.org/docs/current/functions-info.html#FUNCTIONS-INFO-CATALOG-TABLE
	// translates a textual relation name to its OID ... this function will
	// return NULL rather than throwing an error if the name is not found.
	result, err := database.ExecuteQuery(ctx, "SELECT to_regclass('keel_schema') AS name")
	if err != nil {
		return false, err
	}

	return result.Rows[0]["name"] != nil, nil
}

func GetCurrentSchema(ctx context.Context, database db.Database) (*proto.Schema, error) {
	exists, err := keelSchemaTableExists(ctx, database)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	result, err := database.ExecuteQuery(ctx, "SELECT schema FROM keel_schema")
	if err != nil {
		return nil, err
	}

	if len(result.Rows) == 0 {
		return nil, ErrNoStoredSchema
	} else if len(result.Rows) > 1 {
		return nil, ErrMultipleStoredSchemas
	}

	schema, ok := result.Rows[0]["schema"].(string)
	if !ok {
		return nil, errors.New("schema could not be converted to string")
	}

	var protoSchema proto.Schema
	err = protojson.Unmarshal([]byte(schema), &protoSchema)
	if err != nil {
		return nil, err
	}

	return &protoSchema, nil
}

// fkConstraintsForModel generates foreign key constraint statements for each of fields marked as
// being foreign keys in the given model.
func fkConstraintsForModel(model *proto.Model) (fkStatements []string) {
	fkFields := model.ForeignKeyFields()
	for _, field := range fkFields {
		stmt := fkConstraint(field, model)
		fkStatements = append(fkStatements, stmt)
	}
	return fkStatements
}

// fkConstraint generates a foreign key constraint statement for the given foreign key field.
func fkConstraint(field *proto.Field, thisModel *proto.Model) (fkStatement string) {
	fki := field.ForeignKeyInfo
	onDelete := lo.Ternary(field.Optional, "SET NULL", "CASCADE")
	stmt := addForeignKeyConstraintStmt(
		Identifier(thisModel.Name),
		Identifier(field.Name),
		Identifier(fki.RelatedModelName),
		Identifier(fki.RelatedModelField),
		onDelete,
	)
	return stmt
}
