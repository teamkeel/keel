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
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
)

func TestDevelopmentServerAction(t *testing.T) {
	files := codegen.GeneratedFiles{
		{
			Path: "schema.keel",
			Contents: `
				message PersonResponse {
					id Text
				}
				model Person {
					actions {
						read getPerson(id) returns (PersonResponse)
					}
				}
			`,
		},
		{
			Path: "functions/getPerson.ts",
			Contents: `
				import { GetPerson, permissions } from "@teamkeel/sdk";

				export default GetPerson(async (ctx, inputs) => {
					permissions.allow()
					return {id: inputs.id};
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

		body := bytes.NewBufferString(`{"method": "getPerson", "type": "action", "params": {"id": "1234"}, "meta": { "permissionState": { "status": "unknown" }, "headers": {}}}`)
		res, err := http.Post(server.URL, "application/json", body)
		require.NoError(t, err)
		assert.Equal(t, res.StatusCode, 200)

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, `{"jsonrpc":"2.0","result":{"id":"1234"},"meta":{"headers":{}}}`, string(b))
	})
}

func TestDevelopmentServerJob(t *testing.T) {
	files := codegen.GeneratedFiles{
		{
			Path: "schema.keel",
			Contents: `
				job ProcessPeople {
					@permission(roles: [Admin])
				}
				role Admin {}
			`,
		},
		{
			Path: "jobs/processPeople.ts",
			Contents: `
				import { ProcessPeople, models } from "@teamkeel/sdk";

				export default ProcessPeople(async (ctx) => { });
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

		body := bytes.NewBufferString(`{"method": "processPeople", "type": "job", "meta": { "permissionState": { "status": "granted" }}}`)
		res, err := http.Post(server.URL, "application/json", body)
		require.NoError(t, err)
		assert.Equal(t, res.StatusCode, 200)

		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, `{"jsonrpc":"2.0","result":null}`, string(b))
	})
}

func TestDevelopmentServerStartError(t *testing.T) {
	files := codegen.GeneratedFiles{
		{
			Path: "schema.keel",
			Contents: `
				model Person {
					actions {
						get getPerson(id) @function
					}
				}
			`,
		},
		{
			Path: "functions/getPerson.ts",
			Contents: `
				import { permissions, GetPerson } from "@teamkeel/sdk";
				
				console.error('unexpected error')
				process.exit(1);

				export default GetPerson(async (inputs, api, ctx) => {
					permissions.allow();

					return "this is not correct";
				});
			`,
		},
	}

	runDevelopmentServerTest(t, files, func(server *node.DevelopmentServer, err error) {
		assert.Error(t, err)
		assert.Contains(t, server.Output(), "unexpected error")
	})
}

func runDevelopmentServerTest(t *testing.T, files codegen.GeneratedFiles, fn func(*node.DevelopmentServer, error)) {
	tmpDir := t.TempDir()

	require.NoError(t, files.Write(tmpDir))

	wd, err := os.Getwd()
	require.NoError(t, err)

	files, err = node.Bootstrap(tmpDir, node.WithPackagesPath(filepath.Join(wd, "../packages")))
	require.NoError(t, err)
	require.NoError(t, files.Write(tmpDir))

	_, err = testhelpers.NpmInstall(tmpDir)
	require.NoError(t, err)

	err = files.Write(tmpDir)
	require.NoError(t, err)

	b := schema.Builder{}
	schema, err := b.MakeFromDirectory(tmpDir)
	require.NoError(t, err)

	files, err = node.Generate(context.Background(), schema, node.WithDevelopmentServer(true))
	require.NoError(t, err)

	err = files.Write(tmpDir)
	require.NoError(t, err)

	server, err := node.StartDevelopmentServer(tmpDir, &node.ServerOpts{
		EnvVars: map[string]string{
			"KEEL_DB_CONN_TYPE": "pg",
			"KEEL_DB_CONN":      "postgresql://postgres:postgres@localhost:8001/keel",
		},
	})
	t.Cleanup(func() {
		if server != nil {
			_ = server.Kill()
		}
	})
	fn(server, err)
}
