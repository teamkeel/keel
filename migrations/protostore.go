package migrations

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/teamkeel/keel/proto"
)

// InitProtoSchemaStore - sets up the storage for the last-known protobuf schema in the database (as JSON).
// The store comprises a dedicated, single-column table, that we expect to have just a single row.
// If the database is already set up the way we need it to be - by virtue of some previous run,
// the function is harmless and makes no changes.
func InitProtoSchemaStore(db *sql.DB) error {

	// Make the proto table if it does not already exist.
	if err := makeTableForLastKnownProto(db); err != nil {
		return fmt.Errorf("error making table for last known proto: %v", err)
	}

	// Populate the proto table with its one and only row if there are no rows yet.
	empty, err := isProtoTableEmpty(db)
	if err != nil {
		return err
	}
	if empty {
		if err := insertInitialProtoRow(db); err != nil {
			return err
		}
	}
	return nil
}

// SaveProtoToDb updates the database's store of the last known
// schema with the given schema.
func SaveProtoToDb(p *proto.Schema, db *sql.DB) (newProtoJSON string, err error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("could not save protobuf (json marshal): %v", err)
	}
	newProtoJSON = string(b)
	updateSQL := UpdateSingleStringColumn(tableForProtobuf, columnForTheJson, newProtoJSON)

	sqlResult, err := db.Exec(updateSQL)
	if err != nil {
		return "", fmt.Errorf("error saving last-known protobuf: %v", err)
	}
	_ = sqlResult
	return newProtoJSON, nil
}

// FetchProtoFromDb provides the last-known good schema from the database.
// I.e. the one that was used to determine the most recent database migrations.
func FetchProtoFromDb(db *sql.DB) (*proto.Schema, error) {
	theSQL := SelectSingleColumn(tableForProtobuf, columnForTheJson)
	row := db.QueryRow(theSQL)
	var theJSON string
	if err := row.Scan(&theJSON); err != nil {
		return nil, err
	}

	proto := proto.Schema{}
	if err := json.Unmarshal([]byte(theJSON), &proto); err != nil {
		return nil, fmt.Errorf("could not fetch protobuf from local storage (json unmarshal): %v", err)
	}
	return &proto, nil
}

// makeTableForLastKnownProto creates the database table for storing the last-known
// protobuf schema (as JSON) unless that table already exists.
func makeTableForLastKnownProto(db *sql.DB) error {
	fields := []*proto.Field{
		{
			Name: columnForTheJson,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_STRING,
			},
		},
	}
	sql := createTableIfNotExistsStmt(tableForProtobuf, fields)
	if _, err := db.Exec(sql); err != nil {
		return fmt.Errorf("error trying to create protobuf table: %v", err)
	}
	return nil
}

// isProtoTableEmpty returns true if the protobuf table has no rows.
func isProtoTableEmpty(db *sql.DB) (bool, error) {
	theSQL := SelectSingleColumn(tableForProtobuf, columnForTheJson)
	row := db.QueryRow(theSQL)
	var theJSON string
	err := row.Scan(&theJSON)
	switch err {
	case nil: // If it does not error it proves there is a row.
		return false, nil
	case sql.ErrNoRows: // If it tells us there are no rows - we have our answer.
		return true, nil
	default: // We hit some other error
		return false, err
	}
}

// insertInitialProtoRow populates the (empty) last-known protobuf table with a single row,
// that represents a proto.Schema with no contents.
func insertInitialProtoRow(db *sql.DB) error {
	schema := &proto.Schema{}
	theJSON, err := json.Marshal(schema)
	if err != nil {
		return nil
	}
	sql := InsertRowComprisingSingleString(tableForProtobuf, string(theJSON))
	if _, err := db.Exec(sql); err != nil {
		return err
	}
	return nil
}

const tableForProtobuf string = "_protobuf"
const columnForTheJson string = "json" // The column name.
