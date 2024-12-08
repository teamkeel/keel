package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	lambdaevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/events"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type EventPayload struct {
	Subscriber  string        `json:"subscriber,omitempty"`
	Event       *events.Event `json:"event,omitempty"`
	Traceparent string        `json:"traceparent,omitempty"`
}

func initEvents(queueUrl string, awsEndpoint string) (events.EventHandler, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	opts := []func(*sqs.Options){}
	if awsEndpoint != "" {
		opts = append(opts, func(o *sqs.Options) {
			o.BaseEndpoint = &awsEndpoint
		})
	}

	client := sqs.NewFromConfig(cfg, opts...)

	return func(ctx context.Context, subscriber string, event *events.Event, traceparent string) error {
		return sendEvent(ctx, client, queueUrl, subscriber, event, traceparent)
	}, nil
}

func sendEvent(ctx context.Context, client *sqs.Client, queueURL string, subscriber string, event *events.Event, traceparent string) error {
	payload := EventPayload{
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
		QueueUrl:    aws.String(queueURL),
	}

	_, err = client.SendMessage(ctx, input)
	return err
}

func (h *Handler) EventHandler(ctx context.Context, event lambdaevents.SQSEvent) error {
	defer func() {
		if h.tracerProvider != nil {
			h.tracerProvider.ForceFlush(ctx)
		}
	}()

	if len(event.Records) != 1 {
		return fmt.Errorf("event lambda is only designed to process exactly one message at a time, received %v", len(event.Records))
	}

	message := event.Records[0]

	var payload EventPayload
	err := json.Unmarshal([]byte(message.Body), &payload)
	if err != nil {
		return err
	}

	if payload.Event == nil {
		return errors.New("event is nil")
	}

	h.log.WithFields(logrus.Fields{
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

	ctx, span := h.tracer.Start(ctx, "Process event")
	defer span.End()

	span.SetAttributes(
		attribute.String("type", "event"),
		attribute.String("event.name", payload.Event.EventName),
		attribute.String("event.messageId", message.MessageId),
		attribute.String("subscriber.name", payload.Subscriber),
	)

	ctx, err = h.buildContext(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	err = runtime.NewSubscriberHandler(h.schema).RunSubscriber(ctx, payload.Subscriber, payload.Event)
	return err
}
