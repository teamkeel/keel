package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/run"
)

// The Run command does this:
//
// - Starts Postgres locally in a docker container.
// - If this is the first run ever, then:
//		-  perform initial database migrations
//		-  generate and start a GraphQL server that embodies the APIs in the schema
// - Setting up a watcher on the input schema directory with a handler that
//   reacts to changes as follows:
// 		- Parse and validate the input schema files.
// 		- Build the protobuffer schema representation.
// 		- Analyse the differences between the new and previous schema
//		- Generate the database migration SQL required
// 		- Perform this migration on the running database.
//      - Restarts the GraphQL API server

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
