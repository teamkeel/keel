package schema2model

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/internal/testdata"
)

func TestItCompilesAndRuns(t *testing.T) {
	s2m := NewSchema2Model(testdata.ReferenceExample)
	protoModels, err := s2m.Make()
	require.Nil(t, err)

	require.Equal(t, 1, len(protoModels.Models))
}