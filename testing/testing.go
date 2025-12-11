package testing

import (
	"bytes"
	"context"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	lambdaevents "github.com/aws/aws-lambda-go/events"
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
	"go.opentelemetry.io/otel"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/deploy"
	"github.com/teamkeel/keel/deploy/lambdas/runtime"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/apis/httpjson"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/util"
)

const (
	AuthPath        = "auth"
	ActionApiPath   = "testingactionsapi"
	FlowApiPath     = "flows"
	JobPath         = "testingjobs"
	SubscriberPath  = "testingsubscribers"
	JobsWebhookPath = "/webhooks/jobs"
)

type TestOutput struct {
	Output  string
	Success bool
}

type RunnerOpts struct {
	Dir           string
	Pattern       string
	DbConnInfo    *db.ConnectionInfo
	Secrets       map[string]string
	TestGroupName string
	// Generates a Keel client (keelClient.ts) for this test
	GenerateClient bool
}

var tracer = otel.Tracer("github.com/teamkeel/keel/testing")

func Run(ctx context.Context, opts *RunnerOpts) error {
	buildRessult, err := deploy.Build(ctx, &deploy.BuildArgs{
		ProjectRoot: opts.Dir,
		Env:         "test",
		OnLoadSchema: func(schema *proto.Schema) *proto.Schema {
			testApi := &proto.Api{
				Name: ActionApiPath,
			}

			for _, m := range schema.GetModels() {
				apiModel := &proto.ApiModel{
					ModelName:    m.GetName(),
					ModelActions: []*proto.ApiModelAction{},
				}

				testApi.ApiModels = append(testApi.ApiModels, apiModel)
				for _, a := range m.GetActions() {
					apiModel.ModelActions = append(apiModel.ModelActions, &proto.ApiModelAction{ActionName: a.GetName()})
				}
			}

			schema.Apis = append(schema.Apis, testApi)
			return schema
		},
	})
	if err != nil {
		return err
	}

	schema := buildRessult.Schema
	config := buildRessult.Config
	envVars := config.GetEnvVars()

	spanName := opts.TestGroupName
	if spanName == "" {
		spanName = "testing.Run"
	}

	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	dbName := "keel_test"
	database, err := testhelpers.SetupDatabaseForTestCase(ctx, opts.DbConnInfo, schema, dbName, true)
	if err != nil {
		return err
	}
	defer database.Close()

	dbConnString := opts.DbConnInfo.WithDatabase(dbName).String()

	if opts.GenerateClient {
		clientFiles, err := node.GenerateClient(
			ctx,
			schema,
			false,
			ActionApiPath,
		)
		if err != nil {
			return err
		}

		err = clientFiles.Write(opts.Dir)
		if err != nil {
			return err
		}
	}

	runtimePort, err := util.GetFreePort()
	if err != nil {
		return err
	}

	serverURL := fmt.Sprintf("http://localhost:%s", runtimePort)
	bucketName := "testing-bucket-name"
	functionsARN := "arn:test:lambda:functions:function"

	// Generate private key early so it can be passed to functions server
	pk, err := testhelpers.GetEmbeddedPrivateKey()
	if err != nil {
		return err
	}

	pkBytes := x509.MarshalPKCS1PrivateKey(pk)
	pkPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: pkBytes,
		},
	)
	pkBase64 := base64.StdEncoding.EncodeToString(pkPem)

	var functionsServer *node.DevelopmentServer

	if node.HasFunctions(schema, config) {
		functionEnvVars := map[string]string{
			"KEEL_DB_CONN_TYPE":      "pg",
			"KEEL_TRACING_ENABLED":   os.Getenv("TRACING_ENABLED"),
			"KEEL_FILES_BUCKET_NAME": bucketName,

			// Send all AWS API calls to our test server and set some test credentials
			"TEST_AWS_ENDPOINT":     fmt.Sprintf("%s/aws", serverURL),
			"AWS_ACCESS_KEY_ID":     "test",
			"AWS_SECRET_ACCESS_KEY": "test",
			"AWS_SESSION_TOKEN":     "test",
			"AWS_REGION":            "test",

			"OTEL_RESOURCE_ATTRIBUTES": "service.name=functions",

			// Private key for JWT signing (used by tasks SDK withIdentity)
			"KEEL_PRIVATE_KEY": pkBase64,

			// API URL for tasks SDK
			"KEEL_API_URL": serverURL,
		}

		maps.Copy(functionEnvVars, envVars)

		functionsServer, err = node.StartDevelopmentServer(ctx, opts.Dir, &node.ServerOpts{
			EnvVars: functionEnvVars,
			Output:  os.Stdout,
			Debug:   os.Getenv("DEBUG_FUNCTIONS") == "true",
			Watch:   false,
		})

		if err != nil {
			if functionsServer != nil && functionsServer.Output() != "" {
				return errors.New(functionsServer.Output())
			}
			return err
		}

		defer func() {
			_ = functionsServer.Kill()
		}()
	}

	for key, value := range envVars {
		os.Setenv(key, value)
	}

	// This is needed by the auth endpoints to return the right URL's
	os.Setenv("KEEL_API_URL", fmt.Sprintf("http://localhost:%s", runtimePort))

	var lambdaHandler *runtime.Handler

	ssmParams := map[string]string{
		"KEEL_PRIVATE_KEY": string(pkPem),
		"DATABASE_URL":     dbConnString,
	}
	for k, v := range opts.Secrets {
		ssmParams[k] = v
	}

	awsHandler := &AWSAPIHandler{
		PathPrefix:    "/aws/",
		FunctionsARN:  functionsARN,
		SSMParameters: ssmParams,
		// TODO: consider doing this in a go routine to make it async
		// but current tests require it to be sync
		OnSQSEvent: map[string]func(lambdaevents.SQSEvent){
			"https://testing-sqs-queue.com/123456789/events": func(event lambdaevents.SQSEvent) {
				// TODO: consider doing this in a go routine to make it async
				// but current tests require it to be sync
				err := lambdaHandler.EventHandler(ctx, event)
				if err != nil {
					fmt.Printf("error from event handler: %s\nevent:%s\n\n", err.Error(), event.Records[0].Body)
				}
			},
			"https://testing-sqs-queue.com/123456789/flows": func(event lambdaevents.SQSEvent) {
				go func() {
					err := lambdaHandler.FlowHandler(ctx, event)
					if err != nil {
						fmt.Printf("error from flow orchestrator: %s\nevent:%s\n\n", err.Error(), event.Records[0].Body)
					}
				}()
			},
		},
	}
	if functionsServer != nil {
		awsHandler.FunctionsURL = functionsServer.URL
	}

	// This server handles requests from the ActionExecutor, JobExecutor and SubscriberExecutor in the Vitest tests
	// but also AWS API calls which come here because we set a custom endpoint on the clients
	runtimeServer := http.Server{
		Addr: fmt.Sprintf(":%s", runtimePort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), strings.Trim(r.URL.Path, "/"))
			defer span.End()

			// Handle AWS API call
			if strings.HasPrefix(r.URL.Path, "/aws") {
				awsHandler.HandleHTTP(r, w)
				return
			}

			if r.URL.Path == JobsWebhookPath {
				// not doing anything with this for now but can do in the future...
				writeJSON(w, http.StatusOK, map[string]any{})
				return
			}

			// Handle API calls, jobs and subscriber executors
			pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

			switch pathParts[0] {
			case JobPath:
				err := HandleJobExecutorRequest(r, lambdaHandler, awsHandler)
				if err != nil {
					response := httpjson.NewErrorResponse(ctx, err, nil)
					w.WriteHeader(response.Status)
					_, _ = w.Write(response.Body)
					return
				}

				writeJSON(w, http.StatusOK, map[string]any{})
				return
			case SubscriberPath:
				err = HandleSubscriberExecutorRequest(r.WithContext(ctx), lambdaHandler)
				if err != nil {
					response := httpjson.NewErrorResponse(ctx, err, nil)
					w.WriteHeader(response.Status)
					_, _ = w.Write(response.Body)
					return
				}

				writeJSON(w, http.StatusOK, map[string]any{})
				return
			default:
				e, err := toLambdaFunctionURLRequest(r)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, err.Error())
					return
				}
				res, err := lambdaHandler.APIHandler(ctx, e)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, err.Error())
					return
				}
				for k, v := range res.Headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(res.StatusCode)
				_, _ = w.Write([]byte(res.Body))
			}
		}),
	}

	go func() {
		_ = runtimeServer.ListenAndServe()
	}()

	defer func() {
		_ = runtimeServer.Shutdown(ctx)
	}()

	// Small sleep to make sure the server has started as runtime.New will start making requests to it
	time.Sleep(time.Millisecond * 200)

	// We need to set these as even though we are using a custom endpoint and mocking requests the AWS clients
	// still expect to be able to send auth headers, and to do that they read these values from the env. Running locally
	// you might just have these set so it's ok, but in CI they are not available and tests fail
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	os.Setenv("AWS_SESSION_TOKEN", "test")
	os.Setenv("AWS_REGION", "test")

	lambdaHandler, err = runtime.New(ctx, &runtime.HandlerArgs{
		LogLevel:       "warn",
		SchemaPath:     path.Join(opts.Dir, ".build/runtime/schema.json"),
		ConfigPath:     path.Join(opts.Dir, ".build/runtime/config.json"),
		ProjectName:    opts.TestGroupName,
		Env:            "test",
		EventsQueueURL: "https://testing-sqs-queue.com/123456789/events",
		FlowsQueueURL:  "https://testing-sqs-queue.com/123456789/flows",
		FunctionsARN:   functionsARN,
		BucketName:     bucketName,
		SecretNames:    lo.Keys(ssmParams),
		JobsWebhookURL: fmt.Sprintf("%s%s", serverURL, JobsWebhookPath),

		// Send all AWS API calls to our test server
		AWSEndpoint: fmt.Sprintf("%s/aws", serverURL),
	})
	if err != nil {
		fmt.Println("error creating lambda runtime handler:", err)
		return err
	}
	defer func() {
		_ = lambdaHandler.Stop()
	}()

	cmd := exec.Command("./node_modules/.bin/tsc", "--noEmit", "--pretty")
	cmd.Dir = opts.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	if opts.Pattern == "" {
		opts.Pattern = "(.*)"
	}

	cmd = exec.Command("./node_modules/.bin/vitest", "run", "--color", "--reporter", "verbose", "--config", "./.build/vitest.config.mjs", "--silent", "false", "--testNamePattern", opts.Pattern, "--pool=threads", "--poolOptions.threads.singleThread")
	cmd.Dir = opts.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), []string{
		fmt.Sprintf("KEEL_TESTING_API_URL=%s", serverURL),
		fmt.Sprintf("KEEL_TESTING_ACTIONS_API_URL=%s/%s/json", serverURL, ActionApiPath),
		fmt.Sprintf("KEEL_TESTING_FLOWS_API_URL=%s/%s/json", serverURL, FlowApiPath),
		fmt.Sprintf("KEEL_TESTING_JOBS_URL=%s/%s/json", serverURL, JobPath),
		fmt.Sprintf("KEEL_TESTING_SUBSCRIBERS_URL=%s/%s/json", serverURL, SubscriberPath),
		fmt.Sprintf("KEEL_TESTING_CLIENT_API_URL=%s/%s", serverURL, ActionApiPath),
		fmt.Sprintf("KEEL_TESTING_AUTH_API_URL=%s/%s", serverURL, AuthPath),
		"KEEL_DB_CONN_TYPE=pg",
		fmt.Sprintf("KEEL_DB_CONN=%s", dbConnString),
		// Disables experimental fetch warning that pollutes console experience when running tests
		"NODE_NO_WARNINGS=1",
		fmt.Sprintf("KEEL_PRIVATE_KEY=%s", pkBase64),

		// Need to set these so the sdk uses the test endpoint in tests
		fmt.Sprintf("TEST_AWS_ENDPOINT=%s/aws", serverURL),
		fmt.Sprintf("KEEL_FILES_BUCKET_NAME=%s", bucketName),
	}...)

	return cmd.Run()
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	b, _ := json.Marshal(body)
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}

