package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMakeFromDirectoryCompilesAndRuns(t *testing.T) {
	inputDir := "../testdata/schema-dirs/kitchen-sink"
	s2m := Schema{}
	protoModels, err, _ := s2m.MakeFromDirectory(inputDir)
	require.Nil(t, err)
	require.Equal(t, 3, len(protoModels.Models))
}

func TestMakeFromFileCompilesAndRuns(t *testing.T) {
	schemaFile := "../testdata/schema-dirs/kitchen-sink/kitchen-sink.keel"
	s2m := Schema{}
	protoModels, err, _ := s2m.MakeFromFile(schemaFile)
	require.Nil(t, err)
	require.Equal(t, 2, len(protoModels.Models))
}
