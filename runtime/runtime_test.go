package runtime_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/iancoleman/strcase"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/runtimectx"
	rtt "github.com/teamkeel/keel/runtime/runtimetest"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/testhelpers"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NOTE:
// This suite of tests has on the most part been replaced by the integration test framework (see https://github.com/teamkeel/keel/tree/main/integration/testdata)
// HOWEVER, if you want to explicitly test the graphql layer, please add a test here

func TestRuntime(t *testing.T) {
	// We connect to the "main" database here only so we can create a new database
	// for each sub-test
	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			schema := protoSchema(t, tCase.keelSchema)

			// Use the docker compose database
			dbConnInfo := &db.ConnectionInfo{
				Host:     "localhost",
				Port:     "8001",
				Username: "postgres",
				Database: "keel",
				Password: "postgres",
			}

			// Construct the runtime API Handler.
			handler := runtime.NewHandler(schema)

			reqBody := queryAsJSONPayload(t, tCase.gqlOperation, tCase.variables)

			request := &http.Request{
				URL: &url.URL{
					Path: "/test/graphql",
				},
				Method: http.MethodPost,
				Body:   io.NopCloser(strings.NewReader(reqBody)),
			}

			ctx := request.Context()

			dbName := testhelpers.DbNameForTestName(tCase.name)
			testDB, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, dbName)
			require.NoError(t, err)
			defer testDB.Close()

			database, err := db.LocalFromConnection(ctx, testDB)
			require.NoError(t, err)

			ctx = runtimectx.WithDatabase(ctx, database)
			request = request.WithContext(ctx)

			// We are still using gorm database to assert againt the data in the test databases.
			gormDb, err := gorm.Open(postgres.New(postgres.Config{Conn: testDB}), &gorm.Config{})
			require.NoError(t, err)

			// Apply the database prior-set up mandated by this test case.
			if tCase.databaseSetup != nil {
				tCase.databaseSetup(t, gormDb)
			}

			// Call the handler, and capture the response.
			response := handler(request)
			body := string(response.Body)
			bodyFields := respFields{}
			require.NoError(t, json.Unmarshal([]byte(body), &bodyFields))

			// Unless there is a specific assertion for the error returned,
			// check there is no error
			if tCase.assertErrors == nil {
				if len(bodyFields.Errors) != 0 {
					t.Fatalf("response has unexpected errors: %s", litter.Sdump(bodyFields.Errors))
				}
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
				tCase.assertDatabase(t, gormDb, bodyFields.Data)
			}

		})
	}
}

