package node_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/node"
)

// TestBootstrap merely tests that no error occurs during the execution of the function
// More in depth tests for codegen happen in the codegen_test file in this package.
func TestBootstrap(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bootstrap")
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(tmpDir)
	})

	InsertSchemaIntoDir(t, `
		model Post {
			fields {
				title Text
			}

			functions {
				create createPost() with(title)
			}
		}
	`, tmpDir)
	err = node.Bootstrap(tmpDir)
	assert.NoError(t, err)
}

// TestGeneratePackages merely tests that no error occurs during the execution of the function
// More in depth tests for codegen happen in the codegen_test file in this package.
func TestGeneratePackages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generate-packages")
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(tmpDir)
	})

	InsertSchemaIntoDir(t, `
		model Post {
			fields {
				title Text
			}

			functions {
				create createPost() with(title)
			}
		}
	`, tmpDir)
	err = node.GeneratePackages(tmpDir)
	assert.NoError(t, err)
}

func TestGenerateDevelopmentServer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generate-dev-server")
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(tmpDir)
	})

	InsertSchemaIntoDir(t, `
		model Post {
			fields {
				title Text
			}

			functions {
				create createPost() with(title)
			}
		}
	`, tmpDir)
	err = node.GenerateDevelopmentServer(tmpDir)
	assert.NoError(t, err)
}

func TestRunDevelopmentServer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generate-packages")
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(tmpDir)
	})

	InsertSchemaIntoDir(t, `
		model Post {
			fields {
				title Text
			}

			functions {
				create createPost() with(title)
			}
		}
	`, tmpDir)

	err = node.GenerateDevelopmentServer(tmpDir)
	assert.NoError(t, err)

	server, err := node.RunDevelopmentServer(tmpDir, map[string]any{})
	assert.NoError(t, err)
	assert.NoError(t, server.Kill())
}

func InsertSchemaIntoDir(t *testing.T, schema string, dir string) {
	f, err := os.Create(filepath.Join(dir, "schema.keel"))

	assert.NoError(t, err)

	_, err = f.WriteString(schema)

	assert.NoError(t, err)
}