// HandleJobExecutorRequest handles requests to the job module in the testing package.
func HandleJobExecutorRequest(r *http.Request, h *runtime.Handler, awsHandler *AWSAPIHandler) error {
	id := ksuid.New().String()
	key := fmt.Sprintf("jobs/%s", id)

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	awsHandler.S3Bucket[key] = &S3Object{
		Headers: map[string][]string{
			"content-type": {"application/json"},
		},
		Data: b,
	}

	token := ""
	header := r.Header.Get("Authorization")
	if header != "" {
		authParts := strings.Split(header, "Bearer ")
		if len(authParts) == 2 {
			token = authParts[1]
		}
	}

	name := path.Base(r.URL.Path)
	name = strcase.ToCamel(name)

	return h.JobHandler(r.Context(), &runtime.RunJobPayload{
		ID:    id,
		Name:  name,
		Token: token,
	})
}

// HandleSubscriberExecutorRequest handles requests to the subscriber module in the testing package.
func HandleSubscriberExecutorRequest(r *http.Request, h *runtime.Handler) error {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var event *events.Event
	err = json.Unmarshal(b, &event)
	if err != nil {
		return err
	}

	b, err = json.Marshal(runtime.EventPayload{
		Subscriber:  path.Base(r.URL.Path),
		Event:       event,
		Traceparent: "1234",
	})
	if err != nil {
		return err
	}

	return h.EventHandler(r.Context(), lambdaevents.SQSEvent{
		Records: []lambdaevents.SQSMessage{
			{
				MessageId: "",
				Body:      string(b),
			},
		},
	})
}

func toLambdaFunctionURLRequest(r *http.Request) (lambdaevents.LambdaFunctionURLRequest, error) {
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	queryStringParameters := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			queryStringParameters[key] = values[0]
		}
	}

	var body string
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return lambdaevents.LambdaFunctionURLRequest{}, fmt.Errorf("failed to read request body: %w", err)
		}
		body = string(bodyBytes)
		// Reset the body for future use
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	return lambdaevents.LambdaFunctionURLRequest{
		Version:               "2.0",
		RawPath:               r.URL.Path,
		RawQueryString:        r.URL.RawQuery,
		Headers:               headers,
		QueryStringParameters: queryStringParameters,
		RequestContext: lambdaevents.LambdaFunctionURLRequestContext{
			HTTP: lambdaevents.LambdaFunctionURLRequestContextHTTPDescription{
				Method:    r.Method,
				Path:      r.URL.Path,
				Protocol:  r.Proto,
				SourceIP:  r.RemoteAddr,
				UserAgent: r.UserAgent(),
			},
		},
		Body:            body,
		IsBase64Encoded: false,
	}, nil
}
