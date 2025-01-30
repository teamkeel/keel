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
	"github.com/teamkeel/keel/runtime/actions"
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

	// Create the schema table
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

type depPair struct {
	field *proto.Field
	ident *parser.ExpressionIdent
}

// computedFieldDependencies returns a map of computed fields and every field it depends on
func computedFieldDependencies(schema *proto.Schema) (map[*proto.Field][]*depPair, error) {
	dependencies := map[*proto.Field][]*depPair{}

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
				currModel := schema.FindModel(strcase.ToCamel(ident.Fragments[0]))

				for i, f := range ident.Fragments[1:] {
					currField := currModel.FindField(f)

					if i < len(ident.Fragments)-2 {
						currModel = schema.FindModel(currField.Type.ModelName.Value)
						continue
					}

					dep := depPair{
						ident: ident,
						field: currField,
					}

					hasDep := lo.ContainsBy(dependencies[field], func(d *depPair) bool {
						return d.field.Name == currField.Name && d.ident.String() == ident.String()
					})

					if !hasDep {
						dependencies[field] = append(dependencies[field], &dep)
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
	recompute := []*proto.Model{}

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
				strings.HasSuffix(f, "__comp")
		})

		newFns, retiredFns := lo.Difference(lo.Keys(modelFns), existingComputedFnNamesForModel)
		slices.Sort(newFns)
		slices.Sort(retiredFns)

		// Computed functions to be created for each computed field
		for _, fn := range newFns {
			statements = append(statements, modelFns[fn])

			f := fieldFromComputedFnName(schema, fn)
			changes = append(changes, &DatabaseChange{
				Model: f.ModelName,
				Field: f.Name,
				Type:  ChangeTypeModified,
			})
			changedFields[f] = true

			if !lo.Contains(recompute, model) {
				recompute = append(recompute, model)
			}
		}

		// Functions to be dropped
		for _, fn := range retiredFns {
			statements = append(statements, fmt.Sprintf("DROP FUNCTION \"%s\";", fn))

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

		// When all computed fields have been removed
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

	// For each model, we need to create a function which calls all the computed functions for fields on this model
	// Order is important because computed fields can depend on each other - this is catered for
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
				if dep.field.ModelName == field.ModelName {
					visit(dep.field)
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
			s := fmt.Sprintf("NEW.%s := %s(NEW);\n", strcase.ToSnake(field.Name), fieldsFns[field])
			stmts = append(stmts, s)
		}

		// Generate the trigger function which executes all the computed field functions for the model.
		execFnName := computedExecFuncName(model)
		sql := fmt.Sprintf("CREATE OR REPLACE FUNCTION \"%s\"() RETURNS TRIGGER AS $$ BEGIN\n\t%s\tRETURN NEW;\nEND; $$ LANGUAGE plpgsql;", execFnName, strings.Join(stmts, ""))

		// Generate the table trigger which executed the trigger function.
		// This must be a BEFORE trigger because we want to return the row with its computed fields being computed.
		triggerName := computedTriggerName(model)
		trigger := fmt.Sprintf("CREATE OR REPLACE TRIGGER \"%s\" BEFORE INSERT OR UPDATE ON \"%s\" FOR EACH ROW EXECUTE PROCEDURE \"%s\"();", triggerName, strcase.ToSnake(model.Name), execFnName)

		statements = append(statements, sql)
		statements = append(statements, trigger)
	}

	// For computed fields which depend on fields in other models, we need to create triggers which start from the source model and cascades
	// down through the relationship until it reaches the target model (where the computed field is defined). We then perform a fake update on the
	// specific rows in the target model which will then trigger the computed fns.
	depFns := map[string]string{}
	for field, deps := range dependencies {
		for _, dep := range deps {
			// Skip this because the triggers call on the exec functions themselves
			if field.ModelName == dep.field.ModelName {
				continue
			}

			fragments, err := actions.NormalisedFragments(schema, dep.ident.Fragments)
			if err != nil {
				return nil, nil, err
			}

			currentModel := casing.ToCamel(fragments[0])
			for i := 1; i < len(fragments)-1; i++ {
				baseQuery := actions.NewQuery(schema.FindModel(currentModel))
				baseQuery.Select(actions.IdField())

				expr := strings.Join(fragments, ".")
				// Get the fragment pair from the previous model to the current model
				// We need to reset the first fragment to the model name and not the previous model's field name
				subFragments := slices.Clone(fragments[i-1 : i+1])
				subFragments[0] = strcase.ToLowerCamel(currentModel)

				if !proto.ModelHasField(schema, currentModel, fragments[i]) {
					return nil, nil, fmt.Errorf("this model: %s, does not have a field of name: %s", currentModel, subFragments[0])
				}

				// We know that the current fragment is a related model because it's not the last fragment
				relatedModelField := proto.FindField(schema.Models, currentModel, fragments[i])
				foreignKeyField := proto.GetForeignKeyFieldName(schema.Models, relatedModelField)

				previousModel := currentModel
				currentModel = relatedModelField.Type.ModelName.Value
				stmt := ""

				// If the relationship is a belongs to or has many, we need to update the id field on the previous model
				switch {
				case relatedModelField.IsBelongsTo():
					stmt += "UPDATE \"" + strcase.ToSnake(previousModel) + "\" SET id = id WHERE " + strcase.ToSnake(foreignKeyField) + " IN (NEW.id, OLD.id);"
				default:
					stmt += "UPDATE \"" + strcase.ToSnake(previousModel) + "\" SET id = id WHERE id IN (NEW." + strcase.ToSnake(foreignKeyField) + ", OLD." + strcase.ToSnake(foreignKeyField) + ");"
				}

				// Trigger function which will perform a fake update on the earlier model in the expression chain
				fnName := computedDependencyFuncName(schema.FindModel(strcase.ToCamel(previousModel)), schema.FindModel(currentModel), strings.Split(expr, "."))
				sql := fmt.Sprintf("-- %s\nCREATE OR REPLACE FUNCTION \"%s\"() RETURNS TRIGGER AS $$\nBEGIN\n\t%s\nEND; $$ LANGUAGE plpgsql;\n", strings.Join(subFragments, ".")+" in "+strings.Join(fragments, "."), fnName, stmt)

				// For the comp_dep function on the target field's model, we include a filter on the UPDATE trigger to only trigger if the target field has changed
				whenCondition := "TRUE"
				if i == len(fragments)-2 {
					f := strcase.ToSnake(fragments[len(fragments)-1])
					whenCondition = fmt.Sprintf("NEW.%s <> OLD.%s", f, f)

					if !relatedModelField.IsBelongsTo() {
						updatingField := strcase.ToSnake(foreignKeyField)
						whenCondition += fmt.Sprintf(" OR NEW.%s <> OLD.%s", updatingField, updatingField)
					}
				}

				// Must be an AFTER trigger as we need the data to be written in order to perform the joins and for the computation to take into account the updated data
				triggerName := fnName
				sql += fmt.Sprintf("\nCREATE OR REPLACE TRIGGER \"%s\" AFTER INSERT OR DELETE ON \"%s\" FOR EACH ROW EXECUTE PROCEDURE \"%s\"();", triggerName, strcase.ToSnake(currentModel), fnName)
				sql += fmt.Sprintf("\nCREATE OR REPLACE TRIGGER \"%s_update\" AFTER UPDATE ON \"%s\" FOR EACH ROW WHEN(%s) EXECUTE PROCEDURE \"%s\"();", triggerName, strcase.ToSnake(currentModel), whenCondition, fnName)

				depFns[fnName] = sql
			}
		}
	}

	existingDependencyFnNames := lo.FilterMap(existingComputedFns, func(f *FunctionRow, _ int) (string, bool) {
		return f.RoutineName, strings.HasSuffix(f.RoutineName, "__comp_dep")
	})

	newFns, retiredFns := lo.Difference(lo.Keys(depFns), existingDependencyFnNames)
	slices.Sort(newFns)
	slices.Sort(retiredFns)

	// Dependency functions and triggers to be created
	for _, fn := range newFns {
		statements = append(statements, depFns[fn])
	}

	// Dependency functions and triggers to be dropped
	for _, fn := range retiredFns {
		statements = append(statements, fmt.Sprintf("DROP TRIGGER IF EXISTS \"%s\" ON \"%s\";", fn, strings.Split(fn, "__")[0]))
		statements = append(statements, fmt.Sprintf("DROP TRIGGER IF EXISTS \"%s_update\" ON \"%s\";", fn, strings.Split(fn, "__")[0]))
		statements = append(statements, fmt.Sprintf("DROP FUNCTION IF EXISTS \"%s\";", fn))
	}

	// If a computed field has been added or changed, we need to recompute all existing data.
	// This is done by fake updating each row on the table which will cause the triggers to run.
	for _, model := range recompute {
		sql := fmt.Sprintf("UPDATE \"%s\" SET id = id;", strcase.ToSnake(model.Name))
		statements = append(statements, sql)
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
