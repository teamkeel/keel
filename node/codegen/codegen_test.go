package codegenerator_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

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

	tmpDir, err := os.MkdirTemp("", "testing")

	require.NoError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	generator := codegenerator.NewGenerator(proto, tmpDir)

	err = generator.GenerateTesting()

	require.NoError(t, err)

	comparePackageFiles(t, "testing", tmpDir)
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
