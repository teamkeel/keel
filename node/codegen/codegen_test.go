package codegenerator_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/codegen"
	codegenerator "github.com/teamkeel/keel/node/codegen"
	"github.com/teamkeel/keel/schema"
)

func TestSDKGeneration(t *testing.T) {
	builder := schema.Builder{}

	b, err := os.ReadFile(filepath.Join("testdata", "schema.keel"))

	require.NoError(t, err)

	proto, err := builder.MakeFromString(string(b))

	require.NoError(t, err)

	tmpDir, err := os.MkdirTemp("", "sdk")

	fmt.Print(tmpDir)
	require.NoError(t, err)

	// t.Cleanup(func() {
	// 	os.RemoveAll(tmpDir)
	// })

	generator := codegenerator.NewGenerator(proto, tmpDir)

	err = generator.GenerateSDK()

	require.NoError(t, err)

	comparePackageFiles(t, "sdk", tmpDir)
}

func TestTestingGeneration(t *testing.T) {
	builder := schema.Builder{}

	b, err := os.ReadFile(filepath.Join("testdata", "schema.keel"))

	require.NoError(t, err)

	proto, err := builder.MakeFromString(string(b))

	require.NoError(t, err)

	tmpDir, err := os.MkdirTemp("", "sdk")

	require.NoError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	generator := codegenerator.NewGenerator(proto, tmpDir)

	err = generator.GenerateTesting()

	require.NoError(t, err)
}

func comparePackageFiles(t *testing.T, packageName string, tmpDir string) {
	allExpectedFiles := []string{
		"package.json",
		"dist/index.js",
		"dist/index.d.ts",
	}

	for _, f := range allExpectedFiles {
		expectedContents, err := os.ReadFile(filepath.Join("testdata", "artifacts", "sdk", f))

		require.NoError(t, err)

		actualContents, err := os.ReadFile(filepath.Join(tmpDir, "node_modules", "@teamkeel", packageName, f))

		require.NoError(t, err)

		assert.Equal(t, string(actualContents), string(expectedContents))
	}
}

func TestGenerateEntryPointRenderArguments(t *testing.T) {
	testSchema := `
model Post {
  fields {
    title Text
  }

  functions {
    create createPost() with(title)
  }
}
`

	type TestCase struct {
		PathToFunctionsDirArg string
		ExpectedImportPrefix  string
	}

	testCases := []TestCase{
		{
			PathToFunctionsDirArg: ".",
			ExpectedImportPrefix:  "./",
		},
		{
			PathToFunctionsDirArg: "functions",
			ExpectedImportPrefix:  "./functions/",
		},
		{
			PathToFunctionsDirArg: "..",
			ExpectedImportPrefix:  "../",
		},
		{
			PathToFunctionsDirArg: "../../../../functions",
			ExpectedImportPrefix:  "../../../../functions/",
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Input_'%s'", testCase.PathToFunctionsDirArg), func(t *testing.T) {
			scm := schema.Builder{}

			proto, err := scm.MakeFromString(testSchema)

			require.NoError(t, err)

			generator := codegen.NewGenerator(proto)

			renderArguments := generator.GenerateEntryPointRenderArguments(testCase.PathToFunctionsDirArg)

			expectedImports := fmt.Sprintf(`import startRuntimeServer from '@teamkeel/runtime'
import { queryResolverFromEnv } from '@teamkeel/sdk'
import { Logger } from '@teamkeel/sdk'
import { PostApi } from '@teamkeel/sdk'
import createPost from '%screatePost'
import { IdentityApi } from '@teamkeel/sdk'
`, testCase.ExpectedImportPrefix)
			expectedAPI := `models: { post: new PostApi(),
identity: new IdentityApi() },
logger: new Logger({ colorize: true })`
			expectedFunctions := "createPost: { call: createPost, },"

			assert.Equal(t, expectedImports, renderArguments.Imports)
			assert.Equal(t, expectedAPI, renderArguments.API)
			assert.Equal(t, expectedFunctions, renderArguments.Functions)
		})
	}
}
