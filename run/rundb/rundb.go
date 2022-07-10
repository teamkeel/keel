package rundb

import (
	"github.com/teamkeel/keel/migrations"
	keelpostgres "github.com/teamkeel/keel/postgres"
	"github.com/teamkeel/keel/proto"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// launchDB spins up a dockerised Postgres database that is migrated
// to the given schema.
//
// If you specify retainData=true, and its container already
// exists, it will reuse that container with its data volume intact, and therefore
// the migration will be done with respect to the last-known schema stored in the database, and the data inside
// the database will be preserved. In this mode it can also cope with the first-ever
// launch (when there is no previously used container), and does migrations
// with respect to an empty schema - i.e. a schema with no models/APIs etc.
//
// If however you specify retainData==false. It deletes the old container before creating and
// launching a new one. This of course has no incumbent data, and so again the migration is
// performed with respect to the imaginary empty schema.
//
// Nb. The retainData=true mode is intended for the Keel Run command -  when the command comes up.
// Whereas the other mode is intended for automated tests, where we must make sure we don't use a database
// that contains arbitrary data from a previous test and likely a different schema.
func LaunchDB(schemaDir string, retainData bool) (gormDB *gorm.DB, schema *proto.Schema, err error) {
	sqlDB, err := keelpostgres.BringUpPostgresLocally(retainData)
	if err != nil {
		return nil, nil, err
	}
	gormDB, err = gorm.Open(gormpostgres.New(gormpostgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}
	if err := migrations.InitProtoSchemaStore(sqlDB); err != nil {
		return nil, nil, err
	}

	newSchema, err := migrations.DoMigrationBasedOnSchemaChanges(sqlDB, schemaDir)
	if err != nil {
		return nil, nil, err
	}
	return gormDB, newSchema, nil
}
