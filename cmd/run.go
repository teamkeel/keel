package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/formatter"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/postgres"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"

	"github.com/fsnotify/fsnotify"
)

// The Run command does this:
//
// - Starts Postgres locally in a docker container.
// - Setting up a watcher on the input schema directory with a handler that
//   reacts to changes as follows...
//
// 		- Parse and validate the input schema files.
// 		- Build the protobuffer schema representation.
// 		- Generate the SQL to completely remove the existing database and rebuild it
//        from scratch (migration0)
// 		- Perform this migration on the running database.
//
// TODOs these are the major functional todos for the migrations-only first cut...
//
// - How to trigger the database recreation at boot time
// - Stop it making a new postgres docker image every time
// - Clean up when the command terminates (stop postgres)
// - Proper error handling strategy
//
// TODOs these will be the next steps beyond the migrations-only version.
//
// - Auto generate the code to implement the service (GraphQL service)
// - Build the executable service
// - Kill the old version and bring up the new version.
type runCommand struct {
	outputFormatter *formatter.Output
}

var cobraCommandWrapper = &cobra.Command{
	Use:   "run",
	Short: "Run your Keel App locally",
	RunE:  commandImplementation,
}

func commandImplementation(cmd *cobra.Command, args []string) (err error) {
	c := &runCommand{
		outputFormatter: formatter.New(os.Stdout),
	}
	// todo - think this takes a default value, so we can probably
	// not set it up for this case
	switch outputFormat {
	case string(formatter.FormatJSON):
		c.outputFormatter.SetOutput(formatter.FormatJSON, os.Stdout)
	default:
		c.outputFormatter.SetOutput(formatter.FormatText, os.Stdout)
	}

	c.outputFormatter.Write("Starting PostgreSQL")
	db, err := postgres.BringUpPostgresLocally()
	if err != nil {
		return err
	}

	_ = db

	directoryWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating schema change watcher: %v", err)
	}
	defer directoryWatcher.Close()

	handler := NewSchemaChangedHandler()
	// goroutine housekeeping note: This goroutine lives for as long as the Keel-Run command is running, and its
	// resources are release when the command terminates (with CTRL-C).
	go c.reactToSchemaChanges(directoryWatcher, handler)

	// todo: I hate that this is consuming a package-global variable <inputDir>
	// but that's how Cobra command flags are exposed.
	err = directoryWatcher.Add(inputDir)
	if err != nil {
		return fmt.Errorf("error specifying directory to schema watcher: %v", err)
	}

	c.outputFormatter.Write(fmt.Sprintf("Waiting for a schema file to change in %s ...\n", inputDir))

	// Block the main go routine to keep the process alive until the user kills it with CTRL-C.
	ch := make(chan bool)
	<-ch

	// Todo - do some resource housekeeping when the command exits:
	// Do a db.Close() on the database connection
	// Stop the postgres container
	//
	// - It would be good if we had a kill signal handler mechanism at top level
	//   cobra command level.

	return nil
}

func init() {
	rootCmd.AddCommand(cobraCommandWrapper)
	defaultDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("os.Getwd() errored: %v", err))
	}
	// The Keel Run command works by observing a directory, and therefor does not offer a single-file command
	// line flag.
	cobraCommandWrapper.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "schema directory to run")
	cobraCommandWrapper.Flags().StringVarP(&outputFormat, "output", "o", "console", "output format (console, json)")
}

func (c *runCommand) reactToSchemaChanges(watcher *fsnotify.Watcher, handler *SchemaChangedHandler) {
	for {
		select {
		case event := <-watcher.Events:
			// todo horrid hidden use of bitwise and
			if event.Op&fsnotify.Write == fsnotify.Write {
				nameOfFileThatChanged := event.Name
				if err := handler.Handle(nameOfFileThatChanged); err != nil {
					panic(fmt.Errorf("error handling schema change event: %v", err))
				}
			}

		case err := <-watcher.Errors:
			fmt.Printf("XXXX error received on watcher error channel: %v\n", err)
			// Todo Bail out of the Run command if the watcher encounters an error.
			return
		}
	}
}

type SchemaChangedHandler struct{}

func NewSchemaChangedHandler() *SchemaChangedHandler {
	return &SchemaChangedHandler{}
}

func (h *SchemaChangedHandler) Handle(schemaThatHasChanged string) (err error) {
	// todo - feed these user feedback messages through the command's managed formatter.
	fmt.Printf("Reacting to a change in this file: %s, changed\n", schemaThatHasChanged)
	var newProto *proto.Schema
	if newProto, err = makeProtoFromSchemaFiles(); err != nil {
		return fmt.Errorf("error making proto from schema files: %v", err)
	}

	// This is currently just generating the virgin migration SQL, but
	// once that works we'll switch to a schema-delta based migration.

	generatedSQL := migrations.GenerateAllTables(newProto.Models)

	_ = generatedSQL

	// Todo now apply these migrations
	return nil
}

func makeProtoFromSchemaFiles() (proto *proto.Schema, err error) {
	builder := schema.Builder{}
	// todo - inputDir is a cmd package-global variable (because it is a CLI command flag), but we
	// should introduce a pass-by-value copy to pass down the call stack.
	proto, err = builder.MakeFromDirectory(inputDir)
	if err != nil {
		panic(fmt.Sprintf("error making protobuf schema from directory: %v", err))
	}
	return proto, nil
}
