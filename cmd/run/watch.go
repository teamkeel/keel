package run

import (
	"database/sql"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/teamkeel/keel/proto"
)

// reactToSchemaChanges should be called in its own goroutine. It has a blocking
// channel select loop that waits for and receives file system events, or errors.
func reactToSchemaChanges(watcher *fsnotify.Watcher, handler *SchemaChangedHandler) {
	for {
		select {
		case event := <-watcher.Events:
			// The watcher documentation shows how we could inspect the event to discriminate
			// between {create/write/remove/rename/chmod} - but since all of these seem to
			// be valid triggers - we do not.
			nameOfFileThatChanged := event.Name
			// Handle() responds to errors occuring by printing the details
			// to standard out and returning early.
			handler.Handle(nameOfFileThatChanged)

		case err := <-watcher.Errors:
			fmt.Printf("XXXX error received on watcher error channel: %v\n", err)
			// If we get an internal error from the watcher - we simply report the details
			// and allow the watching to continue. We leave it to the user to decide if
			// they want to quit the run command.
		}
	}
}

// A SchemaChangedHandler knows how to react to changes taking place in a schema directory.
type SchemaChangedHandler struct {
	db        *sql.DB
	schemaDir string
}

// NewSchemaChangedHandler provides a SchemaChangedHandler ready to use.
func NewSchemaChangedHandler(schemaDir string, db *sql.DB) *SchemaChangedHandler {
	return &SchemaChangedHandler{
		db:        db,
		schemaDir: schemaDir,
	}
}

// Handle is the callback function that receives change events.
func (h *SchemaChangedHandler) Handle(schemaThatHasChanged string) {
	fmt.Printf("Reacting to a change in this file: %s, changed\n", schemaThatHasChanged)

	// In the context of this Handler - we assume that the oldProto must be available,
	// because we make sure it is before we set up the schema watcher.
	oldProto, err := proto.FetchFromLocalStorage(h.schemaDir)
	if err != nil {
		fmt.Printf("error trying to retreive old protobuf: %v", err)
		return
	}
	fmt.Printf("Migrating your database to the changed schema... ")
	if err = performMigration(oldProto, h.db, h.schemaDir); err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Printf("done\n")
}
