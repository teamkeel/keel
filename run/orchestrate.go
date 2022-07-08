package run

import (
	"fmt"

	"github.com/spf13/cobra"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/fsnotify/fsnotify"
	"github.com/teamkeel/keel/migrations"
	keelpostgres "github.com/teamkeel/keel/postgres"
)

// CommandImplementation is the main call-to-action function for the run command.
// Its responsibility is to orchestrate the command's behaviour.
func CommandImplementation(cmd *cobra.Command, args []string) (err error) {

	schemaDir, _ := cmd.Flags().GetString("dir")

	// Todo - the error handling here, means that if you launch the Run command with
	// an invalid schema - it bombs out. Whereas before it survived and just waited
	// in the watching loop (below) for the schema to become valid.
	// Need to decide if it's ok to bomb out in this situation.
	gormDB, protoSchemaJSON, err := bringUpLocalDBToMatchSchema(schemaDir)
	if err != nil {
		return err
	}

	handler := NewSchemaChangedHandler(schemaDir, gormDB)
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

// bringUpLocalDBToMatchSchema brings up a local, dockerised PostgresSQL database,
// that is fully migrated to match the given Keel Schema. It re-uses the incumbent
// container if it can (including therefore the incumbent database state), but also works
// if it has to do everything from scratch - including fetching the PostgreSQL image.
//
// It is good to use for the Keel Run command, but also to use in test fixtures.
func bringUpLocalDBToMatchSchema(schemaDir string) (gormDB *gorm.DB, protoSchemaJSON string, err error) {
	sqlDB, err := keelpostgres.BringUpPostgresLocally()
	if err != nil {
		return nil, "", err
	}
	gormDB, err = gorm.Open(gormpostgres.New(gormpostgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, "", err
	}
	if err := migrations.InitProtoSchemaStore(sqlDB); err != nil {
		return nil, "", err
	}

	protoSchemaJSON, err = migrations.DoMigrationBasedOnSchemaChanges(sqlDB, schemaDir)
	if err != nil {
		return nil, "", err
	}
	return gormDB, protoSchemaJSON, nil
}
