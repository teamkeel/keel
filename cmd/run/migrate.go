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
// but also for the initial, first-ever migration. In the latter case - the caller should pass
// &proto.Schema{} as the oldProto argument.
func performMigration(oldProto *proto.Schema, db *sql.DB, schemaDir string) (newProtoJSON string, err error) {
	newProto, err := makeProtoFromSchemaFiles(schemaDir)
	if err != nil {
		return "", fmt.Errorf("error making proto from schema files: %v", err)
	}
	migrationsSQL, err := migrations.MakeMigrationsFromSchemaDifference(oldProto, newProto)
	if err != nil {
		return "", fmt.Errorf("could not generate SQL for migrations: %v", err)
	}
	_, err = db.Exec(migrationsSQL)
	if err != nil {
		return "", fmt.Errorf("error trying to perform database migration: %v", err)
	}
	newProtoJSON, err = saveProtoToDb(newProto, db)
	if err != nil {
		return "", fmt.Errorf("error trying to save the new protobuf: %v", err)
	}
	return newProtoJSON, nil
}

func makeProtoFromSchemaFiles(schemaDir string) (*proto.Schema, error) {
	builder := schema.Builder{}
	proto, err := builder.MakeFromDirectory(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("error making protobuf schema from directory: %v", err)
	}
	return proto, nil
}

// doMigrationBasedOnSchemaChanges is a thin wrapper that fetches
// the last-known schema protobuf from the database, before delegating
// the performance of a schema-difference based migration to another function.
func doMigrationBasedOnSchemaChanges(db *sql.DB, schemaDir string) (newProtoJSON string, err error) {
	// This function assumes that the oldProto is available in the database.
	oldProto, err := fetchProtoFromDb(db)
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
