package runtime

import (
	"context"
	"encoding/json"
	"fmt"

	lambdaevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/flows"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func initOrchestrator(ctx context.Context, queueURL string, awsEndpoint string, schema *proto.Schema) (*flows.Orchestrator, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	opts := []func(*sqs.Options){}
	if awsEndpoint != "" {
		opts = append(opts, func(o *sqs.Options) {
			o.BaseEndpoint = &awsEndpoint
		})
	}

	sqsClient := sqs.NewFromConfig(cfg, opts...)

	return flows.NewOrchestrator(schema, flows.WithAsyncQueue(queueURL, sqsClient)), nil
}

func (h *Handler) FlowHandler(ctx context.Context, event lambdaevents.SQSEvent) error {
	defer func() {
		if h.tracerProvider != nil {
			h.tracerProvider.ForceFlush(ctx)
		}
	}()

	if len(event.Records) != 1 {
		return fmt.Errorf("flow lambda is only designed to process exactly one message at a time, received %v", len(event.Records))
	}

	message := event.Records[0]

	var wrapper flows.EventWrapper
	err := json.Unmarshal([]byte(message.Body), &wrapper)
	if err != nil {
		return err
	}

	h.log.WithFields(logrus.Fields{
		"eventName":    wrapper.EventName,
		"eventPayload": wrapper.Payload,
	}).Info("Event received")

	// Use the span context from the event payload, which
	// originates from the runtime execution that triggered the event.
	spanContext := util.ParseTraceparent(wrapper.Traceparent)
	if spanContext.IsValid() {
		ctx = trace.ContextWithSpanContext(ctx, spanContext)
	}

	ctx, span := h.tracer.Start(ctx, "Process event")
	defer span.End()

	span.SetAttributes(
		attribute.String("type", "event"),
		attribute.String("event.name", wrapper.EventName),
		attribute.String("event.messageId", message.MessageId),
		attribute.String("event.payload", wrapper.Payload),
	)

	ctx, err = h.buildContext(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	o, err := flows.GetOrchestrator(ctx)
	if err != nil {
		return err
	}

	return o.HandleEvent(ctx, &wrapper)
}
