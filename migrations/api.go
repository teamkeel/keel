package migrations

import (
	"database/sql"
	"fmt"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
)

// DoMigrationBasedOnSchemaChanges is a high-level orchestration function that wraps all the steps
// necessary to migrate the given database to match the Keel Schema in the given schema directory. It assumes that the
// last known-good protobuf schema is available in the database.
func DoMigrationBasedOnSchemaChanges(db *sql.DB, schemaDir string) (newProtoJSON string, err error) {
	oldProto, err := FetchProtoFromDb(db)
	if err != nil {
		fmt.Printf("error trying to retreive old protobuf: %v", err)
		return "", err
	}
	fmt.Printf("Migrating your database to the latest schema... ")
	newProtoJSON, err = performMigration(oldProto, db, schemaDir)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return "", err
	}
	fmt.Printf("done\n")
	return newProtoJSON, nil
}

// performMigration performs the migration deemed to be necessary to the database
// given the last-used schema, and a (presumed to have changed) user's keel schema directory.
// It is good for both the incremental migrations needed as the user works on their input schema,
// but also for the initial, first-ever migration. In the latter case - the caller should pass
// &proto.Schema{} as the oldProto argument.
func performMigration(oldProto *proto.Schema, db *sql.DB, schemaDir string) (newProtoJSON string, err error) {
	builder := schema.Builder{}
	newProto, err := builder.MakeFromDirectory(schemaDir)
	if err != nil {
		return "", fmt.Errorf("error making protobuf schema from directory: %v", err)
	}
	migrationsSQL, err := MakeMigrationsFromSchemaDifference(oldProto, newProto)
	if err != nil {
		return "", fmt.Errorf("could not generate SQL for migrations: %v", err)
	}
	_, err = db.Exec(migrationsSQL)
	if err != nil {
		return "", fmt.Errorf("error trying to perform database migration: %v", err)
	}
	newProtoJSON, err = SaveProtoToDb(newProto, db)
	if err != nil {
		return "", fmt.Errorf("error trying to save the new protobuf: %v", err)
	}
	return newProtoJSON, nil
}
