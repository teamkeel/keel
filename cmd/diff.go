package cmd

import (
	"context"
	"fmt"
	"github.com/teamkeel/keel/cmd/database"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/schema"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Read database schema, construct the schema and diff the two",
	Run: func(cmd *cobra.Command, args []string) {
		schemaDir, _ := cmd.Flags().GetString("dir")
		b := &schema.Builder{}

		protoSchema, err := b.MakeFromDirectory(schemaDir)
		if err != nil {
			fmt.Println("schema has errors")
			return
		}

		dbConn, _, err := database.Start(true)
		if err != nil {
			fmt.Println("failed to connect to database")
			return
		}
		logger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  logger.Info, // Log level
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
				Colorful:                  true,        // Disable color
			},
		)

		db, err := gorm.Open(postgres.New(postgres.Config{
			Conn: dbConn,
		}), &gorm.Config{
			Logger: logger,
		})
		if err != nil {
			fmt.Println("failed to connect to database")
			return
		}

		currSchema, err := migrations.GetCurrentSchema(context.Background(), db)
		if err != nil {
			fmt.Println("failed to get database schema")
			return
		}

		m := migrations.New(protoSchema, currSchema)

		if len(m.Changes) == 0 {
			fmt.Println("No changes")
			return
		}

		fmt.Println("Changes: ")
		for _, databaseChange := range m.Changes {
			lineStr := databaseChange.Model
			if databaseChange.Field != "" {
				lineStr += " " + databaseChange.Field
			}
			lineStr += databaseChange.Type
			fmt.Printf("* %s\n", lineStr)
		}
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
