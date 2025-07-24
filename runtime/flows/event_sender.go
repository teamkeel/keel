package flows

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/robfig/cron/v3"
)

type EventSender interface {
	// Send sends the given payload onto the flwos queue.
	Send(ctx context.Context, payload *EventWrapper) error
	// Schedule will schedule sending the payload according to the given cron expression.
	Schedule(ctx context.Context, cronExpr string, payload *EventWrapper) error
}

type SQSEventSender struct {
	// Client for sqs messages sent to the flows runtime.
	sqsClient *sqs.Client
	// The Flows runtime queue used to trigger the execution of a flow
	sqsQueueURL string
}

// compile time check that SQSEventSender implement the EventSender interface.
var _ EventSender = &SQSEventSender{}

func NewSQSEventSender(queueURL string, client *sqs.Client) *SQSEventSender {
	return &SQSEventSender{
		sqsClient:   client,
		sqsQueueURL: queueURL,
	}
}

func (s *SQSEventSender) Send(ctx context.Context, payload *EventWrapper) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(bodyBytes)),
		QueueUrl:    aws.String(s.sqsQueueURL),
	}

	_, err = s.sqsClient.SendMessage(ctx, input)
	return err
}

func (s *SQSEventSender) Schedule(ctx context.Context, cronExpr string, payload *EventWrapper) error {
	// TODO: implement via eventbridge
	return nil
}

type NoQueueEventSender struct {
	orchestrator *Orchestrator
	cronRunner   *cron.Cron
}

// compile time check that NoQueueEventSender implement the EventSender interface.
var _ EventSender = &NoQueueEventSender{}

func NewNoQueueEventSender(o *Orchestrator, c *cron.Cron) *NoQueueEventSender {
	return &NoQueueEventSender{
		orchestrator: o,
		cronRunner:   c,
	}
}

func (s *NoQueueEventSender) Send(ctx context.Context, payload *EventWrapper) error {
	go s.orchestrator.HandleEvent(ctx, payload) //nolint we're "simulating" an async queue

	return nil
}

func (s *NoQueueEventSender) Schedule(ctx context.Context, cronExpr string, payload *EventWrapper) error {
	_, err := s.cronRunner.AddFunc(cronExpr, func() {
		s.orchestrator.HandleEvent(ctx, payload) //nolint
	})

	return err
}
