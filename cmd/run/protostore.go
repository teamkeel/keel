package run

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/teamkeel/keel/proto"
)

func saveProtoToDb(p *proto.Schema, db *sql.DB) error {
	b, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("could not save protobuf (json marshal): %v", err)
	}
	updateSQL := makeUpdateSQL(string(b))

	if _, err := db.Exec(updateSQL); err != nil {
		return fmt.Errorf("error saving last-known protobuf: %v", err)
	}
	return nil
}

// fetchProtoFromDb provides the last-known good schema from the database.
// I.e. the one that was used to determine the most recent database migrations.
// When this data is not yet stored - it signals that by returning a nil schema, but no error.
func fetchProtoFromDb(db *sql.DB) (*proto.Schema, error) {
	querySQL := makeQuerySQL()
	row := db.QueryRow(querySQL)
	var theJSON string
	err := row.Scan(&theJSON)
	if err != nil {
		// We deliberately interpret the query failing, as implying that the table does
		// not yet exist. There could be other reasons for the query failing, which we will
		// miss right now. But they will be caught later anyhow.
		return nil, nil
	}

	proto := proto.Schema{}
	if err := json.Unmarshal([]byte(theJSON), &proto); err != nil {
		return nil, fmt.Errorf("could not fetch protobuf from local storage (json unmarshal): %v", err)
	}
	return &proto, nil
}

func initializeRowForLastKnownProto() string {
	return fmt.Sprintf("INSERT INTO %s VALUES ('placeholder');", tableForProtobuf)
}

func makeUpdateSQL(theJSON string) string {
	output := fmt.Sprintf("UPDATE %s SET %s='%s';", tableForProtobuf, columnForTheJson, theJSON)
	return output
}

func makeQuerySQL() string {
	return fmt.Sprintf("SELECT %s FROM %s;", columnForTheJson, tableForProtobuf)
}
