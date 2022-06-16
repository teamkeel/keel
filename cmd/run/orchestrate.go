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
	fmt.Printf("Starting PostgreSQL")
	db, err := postgres.BringUpPostgresLocally()
	if err != nil {
		return fmt.Errorf("could not bring up postgres locally: %v", err)
	}

	// This sets up internal configuration and state in the database - only if it
	// has not been done in an earlier run. It means that all subsequent code can
	// then safely retreive the last-known protobuf used - which makes it simpler.
	if err := initDBIfNecessary(db); err != nil {
		return fmt.Errorf("error initialising the database: %v", err)
	}

	// We refresh the migrations as the command comes up, (before we start the watcher),
	// to make sure the database reflects the current user's schema.
	schemaDir, _ := cmd.Flags().GetString("dir")
	if err := doMigrationBasedOnSchemaChanges(db, schemaDir); err != nil {
		return err
	}

	// The run command remains passive now, until the user changes their schema, so we establish
	// a watcher on the schema directorty.
	directoryWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating schema change watcher: %v", err)
	}
	defer directoryWatcher.Close()

	handler := NewSchemaChangedHandler(schemaDir, db)
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
