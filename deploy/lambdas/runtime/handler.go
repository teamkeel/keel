package runtime

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/storage"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/flows"
	"github.com/teamkeel/keel/runtime/runtimectx"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// Handles API requests.
	RuntimeModeApi = "api"

	// Handles events from SQS.
	RuntimeModeSubscriber = "subscriber"

	// Handles both manual and scheduled jobs.
	RuntimeModeJob = "job"

	// Handles flows.
	RuntimeModeFlow = "flow"
)

type Handler struct {
	args               *HandlerArgs
	log                *logrus.Logger
	schema             *proto.Schema
	config             *config.ProjectConfig
	secrets            map[string]string
	privateKey         *rsa.PrivateKey
	db                 db.Database
	functionsTransport functions.Transport
	sqsEventHandler    events.EventHandler
	filesStorage       *storage.S3BucketStore
	tracer             trace.Tracer
	tracerProvider     *sdktrace.TracerProvider
	flowOrchestrator   *flows.Orchestrator
}

type HandlerArgs struct {
	// One of the logrus log levels, if omitted defaults to "error"
	LogLevel string

	// File-system path for the proto JSON for the Keel schema
	SchemaPath string
	// File-system path to a JSON file containing the Keel config. Note this should be a JSON file, not YAML.
	ConfigPath string
	// The project name. This needs to be provided as for local environments there won't necessarily be a project name in the config.
	ProjectName string
	// The env. For local environments will be "development" or "test", for deployed environments it's the user-provided env name.
	Env string
	// URL of SQS queue for subscriber events
	EventsQueueURL string
	// URL of SQS queue used to trigger orchestrate flows
	FlowsQueueURL string
	// ARN fo the iam role for scheduling
	SchedulerRoleARN string
	// Full ARN of functions Lambda.
	FunctionsARN string
	// Bucket name used for files and job inputs
	BucketName string
	// List of secret names to looad from SSM
	SecretNames []string
	// Webhook URL to use for sending job run updates
	JobsWebhookURL string
	// If true then tracing data will be exported using the GRPC exporter, otherwise a no-nop exporter will be used.
	TracingEnabled bool

	// RDS config (can be omitted if using an external db or local postgres)
	DBEndpoint  string
	DBName      string
	DBSecretArn string

	// If this is set all AWS operations will be directed to this endpoint by configuring the clients to use it.
	// This is used for mocking the API calls in integration tests
	AWSEndpoint string
}

func New(ctx context.Context, args *HandlerArgs) (*Handler, error) {
	tracer, tracerProvider, err := initTracing(args.TracingEnabled)
	if err != nil {
		return nil, err
	}

	s, err := initSchema(args.SchemaPath)
	if err != nil {
		return nil, err
	}

	c, err := initConfig(args.ConfigPath)
	if err != nil {
		return nil, err
	}

	secrets, err := initSecrets(ctx, args.SecretNames, args.ProjectName, args.Env, args.AWSEndpoint)
	if err != nil {
		return nil, err
	}

	pk, err := initPrivateKey(secrets)
	if err != nil {
		return nil, err
	}

	db, err := initDB(secrets, args.DBEndpoint, args.DBName, args.DBSecretArn)
	if err != nil {
		return nil, err
	}

	eventHandler, err := initEvents(args.EventsQueueURL, args.AWSEndpoint)
	if err != nil {
		return nil, err
	}

	files, err := initFiles(ctx, tracer, args.BucketName, args.AWSEndpoint)
	if err != nil {
		return nil, err
	}

	log, err := initLogger(args.LogLevel)
	if err != nil {
		return nil, err
	}

	functionsTransport, err := initFunctions(ctx, args.FunctionsARN, args.AWSEndpoint)
	if err != nil {
		return nil, err
	}

	flowOrchestrator, err := initOrchestrator(ctx, args.FlowsQueueURL, args.AWSEndpoint, args.SchedulerRoleARN, s)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		args:               args,
		log:                log,
		schema:             s,
		config:             c,
		privateKey:         pk,
		secrets:            secrets,
		db:                 db,
		functionsTransport: functionsTransport,
		sqsEventHandler:    eventHandler,
		filesStorage:       files,
		tracer:             tracer,
		tracerProvider:     tracerProvider,
		flowOrchestrator:   flowOrchestrator,
	}

	return h, nil
}

func (h *Handler) Stop() error {
	return h.db.Close()
}

func initSchema(path string) (*proto.Schema, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s proto.Schema
	err = protojson.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func initConfig(path string) (*config.ProjectConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c config.ProjectConfig
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func initPrivateKey(secrets map[string]string) (*rsa.PrivateKey, error) {
	privateKeyPem, ok := secrets["KEEL_PRIVATE_KEY"]
	if !ok {
		return nil, errors.New("missing KEEL_PRIVATE_KEY secret")
	}

	privateKeyBlock, _ := pem.Decode([]byte(privateKeyPem))
	if privateKeyBlock == nil {
		return nil, errors.New("error decoding private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func initLogger(level string) (*logrus.Logger, error) {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)

	if level == "" {
		level = "error"
	}

	l, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	log.SetLevel(l)
	return log, nil
}

func (h *Handler) buildContext(ctx context.Context) (context.Context, error) {
	ctx = runtimectx.WithOAuthConfig(ctx, &h.config.Auth)
	ctx = runtimectx.WithPrivateKey(ctx, h.privateKey)
	ctx = runtimectx.WithSecrets(ctx, h.secrets)
	ctx = runtimectx.WithStorage(ctx, h.filesStorage)
	ctx = db.WithDatabase(ctx, h.db)
	ctx = functions.WithFunctionsTransport(ctx, h.functionsTransport)

	ctx, err := events.WithEventHandler(ctx, h.sqsEventHandler)
	if err != nil {
		return nil, err
	}

	ctx = flows.WithOrchestrator(ctx, h.flowOrchestrator)

	return ctx, nil
}

func SsmParameterName(projectName string, env string, paramName string) string {
	return fmt.Sprintf("/keel/%s/%s/%s", projectName, env, paramName)
}
