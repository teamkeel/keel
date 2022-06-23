package runtime

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
)

func TestJustToDriveInitialCodingAndCompiling(t *testing.T) {

	schemaDir := filepath.Join(".", "testdata")
	s2m := schema.Builder{}
	protoSchema, err := s2m.MakeFromDirectory(schemaDir)
	require.NoError(t, err)
	protoJSON, err := json.Marshal(protoSchema)
	require.NoError(t, err)
	svr, err := NewServer(string(protoJSON))
	require.NoError(t, err)

	_ = svr
}
