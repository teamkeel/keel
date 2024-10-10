package main

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/teamkeel/keel/deploy/lambdas/runtime"
)

func main() {
	h, err := runtime.New(context.Background(), &runtime.HandlerArgs{
		LogLevel:       os.Getenv("KEEL_LOG_LEVEL"),
		SchemaPath:     "/var/task/schema.json",
		ConfigPath:     "/var/task/config.json",
		ProjectName:    os.Getenv("KEEL_PROJECT_NAME"),
		Env:            os.Getenv("KEEL_ENV"),
		QueueURL:       os.Getenv("KEEL_QUEUE_URL"),
		FunctionsARN:   os.Getenv("KEEL_FUNCTIONS_ARN"),
		BucketName:     os.Getenv("KEEL_FILES_BUCKET_NAME"),
		SecretNames:    strings.Split(os.Getenv("KEEL_SECRETS"), ":"),
		JobsWebhookURL: os.Getenv("KEEL_JOBS_WEBHOOK_URL"),
		DBEndpoint:     os.Getenv("KEEL_DATABASE_ENDPOINT"),
		DBName:         os.Getenv("KEEL_DATABASE_DB_NAME"),
		DBSecretArn:    os.Getenv("KEEL_DATABASE_SECRET_ARN"),
		TracingEnabled: os.Getenv("KEEL_TRACING_ENABLED") == "true",
	})
	if err != nil {
		panic(err)
	}

	switch os.Getenv("KEEL_RUNTIME_MODE") {
	case runtime.RuntimeModeApi:
		lambda.Start(h.APIHandler)
	case runtime.RuntimeModeSubscriber:
		lambda.Start(h.EventHandler)
	case runtime.RuntimeModeJob:
		lambda.Start(h.JobHandler)
	}
}
