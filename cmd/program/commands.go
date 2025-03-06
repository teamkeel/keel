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
	"strings"
	"time"

	_ "embed"

	"github.com/Masterminds/semver/v3"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/radovskyb/watcher"
	"github.com/teamkeel/keel/cmd/cliconfig"
	"github.com/teamkeel/keel/cmd/database"
	"github.com/teamkeel/keel/cmd/localTraceExporter"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/proto"
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
		if schema == nil || len(schema.Apis) == 0 {
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

type SetupFunctionsMsg struct {
	Err error
}

func SetupFunctions(dir string, nodePackagesPath string, packageManager string) tea.Cmd {
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
		files, err := node.Generate(context.TODO(), schema, cfg, node.WithDevelopmentServer(true))
		if err != nil {
			return UpdateFunctionsMsg{Err: err}
		}

		err = files.Write(dir)
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
			panic(err.Error())
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
			panic(err.Error())
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

				_ = localTraceExporter.NewClient().UploadTraces(r.Context(), data.ResourceSpans)
			}),
		}

		err := traceServer.ListenAndServe()
		if err != nil {
			err = errors.Join(errors.New("could not start the local tracing server"), err)
			panic(err.Error())
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

func StartWatcher(dir string, ch chan tea.Msg, filter []string) tea.Cmd {
	return func() tea.Msg {
		w := watcher.New()
		w.SetMaxEvents(1)
		w.FilterOps(watcher.Write, watcher.Remove)

		ignored := []string{
			"node_modules",
			"tools",
		}

		w.AddFilterHook(func(info os.FileInfo, fullPath string) error {
			// Skip if any directory component is hidden or is in the ignored list
			pathParts := strings.Split(filepath.Clean(fullPath), string(filepath.Separator))
			for _, part := range pathParts {
				if strings.HasPrefix(part, ".") {
					return watcher.ErrSkip
				}

				for _, v := range ignored {
					if strings.Contains(part, v) {
						return watcher.ErrSkip
					}
				}
			}

			// If there is a filter set then only watch these files
			if len(filter) > 0 {
				for _, v := range filter {
					if !strings.Contains(fullPath, v) {
						return watcher.ErrSkip
					}
				}
			}

			return nil
		})

		go func() {
			for {
				select {
				case event := <-w.Event:
					ch <- WatcherMsg{
						Path:  event.Path,
						Event: event.Op.String(),
					}
				case <-w.Closed:
					return
				}
			}
		}()

		err := w.AddRecursive(dir)
		if err != nil {
			return WatcherMsg{
				Err: err,
			}
		}

		_ = w.Start(time.Millisecond * 100)
		return nil
	}
}

// LoadSecrets lists secrets from the given file and returns a command
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
