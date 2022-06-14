package proto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortedStrings(t *testing.T) {
	require.Equal(t, []string{"a", "b", "c"}, sortedStrings([]string{"b", "c", "a"}))
}
