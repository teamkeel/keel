package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
)

const maxDelaySeconds = 900

type EventSender interface {
	// Send sends the given payload onto the flows queue. Optionally, the payload can be deferred to be sent after the
	// given scheduledAfter time.
	Send(ctx context.Context, payload *EventWrapper, scheduledAfter *time.Time) error
}

type SQSEventSender struct {
	// Client for sqs messages sent to the flows runtime.
	sqsClient *sqs.Client
	// The Flows runtime queue used to trigger the execution of a flow
	sqsQueueURL string
}

// compile time check that SQSEventSender implement the EventSender interface.
var _ EventSender = &SQSEventSender{}

func NewSQSEventSender(queueURL string, sqsClient *sqs.Client) *SQSEventSender {
	return &SQSEventSender{
		sqsClient:   sqsClient,
		sqsQueueURL: queueURL,
	}
}

func (s *SQSEventSender) Send(ctx context.Context, payload *EventWrapper, scheduledAfter *time.Time) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(bodyBytes)),
		QueueUrl:    aws.String(s.sqsQueueURL),
	}

	if scheduledAfter != nil {
		if delaySeconds := time.Until(*scheduledAfter).Seconds(); delaySeconds > 0 {
			if delaySeconds < maxDelaySeconds {
				// sqs message can be delayed
				input.DelaySeconds = int32(delaySeconds)
			} else {
				return fmt.Errorf("delay exceeds maximum supported period of %d seconds", maxDelaySeconds)
			}
		}
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

func (s *NoQueueEventSender) Send(ctx context.Context, payload *EventWrapper, scheduledAfter *time.Time) error {
	go func() {
		if scheduledAfter != nil {
			time.Sleep(time.Until(*scheduledAfter))
		}

		s.orchestrator.HandleEvent(ctx, payload) //nolint we're "simulating" an async queue
	}()

	return nil
}
