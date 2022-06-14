package run

import (
	"database/sql"
	"fmt"

	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
)

// performMigration performs the migration deemed to be necessary to the database
// given the last-used schema, and a (presumed to have changed) user's keel schema directory.
// It is good for both the incremental migrations needed as the user works on their input schema,
// but also for the initial, first-ever migration. In the latter case - the caller must provide
// a valid, but empty last-used schema.
func performMigration(oldProto *proto.Schema, db *sql.DB, schemaDir string) error {
	newProto, err := makeProtoFromSchemaFiles(schemaDir)
	if err != nil {
		return fmt.Errorf("error making proto from schema files: %v", err)
	}
	migrationsSQL, err := migrations.MakeMigrationsFromSchemaDifference(oldProto, newProto)
	if err != nil {
		return fmt.Errorf("could not generate SQL for migrations: %v", err)
	}
	_, err = db.Exec(migrationsSQL)
	if err != nil {
		return fmt.Errorf("error trying to perform database migration: %v", err)
	}
	if err := proto.SaveToLocalStorage(newProto, schemaDir); err != nil {
		return fmt.Errorf("error trying to save the new protobuf: %v", err)
	}
	return nil
}

func makeProtoFromSchemaFiles(schemaDir string) (proto *proto.Schema, err error) {
	builder := schema.Builder{}
	proto, err = builder.MakeFromDirectory(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("error making protobuf schema from directory: %v", err)
	}
	return proto, nil
}

func isFirstEverRun(schemaDir string) (bool, error) {
	proto, err := proto.FetchFromLocalStorage(schemaDir)
	if err != nil {
		return false, fmt.Errorf("error trying to fetch last used protobuf: %v", err)
	}
	return proto == nil, nil
}
