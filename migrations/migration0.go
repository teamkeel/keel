package migrations

import "github.com/teamkeel/keel/proto"

// A Migration0 object knows how to generate the SQL for the inaugral
// migration of the database. This SQL starts with a DROP DATABASE query
// so that it can proceed to recreate it from scratch.
type Migration0 struct {
	schema *proto.Schema
}

func NewMigrationZeroMaker(schema *proto.Schema) *Migration0 {
	return &Migration0{
		schema: schema,
	}
}

func (m0 *Migration0) MakeSQL() (theSQL []string, err error) {
	return nil, nil
}
