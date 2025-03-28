package flows

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/teamkeel/keel/db"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type Status string

const (
	StatusNew       Status = "new"
	StatusRunning   Status = "running"
	StatusWaiting   Status = "waiting"
	StatusFailed    Status = "failed"
	StatusCompleted Status = "completed"
)

type StepType string

const (
	StepTypeFunction StepType = "function"
	StepTypeIO       StepType = "io"
	StepTypeWait     StepType = "wait"
)

type Run struct {
	ID        string    `json:"id" gorm:"primaryKey;not null;default:null"`
	Status    Status    `json:"status"`
	Name      string    `json:"name"`
	Input     *JSONB    `json:"input" gorm:"type:jsonb"`
	Steps     []Step    `json:"steps"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Run) TableName() string {
	return "keel_flow_run"
}

type Step struct {
	ID        string    `json:"id" gorm:"primaryKey;not null;default:null"`
	Name      string    `json:"name"`
	RunID     string    `json:"runId"`
	Status    Status    `json:"status"`
	Type      StepType  `json:"type"`
	Value     *JSONB    `json:"value" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Step) TableName() string {
	return "keel_flow_step"
}

// JSONB Interface for JSONB fields
type JSONB map[string]any

func (jsonField JSONB) Value() (driver.Value, error) {
	b, err := json.Marshal(jsonField)
	return string(b), err
}

func (jsonField *JSONB) Scan(value any) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(data, &jsonField)
}

// GetFlowRun returns the flow run with the given runID and within the given's scope's flow.
// If no flow run found, nil nil is returned
func GetFlowRun(ctx context.Context, scope *Scope, runID string) (*Run, error) {
	if scope.Flow == nil {
		return nil, fmt.Errorf("invalid flow")
	}

	ctx, span := tracer.Start(ctx, "GetFlowRun")
	defer span.End()

	span.SetAttributes(
		attribute.String("flow", scope.Flow.Name),
		attribute.String("flowRunID", runID),
	)

	database, err := db.GetDatabase(ctx)
	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var run Run

	result := database.GetDB().Preload("Steps").Where("id = ? and name = ?", runID, scope.Flow.Name).First(&run)
	if result.Error != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, result.Error.Error())

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}

	return &run, nil
}
