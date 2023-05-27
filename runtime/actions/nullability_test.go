package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"

	"github.com/teamkeel/keel/schema"
)

type testCase struct {
	// Name given to the test case
	name string
	// Valid keel schema for this test case
	keelSchema string
	// Operation name to run test upon
	operationName string
	// Input map for operation
	inputs string
	// Expected SQL template generated (with ? placeholders for values)
	rewrittenInputs string
}

var testCases = []testCase{
	{
		name: "create_op_nullable_field_to_value",
		keelSchema: `
			model Person {
				fields {
					name Text
					nickName Text?
				}
				operations {
					create createPerson() with (name, nickName)
				}
			}`,
		operationName: "createPerson",
		inputs: `
			{
				"name": "Biggy Smalls",
				"nickName": { "value": "Biggy" }
			}`,
		rewrittenInputs: `
			{
				"name": "Biggy Smalls",
				"nickName": "Biggy"
			}`,
	},
	{
		name: "create_op_nullable_field_to_null",
		keelSchema: `
			model Person {
				fields {
					name Text
					nickName Text?
				}
				operations {
					create createPerson() with (name, nickName)
				}
			}`,
		operationName: "createPerson",
		inputs: `
			{
				"name": "Biggy Smalls",
				"nickName": { "isNull": true }
			}`,
		rewrittenInputs: `
			{
				"name": "Biggy Smalls",
				"nickName": null
			}`,
	},
	{
		name: "create_op_nullable_relationship_to_null",
		keelSchema: `
			model Person {
				fields {
					employer Company?
				}
				operations {
					create createPerson() with (employer.id)
				}
			}
			model Company {
				fields {
					name Text
				}
			}`,
		operationName: "createPerson",
		inputs: `
			{
				"employer": {
					"isNull": true
				}
			}`,
		rewrittenInputs: `
			{
				"employer": null
			}`,
	},
	{
		name: "create_op_nullable_relationship_to_existing_id",
		keelSchema: `
			model Person {
				fields {
					employer Company?
				}
				operations {
					create createPerson() with (employer.id)
				}
			}
			model Company {
				fields {
					name Text
				}
			}`,
		operationName: "createPerson",
		inputs: `
			{
				"employer": {
					"value": {
						"id": "2QBXuqlHgJlTAFICwaAK8eeNM7c"
					}
				}
			}`,
		rewrittenInputs: `
			{
				"employer": {
					"id": "2QBXuqlHgJlTAFICwaAK8eeNM7c"
				}
			}`,
	},
	{
		name: "create_op_required_relationship_optional_field_to_null",
		keelSchema: `
			model Person {
				fields {
					employer Company
				}
				operations {
					create createPerson() with (employer.name)
				}
			}
			model Company {
				fields {
					name Text?
				}
			}`,
		operationName: "createPerson",
		inputs: `
			{
				"employer": {
					"name": {
						"isNull": true
					}
				}
			}`,
		rewrittenInputs: `
			{
				"employer": {
					"name": null
				}
			}`,
	},
	{
		name: "create_op_required_relationship_optional_field_to_value",
		keelSchema: `
			model Person {
				fields {
					employer Company
				}
				operations {
					create createPerson() with (employer.name)
				}
			}
			model Company {
				fields {
					name Text?
				}
			}`,
		operationName: "createPerson",
		inputs: `
			{
				"employer": {
					"name": {
						"value": "Big Company"
					}
				}
			}`,
		rewrittenInputs: `
			{
				"employer": {
					"name": "Big Company"
				}
			}`,
	},
	{
		name: "create_op_optional_relationship_optional_field_to_null",
		keelSchema: `
			model Person {
				fields {
					employer Company?
				}
				operations {
					create createPerson() with (employer.name)
				}
			}
			model Company {
				fields {
					name Text?
				}
			}`,
		operationName: "createPerson",
		inputs: `
			{
				"employer": {
					"value": {
						"name": {
							"isNull": true
						}
					}
				}
			}`,
		rewrittenInputs: `
			{
				"employer": {
					"name": null
				}
			}`,
	},
	{
		name: "create_op_optional_relationship_optional_field_to_value",
		keelSchema: `
			model Person {
				fields {
					employer Company?
				}
				operations {
					create createPerson() with (employer.name)
				}
			}
			model Company {
				fields {
					name Text?
				}
			}`,
		operationName: "createPerson",
		inputs: `
			{
				"employer": {
					"value": {
						"name": {
							"value": "Big Company"
						}
					}
				}
			}`,
		rewrittenInputs: `
			{
				"employer": {
					"name": "Big Company"
				}
			}`,
	},
	{
		name: "update_op_nullable_field_to_value",
		keelSchema: `
			model Person {
				fields {
					nickName Text?
				}
				operations {
					update updatePerson(id) with (nickName)
				}
			}`,
		operationName: "updatePerson",
		inputs: `
			{
				"values": {
					"nickName": { "value": "Biggy" }
				}
			}`,
		rewrittenInputs: `
			{
				"values": {
					"nickName": "Biggy"
				}
			}`,
	},
	{
		name: "update_op_nullable_field_to_null",
		keelSchema: `
			model Person {
				fields {
					nickName Text?
				}
				operations {
					update updatePerson(id) with (nickName)
				}
			}`,
		operationName: "updatePerson",
		inputs: `
			{
				"values": {
					"nickName": { "isNull": true }
				}
			}`,
		rewrittenInputs: `
			{
				"values": {
					"nickName": null
				}
			}`,
	},
	{
		name: "update_op_nullable_relationship_to_null",
		keelSchema: `
			model Person {
				fields {
					employer Company?
				}
				operations {
					update updatePerson(id) with (employer.id)
				}
			}
			model Company {
				fields {
					name Text
				}
			}`,
		operationName: "updatePerson",
		inputs: `
			{
				"values": {
					"employer": {
						"isNull": true
					}
				}
			}`,
		rewrittenInputs: `
			{
				"values": {
					"employer": null
				}
			}`,
	},
	{
		name: "update_op_nullable_relationship_to_existing_id",
		keelSchema: `
			model Person {
				fields {
					employer Company?
				}
				operations {
					update updatePerson(id) with (employer.id)
				}
			}
			model Company {
				fields {
					name Text
				}
			}`,
		operationName: "updatePerson",
		inputs: `
			{
				"values": {
					"employer": {
						"value": {
							"id": "2QBXuqlHgJlTAFICwaAK8eeNM7c"
						}
					}
				}
			}`,
		rewrittenInputs: `
			{
				"values": {
					"employer": {
						"id": "2QBXuqlHgJlTAFICwaAK8eeNM7c"
					}
				}
			}`,
	},
}

