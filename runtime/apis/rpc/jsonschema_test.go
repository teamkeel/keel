package rpc_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/apis/rpc"
	"github.com/teamkeel/keel/schema"
)

func TestValidateRequest(t *testing.T) {
	type fixture struct {
		name    string
		schema  string
		opName  string
		request string
		errors  map[string]string
	}

	fixtures := []fixture{
		{
			name: "get action missing input",
			schema: `
				model Person {
					operations {
						get getPerson(id)
					}
				}
			`,
			request: `
				{}
			`,
			opName: "getPerson",
			errors: map[string]string{
				"(root)": "id is required",
			},
		},
		{
			name: "get action type id wrong type",
			schema: `
				model Person {
					operations {
						get getPerson(id)
					}
				}
			`,
			request: `
				{
					"id": 1234
				}
			`,
			opName: "getPerson",
			errors: map[string]string{
				"id": "Invalid type. Expected: string, given: integer",
			},
		},
		{
			name: "get action type text wrong type",
			schema: `
				model Person {
					fields {
						name Text @unique
					}
					operations {
						get getPerson(name)
					}
				}
			`,
			request: `
				{
					"name": 1234
				}
			`,
			opName: "getPerson",
			errors: map[string]string{
				"name": "Invalid type. Expected: string, given: integer",
			},
		},
	}

	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {

			builder := schema.Builder{}
			schema, err := builder.MakeFromString(f.schema)
			require.NoError(t, err)

			var req map[string]any
			err = json.Unmarshal([]byte(f.request), &req)
			require.NoError(t, err)

			op := proto.FindOperation(schema, f.opName)

			result, err := rpc.ValidateRequest(context.Background(), schema, op, req)
			require.NoError(t, err)
			require.NotNil(t, result)

			for _, e := range result.Errors() {
				expected, ok := f.errors[e.Field()]
				if !ok {
					assert.Fail(t, "unexpected error", "%s - %s", e.Field(), e.Description())
					continue
				}

				assert.Equal(t, expected, e.Description(), "error for field %s did not match expected", e.Field())
				delete(f.errors, e.Field())
			}

			// f.errors should now be empty, if not mark test as failed
			for field, description := range f.errors {
				assert.Fail(t, "expected error was not returned", "%s - %s", field, description)
			}
		})
	}

}
