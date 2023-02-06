package node_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/node"
)

func TestDevelopmentServer(t *testing.T) {
	files := node.GeneratedFiles{
		{
			Path: "schema.keel",
			Contents: `
				model Person {
					functions {
						get getPerson(id)
					}
				}
			`,
		},
		{
			Path: "functions/getPerson.ts",
			Contents: `
				import { GetPerson } from "@teamkeel/sdk";

				export default GetPerson(async (inputs) => {
					return {id: inputs.id, createdAt: new Date("2022-01-01"), updatedAt: new Date("2022-01-01")};
				});
			`,
		},
	}

	runDevelopmentServerTest(t, files, func(server *node.DevelopmentServer, err error) {
		if !assert.NoError(t, err) {
			if server != nil {
				fmt.Println("=== Development Server Output ===")
				fmt.Println(server.Output())
			}
			return
		}

		body := bytes.NewBufferString(`{"method": "getPerson", "params": {"id": "1234"}, "meta": { "headers": {}}}`)
		res, err := http.Post(server.URL, "application/json", body)
		require.NoError(t, err)
		assert.Equal(t, res.StatusCode, 200)

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, `{"jsonrpc":"2.0","result":{"id":"1234","createdAt":"2022-01-01T00:00:00.000Z","updatedAt":"2022-01-01T00:00:00.000Z"},"meta":{"headers":{}}}`, string(b))
	})
}

func TestDevelopmentServerStartError(t *testing.T) {
	files := node.GeneratedFiles{
		{
			Path: "schema.keel",
			Contents: `
				model Person {
					functions {
						get getPerson(id)
					}
				}
			`,
		},
		{
			Path: "functions/getPerson.ts",
			Contents: `
				import { GetPerson } from "@teamkeel/sdk";

				export default GetPerson(async (inputs) => {
					return "this is not correct";
				});
			`,
		},
	}

	runDevelopmentServerTest(t, files, func(server *node.DevelopmentServer, err error) {
		assert.Error(t, err)
		assert.Contains(t, server.Output(), "Type 'string' is not assignable to type 'Person'.")
	})
}

func runDevelopmentServerTest(t *testing.T, files node.GeneratedFiles, fn func(*node.DevelopmentServer, error)) {
	tmpDir := t.TempDir()

	wd, err := os.Getwd()
	require.NoError(t, err)

	err = node.Bootstrap(tmpDir, node.WithPackagesPath(filepath.Join(wd, "../packages")))
	require.NoError(t, err)

	for _, f := range files {
		f.Path = filepath.Join(tmpDir, f.Path)
	}
	err = files.Write()
	require.NoError(t, err)

	files, err = node.Generate(context.Background(), tmpDir, node.WithDevelopmentServer(true))
	require.NoError(t, err)

	err = files.Write()
	require.NoError(t, err)

	server, err := node.RunDevelopmentServer(tmpDir, &node.ServerOpts{
		EnvVars: map[string]string{
			"DB_CONN_TYPE": "pg",
			"DB_CONN":      "postgresql://postgres:postgres@localhost:8001/keel",
		},
	})
	t.Cleanup(func() {
		if server != nil {
			_ = server.Kill()
		}
	})
	fn(server, err)
}
