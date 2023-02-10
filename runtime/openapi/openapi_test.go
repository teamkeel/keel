package openapi_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/runtime/openapi"
	"github.com/teamkeel/keel/schema"
)

func TestGeneration(t *testing.T) {

	keelSchema, err := os.ReadFile("./testdata/schema.keel")
	require.NoError(t, err)
	expected, err := os.ReadFile("./testdata/openapi.json")
	require.NoError(t, err)

	builder := schema.Builder{}
	schema, err := builder.MakeFromString(string(keelSchema))
	require.NoError(t, err)

	jsonSchema := openapi.Generate(context.Background(), schema, schema.Apis[0])
	actual, err := json.Marshal(jsonSchema)
	require.NoError(t, err)

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(expected, actual, &opts)

	if diff != jsondiff.FullMatch {
		t.Errorf("actual JSON schema does not match expected: %s", explanation)
	}
}
