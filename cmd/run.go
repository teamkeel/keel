package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/postgres"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"gorm.io/gorm"

	gormpostgres "gorm.io/driver/postgres"
)

// The Run command does this:
//
// - Starts Postgres in a docker container.
// - Loads the Keel schema files, validates them, and watches for changes
// - When the Keel schema files are valid migrations are generated and run
//   against the database and a new runtime handler is created
// - Starts an HTTP server which when the Keel schema files are currently
//   valid delegates the requests to the runtime handler. When there are
//   validation errors in the schema files then an error is returned.

var cobraCommandWrapper = &cobra.Command{
	Use:   "run",
	Short: "Run your Keel App locally",
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaDir, _ := cmd.Flags().GetString("dir")

		dbConn, err := postgres.Start(true)
		if err != nil {
			panic(err)
		}
		defer postgres.Stop()

		db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{
			Conn: dbConn,
		}), &gorm.Config{})
		if err != nil {
			panic(err)
		}

		currSchema, err := migrations.GetCurrentSchema(db)
		if err != nil {
			panic(err)
		}

		reloadSchema := func() {
			clearTerminal()
			printRunHeader(schemaDir)

			fmt.Println("Loading schema ...")
			b := &schema.Builder{}
			protoSchema, err := b.MakeFromDirectory(schemaDir)
			if err != nil {
				errs, ok := err.(errorhandling.ValidationErrors)
				if !ok {
					panic(err)
				}

				out, err := errs.ToConsole(b.SchemaFiles())
				if err != nil {
					panic(err)
				}

				fmt.Println(out)
				fmt.Println("Schema has errors 🚨")
				return
			}

			fmt.Println("Schema is valid ✅")

			m := migrations.New(protoSchema, currSchema)
			if m.SQL != "" {
				fmt.Println("Applying migrations 💿")
				err = m.Apply(db)
				if err != nil {
					panic(err)
				}
			}

			currSchema = protoSchema
			fmt.Println("You're reading to roll 🎉")
		}

		stopWatcher, err := onSchemaFileChanges(schemaDir, reloadSchema)
		if err != nil {
			panic(err)
		}
		defer stopWatcher()

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("Request: %s %s\n", r.Method, r.URL.Path)
			w.Write([]byte("Hello"))
		})

		reloadSchema()

		// Todo - we must not forget housekeeping on close...
		//
		// - the dockerized database
		// - the GraphQL API server

		return http.ListenAndServe(":8000", nil)
	},
}

func clearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func printRunHeader(dir string) {
	fmt.Printf("Waiting for schema files to change in %s ...\n", dir)
	fmt.Printf("Press CTRL-C to exit\n\n")
}

// reactToSchemaChanges should be called in its own goroutine. It has a blocking
// channel select loop that waits for and receives file system events, or errors.
func onSchemaFileChanges(dir string, cb func()) (func() error, error) {
	// The run command remains quiescent now, until the user changes their schema, so we establish
	// a watcher on the schema directorty.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				switch {
				case !strings.HasSuffix(event.Name, ".keel"):
					// Ignore
				case !isRelevantEventType(event.Op):
					// Ignore
				default:
					cb()
				}

			case err := <-watcher.Errors:
				fmt.Printf("error received on watcher error channel: %v\n", err)
				// If we get an internal error from the watcher - we simply report the details
				// and allow the watching to continue. We leave it to the user to decide if
				// they want to quit the run command.
			}
		}
	}()

	// The watcher documentation suggests we tell the watcher about the directory to watch,
	// AFTER we have constructed it, and registered a handler.
	err = watcher.Add(dir)
	if err != nil {
		return nil, err
	}

	return watcher.Close, nil
}

func isRelevantEventType(op fsnotify.Op) bool {
	relevant := []fsnotify.Op{fsnotify.Create, fsnotify.Write, fsnotify.Remove}
	// The irrelevant ones are Rename and Chmod.
	return lo.Contains(relevant, op)
}

func init() {
	rootCmd.AddCommand(cobraCommandWrapper)
	defaultDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("os.Getwd() errored: %v", err))
	}
	cobraCommandWrapper.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "schema directory to run")
}
