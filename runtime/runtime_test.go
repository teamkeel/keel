package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/iancoleman/strcase"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	rtt "github.com/teamkeel/keel/runtime/runtimetest"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRuntime(t *testing.T) {
	// We connect to the "main" database here only so we can create a new database
	// for each sub-test
	mainDB, err := gorm.Open(
		postgres.Open(fmt.Sprintf(dbConnString, "keel")),
		&gorm.Config{})
	require.NoError(t, err)

	var skip string = "" // Name of test case you want to skip, or ""
	var only = ""        // Name of test case you want isolated and alone, or ""

	for _, tCase := range testCases {

		if only != "" && tCase.name != only {
			continue
		}
		if skip == tCase.name {
			continue
		}

		// Run this test case.
		t.Run(tCase.name, func(t *testing.T) {

			// Make a database name for this test
			re := regexp.MustCompile(`[^\w]`)
			dbName := strings.ToLower(re.ReplaceAllString(tCase.name, ""))

			t.Logf("XXXX database name for this test: %s\n", dbName)

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

			// Migrate the database to this test case's schema.
			schema := protoSchema(t, tCase.keelSchema)
			m := migrations.New(schema, nil)
			require.NoError(t, m.Apply(testDB))

			// Construct the runtime API Handler.
			handler := NewHandler(schema)

			// Assemble the query to send from the test case data.
			reqBody := queryAsJSONPayload(t, tCase.gqlOperation, tCase.variables)
			request := Request{
				Context: runtimectx.WithDatabase(context.Background(), testDB),
				URL: url.URL{
					Path: "/Test",
				},
				Body: []byte(reqBody),
			}

			// Apply the database prior-set up mandated by this test case.
			if tCase.databaseSetup != nil {
				tCase.databaseSetup(t, testDB)
			}

			// Call the handler, and capture the response.
			response, err := handler(&request)
			require.NoError(t, err)
			body := string(response.Body)
			bodyFields := respFields{}
			require.NoError(t, json.Unmarshal([]byte(body), &bodyFields))

			// Unless there is a specific assertion for the error returned,
			// check there is no error
			if tCase.assertErrors == nil {
				require.Len(t, bodyFields.Errors, 0, "response has unexpected errors: %+v", bodyFields.Errors)
			}

			// Do the specified assertion on the data returned, if one is specified.
			if tCase.assertData != nil {
				tCase.assertData(t, bodyFields.Data)
			}

			// Do the specified assertion on the errors returned, if one is specified.
			if tCase.assertErrors != nil {
				tCase.assertErrors(t, bodyFields.Errors)
			}

			// Do the specified assertion on the resultant database contents, if one is specified.
			if tCase.assertDatabase != nil {
				tCase.assertDatabase(t, testDB, bodyFields.Data)
			}
		})
	}
}

// respFields is a container to into which a hanlder's response' body can be
// JSON unmarshalled.
type respFields struct {
	Data   map[string]any             `json:"data"`
	Errors []gqlerrors.FormattedError `json:"errors"`
}

const dbConnString = "host=localhost port=8001 user=postgres password=postgres dbname=%s sslmode=disable"

// protoSchema returns a proto.Schema that has been built from the given
// keel schema.
func protoSchema(t *testing.T, keelSchema string) *proto.Schema {
	builder := &schema.Builder{}
	schema, err := builder.MakeFromInputs(&reader.Inputs{
		SchemaFiles: []reader.SchemaFile{
			{
				Contents: keelSchema,
			},
		},
	})
	require.NoError(t, err)
	return schema
}

// queryAsJSONPayload packages up the given gql mutation, alongside the corresponding input
// variables, as JSON that is good to use as the body for a runtime.Request.
func queryAsJSONPayload(t *testing.T, mutationString string, vars map[string]any) (asJSON string) {
	d := map[string]any{
		"query":     mutationString,
		"variables": vars,
	}
	b, err := json.Marshal(d)
	require.NoError(t, err)
	return string(b)
}

// testCase encapsulates the data required to define one particular test case
// as used by the TestRuntime() test suite.
type testCase struct {
	name           string
	keelSchema     string
	databaseSetup  func(t *testing.T, db *gorm.DB)
	gqlOperation   string
	variables      map[string]any
	assertData     func(t *testing.T, data map[string]any)
	assertErrors   func(t *testing.T, errors []gqlerrors.FormattedError)
	assertDatabase func(t *testing.T, db *gorm.DB, data map[string]any)
}

