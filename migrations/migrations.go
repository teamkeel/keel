package migrations

import (
	"github.com/teamkeel/keel/proto"
	"gorm.io/gorm"
)

const (
	ChangeTypeAdded    = "ADDED"
	ChangeTypeRemoved  = "REMOVED"
	ChangeTypeModified = "MODIFIED"
)

type DatabaseChange struct {
	// The model this change applies to
	Model string

	// The field this change applies to (might be empty)
	Field string

	// The type of change
	Type string
}

type Migrations struct {
	// Describes the changes that will be applied to the database
	// if SQL is run
	Changes []*DatabaseChange

	// The SQL to run to execute the database schema changes
	SQL string

	db *gorm.DB
}

// Apply executes the migrations against the database
func (m *Migrations) Apply() error {
	return nil
}

// Create inspects the database using gorm.DB connection
// and creates the required schema migrations that will result in
// the database reflecting the provided proto.Schema
func Create(db *gorm.DB, schema *proto.Schema) *Migrations {
	return &Migrations{
		db: db,
	}
}
