package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/database"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"gorm.io/gorm"

	"gorm.io/driver/postgres"
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
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run your Keel App locally",
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaDir, _ := cmd.Flags().GetString("dir")

		useExistingContainer := !runCmdFlagReset
		dbConn, dbConnInfo, err := database.Start(useExistingContainer)
		if err != nil {
			if portErr, ok := err.(database.ErrPortInUse); ok {
				color.Red("Unable to start database: %s\n", portErr.Error())
				color.Yellow("To create a fresh database container on a different port re-run this command with --reset\n\n")
				return nil
			}
			panic(err)
		}
		defer database.Stop()

		db, err := gorm.Open(postgres.New(postgres.Config{
			Conn: dbConn,
		}), &gorm.Config{})
		if err != nil {
			panic(err)
		}

		var mutex sync.Mutex

		currSchema, err := migrations.GetCurrentSchema(db)
		if err != nil {
			panic(err)
		}

		reloadSchema := func(changedFile string) {
			mutex.Lock()
			defer mutex.Unlock()

			clearTerminal()
			printRunHeader(schemaDir, dbConnInfo)

			if changedFile != "" {
				fmt.Println("Detected change to:", changedFile)
			}

			fmt.Println("📂 Loading schema files")
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
				fmt.Println("🚨 Schema has errors")
				currSchema = nil
				return
			}

			fmt.Println("✅ Schema is valid")

			m := migrations.New(protoSchema, currSchema)
			if m.SQL != "" {
				fmt.Println("💿 Applying migrations")
				err = m.Apply(db)
				if err != nil {
					panic(err)
				}

				printMigrationChanges(m.Changes)
			}

			currSchema = protoSchema
			fmt.Println("🎉 You're ready to roll")
		}

		stopWatcher, err := onSchemaFileChanges(schemaDir, reloadSchema)
		if err != nil {
			panic(err)
		}
		defer stopWatcher()

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			mutex.Lock()
			defer mutex.Unlock()

			fmt.Printf("Request: %s %s\n", r.Method, r.URL.Path)

			if strings.HasSuffix(r.URL.Path, "/graphiql") {
				handler := playground.Handler("GraphiQL", strings.TrimSuffix(r.URL.Path, "/graphiql"))
				handler(w, r)
				return
			}

			if currSchema == nil {
				w.WriteHeader(400)
				w.Write([]byte("Cannot serve requests when schema contains errors"))
				return
			}

			handler := runtime.NewHandler(currSchema)

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(400)
				w.Write([]byte(err.Error()))
				return
			}

			response, err := handler(&runtime.Request{
				Context: r.Context(),
				URL:     *r.URL,
				Body:    body,
			})
			if err != nil {
				w.WriteHeader(400)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(response.Status)
			w.Write(response.Body)
		})

		reloadSchema("")

		go http.ListenAndServe(":"+runCmdFlagPort, nil)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		fmt.Println("\n👋 Bye bye")
		return nil
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

func printRunHeader(dir string, dbConnInfo *database.ConnectionInfo) {
	fmt.Printf("Watching schema files in: %s\n", color.CyanString(dir))

	psql := color.CyanString("psql postgresql://%s:%s@%s:%s/%s",
		dbConnInfo.Username,
		dbConnInfo.Password,
		dbConnInfo.Host,
		dbConnInfo.Port,
		dbConnInfo.Database)

	endpoint := color.CyanString("http://localhost:%s\n", runCmdFlagPort)

	fmt.Printf("Connect to the database: %s\n", psql)
	fmt.Printf("Application running at: %s\n", endpoint)
	fmt.Printf("Press CTRL-C to exit\n\n")
}

func printMigrationChanges(changes []*migrations.DatabaseChange) {
	var t string

	for _, ch := range changes {
		fmt.Printf(" - ")
		switch ch.Type {
		case migrations.ChangeTypeAdded:
			t = color.YellowString(ch.Type)
		case migrations.ChangeTypeRemoved:
			t = color.RedString(ch.Type)
		case migrations.ChangeTypeModified:
			t = color.GreenString(ch.Type)
		}
		fmt.Printf(" %s %s", t, ch.Model)
		if ch.Field != "" {
			fmt.Printf(".%s", ch.Field)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}

// reactToSchemaChanges should be called in its own goroutine. It has a blocking
// channel select loop that waits for and receives file system events, or errors.
func onSchemaFileChanges(dir string, cb func(changedFile string)) (func() error, error) {
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
					cb(event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
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

var runCmdFlagReset bool
var runCmdFlagPort string

func init() {
	rootCmd.AddCommand(runCmd)

	defaultDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	runCmd.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "the directory containing the Keel schema files")
	runCmd.Flags().BoolVar(&runCmdFlagReset, "reset", false, "if set the database will be reset")
	runCmd.Flags().StringVar(&runCmdFlagPort, "port", "8000", "the port to run the Keel application on")
}
