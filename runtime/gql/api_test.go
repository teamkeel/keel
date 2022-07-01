package gql

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
)

// TestHandlerSuite is a table-driven test suite for the Handler type's behaviour.
// The "table" is implied by what it finds in the ../testdata directory.
// That directory has a sub directory for each test case.
// Each such directory should contain:
//	o  The keel schema to use
//  o  The graphql request to send
//  o  Either the expected returned data, or the expected returned errors.
func TestHandlersSuite(t *testing.T) {
	testCasesParent := "../testdata"
	subDirs, err := ioutil.ReadDir(testCasesParent)
	require.NoError(t, err)

	for _, dir := range subDirs {
		if !dir.IsDir() {
			continue
		}
		dirPath := filepath.Join(testCasesParent, dir.Name())
		t.Run(dir.Name(), func(t *testing.T) {
			runTestCase(t, dirPath)
		})
	}
}

func runTestCase(t *testing.T, dirPath string) {
	s2m := schema.Builder{}
	protoSchema, err := s2m.MakeFromDirectory(dirPath)
	require.NoError(t, err)
	protoJSON, err := json.Marshal(protoSchema)
	require.NoError(t, err)

	handlers, err := NewHandlersFromJSON(string(protoJSON))
	require.NoError(t, err)
	chosenHandler, ok := handlers["Web"]
	require.True(t, ok)

	// Ask the handler to respond to the query.
	request := fileContents(t, filepath.Join(dirPath, requestFile))
	result := chosenHandler.Handle(string(request))

	if expectingHappyPath(t, dirPath) {
		expectedData := fileContents(t, filepath.Join(dirPath, expectedDataFile))
		actualData, err := json.MarshalIndent(result.Data, "", "  ")
		require.NoError(t, err)
		require.JSONEq(t, string(expectedData), string(actualData))
	} else {
		expectedErrors := fileContents(t, filepath.Join(dirPath, expectedErrorsFile))
		actualErrors, err := json.MarshalIndent(result.Errors, "", "  ")
		if os.Getenv("DEBUG") != "" {
			t.Logf("Actual error json is: \n%s\n", actualErrors)
		}
		require.NoError(t, err)
		require.JSONEq(t, string(expectedErrors), string(actualErrors))
	}
}

func expectingHappyPath(t *testing.T, dir string) bool {
	return fileExists(t, filepath.Join(dir, expectedDataFile))
}

func fileExists(t *testing.T, filePath string) bool {
	_, err := os.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		t.Fatalf(err.Error())
	}
	return true
}

func fileContents(t *testing.T, filePath string) []byte {
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("cannot read this file: %s, error is: %v", filePath, err)
	}
	return data
}

const expectedDataFile string = "response.json"
const expectedErrorsFile string = "errors.json"
const requestFile string = "request.gql"
