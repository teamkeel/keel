package testing

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/testhelpers"
)

func MakeContext(t *testing.T, ctx context.Context, keelSchema string, resetDatabase bool) (context.Context, db.Database, *proto.Schema) {
	dbConnInfo := &db.ConnectionInfo{
		Host:     "localhost",
		Port:     "8001",
		Username: "postgres",
		Password: "postgres",
		Database: "keel",
	}

	authConfig, _ := runtimectx.GetOAuthConfig(ctx)
	cfg := config.ProjectConfig{
		Auth: *authConfig,
	}

	schemaFiles :=
		&reader.Inputs{
			SchemaFiles: []*reader.SchemaFile{
				{
					Contents: keelSchema,
					FileName: "schema.keel",
				},
			},
		}

	builder := &schema.Builder{Config: &cfg}
	schema, err := builder.MakeFromInputs(schemaFiles)
	require.NoError(t, err)

	// Add private key to context
	pk, err := testhelpers.GetEmbeddedPrivateKey()
	require.NoError(t, err)
	ctx = runtimectx.WithPrivateKey(ctx, pk)

	ctx, err = testhelpers.WithTracing(ctx)
	require.NoError(t, err)

	databaseName := strings.ToLower("keel_test_" + t.Name())

	// Add database to context
	database, err := testhelpers.SetupDatabaseForTestCase(ctx, dbConnInfo, schema, databaseName, resetDatabase)
	require.NoError(t, err)
	ctx = db.WithDatabase(ctx, database)

	return ctx, database, schema
}
