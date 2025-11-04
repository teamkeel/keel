package program

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	_ "embed"

	"github.com/Masterminds/semver/v3"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
	"github.com/teamkeel/keel/cmd/cliconfig"
	"github.com/teamkeel/keel/cmd/database"
	"github.com/teamkeel/keel/cmd/localTraceExporter"
	"github.com/teamkeel/keel/cmd/storage"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/deploy"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/flows"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
	v1 "go.opentelemetry.io/proto/otlp/trace/v1"
	p "google.golang.org/protobuf/proto"
)

//go:embed default.pem
var defaultPem []byte

func NextMsgCommand(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

type FetchLatestVersionMsg struct {
	LatestVersion *semver.Version
}

type LoadSchemaMsg struct {
	Schema      *proto.Schema
	Config      *config.ProjectConfig
	SchemaFiles []*reader.SchemaFile
	Secrets     map[string]string
	Err         error
}

func FetchLatestVersion() tea.Cmd {
	return func() tea.Msg {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd := exec.Command("npm", "view", "keel@latest", "version")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil || stderr.String() != "" {
			return nil
		}

		output := strings.TrimSuffix(stdout.String(), "\n")

		version, err := semver.StrictNewVersion(output)
		if err != nil {
			return nil
		}

		msg := FetchLatestVersionMsg{
			LatestVersion: version,
		}

		return msg
	}
}

func LoadSchema(dir, environment string) tea.Cmd {
	return func() tea.Msg {
		b := schema.Builder{}
		s, err := b.MakeFromDirectory(dir)

		absolutePath, filepathErr := filepath.Abs(dir)
		if filepathErr != nil {
			err = filepathErr
		}

		cliConfig := cliconfig.New(&cliconfig.Options{
			WorkingDir: dir,
		})

		secrets, configErr := cliConfig.GetSecrets(absolutePath, environment)
		if configErr != nil {
			err = configErr
		}

		if b.Config == nil {
			b.Config = &config.ProjectConfig{}
		}

		// Add the OIDC auth provider used on the Console's internal tools
		configErr = b.Config.Auth.AddOidcProvider(consoleAuthProviderName1, consoleAuthProviderIssuer1, consoleAuthProviderClientId1)
		if configErr != nil {
			err = configErr
		}

		// Add the OIDC auth provider used on the Console's internal tools
		configErr = b.Config.Auth.AddOidcProvider(consoleAuthProviderName2, consoleAuthProviderIssuer2, consoleAuthProviderClientId2)
		if configErr != nil {
			err = configErr
		}

		invalid, invalidSecrets := b.Config.ValidateSecrets(secrets)
		if invalid {
			err = fmt.Errorf("missing secrets from local config in ~/.keel/config.yaml: %s", strings.Join(invalidSecrets, ", "))
		}

		msg := LoadSchemaMsg{
			Schema:      s,
			Config:      b.Config,
			SchemaFiles: b.SchemaFiles(),
			Secrets:     secrets,
			Err:         err,
		}

		return msg
	}
}

type GenerateMsg struct {
	Err            error
	Status         int
	GeneratedFiles codegen.GeneratedFiles
	Log            string
}

type GenerateClientMsg struct {
	Err            error
	Status         int
	GeneratedFiles codegen.GeneratedFiles
	Log            string
}

func GenerateClient(dir string, schema *proto.Schema, apiName string, outputDir string, makePackage bool, output chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		if schema == nil || len(schema.GetApis()) == 0 {
			return GenerateMsg{
				Status: StatusNotGenerated,
			}
		}

		output <- GenerateClientMsg{
			Log:    "Generating client SDK",
			Status: StatusGeneratingClient,
		}

		files, err := node.GenerateClient(context.TODO(), schema, makePackage, apiName)

		if err != nil {
			return GenerateClientMsg{
				Err: err,
			}
		}

		o := dir
		if outputDir != "" {
			o = outputDir
		}

		err = files.Write(o)

		if err != nil {
			return GenerateClientMsg{
				Err: err,
			}
		}

		output <- GenerateClientMsg{
			Log:            "Generated @teamkeel/client",
			Status:         StatusGenerated,
			GeneratedFiles: files,
		}

		return nil
	}
}

