package node_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/runtime"
)

func TestBootstrap(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(`
		model Post {
			functions {
				create createPost()
			}
		}
	`), 0777)
	require.NoError(t, err)

	wd, err := os.Getwd()
	require.NoError(t, err)

	err = node.Bootstrap(tmpDir, node.WithPackagesPath(filepath.Join(wd, "../packages")))
	require.NoError(t, err)

	entries, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)

	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	assert.Equal(t, []string{"node_modules", "package-lock.json", "package.json", "schema.keel", "tsconfig.json"}, names)
}

func TestBootstrapVersionInterpolation(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(`
		model Post {
			functions {
				create createPost()
			}
		}
	`), 0777)
	require.NoError(t, err)

	// the current version at time of writing this test
	// we need to set the runtime.Version to a version that actually
	// exists on NPM in order for this test to succeed
	// It doesn't matter what the specific version is for the purposes of the test
	testVersion := "0.322.0"

	runtime.Version = testVersion

	err = node.Bootstrap(tmpDir)
	require.NoError(t, err)

	packageJsonContents, err := os.ReadFile(filepath.Join(tmpDir, "package.json"))
	assert.NoError(t, err)

	m := map[string]any{}

	err = json.Unmarshal(packageJsonContents, &m)
	assert.NoError(t, err)

	dependencies := m["dependencies"].(map[string]interface{})

	assert.Equal(t, testVersion, dependencies["@teamkeel/functions-runtime"])
	assert.Equal(t, testVersion, dependencies["@teamkeel/testing-runtime"])
}

func TestBootstrapPackageJSONExists(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(`
		model Post {
			functions {
				create createPost()
			}
		}
	`), 0777)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0777)
	assert.NoError(t, err)

	err = node.Bootstrap(tmpDir)
	assert.NoError(t, err)

	entries, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)

	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	assert.Equal(t, []string{"package.json", "schema.keel"}, names)
}
