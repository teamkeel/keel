package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/run"
)

// The Run command does this:
//
// - Starts Postgres locally in a docker container.
// - If thus the first run ever, then perform initial database migrations.
// - Setting up a watcher on the input schema directory with a handler that
//   reacts to changes as follows...
//
// 		- Parse and validate the input schema files.
// 		- Build the protobuffer schema representation.
// 		- Analyse the differences between the new and previous schema
//		- Generate the database migration SQL required
// 		- Perform this migration on the running database.
//
// TODOs these are the major functional todos for the migrations-only first cut...
//
// - Clean up when the command terminates (stop postgres)
// - Proper error handling and user feedback strategy
//
// TODOs these will be the next steps beyond the migrations-only version.
//
// - Auto generate the code to implement the service (GraphQL service)
// - Build the executable service
// - Kill the old AP and bring up the new version.

var cobraCommandWrapper = &cobra.Command{
	Use:   "run",
	Short: "Run your Keel App locally",
	RunE:  run.CommandImplementation,
}

func init() {
	rootCmd.AddCommand(cobraCommandWrapper)
	defaultDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("os.Getwd() errored: %v", err))
	}
	cobraCommandWrapper.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "schema directory to run")
}