func TestRuntimeRPC(t *testing.T) {
	// We connect to the "main" database here only so we can create a new database
	// for each sub-test

	for _, tCase := range rpcTestCases {
		t.Run(tCase.name, func(t *testing.T) {
			schema := protoSchema(t, tCase.keelSchema)

			// Use the docker compose database
			dbConnInfo := &db.ConnectionInfo{
				Host:     "localhost",
				Port:     "8001",
				Username: "postgres",
				Database: "keel",
				Password: "postgres",
			}

			handler := runtime.NewHandler(schema)

			request := &http.Request{
				URL: &url.URL{
					Path:     "/test/json/" + tCase.Path,
					RawQuery: tCase.QueryParams,
				},
				Method: tCase.Method,
				Body:   io.NopCloser(strings.NewReader(tCase.Body)),
			}

			ctx := request.Context()

			dbName := testhelpers.DbNameForTestName(tCase.name)
			testDB, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, dbName)
			require.NoError(t, err)

			database, err := db.LocalFromConnection(ctx, testDB)
			require.NoError(t, err)

			ctx = runtimectx.WithDatabase(ctx, database)
			request = request.WithContext(ctx)

			// We are still using gorm database to assert againt the data in the test databases.
			gormDb, err := gorm.Open(postgres.New(postgres.Config{Conn: testDB}), &gorm.Config{})
			require.NoError(t, err)

			// Apply the database prior-set up mandated by this test case.
			if tCase.databaseSetup != nil {
				tCase.databaseSetup(t, gormDb)
			}

			// Call the handler, and capture the response.
			response := handler(request)
			body := string(response.Body)
			var res map[string]any
			require.NoError(t, json.Unmarshal([]byte(body), &res))

			// Do the specified assertion on the resultant database contents, if one is specified.
			if tCase.assertDatabase != nil {
				tCase.assertDatabase(t, gormDb, res)
			}

			if response.Status != 200 && tCase.assertError == nil {
				t.Errorf("method %s returned non-200 (%d) but no assertError function provided", tCase.Path, response.Status)
			}
			if tCase.assertError != nil {
				tCase.assertError(t, res)
			}

			// Do the specified assertion on the data returned, if one is specified.
			if tCase.assertResponse != nil {
				tCase.assertResponse(t, res)
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

type rpcTestCase struct {
	name           string
	keelSchema     string
	databaseSetup  func(t *testing.T, db *gorm.DB)
	Path           string
	QueryParams    string
	Body           string
	Method         string
	assertResponse func(t *testing.T, data map[string]any)
	assertError    func(t *testing.T, data map[string]any)
	assertDatabase func(t *testing.T, db *gorm.DB, data interface{})
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

// basicSchema is a DRY, simplest possible, schema that can be used in test cases.
const basicSchema string = `
	model Person {

		@permission(
			expression: true,
			actions: [create, get, list, update, delete]
		)
		fields {
			name Text 
		}
		operations {
			get getPerson(id)
			create createPerson() with (name)
			update updatePerson(id) with (name)
			list listPeople(name)
			delete deletePerson(id)
		}
	}
	api Test {
		models {
			Person
		}
	}
`

// getWhere is a simple schema that contains a minimal WHERE clause.
const getWhere string = `
	model Person {

		@permission(
			expression: true,
			actions: [create, get, list, update, delete]
		)
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
		models {
			Person
		}
	}
`

const listImplicitAndExplicitInputs string = `
	model Person {

		@permission(
			expression: true,
			actions: [create, get, list, update, delete]
		)
		fields {
			firstName Text
			secondName Text
		}
		operations {
			list listPeople(firstName, secondName: Text) {
				@where(person.secondName == secondName)
			}
		}
	}
	api Test {
		models {
			Person
		}
	}
`

// multiSchema is a schema with a model that exhibits all the simple field types.
const multiSchema string = `
	model Multi {

		@permission(
			expression: true,
			actions: [create, get, list, update, delete]
		)
		fields {
			aText Text
			aBool Boolean
			aNumber Number
		}
		operations {
			get getMulti(id)
			create createMulti() with (aText, aBool, aNumber)
			update updateMulti(id) with (aText, aBool, aNumber)
		}
	}
	api Test {
		models {
			Multi
		}
	}
`

// Schema with all field types
const fieldTypes string = `
	model Thing {

		@permission(
			expression: true,
			actions: [create, get, list, update, delete]
		)
		fields {
			text Text @unique
			bool Boolean
			timestamp Timestamp
			date Date
			number Number
			enum Enums
		}
		operations {
			list listThings(text?, bool?, date?, timestamp?, number?, enum?)
		}
	}
	enum Enums {
		Option1
		Option2
	}
	api Test {
		models {
			Thing
		}
	}
`

const relationships string = `
	model BlogPost {
		fields {
			title Text
			author Author
		}
		operations {
			get getPost(id)

			create createPost() with (title, author.id)
			create createPostWithFk() with (title, authorId)
			update updatePost(id) with (title, author.id)
			update updatePostWithFk(id) with (title, authorId)
		}
		@permission(
			expression: true,
			actions: [get, create, update]
		)
	}
	
	model Author {
		fields {
			name Text
			posts BlogPost[]
			publisher Publisher
		}
		operations {
			get getAuthor(id)
			list listAuthors(name)
			update updateAuthor(id) with (name)
		}
		@permission(
			expression: true,
			actions: [get, list, update]
		)
	}

	model Publisher {
		fields {
			organisation Text
		}
	}

	api Test {
		models {
			BlogPost
			Author
			Publisher
		}
	}		
`

const date_timestamp_parsing = `
	model Thing {
		fields {
			theDate Date
			theTimestamp Timestamp
		}
		operations {
			create createThing() with (theDate, theTimestamp)
			update updateThing(id) with (theDate, theTimestamp)
			get getThing(id, theDate, theTimestamp)
			list listThing(theDate, theTimestamp)
		}
		@permission(
			expression: true,
			actions: [get, create, update, list]
		)
	}
	api Test {
		models {
			Thing
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
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "getPerson", nil)
		},
	},
	{
		name:       "delete_operation_happy",
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
			mutation DeletePerson($id: ID!) {
				deletePerson(input: {id: $id}) {
					success
				}
			}
		`,
		variables: map[string]any{
			"id": "42",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "deletePerson.success", true)
		},
	},
	{
		name:       "update_operation_happy",
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
			mutation UpdatePerson($id: ID!, $name: String!) {
				updatePerson(input: { where: { id: $id }, values: { name: $name }}) {
					id
					name
				}
			}
		`,
		variables: map[string]any{
			"id":   42,
			"name": "Keelson",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "updatePerson.name", "Keelson")
		},
		assertDatabase: func(t *testing.T, db *gorm.DB, data map[string]any) {
			id := rtt.GetValueAtPath(t, data, "updatePerson.id")
			var name string
			err := db.Table("person").Where("id = ?", id).Pluck("name", &name).Error
			require.NoError(t, err)
			require.Equal(t, "Keelson", name)
		},
	},
	{
		name:       "update_operation_errors",
		keelSchema: basicSchema,
		gqlOperation: `
			mutation UpdatePerson($id: ID!, $name: String!) {
				updatePerson(input: { where: { id: $id }, values: { name: $name }}) {
					id
					name
				}
			}
		`,
		variables: map[string]any{
			"id":   42,
			"name": "Fred",
		},
		assertErrors: func(t *testing.T, errors []gqlerrors.FormattedError) {
			require.Len(t, errors, 1)
			require.Equal(t, "record not found", errors[0].Message)
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
		name:       "update_all_field_types",
		keelSchema: multiSchema,
		gqlOperation: `
			mutation UpdateMulti(
				$id: ID!
				$aText: String!
				$aBool: Boolean!
				$aNumber: Int!
			) {
			updateMulti(input: {
				where: {
					id: $id
				},
				values: {
					aText: $aText
					aBool: $aBool
					aNumber: $aNumber
				}
			}) {id aText aBool aNumber}
		}
		`,
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
		variables: map[string]any{
			"id":      "42",
			"aText":   "Keelson",
			"aNumber": int(8001),
			"aBool":   false,
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "updateMulti.aText", "Keelson")
			rtt.AssertValueAtPath(t, data, "updateMulti.aBool", false)
			rtt.AssertValueAtPath(t, data, "updateMulti.aNumber", float64(8001))
		},
		assertDatabase: func(t *testing.T, db *gorm.DB, data map[string]any) {
			id := rtt.GetValueAtPath(t, data, "updateMulti.id")
			record := map[string]any{}
			err := db.Table("multi").Where("id = ?", id).Find(&record).Error
			require.NoError(t, err)

			require.Equal(t, "Keelson", record["a_text"])
			require.Equal(t, false, record["a_bool"])
			require.Equal(t, int32(8001), record["a_number"])
			rtt.AssertIsTimeNow(t, record["created_at"])
			rtt.AssertIsTimeNow(t, record["updated_at"])
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
	{
		name:       "list_operation_generic_and_paging_logic",
		keelSchema: basicSchema,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			for _, nameStub := range []string{"Fred", "Sue"} {
				for i := 0; i < 100; i++ {
					name := fmt.Sprintf("%s_%d", nameStub, i)
					id := fmt.Sprintf("%s_%04d_id", nameStub, i)
					row := initRow(map[string]any{
						"name": name,
						"id":   id,
					})
					require.NoError(t, db.Table("person").Create(row).Error)
				}
			}
		},
		gqlOperation: `
			query ListPeople {
				listPeople(input: { first: 10, after: "Fred_0008_id", where: { name: { startsWith: "Fr" } } })
				{
					pageInfo {
						hasNextPage
						startCursor
						endCursor
					}
					edges {
					  node {
						id
						name
					  }
					}
				  }
		 	}`,

		assertData: func(t *testing.T, data map[string]any) {
			edges := rtt.GetValueAtPath(t, data, "listPeople.edges")
			edgesList, ok := edges.([]any)
			require.True(t, ok)
			// Check conformance with the request asking for the first 10, after id == "Fred_0008_id"
			require.Len(t, edgesList, 10)
			first := edgesList[0]
			edge, ok := first.(map[string]any)
			require.True(t, ok)
			rtt.AssertValueAtPath(t, edge, "node.id", "Fred_0009_id")

			// Check the correctness of the returned page metadata
			pageInfo := rtt.GetValueAtPath(t, data, "listPeople.pageInfo")
			pageInfoMap, ok := pageInfo.(map[string]any)
			require.True(t, ok)
			rtt.AssertValueAtPath(t, pageInfoMap, "startCursor", "Fred_0009_id")
			rtt.AssertValueAtPath(t, pageInfoMap, "endCursor", "Fred_0018_id")
			rtt.AssertValueAtPath(t, pageInfoMap, "hasNextPage", true)

			// todo - we should test hasNextPage when there isn't one - but defer until we switch over to
			// the integration test framework.
		},
	},
	{
		name:       "list_inputs",
		keelSchema: fieldTypes,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			row1 := initRow(map[string]any{
				"id":        "id_123",
				"text":      "some-interesting-text",
				"bool":      true,
				"timestamp": "1970-01-01 00:00:10",
				"date":      "2020-01-02",
				"number":    10,
				"enum":      "Option1",
			})
			require.NoError(t, db.Table("thing").Create(row1).Error)
		},
		gqlOperation: `

		fragment Fields on ThingConnection {
			edges {
				node {
				text
				bool
				# timestamp {seconds}
				# date {day, month, year}
				number
				enum
				}
			}
		}
		{
		string_equals: listThings(input: {where: {text: {equals: "some-interesting-text"}}}) {
			...Fields
		},
		string_startsWith: listThings(input: {where: {text: {startsWith: "some"}}}) {
			...Fields
		},
		string_endWith: listThings(input: {where: {text: {endsWith: "-text"}}}) {
			...Fields
		},
		string_contains: listThings(input: {where: {text: {contains: "interesting"}}}) {
			...Fields
		},
		string_oneOf: listThings(input: {where: {text: {oneOf: ["some-interesting-text", "Another"]}}}) {
			...Fields
		},
		number_equals: listThings(input: {where: {number: {equals: 10}}}) {
			...Fields
		},
		number_gt: listThings(input: {where: {number: {greaterThan: 9}}}) {
			...Fields
		},
		number_gte: listThings(input: {where: {number: {greaterThanOrEquals: 10}}}) {
			...Fields
		},
		number_lt: listThings(input: {where: {number: {lessThan: 11}}}) {
			...Fields
		},
		number_lte: listThings(input: {where: {number: {lessThanOrEquals: 10}}}) {
			...Fields
		},
		enum_equals: listThings(input: {where: {enum: {equals: Option1}}}) {
			...Fields
		},
		enum_oneOf: listThings(input: {where: {enum: {oneOf: [Option1]}}}) {
			...Fields
		},
		timestamp_before: listThings(input: {
			where: {
			timestamp: {
				before: {
					seconds: 11
				}
			}
			}
		}) {
			...Fields
		},
		timestamp_after: listThings(input: {
			where: {
			timestamp: {
				after: {
					seconds: 9
				}
			}
			}
		}) {
			...Fields
		},
		date_before: listThings(input: {where: {date: {before: {
			year: 2020,
			month: 1,
			day: 3
		}}}}) {
			...Fields
		},
		date_after: listThings(input: {where: {date: {after: {
			year: 2020,
			month: 1,
			day: 1
		}}}}) {
			...Fields
		},
		date_onOrbefore: listThings(input: {where: {date: {onOrBefore: {
			year: 2020,
			month: 1,
			day: 2
		}}}}) {
			...Fields
		},
		date_onOrAfter: listThings(input: {where: {date: {onOrAfter: {
			year: 2020,
			month: 1,
			day: 2
		}}}}) {
			...Fields
		},
		date_onOrEquals: listThings(input: {where: {date: {equals: {
			year: 2020,
			month: 1,
			day: 2
		}}}}) {
			...Fields
		},
		bool: listThings(input: {
			where: {
			bool: {
					equals: true
				}
			}
		}) {
			...Fields
		}
		combined: listThings(input: {
			where: {
			bool: {
					equals: true
			},
			enum: {
				equals: Option1
			}
			}
		}) {
			...Fields
		}
		}`,
		assertData: func(t *testing.T, data map[string]any) {

			keys := []string{
				"string_equals",
				"string_startsWith",
				"string_endWith",
				"string_contains",
				"string_oneOf",
				"number_equals",
				"number_gt",
				"number_gte",
				"number_lt",
				"number_lte",
				"enum_equals",
				"enum_oneOf",
				"timestamp_before",
				"timestamp_after",
				"date_before",
				"date_after",
				"date_onOrbefore",
				"date_onOrAfter",
				"date_onOrEquals",
				"bool",
				"combined",
			}

			for _, key := range keys {
				edges := rtt.GetValueAtPath(t, data, key+".edges")
				edgesList, ok := edges.([]any)
				fmt.Println(key)
				require.True(t, ok)
				require.Len(t, edgesList, 1)
			}
		},
	},
	{
		name:       "list_impl_and_expl_inputs",
		keelSchema: listImplicitAndExplicitInputs,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":          "41",
					"first_name":  "Fred",
					"second_name": "Smith",
				}),
				initRow(map[string]any{
					"id":          "42",
					"first_name":  "Francis",
					"second_name": "Smith",
				}),
				initRow(map[string]any{
					"id":          "43",
					"first_name":  "Same",
					"second_name": "Smith",
				}),
				initRow(map[string]any{
					"id":          "44",
					"first_name":  "Fred",
					"second_name": "Bloggs",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("person").Create(row).Error)
			}
		},
		gqlOperation: `
			query ListPeople {
				listPeople(input: { where: {
					firstName: { startsWith: "Fr" }
					secondName: "Smith"
				} }) 	   
				{
					pageInfo {
						hasNextPage
						startCursor
						endCursor
					}
					edges {
					  node {
						id
						firstName
						secondName
					  }
					}
				  }
		 	}`,

		assertData: func(t *testing.T, data map[string]any) {
			edges := rtt.GetValueAtPath(t, data, "listPeople.edges")
			edgesList, ok := edges.([]any)
			require.True(t, ok)
			require.Len(t, edgesList, 2)

			record := edgesList[0]
			edge, ok := record.(map[string]any)
			require.True(t, ok)
			rtt.AssertValueAtPath(t, edge, "node.firstName", "Fred")
			rtt.AssertValueAtPath(t, edge, "node.secondName", "Smith")

			record = edgesList[1]
			edge, ok = record.(map[string]any)
			require.True(t, ok)
			rtt.AssertValueAtPath(t, edge, "node.firstName", "Francis")
			rtt.AssertValueAtPath(t, edge, "node.secondName", "Smith")
		},
	},
	{
		name: "operation_create_set_attribute_with_text_literal",
		keelSchema: `
			model Person {

				@permission(
					expression: true,
					actions: [create, get, list, update, delete]
				)
				fields {
					name Text
					nickname Text?
				}
				operations {
					get getPerson(id)
					create createPerson() with (name) {
						@set(person.nickname = "Joe Soap")
					}	
				}
			}
			api Test {
				models {
					Person
				}
			}
		`,
		gqlOperation: `
			mutation CreatePerson($name: String!) {
				createPerson(input: {name: $name}) {
					id
					name
					nickname
				}
			}
		`,
		variables: map[string]any{
			"name": "Fred",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "createPerson.nickname", "Joe Soap")
		},
		assertDatabase: func(t *testing.T, db *gorm.DB, data map[string]any) {
			id := rtt.GetValueAtPath(t, data, "createPerson.id")
			var name string
			err := db.Table("person").Where("id = ?", id).Pluck("nickname", &name).Error
			require.NoError(t, err)
			require.Equal(t, "Joe Soap", name)
		},
	},
	{
		name: "operation_create_set_attribute_with_number_literal",
		keelSchema: `
			model Person {

				@permission(
					expression: true,
					actions: [create, get, list, update, delete]
				)
				fields {
					name Text
					age Number?
				}
				operations {
					get getPerson(id)
					create createPerson() with (name) {
						@set(person.age = 1)
					}
				}
			}
			api Test {
				models {
					Person
				}
			}
		`,
		gqlOperation: `
			mutation CreatePerson($name: String!) {
				createPerson(input: {name: $name}) {
					id
					name
					age
				}
			}
		`,
		variables: map[string]any{
			"name": "Fred",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "createPerson.age", 1.0)
		},
		assertDatabase: func(t *testing.T, db *gorm.DB, data map[string]any) {
			id := rtt.GetValueAtPath(t, data, "createPerson.id")
			var name string
			err := db.Table("person").Where("id = ?", id).Pluck("age", &name).Error
			require.NoError(t, err)
			require.Equal(t, "1", name)
		},
	},
	{
		name: "operation_create_set_attribute_with_boolean_literal",
		keelSchema: `
			model Person {

				@permission(
					expression: true,
					actions: [create, get, list, update, delete]
				)
				fields {
					name Text
					hasFriends Boolean?
				}
				operations {
					get getPerson(id)
					create createPerson() with (name) {
						@set(person.hasFriends = true)
					}
				}
			}
			api Test {
				models {
					Person
				}
			}
		`,
		gqlOperation: `
			mutation CreatePerson($name: String!) {
				createPerson(input: {name: $name}) {
					id
					name
					hasFriends
				}
			}
		`,
		variables: map[string]any{
			"name": "Fred",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "createPerson.hasFriends", true)
		},
		assertDatabase: func(t *testing.T, db *gorm.DB, data map[string]any) {
			id := rtt.GetValueAtPath(t, data, "createPerson.id")
			var name string
			err := db.Table("person").Where("id = ?", id).Pluck("has_friends", &name).Error
			require.NoError(t, err)
			require.Equal(t, "true", name)
		},
	},
	{
		name: "operation_authenticate_new_user",
		keelSchema: `
			model Person {

				@permission(
					expression: true,
					actions: [create, get, list, update, delete]
				)
				fields {
					name Text
				}
				operations {
					get getPerson(id)
				}
			}
			api Test {
				models {
					Person
				}
			}
		`,
		gqlOperation: `
			mutation {
				authenticate(input: { 
					createIfNotExists: true, 
					emailPassword: { 
						email: "newuser@keel.xyz", 
						password: "1234"
					}
				}) {
					identityCreated
					token
				}
			}
		`,
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "authenticate.identityCreated", true)
			token := rtt.GetValueAtPath(t, data, "authenticate.token")
			require.NotEmpty(t, token)
		},
	},
	{
		name: "operation_authenticate_createifnotexists_false",
		keelSchema: `
			model Person {

				@permission(
					expression: true,
					actions: [create, get, list, update, delete]
				)
				fields {
					name Text
				}
				operations {
					get getPerson(id)
				}
			}
			api Test {
				models {
					Person
				}
			}
		`,
		gqlOperation: `
			mutation {
				authenticate(input: { 
					createIfNotExists: false, 
					emailPassword: { 
						email: "newuser@keel.xyz", 
						password: "1234"
					}
				}) {
					identityCreated
					token
				}
			}
		`,
		assertErrors: func(t *testing.T, errors []gqlerrors.FormattedError) {
			require.Len(t, errors, 1)
			require.Equal(t, "failed to authenticate", errors[0].Message)
		},
	},
	{
		name: "list_no_inputs",
		keelSchema: `
		model Thing {

			@permission(
				expression: true,
				actions: [create, get, list, update, delete]
			)
			fields {
				text Text @unique
				bool Boolean
				timestamp Timestamp
				date Date
				number Number
			}
			operations {
				list listThings()
			}
		}
		api Test {
			models {
				Thing
			}
		}
		`,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			row1 := initRow(map[string]any{
				"id":        "id_123",
				"text":      "some-interesting-text",
				"bool":      true,
				"timestamp": "1970-01-01 00:00:10",
				"date":      "2020-01-02",
				"number":    10,
			})
			require.NoError(t, db.Table("thing").Create(row1).Error)
		},
		gqlOperation: `
		query {
			listThings {
				edges {
					node {
						text
					}
				}
			}
		}
		`,
		assertData: func(t *testing.T, data map[string]any) {
			edges := rtt.GetValueAtPath(t, data, "listThings.edges")
			edgesList, ok := edges.([]any)
			require.True(t, ok)
			require.Len(t, edgesList, 1)

		},
	},
	{
		name:       "get_operation_relationship_belongs_to",
		keelSchema: relationships,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":           "publisher_1",
					"organisation": "Keelson Publishers",
				}),
				initRow(map[string]any{
					"id":           "publisher_2",
					"organisation": "Weaveton Publishers",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("publisher").Create(row).Error)
			}
			rows = []map[string]any{
				initRow(map[string]any{
					"id":          "author_1",
					"name":        "Keelson",
					"publisherId": "publisher_1",
				}),
				initRow(map[string]any{
					"id":          "author_2",
					"name":        "Weaveton",
					"publisherId": "publisher_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("author").Create(row).Error)
			}
		},
		gqlOperation: `
			query GetAuthor($authorId: ID!) {
				getAuthor(input: { id: $authorId }) {
					id
					name
					publisher {
						id
						organisation
					}
				}
		 	}`,
		variables: map[string]any{
			"authorId": "author_1",
		},
		assertData: func(t *testing.T, data map[string]any) {
			authorId := rtt.GetValueAtPath(t, data, "getAuthor.id")
			require.Equal(t, "author_1", authorId)

			rtt.AssertValueAtPath(t, data, "getAuthor.publisher.id", "publisher_1")
			rtt.AssertValueAtPath(t, data, "getAuthor.publisher.organisation", "Keelson Publishers")
		},
	},
	{
		name:       "get_operation_relationships_has_many_page_1",
		keelSchema: relationships,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":           "publisher_1",
					"organisation": "Keelson Publishers",
				}),
				initRow(map[string]any{
					"id":           "publisher_2",
					"organisation": "Weaveton Publishers",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("publisher").Create(row).Error)
			}
			rows = []map[string]any{
				initRow(map[string]any{
					"id":          "author_1",
					"name":        "Keelson",
					"publisherId": "publisher_1",
				}),
				initRow(map[string]any{
					"id":          "author_2",
					"name":        "Weaveton",
					"publisherId": "publisher_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("author").Create(row).Error)
			}

			rows = []map[string]any{
				initRow(map[string]any{
					"id":       "post_1",
					"title":    "Keelson First Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_2",
					"title":    "Keelson Second Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_3",
					"title":    "Keelson Third Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_4",
					"title":    "Weaveton First Second",
					"authorId": "author_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("blog_post").Create(row).Error)
			}
		},
		gqlOperation: `
			query GetAuthor($authorId: ID!, $first: Int!) {
				getAuthor(input: { id: $authorId }) {
					id
					name
					posts(first: $first) {
						edges {
						  node {
							id
							title
						  }
						}
						pageInfo {
							hasNextPage
							startCursor
							endCursor
						}
					}
				}
		 	}`,
		variables: map[string]any{
			"authorId": "author_1",
			"first":    2,
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "getAuthor.id", "author_1")

			posts := rtt.GetValueAtPath(t, data, "getAuthor.posts.edges").([]any)
			require.Len(t, posts, 2)

			first := posts[0].(map[string]any)
			rtt.AssertValueAtPath(t, first, "node.id", "post_1")
			rtt.AssertValueAtPath(t, first, "node.title", "Keelson First Post")

			second := posts[1].(map[string]any)
			rtt.AssertValueAtPath(t, second, "node.id", "post_2")
			rtt.AssertValueAtPath(t, second, "node.title", "Keelson Second Post")

			// Check the correctness of the returned page metadata
			pageInfo := rtt.GetValueAtPath(t, data, "getAuthor.posts.pageInfo")
			pageInfoMap, ok := pageInfo.(map[string]any)
			require.True(t, ok)
			rtt.AssertValueAtPath(t, pageInfoMap, "startCursor", "post_1")
			rtt.AssertValueAtPath(t, pageInfoMap, "endCursor", "post_2")
			rtt.AssertValueAtPath(t, pageInfoMap, "hasNextPage", true)
		},
	},
	{
		name:       "get_operation_relationships_has_many_page_2",
		keelSchema: relationships,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":           "publisher_1",
					"organisation": "Keelson Publishers",
				}),
				initRow(map[string]any{
					"id":           "publisher_2",
					"organisation": "Weaveton Publishers",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("publisher").Create(row).Error)
			}
			rows = []map[string]any{
				initRow(map[string]any{
					"id":          "author_1",
					"name":        "Keelson",
					"publisherId": "publisher_1",
				}),
				initRow(map[string]any{
					"id":          "author_2",
					"name":        "Weaveton",
					"publisherId": "publisher_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("author").Create(row).Error)
			}

			rows = []map[string]any{
				initRow(map[string]any{
					"id":       "post_1",
					"title":    "Keelson First Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_2",
					"title":    "Keelson Second Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_3",
					"title":    "Keelson Third Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_4",
					"title":    "Weaveton First Second",
					"authorId": "author_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("blog_post").Create(row).Error)
			}
		},
		gqlOperation: `
			query GetAuthor($authorId: ID!, $first: Int!, $after: String!) {
				getAuthor(input: { id: $authorId }) {
					id
					name
					posts(first: $first, after: $after) {
						edges {
							node {
							id
							title
							}
						}
						pageInfo {
							hasNextPage
							startCursor
							endCursor
						}
					}
				}
			}`,
		variables: map[string]any{
			"authorId": "author_1",
			"first":    2,
			"after":    "post_2",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "getAuthor.id", "author_1")

			posts := rtt.GetValueAtPath(t, data, "getAuthor.posts.edges").([]any)
			require.Len(t, posts, 1)

			first := posts[0].(map[string]any)
			rtt.AssertValueAtPath(t, first, "node.id", "post_3")
			rtt.AssertValueAtPath(t, first, "node.title", "Keelson Third Post")

			// Check the correctness of the returned page metadata
			pageInfo := rtt.GetValueAtPath(t, data, "getAuthor.posts.pageInfo")
			pageInfoMap, ok := pageInfo.(map[string]any)
			require.True(t, ok)
			rtt.AssertValueAtPath(t, pageInfoMap, "startCursor", "post_3")
			rtt.AssertValueAtPath(t, pageInfoMap, "endCursor", "post_3")
			rtt.AssertValueAtPath(t, pageInfoMap, "hasNextPage", false)
		},
	},
	{
		name:       "list_operation_relationships_has_many_page_1",
		keelSchema: relationships,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":           "publisher_1",
					"organisation": "Keelson Publishers",
				}),
				initRow(map[string]any{
					"id":           "publisher_2",
					"organisation": "Weaveton Publishers",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("publisher").Create(row).Error)
			}
			rows = []map[string]any{
				initRow(map[string]any{
					"id":          "author_1",
					"name":        "Keelson",
					"publisherId": "publisher_1",
				}),
				initRow(map[string]any{
					"id":          "author_2",
					"name":        "Weaveton",
					"publisherId": "publisher_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("author").Create(row).Error)
			}

			rows = []map[string]any{
				initRow(map[string]any{
					"id":       "post_1",
					"title":    "Keelson First Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_2",
					"title":    "Keelson Second Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_3",
					"title":    "Keelson Third Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_4",
					"title":    "Weaveton First Second",
					"authorId": "author_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("blog_post").Create(row).Error)
			}
		},
		gqlOperation: `
			query ListAuthors($authorName: String!, $first: Int!) {
				listAuthors(input: { where: { name: { equals: $authorName } } }) {
					edges {
						node {
							id
							name
							posts(first: $first) {
								edges {
								  node {
									id
									title
								  }
								}
								pageInfo {
									hasNextPage
									startCursor
									endCursor
								}
							}
						}
					}
				}
		 	}`,
		variables: map[string]any{
			"authorName": "Keelson",
			"first":      2,
		},
		assertData: func(t *testing.T, data map[string]any) {
			authors := rtt.GetValueAtPath(t, data, "listAuthors.edges").([]any)
			require.Len(t, authors, 1)

			author := authors[0].(map[string]any)
			rtt.AssertValueAtPath(t, author, "node.id", "author_1")
			rtt.AssertValueAtPath(t, author, "node.name", "Keelson")

			posts := rtt.GetValueAtPath(t, author, "node.posts.edges").([]any)
			require.Len(t, posts, 2)

			first := posts[0].(map[string]any)
			rtt.AssertValueAtPath(t, first, "node.id", "post_1")
			rtt.AssertValueAtPath(t, first, "node.title", "Keelson First Post")

			second := posts[1].(map[string]any)
			rtt.AssertValueAtPath(t, second, "node.id", "post_2")
			rtt.AssertValueAtPath(t, second, "node.title", "Keelson Second Post")

			// Check the correctness of the returned page metadata
			pageInfo := rtt.GetValueAtPath(t, author, "node.posts.pageInfo")
			pageInfoMap, ok := pageInfo.(map[string]any)
			require.True(t, ok)
			rtt.AssertValueAtPath(t, pageInfoMap, "startCursor", "post_1")
			rtt.AssertValueAtPath(t, pageInfoMap, "endCursor", "post_2")
			rtt.AssertValueAtPath(t, pageInfoMap, "hasNextPage", true)
		},
	},
	{
		name:       "list_operation_relationships_has_many_page_2",
		keelSchema: relationships,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":           "publisher_1",
					"organisation": "Keelson Publishers",
				}),
				initRow(map[string]any{
					"id":           "publisher_2",
					"organisation": "Weaveton Publishers",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("publisher").Create(row).Error)
			}
			rows = []map[string]any{
				initRow(map[string]any{
					"id":          "author_1",
					"name":        "Keelson",
					"publisherId": "publisher_1",
				}),
				initRow(map[string]any{
					"id":          "author_2",
					"name":        "Weaveton",
					"publisherId": "publisher_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("author").Create(row).Error)
			}
			rows = []map[string]any{
				initRow(map[string]any{
					"id":       "post_1",
					"title":    "Keelson First Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_2",
					"title":    "Keelson Second Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_3",
					"title":    "Keelson Third Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_4",
					"title":    "Weaveton First Second",
					"authorId": "author_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("blog_post").Create(row).Error)
			}
		},
		gqlOperation: `
			query ListAuthors($authorName: String!, $first: Int!, $after: String!) {
				listAuthors(input: { where: { name: { equals: $authorName } } }) {
					edges {
						node {
							id
							name
							posts(first: $first, after: $after) {
								edges {
								  node {
									id
									title
								  }
								}
								pageInfo {
									hasNextPage
									startCursor
									endCursor
								}
							}
						}
					}
				}
		 	}`,
		variables: map[string]any{
			"authorName": "Keelson",
			"first":      2,
			"after":      "post_2",
		},
		assertData: func(t *testing.T, data map[string]any) {
			authors := rtt.GetValueAtPath(t, data, "listAuthors.edges").([]any)
			require.Len(t, authors, 1)

			author := authors[0].(map[string]any)
			rtt.AssertValueAtPath(t, author, "node.id", "author_1")
			rtt.AssertValueAtPath(t, author, "node.name", "Keelson")

			posts := rtt.GetValueAtPath(t, author, "node.posts.edges").([]any)
			require.Len(t, posts, 1)

			first := posts[0].(map[string]any)
			rtt.AssertValueAtPath(t, first, "node.id", "post_3")
			rtt.AssertValueAtPath(t, first, "node.title", "Keelson Third Post")

			// Check the correctness of the returned page metadata
			pageInfo := rtt.GetValueAtPath(t, author, "node.posts.pageInfo")
			pageInfoMap, ok := pageInfo.(map[string]any)
			require.True(t, ok)
			rtt.AssertValueAtPath(t, pageInfoMap, "startCursor", "post_3")
			rtt.AssertValueAtPath(t, pageInfoMap, "endCursor", "post_3")
			rtt.AssertValueAtPath(t, pageInfoMap, "hasNextPage", false)
		},
	},
	{
		name:       "update_operation_relationships_paging",
		keelSchema: relationships,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":           "publisher_1",
					"organisation": "Keelson Publishers",
				}),
				initRow(map[string]any{
					"id":           "publisher_2",
					"organisation": "Weaveton Publishers",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("publisher").Create(row).Error)
			}
			rows = []map[string]any{
				initRow(map[string]any{
					"id":          "author_1",
					"name":        "Keelson",
					"publisherId": "publisher_1",
				}),
				initRow(map[string]any{
					"id":          "author_2",
					"name":        "Weaveton",
					"publisherId": "publisher_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("author").Create(row).Error)
			}

			rows = []map[string]any{
				initRow(map[string]any{
					"id":       "post_1",
					"title":    "Keelson First Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_2",
					"title":    "Keelson Second Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_3",
					"title":    "Keelson Third Post",
					"authorId": "author_1",
				}),
				initRow(map[string]any{
					"id":       "post_4",
					"title":    "Weaveton First Second",
					"authorId": "author_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("blog_post").Create(row).Error)
			}
		},
		gqlOperation: `
			mutation UpdateAuthor($authorId: ID!, $authorName: String!, $first: Int!) {
				updateAuthor(input: { where: { id: $authorId }, values: { name: $authorName } }) {
					id
					name
					posts(first: $first) {
						edges {
						  node {
							id
							title
						  }
						}
						pageInfo {
							hasNextPage
							startCursor
							endCursor
						}
					}
				}
		 	}`,
		variables: map[string]any{
			"authorId":   "author_1",
			"authorName": "Keeeeelson",
			"first":      2,
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "updateAuthor.id", "author_1")
			rtt.AssertValueAtPath(t, data, "updateAuthor.name", "Keeeeelson")

			posts := rtt.GetValueAtPath(t, data, "updateAuthor.posts.edges").([]any)
			require.Len(t, posts, 2)

			first := posts[0].(map[string]any)
			rtt.AssertValueAtPath(t, first, "node.id", "post_1")
			rtt.AssertValueAtPath(t, first, "node.title", "Keelson First Post")

			second := posts[1].(map[string]any)
			rtt.AssertValueAtPath(t, second, "node.id", "post_2")
			rtt.AssertValueAtPath(t, second, "node.title", "Keelson Second Post")

			// Check the correctness of the returned page metadata
			pageInfo := rtt.GetValueAtPath(t, data, "updateAuthor.posts.pageInfo")
			pageInfoMap, ok := pageInfo.(map[string]any)
			require.True(t, ok)
			rtt.AssertValueAtPath(t, pageInfoMap, "startCursor", "post_1")
			rtt.AssertValueAtPath(t, pageInfoMap, "endCursor", "post_2")
			rtt.AssertValueAtPath(t, pageInfoMap, "hasNextPage", true)
		},
	},
	{
		name: "missing_lookup_in_has_a_relationship",
		keelSchema: `
			model BlogPost {
				fields {
					title Text
					author Author?
				}
				operations {
					get getPost(id)
				}
				@permission(
					expression: true,
					actions: [get]
				)
			}
			
			model Author {
				fields {
					name Text
				}
			}

			api Test {
				models {
					BlogPost
					Author
				}
			}		
		`,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":    "post_1",
					"title": "Without an Author",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("blog_post").Create(row).Error)
			}
		},
		gqlOperation: `
			query GetPost($postId: ID!) {
				getPost(input: { id: $postId }) {
					id
					title
					author {
						name
					}
				}
		 	}`,
		variables: map[string]any{
			"postId": "post_1",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "getPost.id", "post_1")
			rtt.AssertValueAtPath(t, data, "getPost.title", "Without an Author")
			rtt.AssertValueAtPath(t, data, "getPost.author", nil)
		},
	},
	{
		name:       "create_relationship_with_parent_id",
		keelSchema: relationships,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":           "publisher_1",
					"organisation": "Keelson Publishers",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("publisher").Create(row).Error)
			}
			rows = []map[string]any{
				initRow(map[string]any{
					"id":          "author_1",
					"name":        "Keelson",
					"publisherId": "publisher_1",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("author").Create(row).Error)
			}
		},
		gqlOperation: `
			mutation CreatePost($authorId: ID!, $title: String!) {
				createPost(input: { title: $title, authorId: $authorId }) {
					id
					title
					author {
			          id
					  name
					}
				}
		 	}`,
		variables: map[string]any{
			"authorId": "author_1",
			"title":    "Keelson Post",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "createPost.title", "Keelson Post")
			rtt.AssertValueAtPath(t, data, "createPost.author.id", "author_1")
			rtt.AssertValueAtPath(t, data, "createPost.author.name", "Keelson")
		},
	},
	{
		name:       "create_relationship_with_parent_id_foreign_key",
		keelSchema: relationships,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":           "publisher_1",
					"organisation": "Keelson Publishers",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("publisher").Create(row).Error)
			}
			rows = []map[string]any{
				initRow(map[string]any{
					"id":          "author_1",
					"name":        "Keelson",
					"publisherId": "publisher_1",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("author").Create(row).Error)
			}
		},
		gqlOperation: `
			mutation CreatePost($authorId: ID!, $title: String!) {
				createPostWithFk(input: { title: $title, authorId: $authorId }) {
					id
					title
					author {
			          id
					  name
					}
				}
		 	}`,
		variables: map[string]any{
			"authorId": "author_1",
			"title":    "Keelson Post",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "createPostWithFk.title", "Keelson Post")
			rtt.AssertValueAtPath(t, data, "createPostWithFk.author.id", "author_1")
			rtt.AssertValueAtPath(t, data, "createPostWithFk.author.name", "Keelson")
		},
	},
	{
		name:       "update_relationship_with_parent_id",
		keelSchema: relationships,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":           "publisher_1",
					"organisation": "Keelson Publishers",
				}),
				initRow(map[string]any{
					"id":           "publisher_2",
					"organisation": "Weaveton Publishers",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("publisher").Create(row).Error)
			}
			rows = []map[string]any{
				initRow(map[string]any{
					"id":          "author_1",
					"name":        "Keelson",
					"publisherId": "publisher_1",
				}),
				initRow(map[string]any{
					"id":          "author_2",
					"name":        "Weaveton",
					"publisherId": "publisher_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("author").Create(row).Error)
			}

			rows = []map[string]any{
				initRow(map[string]any{
					"id":       "post_1",
					"title":    "Keelson First Post",
					"authorId": "author_1",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("blog_post").Create(row).Error)
			}
		},
		gqlOperation: `
			mutation UpdatePost($postId: ID!, $authorId: ID!, $title: String!) {
				updatePost(input: { where: { id: $postId }, values: { title: $title, authorId: $authorId } }) {
					id
					title
					author {
			          id
					  name
					}
				}
		 	}`,
		variables: map[string]any{
			"postId":   "post_1",
			"authorId": "author_2",
			"title":    "Updated To Weaveton Post",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "updatePost.title", "Updated To Weaveton Post")
			rtt.AssertValueAtPath(t, data, "updatePost.author.id", "author_2")
			rtt.AssertValueAtPath(t, data, "updatePost.author.name", "Weaveton")
		},
	},
	{
		name:       "update_relationship_with_parent_id_foreign_key",
		keelSchema: relationships,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			rows := []map[string]any{
				initRow(map[string]any{
					"id":           "publisher_1",
					"organisation": "Keelson Publishers",
				}),
				initRow(map[string]any{
					"id":           "publisher_2",
					"organisation": "Weaveton Publishers",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("publisher").Create(row).Error)
			}
			rows = []map[string]any{
				initRow(map[string]any{
					"id":          "author_1",
					"name":        "Keelson",
					"publisherId": "publisher_1",
				}),
				initRow(map[string]any{
					"id":          "author_2",
					"name":        "Weaveton",
					"publisherId": "publisher_2",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("author").Create(row).Error)
			}

			rows = []map[string]any{
				initRow(map[string]any{
					"id":       "post_1",
					"title":    "Keelson First Post",
					"authorId": "author_1",
				}),
			}
			for _, row := range rows {
				require.NoError(t, db.Table("blog_post").Create(row).Error)
			}
		},
		gqlOperation: `
				mutation UpdatePost($postId: ID!, $authorId: ID!, $title: String!) {
					updatePostWithFk(input: { where: { id: $postId }, values: { title: $title, authorId: $authorId } }) {
						id
						title
						author {
						  id
						  name
						}
					}
				 }`,
		variables: map[string]any{
			"postId":   "post_1",
			"authorId": "author_2",
			"title":    "Updated To Weaveton Post",
		},
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "updatePostWithFk.title", "Updated To Weaveton Post")
			rtt.AssertValueAtPath(t, data, "updatePostWithFk.author.id", "author_2")
			rtt.AssertValueAtPath(t, data, "updatePostWithFk.author.name", "Weaveton")
		},
	},
	{
		name:       "create_operation_with_date_and_timestamp_implicit_inputs",
		keelSchema: date_timestamp_parsing,
		gqlOperation: `
				mutation CreateThing {
					createThing(input: { 
						theDate: {
							year: 2022,
							month: 6,
							day: 17
						},
						theTimestamp: {
							seconds: 12345
						}
					}) {
						theDate {
							year
							month
							day
						}
						theTimestamp {
							seconds
						}
					}
				 }`,
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "createThing.theDate.year", float64(2022))
			rtt.AssertValueAtPath(t, data, "createThing.theDate.month", float64(6))
			rtt.AssertValueAtPath(t, data, "createThing.theDate.day", float64(17))
			rtt.AssertValueAtPath(t, data, "createThing.theTimestamp.seconds", float64(12345))
		},
	},
	{
		name:       "update_operation_with_date_and_timestamp_implicit_inputs",
		keelSchema: date_timestamp_parsing,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			row := initRow(map[string]any{
				"id":           "thing_1",
				"theDate":      "2022-06-17",
				"theTimestamp": "2022-01-01",
			})
			require.NoError(t, db.Table("thing").Create(row).Error)
		},
		gqlOperation: `
				mutation UpdateThing {
					updateThing(input: { 
						where: {
							id: "thing_1"
						}
						values: {
							theDate: {
								year: 2023,
								month: 7,
								day: 18
							},
							theTimestamp: {
								seconds: 54321
							}
						}
					}) {
						theDate {
							year
							month
							day
						}
						theTimestamp {
							seconds
						}
					}
				 }`,
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "updateThing.theDate.year", float64(2023))
			rtt.AssertValueAtPath(t, data, "updateThing.theDate.month", float64(7))
			rtt.AssertValueAtPath(t, data, "updateThing.theDate.day", float64(18))
			rtt.AssertValueAtPath(t, data, "updateThing.theTimestamp.seconds", float64(54321))
		},
	},
	{
		name:       "get_operation_with_date_and_timestamp_implicit_inputs",
		keelSchema: date_timestamp_parsing,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			row := initRow(map[string]any{
				"id":           "thing_1",
				"theDate":      "2022-06-17",
				"theTimestamp": "2022-01-01",
			})
			require.NoError(t, db.Table("thing").Create(row).Error)
		},
		gqlOperation: `
				query GetThing {
					getThing(input: { 
						id: "thing_1",
						theDate: {
							year: 2022,
							month: 6,
							day: 17
						},
						theTimestamp: {
							seconds: 1640995200
						}
					}) {
						theDate {
							year
							month
							day
						}
						theTimestamp {
							seconds
						}
					}
				 }`,
		assertData: func(t *testing.T, data map[string]any) {
			rtt.AssertValueAtPath(t, data, "getThing.theDate.year", float64(2022))
			rtt.AssertValueAtPath(t, data, "getThing.theDate.month", float64(6))
			rtt.AssertValueAtPath(t, data, "getThing.theDate.day", float64(17))
			rtt.AssertValueAtPath(t, data, "getThing.theTimestamp.seconds", float64(1640995200))
		},
	},
	{
		name:       "list_operation_with_date_and_timestamp_implicit_inputs",
		keelSchema: date_timestamp_parsing,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			row := initRow(map[string]any{
				"id":           "thing_1",
				"theDate":      "2022-06-17",
				"theTimestamp": "2022-01-01",
			})
			require.NoError(t, db.Table("thing").Create(row).Error)
		},
		gqlOperation: `
				query ListThing {
					listThing(input: { 
						where: {
							theDate: {
								equals: {
									year: 2022,
									month: 6,
									day: 17
								}
							},
                            theTimestamp: {
								before: {
                                    seconds: 1640995201
                                },
                                after: {
                                    seconds: 1640995199
                                }
                            }
						}
					}) {
						edges {
							node {
								theDate {
									year
									month
									day
								}
								theTimestamp {
									seconds
								}
							}
						}
					}
				 }`,
		assertData: func(t *testing.T, data map[string]any) {
			things := rtt.GetValueAtPath(t, data, "listThing.edges").([]any)
			require.Len(t, things, 1)

			thing := things[0].(map[string]any)
			rtt.AssertValueAtPath(t, thing, "node.theDate.year", float64(2022))
			rtt.AssertValueAtPath(t, thing, "node.theDate.month", float64(6))
			rtt.AssertValueAtPath(t, thing, "node.theDate.day", float64(17))
			rtt.AssertValueAtPath(t, thing, "node.theTimestamp.seconds", float64(1640995200))
		},
	},
}

var rpcTestCases = []rpcTestCase{
	{
		name: "rpc_list_http_get",
		keelSchema: `
		model Thing {
			operations {
				list listThings()
			}
			@permission(
				expression: true,
				actions: [list]
			)
		}
		api Test {
			models {
				Thing
			}
		}
	`,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			row1 := initRow(map[string]any{
				"id": "id_123",
			})
			require.NoError(t, db.Table("thing").Create(row1).Error)
		},
		Path:   "listThings",
		Method: http.MethodGet,
		assertResponse: func(t *testing.T, res map[string]any) {

			results := res["results"].([]interface{})
			require.Len(t, results, 1)

			hasNextPage := res["hasNextPage"].(bool)
			require.Equal(t, false, hasNextPage)
		},
	},
	{
		name: "rpc_list_http_post",
		keelSchema: `
		model Thing {
			fields {
				text Text
			}
			operations {
				list listThings(text)
			}
			@permission(
				expression: true,
				actions: [list]
			)
		}
		api Test {
			models {
				Thing
			}
		}
	`,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			row := initRow(map[string]any{
				"id":   "id_1",
				"text": "foobar",
			})
			require.NoError(t, db.Table("thing").Create(row).Error)
			row = initRow(map[string]any{
				"id":   "id_2",
				"text": "foobaz",
			})
			require.NoError(t, db.Table("thing").Create(row).Error)
			row = initRow(map[string]any{
				"id":   "id_3",
				"text": "boop",
			})
			require.NoError(t, db.Table("thing").Create(row).Error)
		},
		Path:   "listThings",
		Body:   `{"where": { "text": { "startsWith": "foo" } }}`,
		Method: http.MethodPost,
		assertResponse: func(t *testing.T, res map[string]any) {
			results := res["results"].([]interface{})
			require.Len(t, results, 2)

			hasNextPage := res["hasNextPage"].(bool)
			require.Equal(t, false, hasNextPage)
		},
	},
	{
		name: "rpc_get_http_get",
		keelSchema: `
		model Thing {
			operations {
				get getThing(id)
			}
			@permission(
				expression: true,
				actions: [get]
			)
		}
		api Test {
			models {
				Thing
			}
		}
	`,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			row := initRow(map[string]any{
				"id": "id_1",
			})
			require.NoError(t, db.Table("thing").Create(row).Error)
		},
		Path:        "getThing",
		QueryParams: "id=id_1",
		Method:      http.MethodGet,
		assertResponse: func(t *testing.T, data map[string]any) {
			require.Equal(t, data["id"], "id_1")
		},
	},
	{
		name: "rpc_get_http_post",
		keelSchema: `
		model Thing {
			operations {
				get getThing(id)
			}
			@permission(
				expression: true,
				actions: [get]
			)
		}
		api Test {
			models {
				Thing
			}
		}
	`,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			row := initRow(map[string]any{
				"id": "id_1",
			})
			require.NoError(t, db.Table("thing").Create(row).Error)
		},
		Path:   "getThing",
		Body:   `{"id": "id_1"}`,
		Method: http.MethodPost,
		assertResponse: func(t *testing.T, data map[string]any) {
			require.Equal(t, data["id"], "id_1")
		},
	},
	{
		name: "rpc_create_http_post",
		keelSchema: `
		model Thing {
			fields {
				text Text
			}
			operations {
				create createThing() with (text)
			}
			@permission(
				expression: true,
				actions: [create]
			)
		}
		api Test {
			models {
				Thing
			}
		}
	`,
		Path:   "createThing",
		Body:   `{"text": "foo"}`,
		Method: http.MethodPost,
		assertDatabase: func(t *testing.T, db *gorm.DB, data interface{}) {
			res := data.(map[string]any)
			id := res["id"]

			row := map[string]any{}
			err := db.Table("thing").Where("id = ?", id).Scan(&row).Error
			require.NoError(t, err)

			require.Equal(t, "foo", row["text"])
		},
	},
	{
		name: "rpc_update_http_post",
		keelSchema: `
		model Thing {
			fields {
				text Text
			}
			operations {
				update updateThing(id) with (text)
			}
			@permission(
				expression: true,
				actions: [update]
			)
		}
		api Test {
			models {
				Thing
			}
		}
	`,
		Path:   "updateThing",
		Body:   `{"where": {"id": "id_1"}, "values": {"text": "new value"}}`,
		Method: http.MethodPost,
		databaseSetup: func(t *testing.T, db *gorm.DB) {
			row := initRow(map[string]any{
				"id":   "id_1",
				"text": "foo",
			})
			require.NoError(t, db.Table("thing").Create(row).Error)
			row = initRow(map[string]any{
				"id":   "id_2",
				"text": "bar",
			})
			require.NoError(t, db.Table("thing").Create(row).Error)
		},
		assertDatabase: func(t *testing.T, db *gorm.DB, data interface{}) {
			res := data.(map[string]any)
			// check returned values
			require.Equal(t, "id_1", res["id"])
			require.Equal(t, "new value", res["text"])

			// check row 1 changed
			row := map[string]any{}
			err := db.Table("thing").Where("id = ?", "id_1").Scan(&row).Error
			require.NoError(t, err)
			require.Equal(t, "new value", row["text"])

			// check row 2 did not change
			row = map[string]any{}
			err = db.Table("thing").Where("id = ?", "id_2").Scan(&row).Error
			require.NoError(t, err)
			require.Equal(t, "bar", row["text"])
		},
	},
	{
		name: "rpc_json_schema_errors",
		keelSchema: `
		model Thing {
			operations {
				get getThing(id)
			}
		}
		api Test {
			models {
				Thing
			}
		}
	`,
		Path:   "getThing",
		Body:   `{"total": "nonsense"}`,
		Method: http.MethodPost,
		assertError: func(t *testing.T, data map[string]any) {
			assert.Equal(t, "ERR_INVALID_INPUT", data["code"])
			rtt.AssertValueAtPath(t, data, "data.errors[0].field", "(root)")
			rtt.AssertValueAtPath(t, data, "data.errors[0].error", "id is required")
			rtt.AssertValueAtPath(t, data, "data.errors[1].field", "(root)")
			rtt.AssertValueAtPath(t, data, "data.errors[1].error", "Additional property total is not allowed")
		},
	},
}
