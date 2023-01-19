package testhelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cp "github.com/otiai10/copy"
	"github.com/teamkeel/keel/cmd/database"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// WithTmpDir copies the contents of the src dir to a new temporary directory, returning the tmp dir path
func WithTmpDir(dir string) (string, error) {
	base := filepath.Base(dir)

	tmpDir, err := os.MkdirTemp("", base)

	if err != nil {
		return "", err
	}

	err = cp.Copy(dir, tmpDir)

	if err != nil {
		return "", err
	}

	return tmpDir, nil
}

const dbConnString = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable"

func SetupDatabaseForTestCase(dbConnInfo *database.ConnectionInfo, schema *proto.Schema, dbName string) (*gorm.DB, error) {
	mainDB, err := gorm.Open(
		postgres.Open(fmt.Sprintf(dbConnString, dbConnInfo.Host, dbConnInfo.Port, dbConnInfo.Username, dbConnInfo.Password, dbConnInfo.Database)),
		&gorm.Config{
			Logger: logger.Discard.LogMode(logger.Silent),
		})
	if err != nil {
		return nil, err
	}

	err = mainDB.Exec("select pg_terminate_backend(pg_stat_activity.pid) from pg_stat_activity where datname = '" + dbName + "' and pg_stat_activity.pid <> pg_backend_pid();").Error
	if err != nil {
		return nil, err
	}

	// Drop the database if it already exists. The normal dropping of it at the end of the
	// test case is bypassed if you quit a debug run of the test in VS Code.
	err = mainDB.Exec("DROP DATABASE if exists " + dbName).Error
	if err != nil {
		return nil, err
	}

	// Create the database and drop at the end of the test
	err = mainDB.Exec("CREATE DATABASE " + dbName).Error
	if err != nil {
		return nil, err
	}

	// Connect to the newly created test database and close connection
	// at the end of the test. We need to explicitly close the connection
	// so the mainDB connection can drop the database.
	testDB, err := gorm.Open(
		postgres.Open(fmt.Sprintf(dbConnString, dbConnInfo.Host, dbConnInfo.Port, dbConnInfo.Username, dbConnInfo.Password, dbName)),
	)
	if err != nil {
		return nil, err
	}

	// Migrate the database to this test case's schema.
	m := migrations.New(schema, nil)

	err = m.Apply(testDB)
	if err != nil {
		return nil, err
	}

	return testDB, nil
}

func DbNameForTestName(testName string) string {
	re := regexp.MustCompile(`[^\w]`)
	return strings.ToLower(re.ReplaceAllString(testName, ""))
}
