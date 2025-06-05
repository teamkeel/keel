package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/actions"
	"github.com/teamkeel/keel/runtime/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type RunJobPayload struct {
	// An ID for this job, if provided then it will be sent in the webhook status payload. Will be empty for scheduled jobs.
	ID string `json:"id"`
	// The name of the job to run e.g. MySpecialJob
	Name string `json:"name"`
	// An auth token to use to determine the identity running the job. Will be empty for scheduled jobs.
	Token string `json:"token"`
}

type JobStatusWebhookPayload struct {
	// The ID provided in the payload
	ID string `json:"id"`
	// The name of the job
	Name        string `json:"name"`
	ProjectName string `json:"projectName"`
	Env         string `json:"env"`
	// One of "processing", "success", or "failed"
	Status string `json:"status"`
	// The OTEL trace ID for the job run
	TraceID   string `json:"traceId"`
	Timestamp string `json:"timestamp"`
}

const (
	JobStatusProcessing = "processing"
	JobStatusSuccess    = "success"
	JobStatusFailed     = "failed"
)

func (h *Handler) JobHandler(ctx context.Context, event *RunJobPayload) error {
	defer func() {
		if h.tracerProvider != nil {
			h.tracerProvider.ForceFlush(ctx)
		}
	}()

	ctx, span := h.tracer.Start(ctx, event.Name)
	defer span.End()

	span.SetAttributes(attribute.String("job.name", event.Name))
	span.SetAttributes(attribute.String("job.id", event.ID))

	log := h.log.WithFields(logrus.Fields{
		"jobName": event.Name,
	})

	job := h.schema.FindJob(event.Name)
	if job == nil {
		err := fmt.Errorf("no job found with name %s", event.Name)
		return h.sendJobStatusWebhook(ctx, event, err, JobStatusFailed)
	}

	log.Infof("Running job %s", job.GetName())

	_ = h.sendJobStatusWebhook(ctx, event, nil, JobStatusProcessing)

	ctx, err := h.buildContext(ctx)
	if err != nil {
		return h.sendJobStatusWebhook(ctx, event, err, JobStatusFailed)
	}

	if event.Token != "" {
		identity, err := actions.HandleBearerToken(ctx, h.schema, event.Token)
		if err != nil {
			return h.sendJobStatusWebhook(ctx, event, err, JobStatusFailed)
		}

		if identity != nil {
			ctx = auth.WithIdentity(ctx, identity)
		}
	}

	inputs := map[string]any{}
	if job.GetInputMessageName() != "" {
		if event.ID == "" {
			err = fmt.Errorf("no ref provided but job requires inputs")
			return h.sendJobStatusWebhook(ctx, event, err, JobStatusFailed)
		}

		object, err := h.filesStorage.client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(h.filesStorage.bucketName),
			// Note: this prefix is important and must be used when uploading the job input data to S3
			Key: aws.String(fmt.Sprintf("jobs/%s", event.ID)),
		})
		if err != nil {
			return h.sendJobStatusWebhook(ctx, event, err, JobStatusFailed)
		}

		b, err := io.ReadAll(object.Body)
		if err != nil {
			return h.sendJobStatusWebhook(ctx, event, err, JobStatusFailed)
		}

		err = json.Unmarshal(b, &inputs)
		if err != nil {
			return h.sendJobStatusWebhook(ctx, event, err, JobStatusFailed)
		}
	}

	trigger := functions.ManualTrigger
	if job.GetSchedule() != nil {
		trigger = functions.ScheduledTrigger
	}

	err = runtime.NewJobHandler(h.schema).RunJob(ctx, job.GetName(), inputs, trigger)
	if err != nil {
		return h.sendJobStatusWebhook(ctx, event, err, JobStatusFailed)
	}

	return h.sendJobStatusWebhook(ctx, event, err, JobStatusSuccess)
}

func (h *Handler) sendJobStatusWebhook(ctx context.Context, event *RunJobPayload, err error, status string) error {
	span := trace.SpanFromContext(ctx)

	if err != nil {
		h.log.Errorf("job error: %s", err.Error())
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
	}

	// If no webhook URL configured nothing to do
	if h.args.JobsWebhookURL == "" {
		return nil
	}

	payload := &JobStatusWebhookPayload{
		ID:          event.ID,
		Name:        event.Name,
		ProjectName: h.args.ProjectName,
		Env:         h.args.Env,
		Status:      status,
		TraceID:     span.SpanContext().TraceID().String(),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = http.Post(h.args.JobsWebhookURL, "application/json", bytes.NewBuffer(b))
	return err
}
