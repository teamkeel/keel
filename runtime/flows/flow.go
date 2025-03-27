package flows

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
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

type FlowRun struct {
	ID        string    `json:"id" gorm:"primaryKey;not null;default:null"`
	Status    Status    `json:"status"`
	Name      string    `json:"name"`
	Input     *JSONB    `json:"input" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (FlowRun) TableName() string {
	return "keel_flow"
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
