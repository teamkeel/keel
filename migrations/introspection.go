package migrations

import (
	"context"
	_ "embed"

	"github.com/lib/pq"
	"github.com/teamkeel/keel/db"
)

func getConstraints(ctx context.Context, database db.Database) ([]*ConstraintRow, error) {
	rows := []*ConstraintRow{}
	return rows, database.GetDB().Raw(constraintsQuery).Scan(&rows).Error
}

func getTriggers(ctx context.Context, database db.Database) ([]*TriggerRow, error) {
	rows := []*TriggerRow{}
	return rows, database.GetDB().Raw(triggersQuery).Scan(&rows).Error
}

func getColumns(ctx context.Context, database db.Database) ([]*ColumnRow, error) {
	rows := []*ColumnRow{}
	return rows, database.GetDB().Raw(columnsQuery).Scan(&rows).Error
}

var (
	//go:embed columns.sql
	columnsQuery string

	//go:embed constraints.sql
	constraintsQuery string

	//go:embed triggers.sql
	triggersQuery string
)

type ColumnRow struct {
	TableName    string `json:"table_name"`
	ColumnName   string `json:"column_name"`
	ColumnNum    int    `json:"column_num"`
	NotNull      bool   `json:"not_null"`
	HasDefault   bool   `json:"has_default"`
	DefaultValue string `json:"default_value"`
	DataType     string `json:"data_type"`
}

type ConstraintRow struct {
	TableName          string
	ConstraintName     string
	ConstrainedColumns pq.Int64Array `gorm:"type:smallint[]"`

	// If a foreign key constraint the referenced table and columns
	OnTable           *string
	ReferencesColumns pq.Int64Array `gorm:"type:smallint[]"`

	// c = check constraint,
	// f = foreign key constraint,
	// p = primary key constraint,
	// u = unique constraint,
	// t = constraint trigger,
	// x = exclusion constraint
	ConstraintType string

	// a = no action
	// r = restrict
	// c = cascade
	// n = set null
	// d = set default
	OnDelete string
}

type TriggerRow struct {
	// company_employee_delete
	TriggerName string `json:"trigger_name"`
	// e.g. company_employee
	TableName string `json:"table_name"`
	// e.g. DELETE
	StatementType string `json:"statement_type"`
	// e.g. EXECUTE PROCEDURE process_audit()
	ActionStatement string `json:"action_statement"`
	// e.g. AFTER
	ActionTiming string `json:"action_timing"`
}
