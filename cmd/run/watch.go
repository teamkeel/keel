package run

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/samber/lo"
)

// reactToSchemaChanges should be called in its own goroutine. It has a blocking
// channel select loop that waits for and receives file system events, or errors.
func reactToSchemaChanges(watcher *fsnotify.Watcher, handler *SchemaChangedHandler) {
	for {
		select {
		case event := <-watcher.Events:
			switch {
			case !strings.HasSuffix(event.Name, ".keel"):
				// Ignore
			case !isRelevantEventType(event.Op):
				// Ignore
			default:
				nameOfFileThatChanged := event.Name
				handler.Handle(nameOfFileThatChanged)
			}

		case err := <-watcher.Errors:
			fmt.Printf("error received on watcher error channel: %v\n", err)
			// If we get an internal error from the watcher - we simply report the details
			// and allow the watching to continue. We leave it to the user to decide if
			// they want to quit the run command.
		}
	}
}

// A SchemaChangedHandler knows how to react to changes taking place in a schema directory.
type SchemaChangedHandler struct {
	db *sql.DB

	// We retain a reference to the server, to support a later call to svr.Shutdown().
	apiServer *http.Server
	schemaDir string
}

// NewSchemaChangedHandler provides a SchemaChangedHandler ready to use.
func NewSchemaChangedHandler(schemaDir string, db *sql.DB) *SchemaChangedHandler {
	return &SchemaChangedHandler{
		db:        db,
		schemaDir: schemaDir,
	}
}

// Handle is the callback function that receives notifications that a schema file has changed.
func (h *SchemaChangedHandler) Handle(schemaThatHasChanged string) {
	fmt.Printf("File changed: %s\n", schemaThatHasChanged)

	// Migrate the database to the changed schema
	newSchemaJSON, err := doMigrationBasedOnSchemaChanges(h.db, h.schemaDir)
	if err != nil {
		fmt.Printf("error: database migrations failed with error: %v", err)
		return
	}

	// And restart the API server
	if err = h.retartAPIServer(newSchemaJSON); err != nil {
		fmt.Printf("error: could not restart your API server: %v", err)
		return
	}
}

func isRelevantEventType(op fsnotify.Op) bool {
	relevant := []fsnotify.Op{fsnotify.Create, fsnotify.Write, fsnotify.Remove}
	// The irrelevant ones are Rename and Chmod.
	return lo.Contains(relevant, op)
}
