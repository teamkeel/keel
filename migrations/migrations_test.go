package migrations_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const dbConnString = "host=localhost port=8001 user=postgres password=postgres dbname=%s sslmode=disable"

func TestMigrations(t *testing.T) {
	testCases, err := os.ReadDir("testdata")
	require.NoError(t, err)

	// We connect to the "main" database here only so we can create a new database
	// for each sub-test
	mainDB, err := gorm.Open(
		postgres.Open(fmt.Sprintf(dbConnString, "keel")),
		&gorm.Config{})
	require.NoError(t, err)

	for _, testCase := range testCases {
		t.Run(strings.TrimSuffix(testCase.Name(), ".txt"), func(t *testing.T) {
			// Make a database name for this test
			re := regexp.MustCompile(`[^\w]`)
			dbName := strings.ToLower(re.ReplaceAllString(t.Name(), ""))

			// Drop the database if it already exists. The normal dropping of it at the end of the
			// test case is bypassed if you quit a debug run of the test in VS Code.
			err = mainDB.Exec("DROP DATABASE if exists " + dbName).Error
			require.NoError(t, err)

			// Create the database and drop at the end of the test
			err = mainDB.Exec("CREATE DATABASE " + dbName).Error
			require.NoError(t, err)
			defer func() {
				require.NoError(t, mainDB.Exec("DROP DATABASE "+dbName).Error)
			}()

			// Connect to the newly created test database and close connection
			// at the end of the test. We need to explicitly close the connection
			// so the mainDB connection can drop the database.
			testDB, err := gorm.Open(
				postgres.Open(fmt.Sprintf(dbConnString, dbName)),
				&gorm.Config{})
			require.NoError(t, err)
			defer func() {
				conn, err := testDB.DB()
				require.NoError(t, err)
				conn.Close()
			}()

			// Read the fixture file
			contents, err := os.ReadFile(filepath.Join("testdata", testCase.Name()))
			require.NoError(t, err)

			parts := strings.Split(string(contents), "===")
			parts = lo.Map(parts, func(s string, _ int) string {
				return strings.TrimSpace(s)
			})

			require.Len(t, parts, 4, "migrations test file should contain four sections separated by '==='")

			currSchema, newSchema, expectedSQL, expectedChanges := parts[0], parts[1], parts[2], parts[3]

			// If this test defines a "current schema" then migrate the database to that
			// state first
			var currProto *proto.Schema
			if currSchema != "" {
				currProto = protoSchema(t, currSchema)
				m := migrations.New(currProto, nil)
				require.NoError(t, m.Apply(testDB))
			}

			// Create the new proto
			schema := protoSchema(t, newSchema)

			// Create migrations from old (may be nil) to new
			m := migrations.New(
				schema,
				currProto,
			)

			// Assert correct SQL generated
			assert.Equal(t, expectedSQL, m.SQL)

			actualChanges, err := json.Marshal(m.Changes)
			require.NoError(t, err)

			// Assert changes summary
			assertJSON(t, []byte(expectedChanges), actualChanges)

			// Check the new migrations can be applied without error
			require.NoError(t, m.Apply(testDB))

			// Now fetch the "current" schema from the database, which
			// should be the new one we just applied
			dbSchema, err := migrations.GetCurrentSchema(context.Background(), testDB)
			require.NoError(t, err)

			// Assert it is the new schema
			dbSchemaBytes, err := protojson.Marshal(dbSchema)
			require.NoError(t, err)
			schemaBytes, err := protojson.Marshal(schema)
			require.NoError(t, err)

			assertJSON(t, schemaBytes, dbSchemaBytes)
		})
	}
}

func protoSchema(t *testing.T, s string) *proto.Schema {
	builder := &schema.Builder{}
	schema, err := builder.MakeFromInputs(&reader.Inputs{
		SchemaFiles: []reader.SchemaFile{
			{
				Contents: s,
			},
		},
	})
	require.NoError(t, err)
	return schema
}

func assertJSON(t *testing.T, expected []byte, actual []byte) {
	// assert changes JSON
	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expected, actual, &opts)

	switch diff {
	case jsondiff.FullMatch:
		// success
	case jsondiff.SupersetMatch, jsondiff.NoMatch:
		assert.Fail(t, "changes do not match expected", explanation)
	case jsondiff.FirstArgIsInvalidJson:
		assert.Fail(t, "expected changes JSON is invalid")
	}
}
