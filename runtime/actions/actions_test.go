package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

func TestSuite(t *testing.T) {
	db := connectPg(t)
	_ = db

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
		fixtureData := testData(t, dirFullPath)

		testCase(t, context, dirFullPath, fixtureData)
	}
}

func testCase(t *testing.T, context context.Context, dirFullPath string, fd fixtureData) {

	db := runtimectx.GetDB(context)
	resetDB(t, db)

	emptySchema := &proto.Schema{}
	sqlDB, err := db.DB()
	require.NoError(t, err)
	schema, err := migrations.PerformMigration(emptySchema, sqlDB, dirFullPath)
	require.NoError(t, err)

	model := proto.FindModel(schema.Models, fd.meta.ModelName)
	operation := proto.FindOperation(model, fd.meta.OperationName)

	// locate the model and operation using their names
	// call either Make or Get actions
	// check response
	// check apply the sql query
}

func connectPg(t *testing.T) *gorm.DB {
	// Note expecting the database for this test to be serving on port 8001
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

func testData(t *testing.T, dirPath string) fixtureData {
	fd := fixtureData{}

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
	require.True(t, len(meta.ExpectedActionResponse) > 2)

	fd.meta = meta

	return fd
}

type fixtureData struct {
	schemaJSON string
	meta       Meta
}

type Meta struct {
	ModelName              string
	OperationName          string
	ExpectedActionResponse string
	InterrogationSQL       string
	ExpectedSQLResponse    string
}

func resetDB(t *testing.T, db *gorm.DB) {
	var tables []string
	err := db.Table("pg_tables").
		Where("schemaname = 'public'").
		Pluck("tablename", &tables).Error
	require.NoError(t, err)
	if len(tables) == 0 {
		return
	}
	t.Logf("Found tables: %s", tables)

	dropSQL := "DROP TABLE "
	for _, tbl := range tables {
		dropSQL += fmt.Sprintf(", %s", tbl)
	}
	err = db.Exec(dropSQL).Error
	require.NoError(t, err)
}
