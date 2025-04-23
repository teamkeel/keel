package flows

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/flows")

type Status string

const (
	StatusNew       Status = "NEW"
	StatusRunning   Status = "RUNNING"
	StatusFailed    Status = "FAILED"
	StatusCompleted Status = "COMPLETED"
	StatusCancelled Status = "CANCELLED"
)

type StepType string
type StepStatus string

const (
	StepTypeFunction StepType = "FUNCTION"
	StepTypeUI       StepType = "UI"

	StepStatusPending   StepStatus = "PENDING"
	StepStatusFailed    StepStatus = "FAILED"
	StepStatusCompleted StepStatus = "COMPLETED"
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

// PendingUI tells us if the current run is waiting for UI input
func (r *Run) PendingUI() bool {
	if r == nil || r.Status != StatusRunning {
		return false
	}

	for _, step := range r.Steps {
		if step.Type == StepTypeUI && step.Status == StepStatusPending {
			return true
		}
	}

	return false
}

// SetUIComponent will set the given UI component on the first pending UI step of the flow
func (r *Run) SetUIComponent(ui *JSONB) {
	if r.Status != StatusRunning {
		return
	}

	for i, step := range r.Steps {
		if step.Type == StepTypeUI && step.Status == StepStatusPending {
			r.Steps[i].UI = ui
		}
	}
}

// HasPendingUIStep checks that this run has a pending UI step with the given id
func (r *Run) HasPendingUIStep(stepID string) bool {
	for _, step := range r.Steps {
		if step.ID == stepID {
			return step.Type == StepTypeUI && step.Status == StepStatusPending
		}
	}

	return false
}

type Step struct {
	ID          string     `json:"id" gorm:"primaryKey;not null;default:null"`
	Name        string     `json:"name"`
	RunID       string     `json:"runId"`
	Status      StepStatus `json:"status"`
	Type        StepType   `json:"type"`
	Value       *JSONB     `json:"value" gorm:"type:jsonb"`
	MaxRetries  int        `json:"max_retries"`
	TimeoutInMs int        `json:"timeout_in_ms"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	UI          *JSONB     `json:"ui" gorm:"-"` // UI component, omitted from db operations
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

type paginationFields struct {
	Limit  int
	After  *string
	Before *string
}

// Parse will set the values for the pagination fields from the given map
func (p *paginationFields) Parse(inputs map[string]any) {
	for f, v := range inputs {
		switch f {
		case "limit":
			switch val := v.(type) {
			case int64:
				p.Limit = int(val)
			case int:
				p.Limit = val
			case float64:
				p.Limit = int(val)
			case string:
				if num, err := strconv.Atoi(val); err == nil {
					p.Limit = num
				}
			}
		case "after":
			if val, ok := v.(string); ok {
				p.After = &val
			}
		case "before":
			if val, ok := v.(string); ok {
				p.Before = &val
			}
		}
	}
}

// GetLimit returns a limit of items to be returned. If no limit set in the pagination fields, a default of 10 will be used
func (p *paginationFields) GetLimit() int {
	// default to 10
	if p == nil || p.Limit < 1 {
		return 10
	}

	return p.Limit
}

func (p *paginationFields) IsBackwards() bool {
	if p.After != nil {
		return false
	}

	return true
}

// getRun returns the flow run with the given ID. If no flow run found, nil nil is returned.
func getRun(ctx context.Context, runID string) (*Run, error) {
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

// updateRun will update the status of a flow run
func updateRun(ctx context.Context, runID string, status Status) (*Run, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var run Run
	result := database.GetDB().Model(&run).Clauses(clause.Returning{}).Where("id = ?", runID).Update("status", status)

	return &run, result.Error
}

// createRun will create a new flow run with the given input
func createRun(ctx context.Context, flow *proto.Flow, inputs any) (*Run, error) {
	if flow == nil {
		return nil, fmt.Errorf("invalid flow")
	}

	var jsonInputs JSONB
	if inputsMap, ok := inputs.(map[string]any); ok {
		jsonInputs = inputsMap
	}

	run := Run{
		Status: StatusNew,
		Input:  &jsonInputs,
		Name:   flow.Name,
	}

	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	result := database.GetDB().Create(&run)
	if result.Error != nil {
		return nil, result.Error
	}

	return &run, nil
}

// listRuns will list the flow runs for the given flow using cursor pagination. It defaults to
func listRuns(ctx context.Context, flow *proto.Flow, page *paginationFields) ([]*Run, error) {
	if flow == nil {
		return nil, fmt.Errorf("invalid flow")
	}
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var runs []*Run

	q := database.GetDB().Where("name = ?", flow.Name).Limit(page.GetLimit())

	if page != nil {
		if page.IsBackwards() {
			q = q.Order("id DESC")
		} else {
			q = q.Order("id ASC")
		}

		if page.Before != nil {
			q.Where("id < ?", *page.Before)
		}
		if page.After != nil {
			q.Where("id > ?", *page.After)
		}
	}

	result := q.Find(&runs)
	if result.Error != nil {
		return nil, result.Error
	}

	if !page.IsBackwards() {
		// we always want to return items backwards (i.e. the most recent at the top)
		slices.Reverse(runs)
	}

	return runs, nil
}