type CheckDependenciesMsg struct {
	Err error
}

func CheckDependencies() tea.Cmd {
	return func() tea.Msg {
		err := node.CheckNodeVersion()

		return CheckDependenciesMsg{
			Err: err,
		}
	}
}

type ParsePrivateKeyMsg struct {
	PrivateKey *rsa.PrivateKey
	Err        error
}

func ParsePrivateKey(path string) tea.Cmd {
	return func() tea.Msg {
		// Uses the embedded default private key if a custom key isn't provided
		// This allows for a smooth DX in a local env where the signing of the token isn't important
		// but avoids a code path where we skip token validation
		var privateKeyPem []byte

		if path == "" {
			privateKeyPem = defaultPem
		} else {
			customPem, err := os.ReadFile(path)
			if errors.Is(err, os.ErrNotExist) {
				return ParsePrivateKeyMsg{
					Err: fmt.Errorf("cannot locate private key file at: %s", path),
				}
			} else if err != nil {
				return ParsePrivateKeyMsg{
					Err: fmt.Errorf("cannot read private key file: %s", err.Error()),
				}
			}
			privateKeyPem = customPem
		}

		privateKeyBlock, _ := pem.Decode(privateKeyPem)
		if privateKeyBlock == nil {
			return ParsePrivateKeyMsg{
				Err: errors.New("private key PEM either invalid or empty"),
			}
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			return ParsePrivateKeyMsg{
				Err: err,
			}
		}

		return ParsePrivateKeyMsg{
			PrivateKey: privateKey,
		}
	}
}

type StartDatabaseMsg struct {
	ConnInfo *db.ConnectionInfo
	Err      error
}

func StartDatabase(reset bool, mode int, projectDirectory string) tea.Cmd {
	return func() tea.Msg {
		connInfo, err := database.Start(reset, projectDirectory)
		if err != nil {
			return StartDatabaseMsg{
				Err: err,
			}
		}

		return StartDatabaseMsg{
			ConnInfo: connInfo,
		}
	}
}

type StartStorageMsg struct {
	ConnInfo *storage.ConnectionInfo
	Err      error
}

func StartStorage(projectDirectory string) tea.Cmd {
	return func() tea.Msg {
		conn, err := storage.Start(projectDirectory)
		if err != nil {
			return StartDatabaseMsg{
				Err: err,
			}
		}

		return StartStorageMsg{
			ConnInfo: conn,
		}
	}
}

type SetupFunctionsMsg struct {
	Err error
}

func SetupFunctions(dir string, nodePackagesPath string, packageManager string, mode int) tea.Cmd {
	// If we're not running we can skip the bootstrap
	if mode != ModeRun {
		return func() tea.Msg {
			return SetupFunctionsMsg{}
		}
	}

	return func() tea.Msg {
		err := node.Bootstrap(
			dir,
			node.WithPackageManager(packageManager),
			node.WithPackagesPath(nodePackagesPath),
		)
		if err != nil {
			return SetupFunctionsMsg{
				Err: err,
			}
		}

		return SetupFunctionsMsg{}
	}
}

type UpdateFunctionsMsg struct {
	Err error
}

type TypeScriptError struct {
	Output string
	Err    error
}

func (t *TypeScriptError) Error() string {
	return fmt.Sprintf("TypeScript error: %s", t.Err.Error())
}

func UpdateFunctions(schema *proto.Schema, cfg *config.ProjectConfig, dir string) tea.Cmd {
	return func() tea.Msg {
		ctx := deploy.WithLogLevel(context.Background(), deploy.LogLevelSilent)
		_, err := deploy.Build(ctx, &deploy.BuildArgs{
			ProjectRoot: dir,
			Env:         "development",
		})
		if err != nil {
			return UpdateFunctionsMsg{Err: err}
		}

		cmd := exec.Command("npx", "tsc", "--noEmit", "--pretty")
		cmd.Dir = dir

		b, err := cmd.CombinedOutput()
		if err != nil {
			return UpdateFunctionsMsg{
				Err: &TypeScriptError{
					Output: string(b),
					Err:    err,
				},
			}
		}

		return UpdateFunctionsMsg{}
	}
}

type RunMigrationsMsg struct {
	Err     error
	Changes []*migrations.DatabaseChange
}

