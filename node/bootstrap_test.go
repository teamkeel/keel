package node_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/node"
)

func TestBootstrap(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(`
		model Post {
			functions {
				create createPost()
			}
		}
	`), 0777)
	err := node.Bootstrap(tmpDir)
	assert.NoError(t, err)

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

	os.WriteFile(filepath.Join(tmpDir, "schema.keel"), []byte(`
		model Post {
			functions {
				create createPost()
			}
		}
	`), 0777)
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0777)
	err := node.Bootstrap(tmpDir)
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
