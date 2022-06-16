package run

import (
	"database/sql"
	"fmt"
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
	fmt.Printf("File changed: %s\n", schemaThatHasChanged)

	// Deliberately ignoring errors returned by this call in the context of
	// the handler, because they get printed when they occur, and because the handler
	// is called from the watching goroutine, it has nothing to return them to.
	doMigrationBasedOnSchemaChanges(h.db, h.schemaDir)
}

func isRelevantEventType(op fsnotify.Op) bool {
	relevant := []fsnotify.Op{fsnotify.Create, fsnotify.Write, fsnotify.Remove}
	// The irrelevant ones are Rename and Chmod.
	return lo.Contains(relevant, op)
}
