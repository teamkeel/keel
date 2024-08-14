package openapi_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/runtime/openapi"
	"github.com/teamkeel/keel/schema"
)

func TestGeneration(t *testing.T) {
	entries, err := os.ReadDir("./testdata")
	require.NoError(t, err)

	type testCase struct {
		keelSchema string
		jsonSchema string
	}

	cases := map[string]*testCase{}

	for _, e := range entries {
		caseName := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		c, ok := cases[caseName]
		if !ok {
			c = &testCase{}
			cases[caseName] = c
		}
		b, err := os.ReadFile(filepath.Join("./testdata", e.Name()))
		require.NoError(t, err)
		switch filepath.Ext(e.Name()) {
		case ".keel":
			c.keelSchema = string(b)
		case ".json":
			c.jsonSchema = string(b)
		}
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			builder := schema.Builder{}
			schema, err := builder.MakeFromString(c.keelSchema, config.Empty)
			require.NoError(t, err)

			if len(schema.Jobs) == 0 {
				jsonSchema := openapi.Generate(context.Background(), schema, schema.Apis[0])
				actual, err := json.Marshal(jsonSchema)
				require.NoError(t, err)

				opts := jsondiff.DefaultConsoleOptions()
				diff, explanation := jsondiff.Compare([]byte(c.jsonSchema), actual, &opts)

				if diff != jsondiff.FullMatch {
					t.Error(string(actual))
					t.Errorf("actual JSON schema does not match expected: %s", explanation)
				}
			}

			if len(schema.Jobs) > 0 {
				jsonSchema := openapi.GenerateJob(context.Background(), schema, schema.Jobs[0].Name)
				actual, err := json.Marshal(jsonSchema)
				require.NoError(t, err)

				opts := jsondiff.DefaultConsoleOptions()
				diff, explanation := jsondiff.Compare([]byte(c.jsonSchema), actual, &opts)

				if diff != jsondiff.FullMatch {
					t.Error(string(actual))
					t.Errorf("actual JSON schema does not match expected: %s", explanation)
				}
			}
		})
	}
}
