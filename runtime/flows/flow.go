package flows

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/util"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/flows")

type Status string

const (
	StatusNew           Status = "NEW"
	StatusRunning       Status = "RUNNING"
	StatusAwaitingInput Status = "AWAITING_INPUT"
	StatusFailed        Status = "FAILED"
	StatusCompleted     Status = "COMPLETED"
	StatusCancelled     Status = "CANCELLED"
)

type StepType string
type StepStatus string

const (
	StepTypeFunction StepType = "FUNCTION"
	StepTypeUI       StepType = "UI"
	StepTypeComplete StepType = "COMPLETE"

	StepStatusPending   StepStatus = "PENDING"
	StepStatusFailed    StepStatus = "FAILED"
	StepStatusCompleted StepStatus = "COMPLETED"
)

type Run struct {
	ID          string    `json:"id"        gorm:"primaryKey;not null;default:null"`
	TraceID     string    `json:"traceId"`
	Traceparent string    `json:"-"`
	Status      Status    `json:"status"`
	Name        string    `json:"name"`
	Input       JSON      `json:"input"     gorm:"type:jsonb;serializer:json"`
	Data        JSON      `json:"data"      gorm:"type:jsonb;serializer:json"`
	Steps       []Step    `json:"steps"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Config      JSON      `json:"config"    gorm:"type:jsonb;serializer:json"`
	StartedBy   *string   `json:"startedBy"`
}

func (Run) TableName() string {
	return "keel.flow_run"
}

// HasPendingUIStep tells us if the current run is waiting for UI input.
func (r *Run) HasPendingUIStep() bool {
	if r == nil || (r.Status != StatusAwaitingInput && r.Status != StatusRunning) {
		return false
	}

	for _, step := range r.Steps {
		if step.Type == StepTypeUI && step.Status == StepStatusPending {
			return true
		}
	}

	return false
}

// SetUIComponents will set the given UI component on the first pending UI step of the flow.
func (r *Run) SetUIComponents(c *FlowUIComponents) {
	if c == nil {
		return
	}

	if r.HasPendingUIStep() && c.UI != nil {
		for i, step := range r.Steps {
			if step.Type == StepTypeUI && step.Status == StepStatusPending {
				r.Steps[i].UI = c.UI
			}
		}
	}
}

type Step struct {
	ID        string     `json:"id"        gorm:"primaryKey;not null;default:null"`
	Name      string     `json:"name"`
	RunID     string     `json:"runId"`
	Status    StepStatus `json:"status"`
	Type      StepType   `json:"type"`
	Value     JSON       `json:"value"     gorm:"type:jsonb;serializer:json"`
	Stage     *string    `json:"stage"`
	Error     *string    `json:"error"`
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	UI        JSON       `json:"ui"        gorm:"type:jsonb;serializer:json"`
}

func (Step) TableName() string {
	return "keel.flow_step"
}

type JSON interface{}

type paginationFields struct {
	Limit  int
	After  *string
	Before *string
}

// Parse will set the values for the pagination fields from the given map.
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

type filterFields struct {
	FlowName  *string
	StartedBy *string
	Statuses  []Status
}

// Parse will set the values for the filter fields from the given map; the only applicable field is `Status`.
func (ff *filterFields) Parse(inputs map[string]any) {
	for f, v := range inputs {
		switch f {
		case "status":
			switch val := v.(type) {
			case string:
				sts := strings.Split(val, ",")
				for _, s := range sts {
					ff.Statuses = append(ff.Statuses, Status(s))
				}
			case []string:
				for _, s := range val {
					ff.Statuses = append(ff.Statuses, Status(s))
				}
			}
		}
	}
}

// GetLimit returns a limit of items to be returned. If no limit set in the pagination fields, a default of 10 will be used.
func (p *paginationFields) GetLimit() int {
	// default to 10
	if p == nil || p.Limit < 1 {
		return 10
	}

	return p.Limit
}

func (p *paginationFields) IsBackwards() bool {
	return p.Before != nil
}

// getRun returns the flow run with the given ID. If no flow run found, nil/nil is returned.
func getRun(ctx context.Context, runID string) (*Run, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var run Run
	result := database.GetDB().Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("id = ?", runID).First(&run)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}

	return &run, nil
}

// GetTraceparent returns the traceparent from the db for the given run ID.
func GetTraceparent(ctx context.Context, runID string) (string, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return "", err
	}

	var run Run
	result := database.GetDB().Where("id = ?", runID).First(&run)
	if result.Error != nil {
		return "", result.Error
	}

	return run.Traceparent, nil
}

// updateRun will update the status of a flow run.
func updateRun(ctx context.Context, runID string, status Status, updatedConfig any) (*Run, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	u := Run{
		Status: status,
	}
	if updatedConfig != nil {
		u.Config = updatedConfig
	}

	var run Run
	result := database.GetDB().
		Model(&run).
		Clauses(clause.Returning{}).
		Where("id = ?", runID).
		//Update("status", status)
		Updates(u)

	return &run, result.Error
}

// completeRun will complete a flow run.
func completeRun(ctx context.Context, runID string, data any, config any) (*Run, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var run Run
	result := database.GetDB().
		Model(&run).
		Clauses(clause.Returning{}).
		Where("id = ?", runID).
		Updates(Run{
			Status: StatusCompleted,
			Data:   data,
			Config: config,
		})

	return &run, result.Error
}

// createRun will create a new flow run with the given input.
func createRun(ctx context.Context, flow *proto.Flow, inputs any, traceparent string, identityID *string) (*Run, error) {
	if flow == nil {
		return nil, fmt.Errorf("invalid flow")
	}

	run := Run{
		Status:      StatusNew,
		Input:       inputs,
		Name:        flow.GetName(),
		Traceparent: traceparent,
		TraceID:     util.ParseTraceparent(traceparent).TraceID().String(),
		StartedBy:   identityID,
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

// listRuns will list the flow runs for the given flow using cursor pagination. It defaults to.
func listRuns(ctx context.Context, filters *filterFields, page *paginationFields) ([]*Run, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var runs []*Run

	q := database.GetDB().Limit(page.GetLimit())

	if filters != nil {
		if filters.FlowName != nil {
			q = q.Where("name = ?", filters.FlowName)
		}
		if filters.StartedBy != nil {
			q = q.Where("started_by = ?", filters.StartedBy)
		}
		if len(filters.Statuses) > 0 {
			q = q.Where("status IN ?", filters.Statuses)
		}
	}

	if page != nil {
		if page.IsBackwards() {
			q = q.Order("id ASC")
		} else {
			q = q.Order("id DESC")
		}

		if page.Before != nil {
			q.Where("id > ?", *page.Before)
		}
		if page.After != nil {
			q.Where("id < ?", *page.After)
		}
	} else {
		// default order
		q = q.Order("id DESC")
	}

	result := q.Find(&runs)
	if result.Error != nil {
		return nil, result.Error
	}

	if page.IsBackwards() {
		slices.Reverse(runs)
	}

	return runs, nil
}