// initRow makes a map to represent a database row - that is good to use inside the
// databaseSetup part of a testCase, all it does is augment the map you give it with
// created_at and updated_at fields.
func initRow(with map[string]any) map[string]any {
	res := map[string]any{
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	for k, v := range with {
		res[strcase.ToSnake(k)] = v
	}
	return res
}

// basicSchema is a DRY, simplest possible, schema that can be used in test cases.
const basicSchema string = `
	model Person {
		fields {
			name Text 
		}
		operations {
			get getPerson(id) // short-form filter criterion
			create createPerson() with (name)
		}
	}
	api Test {
		@graphql
		models {
			Person
		}
	}
`

// getWhere is a simple schema that contains a minimal WHERE clause.
const getWhere string = `
	model Person {
		fields {
			name Text @unique
		}
		operations {
			get getPerson(name: Text) {
				@where(person.name == name)
			}
		}
	}
	api Test {
		@graphql
		models {
			Person
		}
	}
`

// multiSchema is a schema with a model that exhibits all the simple field types.
const multiSchema string = `
	model Multi {
		fields {
			aText Text
			aBool Boolean
			aNumber Number
		}
		operations {
			get getMulti(id)
			create createMulti() with (aText, aBool, aNumber)
		}
	}
	api Test {
		@graphql
		models {
			Multi
		}
	}
`

// testCases is a list of testCase that is good for the top level test suite to
// iterate over.
var testCases = []testCase{
	{
		name:       "create_operation_happy",
		keelSchema: basicSchema,
		gqlOperation: `
			mutation CreatePerson($name: String!) {
				createPerson(input: {name: $name}) {
					id
					name
				}
			}
		`,
		variables: map[string]any{
			"name": "Fred",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "createPerson.name", "Fred")
		},
		assertDatabase: func(t *testing.T, db *gorm.DB, data map[string]any) {
			id := rtt.GetValueAtPath(t, data, "createPerson.id")
			var name string
			err := db.Table("person").Where("id = ?", id).Pluck("name", &name).Error
			require.NoError(t, err)
			require.Equal(t, "Fred", name)
		},
	},

	{
		name:       "create_operation_errors",
		keelSchema: basicSchema,
		gqlOperation: `
			mutation CreatePerson($name: String!) {
				createPerson(input: {name: $name}) {
					nosuchfield
				}
			}
		`,
		variables: map[string]any{
			"name": "Fred",
		},
		assertErrors: func(t *testing.T, errors []gqlerrors.FormattedError) {
			require.Len(t, errors, 1)
			require.Equal(t, "Cannot query field \"nosuchfield\" on type \"Person\".", errors[0].Message)
		},
	},
	{
		name:       "get_operation_happy",
		keelSchema: basicSchema,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"name": "Sue",
					"id":   "41",
				}),
				initRow(map[string]any{
					"name": "Fred",
					"id":   "42",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("person").Create(row).Error)
			}
		},
		gqlOperation: `
			query GetPerson($id: ID!) {
				getPerson(input: {id: $id}) {
					name
				}
			}
		`,
		variables: map[string]any{
			"id": "42",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "getPerson.name", "Fred")
		},
	},

	{
		name:       "get_operation_error",
		keelSchema: basicSchema,
		gqlOperation: `
			query GetPerson($id: ID!) {
				getPerson(input: {id: $id}) {
					name
				}
			}
		`,
		variables: map[string]any{
			"id": "42",
		},
		assertErrors: func(t *testing.T, errors []gqlerrors.FormattedError) {
			require.Len(t, errors, 1)
			require.Equal(t, "no records found for Get() operation", errors[0].Message)
		},
	},
	{
		name:       "create_all_field_types",
		keelSchema: multiSchema,
		gqlOperation: `
			mutation CreateMulti(
					$aText: String!
					$aBool: Boolean!
					$aNumber: Int!
				) {
				createMulti(input: {
						aText: $aText
						aBool: $aBool
						aNumber: $aNumber
					}) {id aText aBool aNumber}
			}
		`,
		variables: map[string]any{
			"aText":   "Petunia",
			"aBool":   true,
			"aNumber": 8086,
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "createMulti.aText", "Petunia")
			rtt.AssertValueAtPath(t, data, "createMulti.aBool", true)
			rtt.AssertValueAtPath(t, data, "createMulti.aNumber", float64(8086))
			// todo assert time-based field types - currently don't work properly / not implemented in gql
		},
		assertDatabase: func(t *testing.T, db *gorm.DB, data map[string]any) {
			id := rtt.GetValueAtPath(t, data, "createMulti.id")
			record := map[string]any{}
			err := db.Table("multi").Where("id = ?", id).Find(&record).Error
			require.NoError(t, err)

			require.Equal(t, "Petunia", record["a_text"])
			require.Equal(t, true, record["a_bool"])
			require.Equal(t, int32(8086), record["a_number"])
			rtt.AssertIsTimeNow(t, record["created_at"])
			rtt.AssertIsTimeNow(t, record["updated_at"])
			rtt.AssertKSUIDIsNow(t, record["id"])
		},
	},
	{
		name:       "get_all_field_types",
		keelSchema: multiSchema,
		gqlOperation: `
			query GetMulti($id: ID!) {
				getMulti(input: {id: $id}) {
					aText, aBool, aNumber,
				}
			}
		`,
		variables: map[string]any{
			"id": "42",
		},
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":      "42",
					"aText":   "Petunia",
					"aNumber": int(8086),
					"aBool":   true,
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("multi").Create(row).Error)
			}
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "getMulti.aText", "Petunia")
			rtt.AssertValueAtPath(t, data, "getMulti.aBool", true)
			rtt.AssertValueAtPath(t, data, "getMulti.aNumber", float64(8086))
		},
	},
	{
		name:       "get_where",
		keelSchema: getWhere,
		gqlOperation: `
			query GetPerson($name: String!) {
				getPerson(input: {name: $name}) {
				id, name,
			}
		}
		`,
		variables: map[string]any{
			"name": "Sue",
		},
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":   "41",
					"name": "Fred",
				}),
				initRow(map[string]any{
					"id":   "42",
					"name": "Sue",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("person").Create(row).Error)
			}
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "getPerson.id", "42")
			rtt.AssertValueAtPath(t, data, "getPerson.name", "Sue")
		},
	},
}

/*




 */
