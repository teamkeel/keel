package node_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/node"
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
