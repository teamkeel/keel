package runtime

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
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	rtt "github.com/teamkeel/keel/runtime/runtimetest"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/testhelpers"
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

			testDB, err := testhelpers.SetupDatabaseForTestCase(schema, testhelpers.DbNameForTestName(tCase.name))

			require.NoError(t, err)

			// Construct the runtime API Handler.
			handler := NewHandler(schema)

			reqBody := queryAsJSONPayload(t, tCase.gqlOperation, tCase.variables)

			request := &http.Request{
				URL: &url.URL{
					Path: "/Test",
				},
				Method: http.MethodPost,
				Body:   io.NopCloser(strings.NewReader(reqBody)),
			}

			ctx := request.Context()
			ctx = runtimectx.WithDatabase(ctx, testDB)
			request = request.WithContext(ctx)

			// Apply the database prior-set up mandated by this test case.
			if tCase.databaseSetup != nil {
				tCase.databaseSetup(t, testDB)
			}

			// Call the handler, and capture the response.
			response, err := handler(request)
			require.NoError(t, err)
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
				tCase.assertDatabase(t, testDB, bodyFields.Data)
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

			testDB, err := testhelpers.SetupDatabaseForTestCase(schema, testhelpers.DbNameForTestName(tCase.name))

			require.NoError(t, err)

			handler := NewHandler(schema)

			request := &http.Request{
				URL: &url.URL{
					Path:     "/Test/" + tCase.Path,
					RawQuery: tCase.QueryParams,
				},
				Method: tCase.Method,
				Body:   io.NopCloser(strings.NewReader(tCase.Body)),
			}

			ctx := request.Context()
			ctx = runtimectx.WithDatabase(ctx, testDB)
			request = request.WithContext(ctx)

			// Apply the database prior-set up mandated by this test case.
			if tCase.databaseSetup != nil {
				tCase.databaseSetup(t, testDB)
			}

			// Call the handler, and capture the response.
			response, err := handler(request)
			require.NoError(t, err)
			body := string(response.Body)
			var res interface{}
			require.NoError(t, json.Unmarshal([]byte(body), &res))

			// Do the specified assertion on the data returned, if one is specified.
			if tCase.assertResponse != nil {
				tCase.assertResponse(t, res)
			}

			// Do the specified assertion on the resultant database contents, if one is specified.
			if tCase.assertDatabase != nil {
				tCase.assertDatabase(t, testDB, res)
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
	assertResponse func(t *testing.T, data interface{})
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
		fields {
			name Text 
		}
		operations {
			get getPerson(id)
			create createPerson() with (name)
			list listPeople(name)
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

const listImplicitAndExplicitInputs string = `
	model Person {
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

// Schema with all field types
const fieldTypes string = `
	model Thing {
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
		@graphql
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
	// {
	// 	name:       "list_inputs",
	// 	keelSchema: fieldTypes,
	// 	databaseSetup: func(t *testing.T, db *gorm.DB) {
	// 		row1 := initRow(map[string]any{
	// 			"id":        "id_123",
	// 			"text":      "some-interesting-text",
	// 			"bool":      true,
	// 			"timestamp": "1970-01-01 00:00:10",
	// 			"date":      "2020-01-02",
	// 			"number":    10,
	// 			"enum":      "Option1",
	// 		})
	// 		require.NoError(t, db.Table("thing").Create(row1).Error)
	// 		// require.NoError(t, db.Table("person").Create(row2).Error)
	// 		// require.NoError(t, db.Table("person").Create(row3).Error)
	// 	},
	// 	gqlOperation: `

	// 	fragment Fields on ThingConnection {
	// 		edges {
	// 			node {
	// 			text
	// 			bool
	// 			# timestamp {seconds}
	// 			# date {day, month, year}
	// 			number
	// 			enum
	// 			}
	// 		}
	// 	}

	// 	{
	// 	string_equals: listThings(input: {where: {text: {equals: "some-interesting-text"}}}) {
	// 		...Fields
	// 	},
	// 	string_startsWith: listThings(input: {where: {text: {startsWith: "some"}}}) {
	// 		...Fields
	// 	},
	// 	string_endWith: listThings(input: {where: {text: {endsWith: "-text"}}}) {
	// 		...Fields
	// 	},
	// 	string_contains: listThings(input: {where: {text: {contains: "interesting"}}}) {
	// 		...Fields
	// 	},
	// 	string_oneOf: listThings(input: {where: {text: {oneOf: ["some-interesting-text", "Another"]}}}) {
	// 		...Fields
	// 	},
	// 	number_equals: listThings(input: {where: {number: {equals: 10}}}) {
	// 		...Fields
	// 	},
	// 	number_gt: listThings(input: {where: {number: {greaterThan: 9}}}) {
	// 		...Fields
	// 	},
	// 	number_gte: listThings(input: {where: {number: {greaterThanOrEquals: 10}}}) {
	// 		...Fields
	// 	},
	// 	number_lt: listThings(input: {where: {number: {lessThan: 11}}}) {
	// 		...Fields
	// 	},
	// 	number_lte: listThings(input: {where: {number: {lessThanOrEquals: 10}}}) {
	// 		...Fields
	// 	},
	// 	enum_equals: listThings(input: {where: {enum: {equals: Option1}}}) {
	// 		...Fields
	// 	},
	// 	enum_oneOf: listThings(input: {where: {enum: {oneOf: [Option1]}}}) {
	// 		...Fields
	// 	},
	// 	timestamp_before: listThings(input: {
	// 		where: {
	// 		timestamp: {
	// 			before: {
	// 				seconds: 11
	// 			}
	// 		}
	// 		}
	// 	}) {
	// 		...Fields
	// 	},
	// 	timestamp_after: listThings(input: {
	// 		where: {
	// 		timestamp: {
	// 			after: {
	// 				seconds: 9
	// 			}
	// 		}
	// 		}
	// 	}) {
	// 		...Fields
	// 	},
	// 	date_before: listThings(input: {where: {date: {before: {
	// 		year: 2020,
	// 		month: 1,
	// 		day: 3
	// 	}}}}) {
	// 		...Fields
	// 	},
	// 	date_after: listThings(input: {where: {date: {after: {
	// 		year: 2020,
	// 		month: 1,
	// 		day: 1
	// 	}}}}) {
	// 		...Fields
	// 	},
	// 	date_onOrbefore: listThings(input: {where: {date: {onOrBefore: {
	// 		year: 2020,
	// 		month: 1,
	// 		day: 2
	// 	}}}}) {
	// 		...Fields
	// 	},
	// 	date_onOrAfter: listThings(input: {where: {date: {onOrAfter: {
	// 		year: 2020,
	// 		month: 1,
	// 		day: 2
	// 	}}}}) {
	// 		...Fields
	// 	},
	// 	date_onOrEquals: listThings(input: {where: {date: {equals: {
	// 		year: 2020,
	// 		month: 1,
	// 		day: 2
	// 	}}}}) {
	// 		...Fields
	// 	},
	// 	bool: listThings(input: {
	// 		where: {
	// 		bool: {
	// 				equals: true
	// 			}
	// 		}
	// 	}) {
	// 		...Fields
	// 	}
	// 	combined: listThings(input: {
	// 		where: {
	// 		bool: {
	// 				equals: true
	// 		},
	// 		enum: {
	// 			equals: Option1
	// 		}
	// 		}
	// 	}) {
	// 		...Fields
	// 	}
	// 	}`,
	// 	assertData: func(t *testing.T, data map[string]any) {

	// 		keys := []string{
	// 			"string_equals",
	// 			"string_startsWith",
	// 			"string_endWith",
	// 			"string_contains",
	// 			"string_oneOf",
	// 			"number_equals",
	// 			"number_gt",
	// 			"number_gte",
	// 			"number_lt",
	// 			"number_lte",
	// 			"enum_equals",
	// 			"enum_oneOf",
	// 			"timestamp_before",
	// 			"timestamp_after",
	// 			"date_before",
	// 			"date_after",
	// 			"date_onOrbefore",
	// 			"date_onOrAfter",
	// 			"date_onOrEquals",
	// 			"bool",
	// 			"combined",
	// 		}

	// 		for _, key := range keys {
	// 			edges := rtt.GetValueAtPath(t, data, key+".edges")
	// 			edgesList, ok := edges.([]any)
	// 			fmt.Println(key)
	// 			require.True(t, ok)
	// 			if len(edgesList) != 1 {
	// 				a := 1
	// 				_ = a
	// 			}
	// 			require.Len(t, edgesList, 1)
	// 		}
	// 	},
	// },
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
				@graphql
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
				@graphql
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
				@graphql
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
				fields {
					name Text
				}
				operations {
					get getPerson(id)
				}
			}
			api Test {
				@graphql
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
				fields {
					name Text
				}
				operations {
					get getPerson(id)
				}
			}
			api Test {
				@graphql
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
			@graphql
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
}

var rpcTestCases = []rpcTestCase{
	{
		name: "rpc_list",
		keelSchema: `
		model Thing {
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
			@rpc
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
		Path:   "listThings",
		Body:   "",
		Method: http.MethodGet,
		assertResponse: func(t *testing.T, data interface{}) {
			res := data.([]interface{})
			require.Len(t, res, 1)
		},
	},
	{
		name: "rpc_get",
		keelSchema: `
		model Thing {
			fields {
				text Text @unique
				bool Boolean
				timestamp Timestamp
				date Date
				number Number
			}
			operations {
				get getThing(id)
			}
		}
		api Test {
			@rpc
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
		Path:        "getThing",
		QueryParams: "id=id_123",
		Body:        "",
		Method:      http.MethodGet,
		assertResponse: func(t *testing.T, data interface{}) {
			res := data.(map[string]any)
			require.Equal(t, res["id"], "id_123")
		},
	},
}
