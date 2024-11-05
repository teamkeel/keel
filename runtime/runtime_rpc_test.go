package runtime_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/jsonschema"
	"github.com/teamkeel/keel/runtime/runtimectx"
	rtt "github.com/teamkeel/keel/runtime/runtimetest"
	"github.com/teamkeel/keel/storage"
	"github.com/teamkeel/keel/testhelpers"
	"gorm.io/gorm"
)

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

			handler := runtime.NewApiHandler(schema)

			request := &http.Request{
				URL: &url.URL{
					Path:     "/test/json/" + tCase.Path,
					RawQuery: tCase.QueryParams,
				},
				Method: tCase.Method,
				Body:   io.NopCloser(strings.NewReader(tCase.Body)),
				Header: tCase.Headers,
			}

			ctx := request.Context()

			ctx, err := testhelpers.WithTracing(ctx)
			require.NoError(t, err)

			pk, err := testhelpers.GetEmbeddedPrivateKey()
			require.NoError(t, err)

			ctx = runtimectx.WithPrivateKey(ctx, pk)

			dbName := testhelpers.DbNameForTestName(tCase.name)
			database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, dbName, true)
			require.NoError(t, err)
			defer database.Close()

			ctx = db.WithDatabase(ctx, database)

			storer, err := storage.NewDbStore(ctx, database)
			require.NoError(t, err)
			ctx = runtimectx.WithStorage(ctx, storer)

			request = request.WithContext(ctx)

			// Apply the database prior-set up mandated by this test case.
			if tCase.databaseSetup != nil {
				tCase.databaseSetup(t, database.GetDB())
			}

			// Call the handler, and capture the response.
			response := handler(request)
			body := string(response.Body)
			var res map[string]any
			require.NoError(t, json.Unmarshal([]byte(body), &res))

			// Do the specified assertion on the resultant database contents, if one is specified.
			if tCase.assertDatabase != nil {
				tCase.assertDatabase(t, database.GetDB(), res)
			}

			if response.Status != 200 && tCase.assertError == nil {
				t.Errorf("method %s returned non-200 (%d) but no assertError function provided", tCase.Path, response.Status)
			}
			if tCase.assertError != nil {
				tCase.assertError(t, res, response.Status)
			}

			// Do the specified assertion on the data returned, if one is specified.
			if tCase.assertResponse != nil {
				tCase.assertResponse(t, res)
			}

			action := schema.FindAction(tCase.Path)

			if response.Status == 200 {
				_, result, err := jsonschema.ValidateResponse(ctx, schema, action, res)
				assert.NoError(t, err)

				if !result.Valid() {
					msg := ""

					for _, err := range result.Errors() {
						msg += fmt.Sprintf("%s\n", err.String())
					}
					assert.Fail(t, msg)
				}
			}
		})
	}
}

