package runtime

import (
	"context"
	"encoding/json"
	"fmt"

	lambdaevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/flows"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func initOrchestrator(ctx context.Context, queueURL, awsEndpoint, roleARN string, schema *proto.Schema) (*flows.Orchestrator, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	opts := []func(*sqs.Options){}
	schOpts := []func(*scheduler.Options){}

	if awsEndpoint != "" {
		opts = append(opts, func(o *sqs.Options) {
			o.BaseEndpoint = &awsEndpoint
		})
		schOpts = append(schOpts, func(o *scheduler.Options) {
			o.BaseEndpoint = &awsEndpoint
		})
	}

	sqsClient := sqs.NewFromConfig(cfg, opts...)
	schedulerClient := scheduler.NewFromConfig(cfg, schOpts...)

	// retrieve queue's arn
	attributes, err := sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: &queueURL,
		AttributeNames: []sqstypes.QueueAttributeName{
			sqstypes.QueueAttributeNameQueueArn,
		},
	})
	if err != nil {
		return nil, err
	}

	queueARN := attributes.Attributes["QueueArn"]

	return flows.NewOrchestrator(schema, flows.WithAsyncQueue(queueURL, queueARN, sqsClient, schedulerClient, "", roleARN)), nil
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
