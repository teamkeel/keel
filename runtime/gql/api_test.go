package gql

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/migrations"
	keelpostgres "github.com/teamkeel/keel/postgres"
	"github.com/teamkeel/keel/schema"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestHandlerSuite is a table-driven test suite for the Handler type's behaviour.
// The "table" is implied by what it finds in the ../testdata directory.
// That directory has a sub directory for each test case.
// Each such directory should contain:
//	o  The keel schema to use
//  o  The graphql request(s) to send (in sequence)
//  o  Either the expected responses for each request, or the expected returned errors (for the last of the requests)
// The idea being that the first request might Create some data, and a second might query it.
// Or more generally put, the later requests are dependent on the earlier ones.
func TestHandlersSuite(t *testing.T) {
	testCasesParent := "../testdata"
	subDirs, err := ioutil.ReadDir(testCasesParent)
	require.NoError(t, err)

	// We start a new postgres container for each case in the suite, and stop it again
	// at the end of each case. So we want to start the suite with it stopped.
	keelpostgres.StopThePostgresContainer()

	for _, dir := range subDirs {
		if !dir.IsDir() {
			continue
		}
		dirName := dir.Name()

		// This is to make it quick and easy to isolate just one of the tests during development
		var isolateDir string = ""
		//isolateDir = "create-simplest-happy"
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

	// Make the proto schema from the Keel schema.
	s2m := schema.Builder{}
	protoSchema, err := s2m.MakeFromDirectory(dirPath)
	require.NoError(t, err)
	protoJSON, err := json.Marshal(protoSchema)
	require.NoError(t, err)

	// Bring up a suitable database (that is migrated to this schema)
	gormDB, _, err := bringUpLocalDBToMatchSchema(dirPath)
	require.NoError(t, err)
	sqlDB, err := gormDB.DB()
	require.NoError(t, err)
	defer func() {
		sqlDB.Close()
		keelpostgres.StopThePostgresContainer()
	}()

	// Construct the handers we wish to test.
	handlers, err := NewHandlersFromJSON(string(protoJSON), gormDB)
	require.NoError(t, err)
	chosenHandler, ok := handlers["Web"] // There is one handler per each API defined in the Keel schema.
	require.True(t, ok)

	// Identify the sequence of GraphQL queries we want to execute.
	gqlRequests := splitOutSections(t, dirPath, requestFile)

	// Now we exercise the handler with each request in turn.
	// The checking depends on if this test fixture has implied a happy path or error path check.

	// For happy path tests, we expect the response to each request to match that specified
	// in the test case, and expect no errors.
	if expectingHappyPath(t, dirPath) {
		t.Logf("Is happy path\n")
		expectedDataResponses := splitOutSections(t, dirPath, expectedDataFile)
		for i, req := range gqlRequests {
			t.Logf("Doing request number: %d\n", i)
			result := chosenHandler.Handle(string(req))
			require.Equal(t, 0, len(result.Errors))

			expectedData := expectedDataResponses[i]
			actualData, err := json.MarshalIndent(result.Data, "", "  ")
			if os.Getenv("DEBUG") != "" {
				t.Logf("Actual data json is: \n%s\n", actualData)
			}
			require.NoError(t, err)
			require.JSONEq(t, string(expectedData), string(actualData))
		}
	} else {

		t.Logf("Is error path\n")
		// For error checking tests, we expect all the requests except for the last one
		// to run without errors, and for the last request to return an error as defined by the test case.
		for i, req := range gqlRequests {
			t.Logf("Doing request number: %d\n", i)
			result := chosenHandler.Handle(string(req))
			isFinalRequest := i == len(gqlRequests)-1
			if isFinalRequest {
				expectedErrors := fileContents(t, filepath.Join(dirPath, expectedErrorsFile))
				actualErrors, err := json.MarshalIndent(result.Errors, "", "  ")
				if os.Getenv("DEBUG") != "" {
					t.Logf("Actual error json is: \n%s\n", actualErrors)
				}
				require.NoError(t, err)
				require.JSONEq(t, string(expectedErrors), string(actualErrors))
			} else {
				require.Equal(t, 0, len(result.Errors))
			}
		}
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

// splitOutSections, reads the given file, and splits its contents into sections using
// a dedicated delimiter.
func splitOutSections(t *testing.T, dirPath string, file string) (sections []string) {
	contents := string(fileContents(t, filepath.Join(dirPath, file)))
	return strings.Split(contents, delimiter)
}

// bringUpLocalDBToMatchSchema brings up a local, dockerised PostgresSQL database,
// that is fully migrated to match the given Keel Schema. It re-uses the incumbent
// container if it can (including therefore the incumbent database state), but also works
// if it has to do everything from scratch - including fetching the PostgreSQL image.
//
// It is good to use for the Keel Run command, but also to use in test fixtures.
func bringUpLocalDBToMatchSchema(schemaDir string) (gormDB *gorm.DB, protoSchemaJSON string, err error) {
	sqlDB, err := keelpostgres.BringUpPostgresLocally()
	if err != nil {
		return nil, "", err
	}
	gormDB, err = gorm.Open(gormpostgres.New(gormpostgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, "", err
	}
	if err := migrations.InitProtoSchemaStore(sqlDB); err != nil {
		return nil, "", err
	}

	protoSchemaJSON, err = migrations.DoMigrationBasedOnSchemaChanges(sqlDB, schemaDir)
	if err != nil {
		return nil, "", err
	}
	return gormDB, protoSchemaJSON, nil
}

const expectedDataFile string = "response.json"
const expectedErrorsFile string = "errors.json"
const requestFile string = "request.gql"

const delimiter string = `--next section--`
