package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/segmentio/ksuid"
	"go.opentelemetry.io/otel/attribute"
)

const maxDelaySeconds = 20 // TODO: change this back to 900. currently set for testing purposes

type EventSender interface {
	// Send sends the given payload onto the flows queue. Optionally, the payload can be deferred to be sent after the
	// given scheduledAfter time.
	Send(ctx context.Context, payload *EventWrapper, scheduledAfter *time.Time) error
}

type SQSEventSender struct {
	// Client for sqs messages sent to the flows runtime.
	sqsClient *sqs.Client
	// Client for eventbridge schedules
	schedulerClient *scheduler.Client
	// The URL for the Flows runtime queue used to trigger the execution of a flow
	sqsQueueURL string
	// The ARN for the Flows runtime queue used to trigger the execution of a flow
	sqsQueueARN string
	// The ARN for the role used to schedule sqs messages
	scheduleRoleARN string
	// A prefix used for any one-off schedules names
	schedulePrefix string
}

// compile time check that SQSEventSender implement the EventSender interface.
var _ EventSender = &SQSEventSender{}

func NewSQSEventSender(queueURL, queueARN string, sqsClient *sqs.Client, schedulerClient *scheduler.Client, schedulePrefix string, scheduleRoleARN string) *SQSEventSender {
	return &SQSEventSender{
		sqsClient:       sqsClient,
		schedulerClient: schedulerClient,
		sqsQueueURL:     queueURL,
		sqsQueueARN:     queueARN,
		schedulePrefix:  schedulePrefix,
		scheduleRoleARN: scheduleRoleARN,
	}
}

func (s *SQSEventSender) Send(ctx context.Context, payload *EventWrapper, scheduledAfter *time.Time) error {
	// send with no delay
	if scheduledAfter == nil {
		return s.sendWithDelay(ctx, payload, 0)
	}

	// delayed for up to maxDelaySeconds
	if delaySeconds := math.Ceil(time.Until(*scheduledAfter).Seconds()); delaySeconds < maxDelaySeconds {
		return s.sendWithDelay(ctx, payload, int32(delaySeconds))
	}

	// delayed for more than maxDelaySeconds, this message needs to be scheduled
	return s.schedule(ctx, payload, *scheduledAfter)
}

// sendWithDelay will send the given payload with the given delay.
// If the delay is less than 1, payload will be sent immediately.
func (s *SQSEventSender) sendWithDelay(ctx context.Context, payload *EventWrapper, delaySeconds int32) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(bodyBytes)),
		QueueUrl:    aws.String(s.sqsQueueURL),
	}

	if delaySeconds > 0 {
		input.DelaySeconds = delaySeconds
	}

	_, err = s.sqsClient.SendMessage(ctx, input)
	return err
}

func (s *SQSEventSender) schedule(ctx context.Context, payload *EventWrapper, scheduledAt time.Time) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, span := tracer.Start(ctx, "scheduling event")
	span.SetAttributes(attribute.String("time", fmt.Sprintf("at(%s)", scheduledAt.Format("2006-01-02T15:04:05"))))
	defer span.End()

	input := scheduler.CreateScheduleInput{
		FlexibleTimeWindow: &types.FlexibleTimeWindow{
			Mode: types.FlexibleTimeWindowModeOff,
		},
		Name:               aws.String(fmt.Sprintf("%s-%s", s.schedulePrefix, ksuid.New().String())),
		ScheduleExpression: aws.String(fmt.Sprintf("at(%s)", scheduledAt.Format("2006-01-02T15:04:05"))),
		Target: &types.Target{
			Arn:     &s.sqsQueueARN,
			RoleArn: aws.String(s.scheduleRoleARN),
			Input:   aws.String(string(bodyBytes)),
		},
		ScheduleExpressionTimezone: aws.String("UTC"),
		StartDate:                  aws.Time(time.Now()),
		State:                      types.ScheduleStateEnabled,
		ActionAfterCompletion:      types.ActionAfterCompletionDelete,
	}

	_, err = s.schedulerClient.CreateSchedule(ctx, &input)
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
