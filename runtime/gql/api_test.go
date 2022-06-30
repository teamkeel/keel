package gql

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
)

func TestHandlersHappyPath(t *testing.T) {
	schemaDir := filepath.Join("..", "testdata", "get-simplest")
	s2m := schema.Builder{}
	protoSchema, err := s2m.MakeFromDirectory(schemaDir)
	require.NoError(t, err)
	protoJSON, err := json.Marshal(protoSchema)
	require.NoError(t, err)

	// We'll use the highest level wrapper constructor (NewHandlersFromJSON) - because this
	// exercises the stack of lower level constructors under the hood.
	handlers, err := NewHandlersFromJSON(string(protoJSON))
	require.NoError(t, err)
	chosenHandler, ok := handlers["Web"]
	require.True(t, ok)
	result := chosenHandler.Handle(exampleQuery)
	require.Equal(t, 0, len(result.Errors))
	require.Equal(t, expected, result.Data)
}

func TestHandlersErrorPath(t *testing.T) {
	schemaDir := filepath.Join("..", "testdata", "get-simplest")
	s2m := schema.Builder{}
	protoSchema, err := s2m.MakeFromDirectory(schemaDir)
	require.NoError(t, err)
	protoJSON, err := json.Marshal(protoSchema)
	require.NoError(t, err)

	handlers, err := NewHandlersFromJSON(string(protoJSON))
	require.NoError(t, err)
	chosenHandler, ok := handlers["Web"]
	require.True(t, ok)
	result := chosenHandler.Handle(malformedQuery)
	require.Equal(t, 1, len(result.Errors))

	errorMsg := result.Errors[0].Message
	require.Equal(t, `Unknown argument "nosuchfield" on field "getAuthor" of type "Query".`, errorMsg)
}

const exampleQuery string = `{ getAuthor(name: "fred") { name } }`
const malformedQuery string = `{ getAuthor(nosuchfield: "fred") { name } }`

const expectedKey string = `getAuthor`

var expectedData map[string]any = map[string]any{"name": "Harriet"}
var expected map[string]any = map[string]any{expectedKey: expectedData}
