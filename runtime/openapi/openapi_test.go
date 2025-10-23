package openapi_test

import (
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

// There are multiple types of openAPI schemas:
// * API schema - the openAPI schema generated for the runtime API
// * Job schemas - generated for each job defined in the keel schema
// * Flow schema - generated for the Flows API for all the flows defined in your schema
// * Task schema - generated for the Tasks API for tasks defined in your schema
//
// Test cases will detect what type of schema is asserted by looking at the prefix of the files:
//   - if the file starts with "job", the schema for the first job defined will be tested
//   - if the file starts with "flow", the schema for all the flows will be tested
//   - if the file starts with "task", the schema for the tasks API will be tested
//   - then default to the keel api schema
func TestGeneration(t *testing.T) {
	entries, err := os.ReadDir("./testdata")
	require.NoError(t, err)

	type schemaType string
	const (
		jobType  schemaType = "job"
		apiType  schemaType = "api"
		flowType schemaType = "flow"
		taskType schemaType = "task"
	)

	type testCase struct {
		keelSchema string
		jsonSchema string
		schemaType schemaType // job, flow, task or keel
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

		switch {
		case strings.HasPrefix(caseName, "job"):
			c.schemaType = jobType
		case strings.HasPrefix(caseName, "flow"):
			c.schemaType = flowType
		case strings.HasPrefix(caseName, "task"):
			c.schemaType = taskType
		default:
			c.schemaType = apiType
		}
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			builder := schema.Builder{}
			schema, err := builder.MakeFromString(c.keelSchema, config.Empty)
			require.NoError(t, err)

			jsonSchema := openapi.OpenAPI{}
			switch c.schemaType {
			case apiType:
				jsonSchema = openapi.Generate(t.Context(), schema, schema.GetApis()[0])
			case jobType:
				jsonSchema = openapi.GenerateJob(t.Context(), schema, schema.GetJobs()[0].GetName())
			case flowType:
				jsonSchema = openapi.GenerateFlows(t.Context(), schema)
			case taskType:
				jsonSchema = openapi.GenerateTasks(t.Context(), schema)
			}

			actual, err := json.Marshal(jsonSchema)
			require.NoError(t, err)

			opts := jsondiff.DefaultConsoleOptions()
			diff, explanation := jsondiff.Compare([]byte(c.jsonSchema), actual, &opts)

			if diff != jsondiff.FullMatch {
				t.Error(string(actual))
				t.Errorf("actual JSON schema does not match expected: %s", explanation)
			}
		})
	}
}
