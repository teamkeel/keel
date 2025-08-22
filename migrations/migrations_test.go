package migrations_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestMigrations(t *testing.T) {
	dbConnInfo := &db.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Password: "postgres",
		Database: "keel",
	}

	testCases, err := os.ReadDir("testdata")
	require.NoError(t, err)

	// We connect to the "main" database here only so we can create a new database
	// for each sub-test
	mainDB, err := sql.Open("pgx/v5", dbConnInfo.String())
	require.NoError(t, err)
	defer func() {
		mainDB.Close()
	}()

	for _, testCase := range testCases {
		// if testCase.Name() != "computed_field_many_to_one.txt" {
		// 	continue
		// }
		t.Run(strings.TrimSuffix(testCase.Name(), ".txt"), func(t *testing.T) {
			// Make a database name for this test
			re := regexp.MustCompile(`[^\w]`)
			dbName := strings.ToLower(re.ReplaceAllString(t.Name(), ""))

			// Drop the database if it already exists. The normal dropping of it at the end of the
			// test case is bypassed if you quit a debug run of the test in VS Code.
			_, err = mainDB.Exec("DROP DATABASE if exists " + dbName)
			require.NoError(t, err)

			// Create the database and drop at the end of the test
			_, err = mainDB.Exec("CREATE DATABASE " + dbName)
			require.NoError(t, err)

			ctx := context.Background()

			database, err := db.New(ctx, dbConnInfo.WithDatabase(dbName).String())
			require.NoError(t, err)

			// Read the fixture file
			contents, err := os.ReadFile(filepath.Join("testdata", testCase.Name()))
			require.NoError(t, err)

			parts := strings.Split(string(contents), "===")
			parts = lo.Map(parts, func(s string, _ int) string {
				return strings.TrimSpace(s)
			})

			require.Len(t, parts, 4, "migrations test file should contain four sections separated by '==='")

			currSchema, newSchema, expectedSQL, expectedChanges := parts[0], parts[1], parts[2], parts[3]

			ctx, err = testhelpers.WithTracing(ctx)
			require.NoError(t, err)

			// If this test defines a "current schema" then migrate the database to that
			// state first
			var currProto *proto.Schema
			if currSchema != "" {
				builder := &schema.Builder{}
				currProto, err = builder.MakeFromString(currSchema, config.Empty)
				require.NoError(t, err)

				m, err := migrations.New(ctx, currProto, database)
				require.NoError(t, err)
				err = m.Apply(ctx, false)
				require.NoError(t, err)
			}

			// Create the new proto
			builder := &schema.Builder{}
			schema, err := builder.MakeFromString(newSchema, config.Empty)
			require.NoError(t, err)

			// Create migrations from old (may be nil) to new
			m, err := migrations.New(
				ctx,
				schema,
				database,
			)
			require.NoError(t, err)

			// Assert correct SQL generated
			equal := assert.Equal(t, expectedSQL, m.SQL)

			if !equal {
				fmt.Println(m.SQL)
			}

			actualChanges, err := json.Marshal(m.Changes)
			require.NoError(t, err)

			// Assert changes summary
			assertJSON(t, []byte(expectedChanges), actualChanges)

			// Check the new migrations can be applied without error
			require.NoError(t, m.Apply(ctx, false))

			// Now fetch the "current" schema from the database, which
			// should be the new one we just applied
			dbSchema, err := migrations.GetCurrentSchema(ctx, database)
			require.NoError(t, err)

			// Assert it is the new schema
			dbSchemaBytes, err := protojson.Marshal(dbSchema)
			require.NoError(t, err)
			schemaBytes, err := protojson.Marshal(schema)
			require.NoError(t, err)

			assertJSON(t, schemaBytes, dbSchemaBytes)

			// Test ksuid function
			r, err := database.ExecuteQuery(ctx, `select ksuid()`)
			require.NoError(t, err)
			require.Equal(t, 1, len(r.Rows))
			k, err := ksuid.Parse(r.Rows[0]["ksuid"].(string))
			require.NoError(t, err)
			assert.False(t, k.IsNil())
		})
	}
}

func assertJSON(t *testing.T, expected []byte, actual []byte) {
	// assert changes JSON
	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expected, actual, &opts)

	switch diff {
	case jsondiff.FullMatch:
		// success
		return
	case jsondiff.SupersetMatch, jsondiff.NoMatch:
		assert.Fail(t, "changes do not match expected", explanation)
	case jsondiff.FirstArgIsInvalidJson:
		assert.Fail(t, "expected changes JSON is invalid")
	}
}
