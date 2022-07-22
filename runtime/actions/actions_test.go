package actions

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
	"gorm.io/gorm"
)

// TestSuite is table driven test for Create and Get etc.
//
// The tests are dependent on a PostgreSQL service - see connectPg() below.
func TestSuite(t *testing.T) {

	// todo this needs re-writing to follow the db housekeeping approach
	// used by Jon elsewhere - so skipping at the moment.

	t.Skip()

}

var _ = runTestCase

// runTestCase is a helper for TestSuite that performs the tests on
// the given test case.
func runTestCase(t *testing.T, ctx context.Context, testCase TestCase) {

	// Acquire a connect to Postgres, and clear down any existing tables.
	db := runtimectx.GetDB(ctx)
	defer resetDB(t, db)

	schema := makeProtoSchema(t, testCase.KeelSchema)

	ctx = runtimectx.WithSchema(ctx, schema)

	// Migrate the DB to this schema.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	err = migrations.PerformInitialMigration(sqlDB, schema)
	require.NoError(t, err)

	// Acquire the test parameters.
	model := proto.FindModel(schema.Models, testCase.ModelName)
	operation := proto.FindOperation(model, testCase.OperationName)

	// Call the operation function that is under test.
	switch operation.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE:
		res, err := Create(ctx, operation, testCase.OperationInputs)
		require.NoError(t, err)

		_ = res
	default:
		t.Fatalf("operation type: %s, not yet supported", operation.Type)
	}

	// todo
	// check response
	// check apply the sql query
}

// makeProtoSchema generates a proto.Schema fom the Keel schema in the given
// string.
func makeProtoSchema(t *testing.T, keelSchema string) *proto.Schema {
	builder := schema.Builder{}
	proto, err := builder.MakeFromInputs(&reader.Inputs{
		SchemaFiles: []reader.SchemaFile{
			{
				Contents: keelSchema,
			},
		},
	})
	require.NoError(t, err)
	return proto
}

// TestCase provides a Schema, and identifies various artefacts from it that should be used in
// a test, along with some request / expected responses for that test.
type TestCase struct {
	KeelSchema             string
	ModelName              string
	OperationName          string
	OperationInputs        map[string]any
	ExpectedActionResponse string

	// This is an arbitrary SQL query that the test should make to the database
	// after the Action has been run, to acquire data that can be used to verify the
	// correct operation of mutation operations.
	InterrogationSQL    string
	ExpectedSQLResponse string
}

// resetDB drops all the public tables from the database.
func resetDB(t *testing.T, db *gorm.DB) {
	var tables []string
	err := db.Table("pg_tables").
		Where("schemaname = 'public'").
		Pluck("tablename", &tables).Error
	require.NoError(t, err)
	if len(tables) == 0 {
		return
	}
	t.Logf("Resetting db by deleting all existng tables: %s", tables)

	dropSQL := `DROP TABLE `
	for i, tbl := range tables {
		dropSQL += doubleQuoteString(tbl)
		if i != len(tables)-1 {
			dropSQL += `, `
		}
	}
	dropSQL += `;`
	err = db.Exec(dropSQL).Error
	require.NoError(t, err)
}

func doubleQuoteString(s string) string {
	const dq string = `"`
	return dq + s + dq
}
