package migrations

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"
)

const (
	ChangeTypeAdded    = "ADDED"
	ChangeTypeRemoved  = "REMOVED"
	ChangeTypeModified = "MODIFIED"
)

var ErrNoStoredSchema = errors.New("no schema stored in keel_schema table")
var ErrMultipleStoredSchemas = errors.New("more than one schema found in keel_schema table")

type DatabaseChange struct {
	// The model this change applies to
	Model string

	// The field this change applies to (might be empty)
	Field string

	// The type of change
	Type string
}

type Migrations struct {
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
func (m *Migrations) Apply(db *gorm.DB) error {

	sql := strings.Builder{}

	sql.WriteString(m.SQL)
	sql.WriteString("\n")

	sql.WriteString("CREATE TABLE IF NOT EXISTS keel_schema ( schema TEXT NOT NULL );\n")
	sql.WriteString("DELETE FROM keel_schema;\n")

	b, err := protojson.Marshal(m.Schema)
	if err != nil {
		return err
	}

	// Cannot use parameters as then you get an error:
	//   ERROR: cannot insert multiple commands into a prepared statement (SQLSTATE 42601)
	escapedJSON := pq.QuoteLiteral(string(b))
	insertStmt := fmt.Sprintf("INSERT INTO keel_schema (schema) VALUES (%s);", escapedJSON)
	sql.WriteString(insertStmt)

	tx := db.Session(&gorm.Session{
		PrepareStmt: false,
	})

	return tx.Exec(sql.String()).Error
}

// Create inspects the database using gorm.DB connection
// and creates the required schema migrations that will result in
// the database reflecting the provided proto.Schema
func New(newSchema *proto.Schema, currSchema *proto.Schema) *Migrations {

	if currSchema == nil {
		currSchema = &proto.Schema{}
	}

	statements := []string{}

	changes := []*DatabaseChange{}

	currModels := proto.ModelNames(currSchema)
	newModels := proto.ModelNames(newSchema)
	modelsInCommon := lo.Intersect(newModels, currModels)

	// Models added or removed.
	modelsAdded, modelsRemoved := lo.Difference(newModels, currModels)

	for _, modelName := range modelsAdded {
		model := proto.FindModel(newSchema.Models, modelName)
		statements = append(statements, createTableStmt(model))
		changes = append(changes, &DatabaseChange{
			Model: modelName,
			Type:  ChangeTypeAdded,
		})
	}

	for _, modelName := range modelsRemoved {
		statements = append(statements, dropTableStmt(modelName))
		changes = append(changes, &DatabaseChange{
			Model: modelName,
			Type:  ChangeTypeRemoved,
		})
	}

	// Fields added or removed
	for _, modelName := range modelsInCommon {
		currFieldNames := proto.FieldNames(proto.FindModel(currSchema.Models, modelName))
		newFieldNames := proto.FieldNames(proto.FindModel(newSchema.Models, modelName))
		fieldsAdded, fieldsRemoved := lo.Difference(newFieldNames, currFieldNames)

		for _, fieldName := range fieldsAdded {
			field := proto.FindField(newSchema.Models, modelName, fieldName)
			statements = append(statements, addColumnStmt(modelName, field))
			changes = append(changes, &DatabaseChange{
				Model: modelName,
				Field: fieldName,
				Type:  ChangeTypeAdded,
			})
		}

		for _, fieldName := range fieldsRemoved {
			statements = append(statements, dropColumnStmt(modelName, fieldName))
			changes = append(changes, &DatabaseChange{
				Model: modelName,
				Field: fieldName,
				Type:  ChangeTypeRemoved,
			})
		}

		fieldsInCommon := lo.Intersect(newFieldNames, currFieldNames)
		for _, fieldName := range fieldsInCommon {
			newField := proto.FindField(newSchema.Models, modelName, fieldName)
			currField := proto.FindField(currSchema.Models, modelName, fieldName)

			alterSQL := alterColumnStmt(modelName, newField, currField)
			if alterSQL == "" {
				continue
			}

			statements = append(statements, alterSQL)
			changes = append(changes, &DatabaseChange{
				Model: modelName,
				Field: fieldName,
				Type:  ChangeTypeModified,
			})
		}
	}

	return &Migrations{
		Schema:  newSchema,
		Changes: changes,
		SQL:     strings.TrimSpace(strings.Join(statements, "\n")),
	}
}

func keelSchemaTableExists(db *gorm.DB) (bool, error) {
	var rows []struct {
		Name *string
	}

	// to_regclass docs - https://www.postgresql.org/docs/current/functions-info.html#FUNCTIONS-INFO-CATALOG-TABLE
	// translates a textual relation name to its OID ... this function will
	// return NULL rather than throwing an error if the name is not found.
	err := db.Raw("SELECT to_regclass('keel_schema') AS name").Scan(&rows).Error
	if err != nil {
		return false, err
	}

	return rows[0].Name != nil, nil
}

func GetCurrentSchema(db *gorm.DB) (*proto.Schema, error) {
	exists, err := keelSchemaTableExists(db)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	var rows [][]byte
	err = db.Raw("SELECT schema FROM keel_schema").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, ErrNoStoredSchema
	}

	if len(rows) > 1 {
		return nil, ErrMultipleStoredSchemas
	}

	var protoSchema proto.Schema
	err = protojson.Unmarshal(rows[0], &protoSchema)
	if err != nil {
		return nil, err
	}

	return &protoSchema, nil
}
