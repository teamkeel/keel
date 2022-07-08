package run

import (
	"database/sql"

	"github.com/teamkeel/keel/migrations"
	keelpostgres "github.com/teamkeel/keel/postgres"
	gormpostgres "gorm.io/driver/postgres"

	"gorm.io/gorm"
)

// BringUpLocalDBToMatchSchema brings up a local, dockerised PostgresSQL database,
// that is fully migrated to match the given Keel Schema. It re-uses the incumbent
// container if it can (including therefore the incumbent database state), but also works
// if it has to do everything from scratch - including fetching the PostgreSQL image.
//
// It is good to use for the Keel Run command, but also to use in test fixtures.
func BringUpLocalDBToMatchSchema(schemaDir string) (sqlDB *sql.DB, gormDB *gorm.DB, protoSchemaJSON string, err error) {
	sqlDB, err = keelpostgres.BringUpPostgresLocally()
	if err != nil {
		return nil, nil, "", err
	}
	gormDB, err = gorm.Open(gormpostgres.New(gormpostgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, "", err
	}
	if err := migrations.InitProtoSchemaStore(sqlDB); err != nil {
		return nil, nil, "", err
	}

	protoSchemaJSON, err = migrations.DoMigrationBasedOnSchemaChanges(sqlDB, schemaDir)
	if err != nil {
		return nil, nil, "", err
	}
	return sqlDB, gormDB, protoSchemaJSON, nil
}
