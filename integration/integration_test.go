package integration_test

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	gotest "testing"

	cp "github.com/otiai10/copy"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/testing"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var pattern = flag.String("pattern", "", "Pattern to match individual test case names")

func TestIntegration(t *gotest.T) {
	t.Parallel()
	entries, err := os.ReadDir("./testdata")
	require.NoError(t, err)

	tmpDir := t.TempDir()

	wd, err := os.Getwd()
	require.NoError(t, err)

	err = node.Bootstrap(
		tmpDir,
		node.WithPackagesPath(filepath.Join(wd, "../packages")),
		node.WithLogger(func(s string) {}),
	)
	require.NoError(t, err)

	_, err = testhelpers.NpmInstall(tmpDir)
	require.NoError(t, err)

	// Whatever files/dirs are present now can stay between tests
	// e.g. node_modules, package.json
	genericEntries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())
	if err != nil {
		panic(err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewSchemaless(attribute.String("service.name", "runtime")),
		),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	for _, e := range entries {
		t.Run(e.Name(), func(t *gotest.T) {
			testDir := filepath.Join("./testdata", e.Name())

			defer provider.ForceFlush(ctx)

			// These files might be present when someone is working on tests
			// but we don't want to copy them over to the test dir
			skipEntries := []string{
				"/.build",
				"/node_modules",
				"/package.json",
				"/package-lock.json",
				"/tsconfig.json",
			}

			// Copy test files to temp dir
			require.NoError(t, cp.Copy(testDir, tmpDir, cp.Options{
				Skip: func(srcinfo os.FileInfo, s, d string) (bool, error) {
					for _, v := range skipEntries {
						if strings.HasSuffix(d, v) {
							return true, nil
						}
					}
					return false, nil
				},
			}))

			// At the end of this tests remove all the test files
			t.Cleanup(func() {
				entries, err := os.ReadDir(tmpDir)
				require.NoError(t, err)
			outer:
				for _, entry := range entries {
					for _, g := range genericEntries {
						if g.Name() == entry.Name() {
							continue outer
						}
					}
					os.RemoveAll(filepath.Join(tmpDir, entry.Name()))
				}
			})

			// Use the docker compose database
			dbConnInfo := &db.ConnectionInfo{
				Host:     "localhost",
				Port:     "8001",
				Username: "postgres",
				Database: "keel",
				Password: "postgres",
			}

			err = testing.Run(ctx, &testing.RunnerOpts{
				Dir:        tmpDir,
				DbConnInfo: dbConnInfo,
				Secrets: map[string]string{
					"TEST_API_KEY": "1232132_2323",
					"NAME_API_KEY": "worf",
				},
				Pattern:        *pattern,
				TestGroupName:  e.Name(),
				GenerateClient: true,
			})
			require.NoError(t, err)
		})
	}
}
