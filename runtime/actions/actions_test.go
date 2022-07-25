package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const dbConnString = "host=localhost port=8001 user=postgres password=postgres dbname=%s sslmode=disable"

func TestCreate(t *testing.T) {
	testCases, err := ioutil.ReadDir(filepath.Join(".", "testdata", "create"))
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
			require.NoError(t, mainDB.Exec("DROP DATABASE if exists "+dbName).Error)

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

			// Access the test instructions.
			contents, err := ioutil.ReadFile(filepath.Join("testdata", "create", testCase.Name()))
			require.NoError(t, err)
			parts := strings.Split(string(contents), "===")
			parts = lo.Map(parts, func(s string, _ int) string {
				return strings.TrimSpace(s)
			})
			require.Len(t, parts, 3, "create test file should contain three sections separated by '==='")
			keelSchema, operationName, inputArgsJSON := parts[0], parts[1], parts[2]

			// Compose the things we need to call the Create function.
			schema := protoSchema(t, keelSchema)
			createOp := findOp(t, schema, operationName)
			args := inputArgs(t, inputArgsJSON)

			// Migrate the database to this schema, in readiness for the Create Action.
			m := migrations.New(schema, nil)
			require.NoError(t, m.Apply(testDB))

			// Call the Create Operation.
			response, err := Create(testDB, createOp, schema, args)
			require.NoError(t, err)

			// Check we got the correct return value.
			// Todo hard-coded placeholder for just the one test case.
			require.Equal(t, "foo@bar.com", response["email"])
			require.IsType(t, time.Time{}, response["created_at"])

			// Check the correct row got added to the database.
			// Todo hard-coded placeholder for just the one test case.
			row := map[string]any{}
			require.NoError(t, testDB.Table("person").Where("email = ?", "foo@bar.com").Find(row).Error)
			require.Equal(t, "foo@bar.com", row["email"])
			require.IsType(t, time.Time{}, row["created_at"])
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

func findOp(t *testing.T, schema *proto.Schema, operationName string) *proto.Operation {
	// Only reliable if there are no duplicate operation names in the schema!
	for _, model := range schema.Models {
		for _, op := range model.Operations {
			if op.Name == operationName {
				return op
			}
		}
	}
	t.Fatalf("cannot find operation: %s", operationName)
	return nil
}

func inputArgs(t *testing.T, inputArgs string) map[string]any {
	args := map[string]any{}
	err := json.Unmarshal([]byte(inputArgs), &args)
	require.NoError(t, err)
	return args
}