func TestNullabilityUnwrapping(t *testing.T) {
	for _, testCase := range testCases {
		// if testCase.name != "create_op_optional_relationship_optional_field_to_null" {
		// 	continue
		// }
		t.Run(testCase.name, func(t *testing.T) {
			scope, err := generateScope(context.Background(), testCase.keelSchema, testCase.operationName)
			if err != nil {
				require.NoError(t, err)
			}

			inputsAsMap := map[string]any{}

			err = json.Unmarshal([]byte(testCase.inputs), &inputsAsMap)
			if err != nil {
				require.NoError(t, err)
			}

			err = rewriteNullableInputs(scope, inputsAsMap)
			if err != nil {
				require.NoError(t, err)
			}

			json, err := json.Marshal(inputsAsMap)
			if err != nil {
				require.NoError(t, err)
			}

			opts := jsondiff.DefaultConsoleOptions()
			diff, explanation := jsondiff.Compare([]byte(testCase.rewrittenInputs), json, &opts)

			if diff != jsondiff.FullMatch {
				t.Errorf("rewritten inputs do not match expected: %s", explanation)
				fmt.Println(string(json))
			}

		})
	}
}

func generateScope(ctx context.Context, schemaText string, operationName string) (*Scope, error) {
	builder := &schema.Builder{}
	schema, err := builder.MakeFromString(schemaText)
	if err != nil {
		return nil, err
	}

	operation := proto.FindOperation(schema, operationName)
	if operation == nil {
		return nil, fmt.Errorf("operation not found in schema: %s", operationName)
	}

	scope := NewScope(ctx, operation, schema)

	return scope, nil
}
