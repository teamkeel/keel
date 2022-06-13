package cmd

import (
	"database/sql"
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
		return fmt.Errorf("could not bring up postgres locally: %v", err)
	}

	// If this is the first ever run - do the initial migration to the database, before
	// we drop into the watcher that reacts to live schema changes.
	firstRun, err := isFirstEverRun(inputDir)
	if err != nil {
		return fmt.Errorf("error assessing if first ever run: %v", err)
	}
	if firstRun {
		fmt.Printf("This is the first ever run, so performing initial database migration... ")
		err := performInitialDatabaseMigration(inputDir, db)
		if err != nil {
			return fmt.Errorf("error trying to perform initial database migrations: %v", err)
		}
		fmt.Printf("done\n")
	}

	directoryWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating schema change watcher: %v", err)
	}
	defer directoryWatcher.Close()

	handler := NewSchemaChangedHandler(db)
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
				// todo the Handle() function responds to internal errors by printing errors
				// to standard out and returning early.
				handler.Handle(nameOfFileThatChanged)
			}

		case err := <-watcher.Errors:
			fmt.Printf("XXXX error received on watcher error channel: %v\n", err)
			// Todo Bail out of the Run command if the watcher encounters an error.
			return
		}
	}
}

type SchemaChangedHandler struct {
	db *sql.DB
}

func NewSchemaChangedHandler(db *sql.DB) *SchemaChangedHandler {
	return &SchemaChangedHandler{
		db: db,
	}
}

func (h *SchemaChangedHandler) Handle(schemaThatHasChanged string) {
	fmt.Printf("Reacting to a change in this file: %s, changed\n", schemaThatHasChanged)

	// In the context of this Handler - we assume that the oldProto is available,
	// because we make sure it is before we set up the schema watcher.
	oldProto, err := proto.FetchFromLocalStorage(inputDir)
	if err != nil {
		fmt.Printf("error trying to retreive old protobuf: %v", err)
		return
	}
	fmt.Printf("Migrating your database to the changed schema... ")
	if err = performMigration(oldProto, h.db); err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Printf("done\n")
}

func performMigration(oldProto *proto.Schema, db *sql.DB) error {
	newProto, err := makeProtoFromSchemaFiles()
	if err != nil {
		return fmt.Errorf("error making proto from schema files: %v", err)
	}
	migrationsSQL, err := migrations.MakeMigrationsFromSchemaDifference(oldProto, newProto)
	if err != nil {
		return fmt.Errorf("could not generate SQL for migrations: %v", err)
	}
	_, err = db.Exec(migrationsSQL)
	if err != nil {
		return fmt.Errorf("error trying to perform database migration: %v", err)
	}
	if err := proto.SaveToLocalStorage(newProto, inputDir); err != nil {
		return fmt.Errorf("error trying to save the new protobuf: %v", err)
	}
	return nil
}

func makeProtoFromSchemaFiles() (proto *proto.Schema, err error) {
	builder := schema.Builder{}
	// todo - inputDir is a cmd package-global variable (because it is a CLI command flag), but we
	// should introduce a pass-by-value copy to pass down the call stack.
	proto, err = builder.MakeFromDirectory(inputDir)
	if err != nil {
		return nil, fmt.Errorf("error making protobuf schema from directory: %v", err)
	}
	return proto, nil
}

func isFirstEverRun(schemaDir string) (bool, error) {
	proto, err := proto.FetchFromLocalStorage(schemaDir)
	if err != nil {
		return false, fmt.Errorf("error trying to fetch last used protobuf: %v", err)
	}
	return proto == nil, nil
}

func performInitialDatabaseMigration(inputDir string, db *sql.DB) error {
	// We re-use the performMigration() function, but have to provide it with
	// a valid, but completely empty schema.
	emptySchema := &proto.Schema{}
	if err := performMigration(emptySchema, db); err != nil {
		return fmt.Errorf("error: %v\n", err)
	}
	return nil
}
