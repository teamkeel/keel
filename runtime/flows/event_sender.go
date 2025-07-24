package flows

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
)

type EventSender interface {
	Send(ctx context.Context, payload *EventWrapper) error
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

type NoQueueEventSender struct {
	orchestrator *Orchestrator
}

// compile time check that NoQueueEventSender implement the EventSender interface.
var _ EventSender = &NoQueueEventSender{}

func NewNoQueueEventSender(o *Orchestrator) *NoQueueEventSender {
	return &NoQueueEventSender{
		orchestrator: o,
	}
}

func (s *NoQueueEventSender) Send(ctx context.Context, payload *EventWrapper) error {
	go s.orchestrator.HandleEvent(ctx, payload) //nolint we're "simulating" an async queue

	return nil
}
