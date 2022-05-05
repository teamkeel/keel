package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/testdata"
)

func TestItCompilesAndRuns(t *testing.T) {
	s2m := NewSchema(testdata.ReferenceExample)
	protoModels, err := s2m.Make()
	require.Nil(t, err)

	require.Equal(t, 1, len(protoModels.Models))
}