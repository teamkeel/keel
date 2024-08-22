package proto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFieldNames(t *testing.T) {
	t.Parallel()
	require.Equal(t, []string{"Field1", "Field2"}, referenceSchema.Models[0].FieldNames())
}
