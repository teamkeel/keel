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
	if err := saveProtoToDb(newProto, db); err != nil {
		return fmt.Errorf("error trying to save the new protobuf: %v", err)
	}
	return nil
}

func makeProtoFromSchemaFiles(schemaDir string) (*proto.Schema, error) {
	builder := schema.Builder{}
	proto, err := builder.MakeFromDirectory(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("error making protobuf schema from directory: %v", err)
	}
	return proto, nil
}

func isFirstEverRun(db *sql.DB) (bool, error) {
	proto, err := fetchProtoFromDb(db)
	if err != nil {
		return false, fmt.Errorf("error trying to fetch last used protobuf: %v", err)
	}
	return proto == nil, nil
}

// performFirstEverMigration is a minor variation to the sister method performMigration.
// It differs does not depend on a last-known protobuf input, and instead uses a
// valid, but empty protobuf to drive the standard deltas-based migration process.
// It also (uniquely) creates the table in which we store the last-known protobuf
// for subsequent migrations.
func performFirstEverMigration(db *sql.DB, schemaDir string) error {

	// Make a table to store the last-known protobuf.
	sql := makeTableForLastKnownProto()
	if _, err := db.Exec(sql); err != nil {
		return err
	}

	sql = initializeRowForLastKnownProto()
	if _, err := db.Exec(sql); err != nil {
		return err
	}

	// Give the db a structure that represents the user's schema.
	dummyLastKnownProto := &proto.Schema{}
	if err := performMigration(dummyLastKnownProto, db, schemaDir); err != nil {
		return err
	}
	return nil
}

// makeTableForLastKnownProto generates the SQL to make a database table
// suitable for storing the last-known protobuf schema (as JSON)
func makeTableForLastKnownProto() string {
	output := fmt.Sprintf("CREATE TABLE %s(\n", tableForProtobuf)
	f := fmt.Sprintf("%s %s\n", columnForTheJson, "TEXT")
	output += f
	output += ");"
	return output
}

const tableForProtobuf string = "_protobuf"
const columnForTheJson string = "json" // The column name.
