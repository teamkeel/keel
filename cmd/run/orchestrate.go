package run

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/localdb"

	"github.com/fsnotify/fsnotify"
)

// CommandImplementation is the main call-to-action function for the run command.
// Its responsibility is to orchestrate the command's behaviour.
func CommandImplementation(cmd *cobra.Command, args []string) (err error) {

	schemaDir, _ := cmd.Flags().GetString("dir")

	sqlDB, gormDB, protoSchemaJSON, err := localdb.BringUpLocalDBToMatchSchema(schemaDir)

	handler := NewSchemaChangedHandler(schemaDir, sqlDB, gormDB)
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

	// Todo - we must not forget housekeeping on close...
	//
	// - the dockerized database
	// - the GraphQL API server

	return nil
}
