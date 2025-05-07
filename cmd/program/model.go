package program

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Masterminds/semver/v3"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/cmd/database"
	"github.com/teamkeel/keel/cmd/localTraceExporter"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/mail"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/rpc/rpc"
	rpcApi "github.com/teamkeel/keel/rpc/rpcApi"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/flows"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/storage"
	"github.com/twitchtv/twirp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	ModeRun = iota
	ModeSeed
	ModeReset
)

const (
	StatusCheckingDependencies = iota
	StatusParsePrivateKey
	StatusSetupDatabase
	StatusSetupFunctions
	StatusLoadSchema
	StatusRunMigrations
	StatusSeedData
	StatusUpdateFunctions
	StatusStartingFunctions
	StatusRunning
	StatusQuitting
	StatusSeedCompleted
	StatusSnapshotDatabase
	StatusSnapshotCompleted
	StatusErrorStartingServers
)

const (
	consoleAuthProviderName1     = "keel_console_auth_1"
	consoleAuthProviderIssuer1   = "https://auth.keel.xyz/"
	consoleAuthProviderClientId1 = "KvnOmNPy17WtMGBseDZRw3Hgn0ZOQpDd"
)

const (
	consoleAuthProviderName2     = "keel_console_auth_2"
	consoleAuthProviderIssuer2   = "https://auth.staging.keel.xyz/"
	consoleAuthProviderClientId2 = "mXXReYQdTSIm2UhTRzMkYmMQGU1GC6wp"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/db")

func Run(model *Model) {
	// The runtime currently does logging with logrus, which is super noisy.
	// For now we just discard the logs as they are not useful in the CLI
	logrus.SetOutput(io.Discard)

	var exporter *otlptrace.Exporter
	var err error
	if model.CustomTracing {
		exporter, err = otlptracehttp.New(context.Background(), otlptracehttp.WithInsecure())
		if err != nil {
			panic(err.Error())
		}
	} else {
		exporter, err = localTraceExporter.New(context.Background())
		if err != nil {
			panic(err.Error())
		}
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewSchemaless(attribute.String("service.name", "cli")),
		),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	defer func() {
		_ = database.Stop()
		if model.FunctionsServer != nil {
			_ = model.FunctionsServer.Kill()
		}
	}()

	_, err = tea.NewProgram(model).Run()
	if err != nil {
		panic(err)
	}

	if model.Err != nil {
		os.Exit(1)
	}
}

type Model struct {
	// The directory of the Keel project
	ProjectDir string

	// The mode the Model is running in
	Mode int

	// Port to run the runtime servers on in ModeRun
	Port      string
	RpcPort   string
	TracePort string

	// If true then the database will be reset. Only
	// applies to ModeRun.
	ResetDatabase bool

	// If set then @teamkeel/* npm packages will be installed
	// from this path, rather than NPM.
	NodePackagesPath string

	// Either 'npm' or 'pnpm'
	PackageManager string

	// If set then runtime will be configured with private key
	// located at this path in pem format.
	PrivateKeyPath string

	// The private key to configure on runtime, or nil.
	PrivateKey *rsa.PrivateKey

	// Pattern to pass to vitest to isolate specific tests
	TestPattern string

	// If true then traces will be sent to localhost:4318 instead of the default exporter
	// This does mean that the Console Monitoring page will not reflect any traces.
	CustomTracing bool

	// If true, do not filter events in the local trace exporter.
	// This will then show all system events in the local console.
	VerboseTracing bool

	// A custom configured hostname, which may be necessary to change for SSO callback.
	CustomHostname string

	// Model state - used in View()
	Status            int
	Err               error
	Schema            *proto.Schema
	Config            *config.ProjectConfig
	SchemaFiles       []*reader.SchemaFile
	Database          db.Database
	DatabaseConnInfo  *db.ConnectionInfo
	GeneratedFiles    codegen.GeneratedFiles
	MigrationChanges  []*migrations.DatabaseChange
	FunctionsServer   *node.DevelopmentServer
	RuntimeHandler    http.Handler
	JobHandler        runtime.JobHandler
	SubscriberHandler runtime.SubscriberHandler
	RpcHandler        http.Handler
	RuntimeRequests   []*RuntimeRequest
	FunctionsLog      []*FunctionLog
	Storage           storage.Storer
	TestOutput        string
	Secrets           map[string]string
	Environment       string
	SeedData          bool
	SeededFiles       []string
	SnapshotDatabase  bool

	// The current latest version of Keel in NPM
	LatestVersion *semver.Version

	// Channels for communication between long-running
	// commands and the Bubbletea program
	runtimeRequestsCh chan tea.Msg
	functionsLogCh    chan tea.Msg
	rpcRequestsCh     chan tea.Msg
	watcherCh         chan tea.Msg

	// Maintain the current dimensions of the user's terminal
	width  int
	height int

	Debug   bool
	Timings map[string]time.Time
}

type RuntimeRequest struct {
	Time   time.Time
	Method string
	Path   string
}

type FunctionLog struct {
	Time  time.Time
	Value string
}

var _ tea.Model = &Model{}

func (m *Model) Init() tea.Cmd {
	m.runtimeRequestsCh = make(chan tea.Msg, 1)
	m.rpcRequestsCh = make(chan tea.Msg, 1)
	m.functionsLogCh = make(chan tea.Msg, 1)
	m.watcherCh = make(chan tea.Msg, 1)
	m.Environment = "development"
	m.RpcHandler = rpc.NewAPIServer(&rpcApi.Server{}, twirp.WithServerPathPrefix("/rpc"))
	m.RpcPort = "34087"
	m.TracePort = "4318"

	m.Status = StatusCheckingDependencies

	cmds := []tea.Cmd{
		FetchLatestVersion(),
		CheckDependencies(),
	}

	m.Timings = make(map[string]time.Time)
	m.Timings["start"] = time.Now()

	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Status = StatusQuitting
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		// This msg is sent once on program start
		// and then subsequently every time the user
		// resizes their terminal window.
		m.width = msg.Width
		m.height = msg.Height

		return m, nil

	case FetchLatestVersionMsg:
		m.LatestVersion = msg.LatestVersion

		return m, nil
	case CheckDependenciesMsg:
		m.Err = msg.Err

		if m.Err != nil {
			return m, tea.Quit
		}

		m.Status = StatusSetupDatabase
		return m, StartDatabase(m.ResetDatabase, m.Mode, m.ProjectDir)
	case StartServerError:
		m.Err = msg.Err
		// If the servers can't be started we exit
		if m.Err != nil {
			m.Status = StatusErrorStartingServers
			return m, tea.Quit
		}
	case StartDatabaseMsg:
		m.DatabaseConnInfo = msg.ConnInfo
		m.Err = msg.Err

		m.Timings["startDatabase"] = time.Now()

		// If the database can't be started we exit
		if m.Err != nil {
			return m, tea.Quit
		}

		if msg.ConnInfo.IsNewDatabase {
			m.SeedData = true
		}

		database, err := db.New(context.Background(), m.DatabaseConnInfo.String())
		if err != nil {
			m.Err = err
			return m, tea.Quit
		}

		m.Database = database
		m.Status = StatusParsePrivateKey
		return m, ParsePrivateKey(m.PrivateKeyPath)
	case ParsePrivateKeyMsg:
		m.Err = msg.Err

		m.Timings["parsePrivateKey"] = time.Now()

		// If the private key can't be parsed we exit
		if m.Err != nil {
			return m, tea.Quit
		}

		m.PrivateKey = msg.PrivateKey

		m.Status = StatusSetupFunctions
		return m, SetupFunctions(m.ProjectDir, m.NodePackagesPath, m.PackageManager, m.Mode)
	case SetupFunctionsMsg:
		m.Err = msg.Err

		m.Timings["setupFunctions"] = time.Now()

		// If something failed here (most likely npm install) we exit
		if m.Err != nil {
			return m, tea.Quit
		}

		m.Status = StatusLoadSchema

		cmds := []tea.Cmd{
			StartRuntimeServer(m.Port, m.CustomHostname, m.runtimeRequestsCh),
			StartRpcServer(m.RpcPort, m.rpcRequestsCh),
			NextMsgCommand(m.runtimeRequestsCh),
			NextMsgCommand(m.rpcRequestsCh),
			LoadSchema(m.ProjectDir, m.Environment),
		}

		if !m.CustomTracing {
			cmds = append(cmds, StartTraceServer(m.TracePort))
		}

		if m.Mode == ModeRun {
			cmds = append(
				cmds,
				StartWatcher(m.ProjectDir, m.watcherCh, nil),
				NextMsgCommand(m.watcherCh),
			)
		}

		return m, tea.Batch(cmds...)
	case LoadSchemaMsg:
		m.Schema = msg.Schema
		m.SchemaFiles = msg.SchemaFiles
		m.Config = msg.Config
		m.Err = msg.Err
		m.Secrets = msg.Secrets

		m.Timings["loadSchema"] = time.Now()

		if m.Err != nil {
			return m, nil
		}

		cors := cors.New(cors.Options{
			AllowOriginFunc: func(origin string) bool {
				return true
			},
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,

				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		})

		m.RuntimeHandler = cors.Handler(runtime.NewHttpHandler(m.Schema))
		m.Status = StatusRunMigrations
		return m, RunMigrations(m.Schema, m.Database)
	case RunMigrationsMsg:
		m.Err = msg.Err
		m.MigrationChanges = msg.Changes

		m.Timings["runMigrations"] = time.Now()

		// we now set the file Storage using a dbstore
		storer, err := storage.NewDbStore(context.Background(), m.Database)
		if err != nil {
			m.Err = err
			return m, tea.Quit
		}
		m.Storage = storer

		if m.Err != nil {
			return m, nil
		}

		if m.Mode == ModeSeed {
			if m.SnapshotDatabase {
				m.Status = StatusSnapshotDatabase
				return m, SnapshotDatabase(m.ProjectDir, m.Schema, m.Database)
			}
		}

		// If seed data flag is set, run seed data
		if m.SeedData {
			// If seed data directory exists, run seed data
			seedDir := filepath.Join(m.ProjectDir, "seed")
			if _, err := os.Stat(seedDir); err == nil {
				m.Status = StatusSeedData
				return m, SeedData(m.ProjectDir, m.Schema, m.Database)
			}
		}

		// Otherwise carry on as normal
		if m.Mode == ModeRun && !node.HasFunctions(m.Schema, m.Config) {
			m.Status = StatusRunning
			return m, nil
		}

		m.Status = StatusUpdateFunctions
		return m, UpdateFunctions(m.Schema, m.Config, m.ProjectDir)

	case SeedDataMsg:
		m.SeededFiles = msg.SeededFiles
		m.Err = msg.Err
		if m.Err != nil {
			return m, nil
		}

		m.Timings["seedData"] = time.Now()

		// If seeding data, we quit after seeding
		if m.Mode == ModeSeed {
			m.Status = StatusSeedCompleted
			return m, tea.Quit
		}

		if m.Mode == ModeRun && !node.HasFunctions(m.Schema, m.Config) {
			m.Status = StatusRunning
			return m, nil
		}

		m.Status = StatusUpdateFunctions
		return m, UpdateFunctions(m.Schema, m.Config, m.ProjectDir)

	case SnapshotDatabaseMsg:
		m.Err = msg.Err
		if m.Err != nil {
			return m, nil
		}

		m.Timings["snapshotDatabase"] = time.Now()

		m.Status = StatusSnapshotCompleted
		time.Sleep(1 * time.Second)
		return m, tea.Quit

	case UpdateFunctionsMsg:
		m.Err = msg.Err
		if m.Err != nil {
			return m, nil
		}

		m.Timings["updateFunctions"] = time.Now()

		// If functions already running nothing to do
		if m.FunctionsServer != nil {
			_ = m.FunctionsServer.Rebuild()
			m.Status = StatusRunning
			return m, nil
		}

		// Start functions if needed
		if node.HasFunctions(m.Schema, m.Config) {
			m.Status = StatusStartingFunctions
			return m, tea.Batch(
				StartFunctions(m),
				NextMsgCommand(m.functionsLogCh),
			)
		}

		return m, nil

	case StartFunctionsMsg:
		m.Err = msg.Err
		m.FunctionsServer = msg.Server

		m.Timings["startFunctions"] = time.Now()

		if msg.Err == nil {
			m.Status = StatusRunning
		}

		return m, nil
	case FunctionsOutputMsg:
		log := &FunctionLog{
			Time:  time.Now(),
			Value: msg.Output,
		}
		m.FunctionsLog = append(m.FunctionsLog, log)

		cmds := []tea.Cmd{
			NextMsgCommand(m.functionsLogCh),
		}

		if m.Mode == ModeRun {
			cmds = append(cmds, tea.Println(renderFunctionLog(log)))
		}

		return m, tea.Batch(cmds...)
	case RuntimeRequestMsg:
		r := msg.r
		w := msg.w

		request := &RuntimeRequest{
			Time:   time.Now(),
			Method: r.Method,
			Path:   r.URL.Path,
		}

		cmds := []tea.Cmd{
			NextMsgCommand(m.runtimeRequestsCh),
		}

		m.RuntimeRequests = append(m.RuntimeRequests, request)

		// log runtime requests for the run cmd
		if m.Mode == ModeRun && m.Err == nil && m.Status >= StatusLoadSchema {
			cmds = append(cmds, tea.Println(renderRequestLog(request)))
		}

		if strings.HasSuffix(r.URL.Path, "/graphiql") {
			handler := playground.Handler("GraphiQL", strings.TrimSuffix(r.URL.Path, "/graphiql")+"/graphql")
			handler(w, r)
			msg.done <- true
			return m, NextMsgCommand(m.runtimeRequestsCh)
		}

		if m.RuntimeHandler == nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Cannot serve requests while there are schema errors. Please see the CLI output for more info."))
			msg.done <- true
			return m, NextMsgCommand(m.runtimeRequestsCh)
		}

		ctx := msg.r.Context()

		ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", request.Method, request.Path))
		defer span.End()

		span.SetAttributes(
			attribute.String("type", "request"),
			attribute.String("http.method", request.Method),
			attribute.String("http.path", request.Path),
		)

		if m.PrivateKey != nil {
			ctx = runtimectx.WithPrivateKey(ctx, m.PrivateKey)
		}

		ctx = db.WithDatabase(ctx, m.Database)

		m.Secrets["KEEL_DB_CONN"] = m.DatabaseConnInfo.String()
		ctx = runtimectx.WithSecrets(ctx, m.Secrets)

		ctx = runtimectx.WithOAuthConfig(ctx, &m.Config.Auth)
		if m.Storage != nil {
			ctx = runtimectx.WithStorage(ctx, m.Storage)
		}

		mailClient := mail.NewSMTPClientFromEnv()
		if mailClient != nil {
			ctx = runtimectx.WithMailClient(ctx, mailClient)
		} else {
			ctx = runtimectx.WithMailClient(ctx, mail.NoOpClient())
		}

		if m.FunctionsServer != nil {
			ctx = functions.WithFunctionsTransport(
				ctx,
				functions.NewHttpTransport(m.FunctionsServer.URL),
			)
		}

		envVars := m.Config.GetEnvVars()
		for k, v := range envVars {
			os.Setenv(k, v)
		}

		// Synchronous event handling for keel run.
		// TODO: make asynchronous
		ctx, err := events.WithEventHandler(ctx, func(ctx context.Context, subscriber string, event *events.Event, traceparent string) error {
			return runtime.NewSubscriberHandler(m.Schema).RunSubscriber(ctx, subscriber, event)
		})
		if err != nil {
			m.Err = err
			return m, tea.Quit
		}

		// Setting the flows orchestrator
		ctx = flows.WithOrchestrator(ctx, flows.NewOrchestrator(m.Schema))

		r = msg.r.WithContext(ctx)
		m.RuntimeHandler.ServeHTTP(msg.w, r)

		for k := range envVars {
			os.Unsetenv(k)
		}

		msg.done <- true
		return m, tea.Batch(cmds...)
	case RpcRequestMsg:
		ctx := msg.r.Context()
		ctx = db.WithDatabase(ctx, m.Database)
		ctx = rpcApi.WithSchema(ctx, m.Schema)
		ctx = rpcApi.WithConfig(ctx, m.Config)
		ctx = rpcApi.WithProjectDir(ctx, m.ProjectDir)
		r := msg.r.WithContext(ctx)
		w := msg.w

		cmds := []tea.Cmd{
			NextMsgCommand(m.rpcRequestsCh),
		}

		if m.RpcHandler == nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Cannot serve requests while there are schema errors. Please see the CLI output for more info."))
			msg.done <- true
			return m, NextMsgCommand(m.runtimeRequestsCh)
		}

		cors := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		})
		cors.Handler(m.RpcHandler).ServeHTTP(msg.w, r)
		msg.done <- true
		return m, tea.Batch(cmds...)
	case WatcherMsg:
		m.Err = msg.Err
		m.Status = StatusLoadSchema

		// If the watcher errors then probably best to exit
		if m.Err != nil {
			return m, tea.Quit
		}

		m.Timings["watcherEvent"] = time.Now()

		return m, tea.Batch(
			NextMsgCommand(m.watcherCh),
			LoadSchema(m.ProjectDir, m.Environment),
		)
	}

	return m, nil
}

func (m *Model) View() string {
	b := strings.Builder{}

	// lipgloss will automatically wrap any text based on the current dimensions of the user's term.
	s := lipgloss.
		NewStyle().
		MaxWidth(m.width).
		MaxHeight(m.height)

	b.WriteString(renderRun(m))

	if m.Err != nil && m.Status != StatusQuitting {
		b.WriteString(renderError(m))
	}

	// The final "\n" is important as when Bubbletea exists it resets the last
	// line of output, meaning without a new line we'd lose the final line
	return s.Render(b.String() + "\n")
}