type ApplyMigrationsError struct {
	Err error
}

func (a *ApplyMigrationsError) Error() string {
	return a.Err.Error()
}

func (e *ApplyMigrationsError) Unwrap() error {
	return e.Err
}

func RunMigrations(schema *proto.Schema, database db.Database) tea.Cmd {
	return func() tea.Msg {
		m, err := migrations.New(context.Background(), schema, database)
		if err != nil {
			return RunMigrationsMsg{
				Err: &ApplyMigrationsError{
					Err: err,
				},
			}
		}

		msg := RunMigrationsMsg{
			Changes: m.Changes,
		}

		err = m.Apply(context.Background(), false)
		if err != nil {
			msg.Err = &ApplyMigrationsError{
				Err: err,
			}
		}

		return msg
	}
}

type StartFunctionsMsg struct {
	Err    error
	Server *node.DevelopmentServer
}

type StartFunctionsError struct {
	Err    error
	Output string
}

func (s *StartFunctionsError) Error() string {
	return s.Err.Error()
}

type FunctionsOutputMsg struct {
	Output string
}

func StartFunctions(m *Model) tea.Cmd {
	return func() tea.Msg {
		envVars := m.Config.GetEnvVars()
		envVars["KEEL_DB_CONN_TYPE"] = "pg"
		// KEEL_DB_CONN is passed via a secret but for backwards compatibility with old functions-runtimes
		// we'll also pass as an env var for now
		envVars["KEEL_DB_CONN"] = m.DatabaseConnInfo.String()
		envVars["KEEL_TRACING_ENABLED"] = "true"
		envVars["OTEL_RESOURCE_ATTRIBUTES"] = "service.name=functions"

		if m.StorageConnInfo != nil {
			envVars["KEEL_FILES_BUCKET_NAME"] = m.StorageConnInfo.Bucket
			envVars["KEEL_S3_ENDPOINT"] = fmt.Sprintf("http://%s:%s", m.StorageConnInfo.Host, m.StorageConnInfo.Port)
		}

		output := &FunctionsOutputWriter{
			// Initially buffer output inside the writer in case there's an error
			Buffer: true,
			ch:     m.functionsLogCh,
		}
		server, err := node.StartDevelopmentServer(context.Background(), m.ProjectDir, &node.ServerOpts{
			EnvVars: envVars,
			Output:  output,
			Watch:   true,
		})
		if err != nil {
			return StartFunctionsMsg{
				Err: &StartFunctionsError{
					Output: strings.Join(output.Output, "\n"),
					Err:    err,
				},
			}
		}

		// Stop buffering output now we know the process started.
		// All future output will be written to the given channel
		output.Buffer = false

		return StartFunctionsMsg{
			Server: server,
		}
	}
}

type RuntimeRequestMsg struct {
	w    http.ResponseWriter
	r    *http.Request
	done chan bool
}

type StartServerError struct {
	Err error
}

func StartRuntimeServer(port string, customHostname string, ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		if customHostname != "" {
			os.Setenv("KEEL_API_URL", customHostname)
		} else {
			os.Setenv("KEEL_API_URL", fmt.Sprintf("http://localhost:%s", port))
		}

		runtimeServer := http.Server{
			Addr: fmt.Sprintf(":%s", port),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				done := make(chan bool, 1)
				ch <- RuntimeRequestMsg{
					w:    w,
					r:    r,
					done: done,
				}
				<-done
			}),
		}

		err := runtimeServer.ListenAndServe()
		if err != nil {
			err = errors.Join(errors.New("could not start the http server"), err)
			return StartServerError{
				Err: err,
			}
		}

		return nil
	}
}

type RpcRequestMsg struct {
	w    http.ResponseWriter
	r    *http.Request
	done chan bool
}

func StartRpcServer(port string, ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		rpcServer := http.Server{
			Addr: fmt.Sprintf(":%s", port),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				done := make(chan bool, 1)
				ch <- RpcRequestMsg{
					w:    w,
					r:    r,
					done: done,
				}
				<-done
			}),
		}

		err := rpcServer.ListenAndServe()
		if err != nil {
			err = errors.Join(errors.New("could not start the local rpc server"), err)
			return StartServerError{
				Err: err,
			}
		}

		return nil
	}
}

