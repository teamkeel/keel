package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	ebTypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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
	// Client for eventbridge
	ebClient *eventbridge.Client
	// The Flows runtime queue used to trigger the execution of a flow
	sqsQueueURL string
}

// compile time check that SQSEventSender implement the EventSender interface.
var _ EventSender = &SQSEventSender{}

func NewSQSEventSender(queueURL string, sqsClient *sqs.Client, ebClient *eventbridge.Client) *SQSEventSender {
	return &SQSEventSender{
		sqsClient:   sqsClient,
		sqsQueueURL: queueURL,
		ebClient:    ebClient,
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
	// retrieve the sqs queue ARN
	resp, err := s.sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: &s.sqsQueueURL,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameQueueArn,
		},
	})
	if err != nil {
		return fmt.Errorf("failed retrieving queue ARN: %w", err)
	}
	queueARN := resp.Attributes[string(types.QueueAttributeNameQueueArn)]

	var ev FlowRunStarted
	if err := ev.ReadPayload(payload); err != nil {
		return err
	}

	ruleName := "ScheduledFlow" + ev.Name

	_, err = s.ebClient.PutRule(ctx, &eventbridge.PutRuleInput{
		Name:               &ruleName,
		ScheduleExpression: &cronExpr,
		State:              ebTypes.RuleStateEnabled,
	})
	if err != nil {
		return fmt.Errorf("creating scheduled rule: %w", err)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = s.ebClient.PutTargets(ctx, &eventbridge.PutTargetsInput{
		Rule: &ruleName,
		Targets: []ebTypes.Target{
			{
				Id:    &ruleName,
				Arn:   &queueARN,
				Input: aws.String(string(bodyBytes)),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("setting rule target: %w", err)
	}

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
	// Our cron expressions for schedules include the year, which is not relevant to our use case.
	schedule := strings.TrimSuffix(cronExpr, " *")

	_, err := s.cronRunner.AddFunc(schedule, func() {
		s.orchestrator.HandleEvent(ctx, payload) //nolint
	})

	return err
}
