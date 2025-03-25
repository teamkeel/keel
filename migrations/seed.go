package migrations

import (
	"context"
	"fmt"
	"os"
)

func (m *Migrations) ApplySeedData(ctx context.Context, files []string) error {
	ctx, span := tracer.Start(ctx, "Apply Seed Data")
	defer span.End()

	// TODO should we run all files in a single transaction?

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