func StartTraceServer(port string) tea.Cmd {
	return func() tea.Msg {
		traceServer := http.Server{
			Addr: fmt.Sprintf(":%s", port),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				if err != nil {
					return
				}

				data := &v1.TracesData{}
				err = p.Unmarshal(b, data)
				if err != nil {
					return
				}

				_ = localTraceExporter.NewClient().UploadTraces(r.Context(), data.GetResourceSpans())
			}),
		}

		err := traceServer.ListenAndServe()
		if err != nil {
			err = errors.Join(errors.New("could not start the local tracing server"), err)
			return StartServerError{
				Err: err,
			}
		}

		return nil
	}
}

type FunctionsOutputWriter struct {
	Output []string
	Buffer bool
	ch     chan tea.Msg
}

func (f *FunctionsOutputWriter) Write(p []byte) (n int, err error) {
	str := string(p)

	if len(strings.Split(str, "\n")) > 1 {
		str = fmt.Sprintf("\n%s", str)
	}

	if f.Buffer {
		f.Output = append(f.Output, str)
	} else {
		f.ch <- FunctionsOutputMsg{
			Output: str,
		}
	}

	return len(p), nil
}

type WatcherMsg struct {
	Err   error
	Path  string
	Event string
}

func shouldSkipWatch(path string) bool {
	// List of ignored directory components
	ignored := []string{
		"node_modules",
		"tools",
	}

	// Skip if any directory component is hidden or is in the ignored list
	pathParts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for _, part := range pathParts {
		if strings.HasPrefix(part, ".") {
			return true
		}

		for _, v := range ignored {
			if strings.Contains(part, v) {
				return true
			}
		}
	}

	return false
}

func StartWatcher(dir string, ch chan tea.Msg, filter []string) tea.Cmd {
	return func() tea.Msg {
		// Convert relative path to absolute path
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return WatcherMsg{
				Err: err,
			}
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return WatcherMsg{
				Err: err,
			}
		}

		// Walk through the directory tree and add directories to watch
		err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				return nil // Only add directories to the watcher
			}

			if shouldSkipWatch(path) {
				return filepath.SkipDir
			}

			// If there is a filter set, check if we should watch this directory
			if len(filter) > 0 {
				// For filters, we check if any of the files in the directory match
				// the filter. If not, we skip the directory.
				shouldWatch := false
				for _, v := range filter {
					if strings.Contains(path, v) {
						shouldWatch = true
						break
					}
				}

				if !shouldWatch {
					return filepath.SkipDir
				}
			}

			// Add the directory to the watcher
			return watcher.Add(path)
		})

		if err != nil {
			return WatcherMsg{
				Err: err,
			}
		}

		// Start watching for events
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}

					var eventType string
					switch {
					case event.Op&fsnotify.Write == fsnotify.Write:
						eventType = "Write"
					case event.Op&fsnotify.Remove == fsnotify.Remove:
						eventType = "Remove"
					case event.Op&fsnotify.Rename == fsnotify.Rename:
						eventType = "Rename"
					case event.Op&fsnotify.Create == fsnotify.Create:
						eventType = "Create"
					default:
						continue
					}

					if info, err := os.Stat(event.Name); err == nil {
						// If a directory is added or renamed then we need to also watch it for changes to its contents
						if info.IsDir() {
							if shouldSkipWatch(event.Name) {
								continue
							}

							err = watcher.Add(event.Name)
							if err != nil {
								continue
							}
						}
					}

					ch <- WatcherMsg{
						Path:  event.Name,
						Event: eventType,
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					ch <- WatcherMsg{
						Err: err,
					}
				}
			}
		}()

		// Return nil to indicate successful setup
		return nil
	}
}

