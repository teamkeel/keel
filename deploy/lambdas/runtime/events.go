package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	lambdaevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var sqsEventHandler events.EventHandler

type Payload struct {
	Subscriber  string        `json:"subscriber,omitempty"`
	Event       *events.Event `json:"event,omitempty"`
	Traceparent string        `json:"traceparent,omitempty"`
}

func initEvents() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	sqsEventHandler = NewSqsEventHandler(sqs.NewFromConfig(cfg))
}

func NewSqsEventHandler(client *sqs.Client) events.EventHandler {
	return func(ctx context.Context, subscriber string, event *events.Event, traceparent string) error {
		return sendEvent(ctx, client, subscriber, event, traceparent)
	}
}

func sendEvent(ctx context.Context, client *sqs.Client, subscriber string, event *events.Event, traceparent string) error {
	payload := Payload{
		Subscriber:  subscriber,
		Event:       event,
		Traceparent: traceparent,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(bodyBytes)),
		QueueUrl:    aws.String(os.Getenv("KEEL_QUEUE_URL")),
	}

	_, err = client.SendMessage(ctx, input)
	return err
}

func eventHandler(ctx context.Context, event lambdaevents.SQSEvent) error {
	defer func() {
		if tracerProvider != nil {
			tracerProvider.ForceFlush(ctx)
		}
	}()

	if len(event.Records) != 1 {
		return fmt.Errorf("event lambda is only designed to process exactly one message at a time, received %v", len(event.Records))
	}

	message := event.Records[0]

	var payload Payload
	err := json.Unmarshal([]byte(message.Body), &payload)
	if err != nil {
		return err
	}

	if payload.Event == nil {
		return errors.New("event is nil")
	}

	log.WithFields(logrus.Fields{
		"subscriber": payload.Subscriber,
		"eventName":  payload.Event.EventName,
		"type":       payload.Event.Target.Type,
		"id":         payload.Event.Target.Id,
	}).Info("Event received")

	// Use the span context from the event payload, which
	// originates from the runtime execution that triggered the event.
	spanContext := util.ParseTraceparent(payload.Traceparent)
	if spanContext.IsValid() {
		ctx = trace.ContextWithSpanContext(ctx, spanContext)
	}

	ctx, span := tracer.Start(ctx, "Process event")
	defer span.End()

	span.SetAttributes(
		attribute.String("type", "event"),
		attribute.String("event.name", payload.Event.EventName),
		attribute.String("event.messageId", message.MessageId),
		attribute.String("subscriber.name", payload.Subscriber),
	)

	ctx = runtimectx.WithOAuthConfig(ctx, &keelConfig.Auth)
	ctx = runtimectx.WithPrivateKey(ctx, privateKey)
	ctx = runtimectx.WithSecrets(ctx, secrets)
	ctx = runtimectx.WithStorage(ctx, NewS3BucketStore(ctx))
	ctx = db.WithDatabase(ctx, dbConn)
	ctx = functions.WithFunctionsTransport(ctx, NewLambdaInvokeTransport(os.Getenv("KEEL_FUNCTIONS_ARN")))

	ctx, err = events.WithEventHandler(ctx, sqsEventHandler)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	err = runtime.NewSubscriberHandler(schema).RunSubscriber(ctx, payload.Subscriber, payload.Event)
	return err
}
