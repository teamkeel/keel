package migrations

import (
	"context"
	_ "embed"

	"github.com/lib/pq"
	"github.com/teamkeel/keel/db"
)

func getConstraints(ctx context.Context, database db.Database) ([]*ContraintRow, error) {
	rows := []*ContraintRow{}
	return rows, database.GetDB().Raw(constraintsQuery).Scan(&rows).Error
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

type ContraintRow struct {
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
