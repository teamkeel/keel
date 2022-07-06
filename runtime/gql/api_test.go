package gql

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
	"gorm.io/gorm"
)

// TestHandlerSuite is a table-driven test suite for the Handler type's behaviour.
// The "table" is implied by what it finds in the ../testdata directory.
// That directory has a sub directory for each test case.
// Each such directory should contain:
//	o  The keel schema to use
//  o  The graphql request(s) to send (in sequence)
//  o  Either the expected returned data, or the expected returned errors (for the last of those requests)
func TestHandlersSuite(t *testing.T) {
	testCasesParent := "../testdata"
	subDirs, err := ioutil.ReadDir(testCasesParent)
	require.NoError(t, err)

	for _, dir := range subDirs {
		if !dir.IsDir() {
			continue
		}
		dirName := dir.Name()

		// This is isolate just one of the tests during development
		var isolateDir string
		isolateDir = "create-simplest-error"
		isolateDir = ""
		if isolateDir != "" && dirName != isolateDir {
			continue
		}

		dirPath := filepath.Join(testCasesParent, dirName)
		t.Logf("Starting test for directory: %s\n", dirPath)

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

	var gormDB *gorm.DB = nil // todo provide a suitably initialised db to the this test fixture

	handlers, err := NewHandlersFromJSON(string(protoJSON), gormDB)
	require.NoError(t, err)
	chosenHandler, ok := handlers["Web"] // There is one handler per each API defined in the Keel schema.
	require.True(t, ok)

	// Fetch the list of GraphQL queries required for this test case, and execute them
	// in turn. If there are more than one, the earlier ones will be setup for the final one, and
	// therefore, we are only interested in the result from the last one.
	gqlRequests := assembleRequests(t, dirPath, requestFile)
	var result *graphql.Result
	finalRequest := len(gqlRequests) - 1
	for i, req := range gqlRequests {
		result = chosenHandler.Handle(string(req))
		if i != finalRequest { // All but the last request must always work without error.
			if len(result.Errors) != 0 {
				t.Fatalf("error encountered on one of the set-up gql requests: %v", result.Errors)
			}
		}
	}

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

// 	request := fileContents(t, filepath.Join(dirPath, requestFile))

// assembleRequests expects to find 1 or more graphql requests in the request file -
// delimitted by a special reserved string token. It returns the requests thus
// delimited as strings.
func assembleRequests(t *testing.T, dirPath string, requestFile string) (requests []string) {
	contents := string(fileContents(t, filepath.Join(dirPath, requestFile)))
	return strings.Split(contents, requestDelim)
}

const expectedDataFile string = "response.json"
const expectedErrorsFile string = "errors.json"
const requestFile string = "request.gql"

const requestDelim string = `--nextrequest--`
