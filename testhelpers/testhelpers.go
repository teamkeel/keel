package testhelpers

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	cp "github.com/otiai10/copy"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// WithTmpDir copies the contents of the src dir to a new temporary directory, returning the tmp dir path
func WithTmpDir(dir string) (string, error) {
	base := filepath.Base(dir)

	tmpDir, err := ioutil.TempDir("", base)

	if err != nil {
		return "", err
	}

	err = cp.Copy(dir, tmpDir)

	if err != nil {
		return "", err
	}

	return tmpDir, nil
}

const dbConnString = "host=localhost port=8001 user=postgres password=postgres dbname=%s sslmode=disable"

func SetupDatabaseForTestCase(t *testing.T, schema *proto.Schema, dbName string) *gorm.DB {
	mainDB, err := gorm.Open(
		postgres.Open(fmt.Sprintf(dbConnString, "keel")),
		&gorm.Config{})
	require.NoError(t, err)

	// Drop the database if it already exists. The normal dropping of it at the end of the
	// test case is bypassed if you quit a debug run of the test in VS Code.
	require.NoError(t, mainDB.Exec("DROP DATABASE if exists "+dbName).Error)

	// Create the database and drop at the end of the test
	err = mainDB.Exec("CREATE DATABASE " + dbName).Error
	require.NoError(t, err)
	// t.Cleanup(func() {
	// 	require.NoError(t, mainDB.Exec("DROP DATABASE "+dbName).Error)
	// })

	// Connect to the newly created test database and close connection
	// at the end of the test. We need to explicitly close the connection
	// so the mainDB connection can drop the database.
	testDB, err := gorm.Open(
		postgres.Open(fmt.Sprintf(dbConnString, dbName)),
		&gorm.Config{})
	require.NoError(t, err)

	t.Cleanup(func() {
		conn, err := testDB.DB()
		require.NoError(t, err)
		conn.Close()
	})

	// Migrate the database to this test case's schema.
	m := migrations.New(schema, nil)

	require.NoError(t, m.Apply(testDB))

	return testDB
}

func DbNameForTestName(testName string) string {
	re := regexp.MustCompile(`[^\w]`)
	return strings.ToLower(re.ReplaceAllString(testName, ""))
}
