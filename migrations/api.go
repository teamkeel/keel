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
func DoMigrationBasedOnSchemaChanges(db *sql.DB, schemaDir string) (newSchema *proto.Schema, err error) {
	oldProto, err := FetchProtoFromDb(db)
	if err != nil {
		fmt.Printf("error trying to retreive old protobuf: %v", err)
		return nil, err
	}
	fmt.Printf("Migrating your database to the latest schema\n")
	newSchema, err = PerformMigration(oldProto, db, schemaDir)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}
	if _, err := SaveProtoToDb(newSchema, db); err != nil {
		return nil, err
	}
	return newSchema, nil
}

// PerformMigration performs the migration deemed to be necessary to the database
// given the last-used schema, and a (presumed to have changed) user's keel schema directory.
// It is good for both the incremental migrations needed as the user works on their input schema,
// but also for the initial, first-ever migration. In the latter case - the caller should pass
// &proto.Schema{} as the oldProto argument.
func PerformMigration(oldProto *proto.Schema, db *sql.DB, schemaDir string) (newSchema *proto.Schema, err error) {
	builder := schema.Builder{}
	newSchema, err = builder.MakeFromDirectory(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("error making protobuf schema from directory: %v", err)
	}
	migrationsSQL, err := MakeMigrationsFromSchemaDifference(oldProto, newSchema)
	if err != nil {
		return nil, fmt.Errorf("could not generate SQL for migrations: %v", err)
	}
	_, err = db.Exec(migrationsSQL)
	if err != nil {
		return nil, fmt.Errorf("error trying to perform database migration: %v", err)
	}
	return newSchema, nil
}

// PerformInitialMigration performs an initial migration on the database to make it
// match the given schema. It assumes the database has no tables in to start with, and
// should not be used unless this is known.
func PerformInitialMigration(db *sql.DB, schema *proto.Schema) error {
	oldSchema := &proto.Schema{}
	migrationsSQL, err := MakeMigrationsFromSchemaDifference(oldSchema, schema)
	if err != nil {
		return fmt.Errorf("could not generate SQL for migrations: %v", err)
	}
	_, err = db.Exec(migrationsSQL)
	if err != nil {
		return fmt.Errorf("error trying to perform database migration: %v", err)
	}
	return nil
}
