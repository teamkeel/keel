package schema2model

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/testdata"
)

func TestItCompilesAndRuns(t *testing.T) {
	s2m := NewSchema(testdata.ReferenceExample)
	protoModels, err := s2m.Make()

	require.Nil(t, err)

	require.Equal(t, 2, len(protoModels.Models))
}
