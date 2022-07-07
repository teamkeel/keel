package runtime

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
	"gorm.io/gorm"
)

func TestServer(t *testing.T) {
	// todo - this test is problematic when you run with go test ./... because its use
	// of the postgres database (singleton) - e.g. starting and stopping the db container
	// and doing migrations clash with any other tests that use the database at the same
	// time.

	// Temporarily we only test that the server compiles, can be told to listen and serve,
	// and can be shut down.

	schemaDir := filepath.Join(".", "testdata", "get-simplest-happy")

	s2m := schema.Builder{}
	protoSchema, err := s2m.MakeFromDirectory(schemaDir)
	require.NoError(t, err)
	protoJSON, err := json.Marshal(protoSchema)
	require.NoError(t, err)

	var gormDB *gorm.DB = nil

	svr, err := NewServer(string(protoJSON), gormDB)
	require.NoError(t, err)
	defer svr.Shutdown(context.Background())

	go svr.ListenAndServe()
}
