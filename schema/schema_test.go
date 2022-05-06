package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestItCompilesAndRuns(t *testing.T) {
	inputDir := "./testdata/schema-dirs/kitchen-sink"
	s2m := NewSchema(inputDir)
	protoModels, err := s2m.Make()

	require.Nil(t, err)

	require.Equal(t, 2, len(protoModels.Models))
}
