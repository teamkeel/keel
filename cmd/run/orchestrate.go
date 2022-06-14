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

	schemaDir, _ := cmd.Flags().GetString("dir")

	// If this is the first ever run - do the initial migrations on the database.
	isFirstRun, err := isFirstEverRun(db)
	if err != nil {
		return fmt.Errorf("error while assessing if first ever run: %v", err)
	}
	if isFirstRun {
		fmt.Printf("This is the first ever run, so performing initial database migration... ")
		if err := performFirstEverMigration(db, schemaDir); err != nil {
			return fmt.Errorf("error trying to perform initial database migrations: %v", err)
		}
		fmt.Printf("done\n")
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
	// resources are release when the command terminates (with CTRL-C).
	go reactToSchemaChanges(directoryWatcher, handler)

	// The watcher documentation suggests we tell the watcher about the directory to watch,
	// AFTER we have constructed it, and registered a handler.
	err = directoryWatcher.Add(schemaDir)
	if err != nil {
		return fmt.Errorf("error specifying directory to schema watcher: %v", err)
	}

	fmt.Printf("Waiting for a schema file to change in %s ...\n", schemaDir)

	// Block the main go routine to keep the process alive until the user kills it with CTRL-C.
	ch := make(chan bool)
	<-ch

	return nil
}
