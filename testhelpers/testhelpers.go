package testhelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cp "github.com/otiai10/copy"
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

const dbConnString = "host=localhost port=8001 user=postgres password=postgres dbname=%s sslmode=disable"

var mainDB *gorm.DB

func TruncateTables(db *gorm.DB) error {
	var tables []string
	err := db.Table("pg_tables").
		Where("schemaname = 'public' and tablename != 'keel_schema'").
		Pluck("tablename", &tables).Error
	if err != nil {
		return err
	}

	err = db.Exec("TRUNCATE TABLE " + strings.Join(tables, ",") + " CASCADE").Error
	if err != nil {
		return err
	}

	return nil
}

func SetupDatabaseForTestCase(schema *proto.Schema, dbName string) (*gorm.DB, error) {
	if mainDB == nil {
		var err error
		mainDB, err = gorm.Open(
			postgres.Open(fmt.Sprintf(dbConnString, "keel")),
			&gorm.Config{
				Logger: logger.Discard.LogMode(logger.Silent),
			})

		if err != nil {
			return nil, err
		}
	}

	err := mainDB.Exec("select pg_terminate_backend(pg_stat_activity.pid) from pg_stat_activity where datname = '" + dbName + "' and pg_stat_activity.pid <> pg_backend_pid();").Error

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
		postgres.Open(fmt.Sprintf(dbConnString, dbName)))

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

func CleanupDatabaseSetup(main *gorm.DB, testDB *gorm.DB, dbName string) error {
	err := main.Exec("select pg_terminate_backend(pg_stat_activity.pid) from pg_stat_activity where datname = '" + dbName + "' and pg_stat_activity.pid <> pg_backend_pid();").Error

	if err != nil {
		return err
	}

	err = main.Exec("DROP DATABASE if exists " + dbName).Error

	if err != nil {
		return err
	}
	conn, err := testDB.DB()
	if err != nil {
		return err
	}

	err = conn.Close()

	if err != nil {
		return err
	}
	return nil
}

func DbNameForTestName(testName string) string {
	re := regexp.MustCompile(`[^\w]`)
	return strings.ToLower(re.ReplaceAllString(testName, ""))
}
