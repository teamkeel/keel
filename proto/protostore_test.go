package proto

import (
	"testing"

	"os"

	"github.com/stretchr/testify/require"
)

func TestRoundTripHappyPath(t *testing.T) {
	schemaDir, err := os.MkdirTemp("", "keel")
	require.NoError(t, err)
	proto := Schema{
		Models: []*Model{
			{
				Name: "foo",
			},
		},
	}
	err = SaveToLocalStorage(&proto, schemaDir)
	require.NoError(t, err)

	retreivedProto, err := FetchFromLocalStorage(schemaDir)
	require.NoError(t, err)
	require.Equal(t, "foo", retreivedProto.Models[0].Name)

	// cleanup
	err = os.RemoveAll(schemaDir)
	require.NoError(t, err)

	// todo - put in error handling tests
}
