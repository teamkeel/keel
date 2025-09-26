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

	StepStatusCancelled StepStatus = "CANCELLED"
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
	if s := r.PendingUIStep(); s != nil {
		return true
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

// PendingUIStep returns the pending UI step for this flow run.
// If the flow does not have a pending UI step then nil will be returned.
func (r *Run) PendingUIStep() *Step {
	if r == nil || (r.Status != StatusAwaitingInput && r.Status != StatusRunning) {
		return nil
	}

	for _, step := range r.Steps {
		if step.Type == StepTypeUI && step.Status == StepStatusPending {
			return &step
		}
	}

	return nil
}

// LastCompletedUIStep returns the last completed step for this flow run.
func (r *Run) LastCompletedUIStep() *Step {
	if r == nil {
		return nil
	}

	var lastStep *Step

	for _, step := range r.Steps {
		if step.Status == StepStatusCompleted && step.Type == StepTypeUI {
			if lastStep == nil {
				lastStep = &step
			} else {
				if step.CreatedAt.After(lastStep.CreatedAt) {
					lastStep = &step
				}
			}
		}
	}

	return lastStep
}

type FlowStats struct {
	Name           string             `json:"name"`
	LastRun        *time.Time         `json:"lastRun"`
	TotalRuns      int                `json:"totalRuns"`
	ErrorRate      float32            `json:"errorRate"`
	ActiveRuns     int                `json:"activeRuns"`
	CompletedToday int                `json:"completedToday"`
	TimeSeries     []*FlowStatsBucket `json:"timeSeries,omitempty" gorm:"-"`
}

// PopulateTimeSeries will take the given FlowStatsBuckets and set the applicable one's onto the FlowStats.
func (s *FlowStats) PopulateTimeSeries(buckets []*FlowStatsBucket) {
	for _, b := range buckets {
		if b.Name == s.Name {
			s.TimeSeries = append(s.TimeSeries, b)
		}
	}
}

type FlowStatsBucket struct {
	Name       string    `json:"-"` // omitted
	Time       time.Time `json:"time"`
	TotalRuns  int       `json:"totalRuns"`
	FailedRuns int       `json:"failedRuns"`
}

const (
	StatsIntervalDaily  = "daily"
	StatsIntervalHourly = "hourly"
)

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

type statsFilters struct {
	FlowNames []string
	Before    *time.Time
	After     *time.Time
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
// If the flow's status is set to cancelled, any pending steps of that flow's run will be set to cancelled as well.
func updateRun(ctx context.Context, runID string, status Status, config any) (*Run, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var run Run

	err = database.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&run).
			Clauses(clause.Returning{}).
			Where("id = ?", runID).
			Updates(Run{
				Status: status,
				Config: config,
			}).Error; err != nil {
			return err
		}

		// if we've cancelled the flow, we need to cancel any UI steps that are PENDING
		if status == StatusCancelled {
			if err := tx.Model(&Step{}).
				Where("run_id = ? AND type = ? AND status = ?", runID, StepTypeUI, StepStatusPending).
				Updates(Step{
					Status: StepStatusCancelled,
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return &run, err
}

// resetSteps will delete the given stepsand reset the last step to pending.
func resetSteps(ctx context.Context, runID string, deleteSteps []string, lastStepID string) error {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return err
	}

	return database.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Step{}).
			Where("id = ?", lastStepID).
			Updates(map[string]any{
				"value":    nil,
				"end_time": nil,
				"status":   StepStatusPending,
			}).Error; err != nil {
			return err
		}

		if err := tx.Where("run_id = ? AND id IN ?", runID, deleteSteps).Delete(&Step{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// completeRun will complete a flow run.
func completeRun(ctx context.Context, runID string, config any, data any) (*Run, error) {
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
			q = q.Order("created_at ASC")
		} else {
			q = q.Order("created_at DESC")
		}

		if page.Before != nil {
			q.Where("created_at > (?)", database.GetDB().Model(&Run{}).Select("created_at").Where("id = ?", *page.Before))
		}
		if page.After != nil {
			q.Where("created_at < (?)", database.GetDB().Model(&Run{}).Select("created_at").Where("id = ?", *page.After))
		}
	} else {
		// default order
		q = q.Order("created_at DESC")
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

// listFlowStats will list generic flow run stats for the given filters.
func listFlowStats(ctx context.Context, filters statsFilters) ([]*FlowStats, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var stats []*FlowStats

	query := database.GetDB().Table(Run{}.TableName()).Select(`
		name,
		COUNT(*) AS total_runs,
		COUNT(*) FILTER (WHERE status = 'FAILED')::float / NULLIF(COUNT(*), 0) AS error_rate,
		COUNT(*) FILTER (WHERE status IN ('RUNNING', 'AWAITING_INPUT')) AS active_runs,
		COUNT(*) FILTER (WHERE status = 'COMPLETED' AND created_at::date = CURRENT_DATE) AS completed_today,
		MAX(created_at) AS last_run
	`).Where("name IN ?", filters.FlowNames)

	if filters.Before != nil && !filters.Before.IsZero() {
		query = query.Where("created_at <= ?", *filters.Before)
	}
	if filters.After != nil && !filters.After.IsZero() {
		query = query.Where("created_at >= ?", *filters.After)
	}

	err = query.
		Group("name").
		Order("name ASC").
		Scan(&stats).Error

	return stats, err
}

// listFlowStatsSeries will list flow run stats grouped by day for the given filters.
func listFlowStatsSeries(ctx context.Context, filters statsFilters, interval string) ([]*FlowStatsBucket, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	// default to daily
	intervalStr := "created_at::date"
	if interval == StatsIntervalHourly {
		intervalStr = `date_trunc('hour', created_at)`
	}

	var buckets []*FlowStatsBucket

	query := database.GetDB().Table(Run{}.TableName()).Select(`
		name,
		`+intervalStr+` AS time,
		COUNT(*) AS total_runs,
		COUNT(*) FILTER (WHERE status = 'FAILED') AS failed_runs
	`).Where("name IN ?", filters.FlowNames)

	if filters.Before != nil && !filters.Before.IsZero() {
		query = query.Where("created_at <= ?", *filters.Before)
	}
	if filters.After != nil && !filters.After.IsZero() {
		query = query.Where("created_at >= ?", *filters.After)
	}

	err = query.
		Group("name, time").
		Order("name ASC").
		Scan(&buckets).Error

	return buckets, err
}
