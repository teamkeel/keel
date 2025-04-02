package flows

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/teamkeel/keel/db"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/flows")

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

// GetFlowRun returns the flow run with the given ID. If no flow run found, nil nil is returned.
func GetFlowRun(ctx context.Context, runID string) (*Run, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var run Run
	result := database.GetDB().Preload("Steps").Where("id = ?", runID).First(&run)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}

	return &run, nil
}

// UpdateRun will update the status of a flow run
func UpdateRun(ctx context.Context, runID string, status Status) (*Run, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var run Run
	result := database.GetDB().Model(&run).Clauses(clause.Returning{}).Where("id = ?", runID).Update("status", status)

	return &run, result.Error
}