// LoadSecrets lists secrets from the given file and returns a command.
func LoadSecrets(path, environment string) (map[string]string, error) {
	projectPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	config := cliconfig.New(&cliconfig.Options{
		WorkingDir: projectPath,
	})

	secrets, err := config.GetSecrets(path, environment)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func SetSecret(path, environment, key, value string) error {
	projectPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	config := cliconfig.New(&cliconfig.Options{
		WorkingDir: projectPath,
	})

	return config.SetSecret(path, environment, key, value)
}

func RemoveSecret(path, environment, key string) error {
	projectPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	config := cliconfig.New(&cliconfig.Options{
		WorkingDir: projectPath,
	})

	return config.RemoveSecret(path, environment, key)
}

type SeedDataMsg struct {
	Err         error
	SeededFiles []string
}

type SeedDataError struct {
	Err error
}

func (a *SeedDataError) Error() string {
	return a.Err.Error()
}

func (e *SeedDataError) Unwrap() error {
	return e.Err
}

func SeedData(dir string, schema *proto.Schema, database db.Database) tea.Cmd {
	return func() tea.Msg {
		m, err := migrations.New(context.Background(), schema, database)
		if err != nil {
			return SeedDataMsg{
				Err: &SeedDataError{
					Err: err,
				},
			}
		}

		msg := SeedDataMsg{
			Err: err,
		}

		seedDir := filepath.Join(dir, "seed")

		if _, err := os.Stat(seedDir); os.IsNotExist(err) {
			return msg
		}

		// Get all .sql files from seed directory
		files, err := filepath.Glob(filepath.Join(seedDir, "*.sql"))
		if err != nil {
			return SeedDataMsg{
				Err: &SeedDataError{
					Err: err,
				},
			}
		}

		if len(files) == 0 {
			return msg
		}

		// Sort files for deterministic seeding
		sort.Strings(files)

		err = m.ApplySeedData(context.Background(), files)
		if err != nil {
			msg.Err = &SeedDataError{
				Err: err,
			}
		}

		msg.SeededFiles = files

		return msg
	}
}

type SnapshotDatabaseMsg struct {
	Err error
}

func SnapshotDatabase(dir string, schema *proto.Schema, database db.Database) tea.Cmd {
	return func() tea.Msg {
		m, err := migrations.New(context.Background(), schema, database)
		if err != nil {
			return SnapshotDatabaseMsg{
				Err: err,
			}
		}

		snapshot, err := m.SnapshotDatabase(context.Background())
		if err != nil {
			return SnapshotDatabaseMsg{
				Err: err,
			}
		}

		seedDir := filepath.Join(dir, "seed")

		err = os.MkdirAll(seedDir, 0755)
		if err != nil {
			return SnapshotDatabaseMsg{
				Err: err,
			}
		}

		err = os.WriteFile(filepath.Join(seedDir, "snapshot.sql"), []byte(snapshot), 0644)
		if err != nil {
			return SnapshotDatabaseMsg{
				Err: err,
			}
		}

		return SnapshotDatabaseMsg{}
	}
}

type CronRunnerMsg struct {
	Err error
}

func SetupCron(schema *proto.Schema, database db.Database, functionsServer *node.DevelopmentServer, cronRunner *cron.Cron, secrets map[string]string) tea.Cmd {
	return func() tea.Msg {
		cronRunner.Stop()
		// restart cron jobs
		for _, e := range cronRunner.Entries() {
			cronRunner.Remove(e.ID)
		}

		if !schema.HasScheduledFlows() {
			// no scheduled flows
			return nil
		}
		ctx := context.Background()
		ctx = db.WithDatabase(ctx, database)
		ctx = functions.WithFunctionsTransport(ctx, functions.NewHttpTransport(functionsServer.URL))
		ctx = runtimectx.WithSecrets(ctx, secrets)

		o := flows.NewOrchestrator(schema, flows.WithNoQueueEventSender())

		for _, f := range schema.ScheduledFlows() {
			// make the event payload
			ev := flows.FlowRunStarted{Name: f.GetName(), Inputs: map[string]any{}}

			payload, err := ev.Wrap()
			if err != nil {
				return fmt.Errorf("wrapping event: %w", err)
			}

			// Our cron expressions for schedules include the year, which is not relevant to our use case.
			schedule := strings.TrimSuffix(f.GetSchedule().GetExpression(), " *")

			if _, err := cronRunner.AddFunc(schedule, func() {
				o.HandleEvent(ctx, payload) //nolint
			}); err != nil {
				return CronRunnerMsg{
					Err: fmt.Errorf("scheduling flow: %w", err),
				}
			}
		}

		cronRunner.Start()

		return CronRunnerMsg{}
	}
}
