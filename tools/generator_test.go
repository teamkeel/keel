package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/rpc/rpc"
	"github.com/teamkeel/keel/schema"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestWriteTableInterface(t *testing.T) {
	t.Parallel()

	schema := `
model Thing {
	fields {
		name Text
	}
	actions {
		create createThing() with (name)
		update updateThing(id) with (name)
	}
}`

	expected := `
{"id":"createThing", "name":"Create thing", "actionName":"createThing", "apiNames":["Api"], "modelName":"Thing", "actionType":"ACTION_TYPE_CREATE", "implementation":"ACTION_IMPLEMENTATION_AUTO", "inputs":[{"fieldLocation":{"path":"$.name"}, "fieldType":"TYPE_STRING", "displayName":"Name", "visible":true}], "response":[{"fieldLocation":{"path":"$.name"}, "fieldType":"TYPE_STRING", "displayName":"Name", "visible":true}, {"fieldLocation":{"path":"$.id"}, "fieldType":"TYPE_ID", "displayName":"Id", "displayOrder":2, "visible":true}, {"fieldLocation":{"path":"$.createdAt"}, "fieldType":"TYPE_DATETIME", "displayName":"Created at", "displayOrder":3, "visible":true}, {"fieldLocation":{"path":"$.updatedAt"}, "fieldType":"TYPE_DATETIME", "displayName":"Updated at", "displayOrder":4, "visible":true}], "title":{"template":"Create thing"}, "entitySingle":"thing", "entityPlural":"things", "capabilities":{}},
{
	"id":"updateThing", 
	"name":"Update thing", 
	"actionName":"updateThing", 
	"apiNames":["Api"], 
	"modelName":"Thing", 
	"actionType":"ACTION_TYPE_UPDATE", 
	"implementation":"ACTION_IMPLEMENTATION_AUTO", 
	"inputs":[
		{"fieldLocation":{"path":"$.where"}, "fieldType":"TYPE_MESSAGE", "displayName":"Where", "visible":true}, 
		{"fieldLocation":{"path":"$.where.id"}, "fieldType":"TYPE_ID", "displayName":"Id", "visible":true}, 
		{"fieldLocation":{"path":"$.values"}, "fieldType":"TYPE_MESSAGE", "displayName":"Values", "displayOrder":1, "visible":true}, 
		{"fieldLocation":{"path":"$.values.name"}, "fieldType":"TYPE_STRING", "displayName":"Name", "visible":true}
	], 
	"response":[
		{"fieldLocation":{"path":"$.name"}, "fieldType":"TYPE_STRING", "displayName":"Name", "visible":true}, 
		{"fieldLocation":{"path":"$.id"}, "fieldType":"TYPE_ID", "displayName":"Id", "displayOrder":2, "visible":true}, 
		{"fieldLocation":{"path":"$.createdAt"}, "fieldType":"TYPE_DATETIME", "displayName":"Created at", "displayOrder":3, "visible":true}, 
		{"fieldLocation":{"path":"$.updatedAt"}, "fieldType":"TYPE_DATETIME", "displayName":"Updated at", "displayOrder":4, "visible":true}
	], 
	"title":{"template":"Update thing"}, 
	"entitySingle":"thing", 
	"entityPlural":"things", 
	"capabilities":{}
}
`
	runGeneratorTest(t, schema, expected)
}

func normalise(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), "\t", "    ")
}

func runGeneratorTest(t *testing.T, schemaString string, expected string) {
	b := schema.Builder{}
	s, err := b.MakeFromString(schemaString, config.Empty)
	require.NoError(t, err)

	gen, err := NewGenerator(s)
	require.NoError(t, err)

	err = gen.Generate(context.Background())
	require.NoError(t, err)

	response := &rpc.ListToolsResponse{
		Tools: gen.GetConfigs(),
	}

	actual, err := protojson.Marshal(response)
	require.NoError(t, err)

	json := string(actual)

	diff := diffmatchpatch.New()
	diffs := diff.DiffMain(normalise(expected), normalise(json), true)
	if !strings.Contains(normalise(json), normalise(expected)) {
		t.Errorf("generated code does not match expected:\n%s", diffs)
		t.Errorf("\nExpected:\n---------\n%s", normalise(expected))
		t.Errorf("\nActual:\n---------\n%s", normalise(json))
	}
}
