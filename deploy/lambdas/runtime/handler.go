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

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/runtimectx"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	RuntimeModeApi        = "api"
	RuntimeModeSubscriber = "subscriber"
	RuntimeModeJob        = "job"
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
	filesStorage       *S3BucketStore
	tracer             trace.Tracer
	tracerProvider     *sdktrace.TracerProvider
}

type HandlerArgs struct {
	LogLevel       string
	SchemaPath     string
	ConfigPath     string
	ProjectName    string
	Env            string
	QueueURL       string
	FunctionsARN   string
	BucketName     string
	SecretNames    []string
	JobsWebhookURL string
	TracingEnabled bool

	// For RDS
	DBEndpoint  string
	DBName      string
	DBSecretArn string

	// For testing
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

	eventHandler, err := initEvents(args.QueueURL, args.AWSEndpoint)
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

	return ctx, nil
}

func SsmParameterName(projectName string, env string, paramName string) string {
	// TODO: add /secret/ before param name
	return fmt.Sprintf("/keel/%s/%s/%s", projectName, env, paramName)
}