var rpcTestCases = []rpcTestCase{
	{
		name: "json_invalid_token_missing_bearer",
		keelSchema: `
			model Thing {
				actions {
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
		Path:   "listThings",
		Method: http.MethodGet,
		Headers: map[string][]string{
			"Authorization": {"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIyVUtUZ1kyanY3S0dBSlpHdjJYdGlybnBRSlciLCJleHAiOjE2OTM0OTEyMjIsImlhdCI6MTY5MzQwNDgyMn0.C3DH-k8vcKoVNkJ2bWp5v84tpOu4KPyVEWtJMoE_4Ys"},
		},
		assertError: func(t *testing.T, data map[string]any, statusCode int) {
			assert.Equal(t, statusCode, http.StatusUnauthorized)
			assert.Equal(t, "ERR_AUTHENTICATION_FAILED", data["code"])
			assert.Equal(t, "no 'Bearer' prefix in the Authorization header", data["message"])
		},
	},
	{
		name: "json_invalid_token_not_jwt",
		keelSchema: `
			model Thing {
				actions {
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
		Path:   "listThings",
		Method: http.MethodGet,
		Headers: map[string][]string{
			"Authorization": {"Bearer invalid.token"},
		},
		assertError: func(t *testing.T, data map[string]any, statusCode int) {
			assert.Equal(t, statusCode, http.StatusUnauthorized)
			assert.Equal(t, "ERR_AUTHENTICATION_FAILED", data["code"])
			assert.Equal(t, "cannot be parsed or verified as a valid JWT", data["message"])
		},
	},
	{
		name: "json_invalid_token_not_authenticated",
		keelSchema: `
			model Thing {
				actions {
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
		Path:   "listThings",
		Method: http.MethodGet,
		Headers: map[string][]string{
			"Authorization": {"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIyVUtUZ1kyanY3S0dBSlpHdjJYdGlybnBRSlciLCJleHAiOjE2OTM0OTEyMjIsImlhdCI6MTY5MzQwNDgyMn0.C3DH-k8vcKoVNkJ2bWp5v84tpOu4KPyVEWtJMoE_4Ys"},
		},
		assertError: func(t *testing.T, data map[string]any, statusCode int) {
			assert.Equal(t, statusCode, http.StatusUnauthorized)
			assert.Equal(t, "ERR_AUTHENTICATION_FAILED", data["code"])
			assert.Equal(t, "cannot be parsed or verified as a valid JWT", data["message"])
		},
	},
	{
		name: "json_not_permitted",
		keelSchema: `
			model Thing {
				actions {
					list listThings()
				}
				@permission(
					expression: false,
					actions: [list]
				)
			}
			api Test {
				models {
					Thing
				}
			}
		`,
		Path:   "listThings",
		Method: http.MethodGet,
		assertError: func(t *testing.T, data map[string]any, statusCode int) {
			assert.Equal(t, statusCode, http.StatusForbidden)
			assert.Equal(t, "ERR_PERMISSION_DENIED", data["code"])
			assert.Equal(t, "not authorized to access this action", data["message"])
		},
	},
	{
		name: "rpc_list_http_get",
		keelSchema: `
			model Thing {
				actions {
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
			pageInfo := res["pageInfo"].(map[string]any)

			hasNextPage := pageInfo["hasNextPage"].(bool)
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
			actions {
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
			pageInfo := res["pageInfo"].(map[string]any)

			hasNextPage := pageInfo["hasNextPage"].(bool)
			require.Equal(t, false, hasNextPage)
		},
	},
	{
		name: "rpc_list_paging",
		keelSchema: `
		model Thing {
			fields {
				text Text
			}
			actions {
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
				"id":   "id_1",
				"text": "foobar",
			})
			require.NoError(t, db.Table("thing").Create(row1).Error)
			row2 := initRow(map[string]any{
				"id":   "id_2",
				"text": "foobaz",
			})
			require.NoError(t, db.Table("thing").Create(row2).Error)
			row3 := initRow(map[string]any{
				"id":   "id_3",
				"text": "boop",
			})
			require.NoError(t, db.Table("thing").Create(row3).Error)
			row4 := initRow(map[string]any{
				"id":   "id_4",
				"text": "boop",
			})
			require.NoError(t, db.Table("thing").Create(row4).Error)
		},
		Path:   "listThings",
		Body:   `{"where": { }, "first": 2}`,
		Method: http.MethodPost,
		assertResponse: func(t *testing.T, res map[string]any) {
			results := res["results"].([]interface{})
			require.Len(t, results, 2)

			pageInfo := res["pageInfo"].(map[string]any)

			hasNextPage := pageInfo["hasNextPage"].(bool)
			require.Equal(t, true, hasNextPage)

			assert.Equal(t, "id_2", pageInfo["endCursor"].(string))

			totalCount := pageInfo["totalCount"].(float64)
			assert.Equal(t, float64(4), totalCount)
		},
	},
	{
		name: "rpc_get_http_get",
		keelSchema: `
		model Thing {
			actions {
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
			actions {
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
				decimal Decimal
			}
			actions {
				create createThing() with (text, decimal)
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
		Body:   `{"text": "foo", "decimal": 1.3}`,
		Method: http.MethodPost,
		assertDatabase: func(t *testing.T, db *gorm.DB, data interface{}) {
			res := data.(map[string]any)
			id := res["id"]

			row := map[string]any{}
			err := db.Table("thing").Where("id = ?", id).Scan(&row).Error
			require.NoError(t, err)

			require.Equal(t, "foo", row["text"])
			require.Equal(t, 1.3, row["decimal"])
		},
	},
	{
		name: "rpc_update_http_post",
		keelSchema: `
		model Thing {
			fields {
				text Text
			}
			actions {
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
				actions {
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
		assertError: func(t *testing.T, data map[string]any, statusCode int) {
			assert.Equal(t, statusCode, http.StatusBadRequest)
			assert.Equal(t, "ERR_INVALID_INPUT", data["code"])
			rtt.AssertValueAtPath(t, data, "data.errors[0].field", "(root)")
			rtt.AssertValueAtPath(t, data, "data.errors[0].error", "id is required")
			rtt.AssertValueAtPath(t, data, "data.errors[1].field", "(root)")
			// TODO: gojsonschema doesnt support the latest json schema spec
			// and unevaluatedProperties isn't supported. we need to change out
			// this library to something which supports unevaluatedProperties
			//rtt.AssertValueAtPath(t, data, "data.errors[1].error", "Additional property total is not allowed")
		},
	},
	{
		name: "create_arrays",
		keelSchema: `
			model Thing {
				fields {
					texts Text[]
					bools Boolean[]
					numbers Number[]
					dates Date[]
					timestamps Timestamp[]
				}
				actions {
					create createThing() with (texts, bools, numbers, dates, timestamps) {
						@permission(expression: true)
					}
				}
			}
			api Test {
				models {
					Thing
				}
			}
		`,
		Path: "createThing",
		Body: `{
				"texts": ["science", "technology"],
				"bools": [true, true, false],
				"numbers": [1, 2, 3],
				"dates": ["2023-03-13T17:00:00.00+07:00", "2024-01-01T00:00:00.00Z"],
				"timestamps": ["2023-03-13T17:00:45+07:00", "2024-01-01T00:00:00.3Z"]
			}`,
		Method: http.MethodPost,
		assertError: func(t *testing.T, data map[string]any, statusCode int) {
			assert.Equal(t, statusCode, http.StatusOK)
			rtt.AssertValueAtPath(t, data, "texts", []any{"science", "technology"})
			rtt.AssertValueAtPath(t, data, "bools", []any{true, true, false})
			rtt.AssertValueAtPath(t, data, "numbers", []any{1.0, 2.0, 3.0})
			rtt.AssertValueAtPath(t, data, "dates", []any{"2023-03-13T00:00:00Z", "2024-01-01T00:00:00Z"})
			rtt.AssertValueAtPath(t, data, "timestamps", []any{"2023-03-13T10:00:45Z", "2024-01-01T00:00:00.3Z"})
		},
	},
}
