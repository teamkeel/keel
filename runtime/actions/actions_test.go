package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestSuite is table driven test for Create and Get etc. functions.
// The table is implied implied by a set of directories in ./testdata.
// Each directory requires two files: a keel schema, and meta.json file.
// The meta.json file contains various sections that define test set up and inputs
// or expected behaviour.
//
// The tests are dependent on a PostgreSQL service - see connectPg() below.
func TestSuite(t *testing.T) {
	db := connectPg(t)

	context := runtimectx.ContextWithDB(context.Background(), db)
	_ = context
	testCases, err := ioutil.ReadDir("./testdata")
	require.NoError(t, err)

	for _, dir := range testCases {
		if !dir.IsDir() {
			continue
		}
		dirName := dir.Name()
		dirFullPath := filepath.Join("./testdata", dirName)
		t.Logf("Processing dir: %s", dirName)

		runTestCase(t, context, dirFullPath)
	}
}

// runTestCase is a helper for TestSuite that performs the tests on
// a given directory.
func runTestCase(t *testing.T, ctx context.Context, dirFullPath string) {

	fixture := fixtureData(t, dirFullPath)

	db := runtimectx.GetDB(ctx)
	resetDB(t, db)

	emptySchema := &proto.Schema{}
	sqlDB, err := db.DB()
	require.NoError(t, err)
	schema, err := migrations.PerformMigration(emptySchema, sqlDB, dirFullPath)
	require.NoError(t, err)

	model := proto.FindModel(schema.Models, fixture.meta.ModelName)
	operation := proto.FindOperation(model, fixture.meta.OperationName)

	switch operation.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE:
		res, err := Create(ctx, model, fixture.meta.OperationInputs)
		require.NoError(t, err)

		_ = res
	default:
		t.Fatalf("operation type: %s, not yet supported", operation.Type)
	}

	// todo
	// check response
	// check apply the sql query
}

// connectPg establishes a connection to a local PostgreSQL service, which
// it expects to find running on port 8081. The /docker.compose file facilitates
// this. (docker compose up).
func connectPg(t *testing.T) *gorm.DB {
	psqlInfo := "host=localhost port=8001 user=postgres password=postgres dbname=keel sslmode=disable"
	sqlDB, err := sql.Open("postgres", psqlInfo)
	require.NoError(t, err)
	gormDB, err := gorm.Open(gormpostgres.New(gormpostgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		pingError := sqlDB.Ping()
		if pingError == nil {
			return gormDB
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("Failed to ping db")
	return nil
}

// fixtureData composes a FixtureData by reading files in the given directory.
func fixtureData(t *testing.T, dirPath string) FixtureData {
	fd := FixtureData{}

	b, err := ioutil.ReadFile(filepath.Join(dirPath, "schema.keel"))
	require.NoError(t, err)
	fd.schemaJSON = string(b)

	b, err = ioutil.ReadFile(filepath.Join(dirPath, "meta.json"))
	require.NoError(t, err)
	var meta Meta
	err = json.Unmarshal(b, &meta)
	require.NoError(t, err)

	require.True(t, len(meta.ModelName) > 2)
	require.True(t, len(meta.OperationName) > 2)
	require.True(t, len(meta.OperationInputs) > 0)
	require.True(t, len(meta.ExpectedActionResponse) > 2)

	// todo: decide what or if to check the other meta.json fields

	fd.meta = meta

	return fd
}

// FixtureData encapsulates all the test fixture definition for the
// tests on a given test directory.
type FixtureData struct {
	schemaJSON string
	meta       Meta
}

// Meta identifies various artefacts from a schema that should be used in
// a test, along with some request / expected responses for that test.
type Meta struct {
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
