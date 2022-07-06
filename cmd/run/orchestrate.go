package run

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/postgres"

	"github.com/fsnotify/fsnotify"
)

// CommandImplementation is the main call-to-action function for the run command.
// Its responsibility is to orchestrate the command's behaviour.
func CommandImplementation(cmd *cobra.Command, args []string) (err error) {
	fmt.Printf("Starting PostgreSQL\n")
	db, err := postgres.BringUpPostgresLocally()
	if err != nil {
		return fmt.Errorf("could not bring up postgres locally: %v", err)
	}

	// This sets up internal configuration and state in the database - only if it
	// has not been done in an earlier run. It means that all subsequent code can
	// then safely retreive the last-known protobuf used - which makes the code simpler.
	if err := initDBIfNecessary(db); err != nil {
		return fmt.Errorf("error initialising the database: %v", err)
	}

	// Before we enter schema-watching mode, we have to consider that this might be
	// the first run ever, or the user's schema may have changed on disk *after* they last
	// launched the Run command. So we do a migration just in case, and (providing the migration was
	// successful, we also bring up their GraphQL API server representing the current state of the
	// schema.
	schemaDir, _ := cmd.Flags().GetString("dir")
	protoSchemaJSON, err := doMigrationBasedOnSchemaChanges(db, schemaDir)
	if err != nil {
		return err
	}
	handler := NewSchemaChangedHandler(schemaDir, db)
	if err := handler.retartAPIServer(protoSchemaJSON); err != nil {
		return err
	}

	// The run command remains quiescent now, until the user changes their schema, so we establish
	// a watcher on the schema directorty.
	directoryWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating schema change watcher: %v", err)
	}
	defer directoryWatcher.Close()

	// goroutine housekeeping note: This goroutine lives for as long as the Keel-Run command is running, and its
	// resources are released when the command terminates (with CTRL-C).
	go reactToSchemaChanges(directoryWatcher, handler)

	// The watcher documentation suggests we tell the watcher about the directory to watch,
	// AFTER we have constructed it, and registered a handler.
	err = directoryWatcher.Add(schemaDir)
	if err != nil {
		return fmt.Errorf("error specifying directory to schema watcher: %v", err)
	}

	fmt.Printf("Waiting for schema files to change in %s ...\n", schemaDir)
	fmt.Printf("Press CTRL-C to exit\n")

	// Block the main go routine to keep the process alive until the user kills it with CTRL-C.
	ch := make(chan bool)
	<-ch

	return nil
}
