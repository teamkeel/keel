package migrations

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// ApplySeedData executes all SQL files in the seed directory in lexicographic order
// If dryRun is true, then the changes are rolled back
func (m *Migrations) ApplySeedData(ctx context.Context, files []string) error {
	ctx, span := tracer.Start(ctx, "Apply Seed Data")
	defer span.End()

	// TODO run in a transaction

	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading seed file %s: %w", file, err)
		}

		if _, err := m.database.ExecuteStatement(ctx, string(contents)); err != nil {
			return fmt.Errorf("error executing seed file %s: %w", file, err)
		}
	}

	return nil
}

func (m *Migrations) SnapshotDatabase(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "Snapshot Database")
	defer span.End()

	var output strings.Builder
	models := m.Schema.ModelNames()

	for i, model := range models {
		table := Identifier(model)

		result, err := m.database.ExecuteQuery(ctx, fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			return "", fmt.Errorf("error querying table %s: %w", table, err)
		}

		if len(result.Rows) == 0 {
			continue
		}

		if i > 0 {
			output.WriteString("\n")
		}

		insertStmt := fmt.Sprintf("INSERT INTO %s (\n    %s\n) VALUES\n",
			table,
			strings.Join(result.Columns, ",\n    "),
		)
		output.WriteString(insertStmt)

		valueTuples := make([]string, len(result.Rows))

		for i, row := range result.Rows {
			values := make([]string, len(result.Columns))

			for j, col := range result.Columns {
				val := row[col]
				switch v := val.(type) {
				case string:
					values[j] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
				case nil:
					values[j] = "NULL"
				case time.Time:
					values[j] = fmt.Sprintf("'%s'", v.Format(time.RFC3339Nano))
				default:
					values[j] = fmt.Sprintf("%v", v)
				}
			}
			valueTuples[i] = fmt.Sprintf("    (%s)", strings.Join(values, ", "))
		}

		output.WriteString(strings.Join(valueTuples, ",\n"))
		output.WriteString(";\n")
	}

	return output.String(), nil
}
