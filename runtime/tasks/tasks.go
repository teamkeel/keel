package tasks

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var tracer = otel.Tracer("github.com/teamkeel/keel/runtime/tasks")

type Status string

const (
	StatusNew       Status = "NEW"
	StatusAssigned  Status = "ASSIGNED"
	StatusDeferred  Status = "DEFERRED"
	StatusCompleted Status = "COMPLETED"
)

type Task struct {
	ID            string     `gorm:"column:id;primaryKey" json:"id"`
	Name          string     `gorm:"column:name" json:"name"`
	Status        Status     `gorm:"column:status" json:"status"`
	FlowRunID     *string    `gorm:"column:flow_run_id" json:"flowRunId,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at;->" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;->" json:"updatedAt"`
	AssignedTo    *string    `gorm:"column:assigned_to" json:"assignedTo,omitempty"`
	AssignedAt    *time.Time `gorm:"column:assigned_at" json:"assignedAt,omitempty"`
	ResolvedAt    *time.Time `gorm:"column:resolved_at" json:"resolvedAt,omitempty"`
	DeferredUntil *time.Time `gorm:"column:deferred_until" json:"deferredUntil,omitempty"`
}

func (Task) TableName() string {
	return "keel.task"
}

func (t *Task) isCompleted() bool {
	return t.Status == StatusCompleted
}

// TaskStatus represents a log of status updates for a particular task.
type TaskStatus struct {
	ID         string    `gorm:"column:id;primaryKey;->" json:"id"`
	TaskID     string    `gorm:"column:keel_task_id" json:"taskId"`
	Status     Status    `gorm:"column:status" json:"status"`
	AssignedTo *string   `gorm:"column:assigned_to" json:"assignedTo,omitempty"`
	SetBy      string    `gorm:"column:set_by" json:"setBy"`
	CreatedAt  time.Time `gorm:"column:created_at;->" json:"createdAt"`
}

func (TaskStatus) TableName() string {
	return "keel.task_status"
}

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

type filterFields struct {
	TopicName string
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

var ErrTaskNotFound = errors.New("task not found")

// getTask returns the task with the given ID and topic
func getTask(ctx context.Context, pbTask *proto.Task, id string) (*Task, error) {
	if pbTask == nil {
		return nil, nil
	}
	dbase, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var task Task
	err = dbase.GetDB().Where("name = ? AND id = ?", pbTask.GetName(), id).First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		} else {
			return nil, err
		}
	}

	return &task, nil
}

// getTasks will list the tasks according to the given filters using cursor pagination.
func getTasks(ctx context.Context, filters *filterFields, page *paginationFields) ([]*Task, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var tasks []*Task

	q := database.GetDB().Limit(page.GetLimit())

	if filters != nil {
		if filters.TopicName != "" {
			q = q.Where("name = ?", filters.TopicName)
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
			q.Where("created_at > (?)", database.GetDB().Model(&Task{}).Select("created_at").Where("id = ?", *page.Before))
		}
		if page.After != nil {
			q.Where("created_at < (?)", database.GetDB().Model(&Task{}).Select("created_at").Where("id = ?", *page.After))
		}
	} else {
		// default order
		q = q.Order("created_at DESC")
	}

	result := q.Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}

	if page.IsBackwards() {
		slices.Reverse(tasks)
	}

	return tasks, nil
}

// ListTasks for a given topic
func ListTasks(ctx context.Context, pbTask *proto.Task, inputs map[string]any) (tasks []*Task, err error) {
	ctx, span := tracer.Start(ctx, "ListTasks")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	pf := paginationFields{}
	pf.Parse(inputs)

	ff := filterFields{TopicName: pbTask.GetName()}
	ff.Parse(inputs)

	tasks, err = getTasks(ctx, &ff, &pf)

	return
}

// CompleteTask marks the given task as completed and returns it.
func CompleteTask(ctx context.Context, pbTask *proto.Task, id string, identityID string) (task *Task, err error) {
	ctx, span := tracer.Start(ctx, "CompleteTask")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	task, err = getTask(ctx, pbTask, id)
	if err != nil {
		return nil, err
	}

	// if task already completed, return it
	if task.isCompleted() {
		return task, nil
	}

	// mark task as completed
	dbase, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	err = dbase.GetDB().Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		errTx := tx.
			Model(&task).
			Clauses(clause.Returning{}).
			Where("name = ? AND id = ?", pbTask.GetName(), id).
			Updates(Task{Status: StatusCompleted, ResolvedAt: &now}).
			Error
		if errTx != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTaskNotFound
			} else {
				return errTx
			}
		}

		// we now save the status in the log
		return tx.Save(TaskStatus{
			TaskID: task.ID,
			Status: StatusCompleted,
			SetBy:  identityID,
		}).Error
	})
	if err != nil {
		return nil, err
	}

	return
}

// DeferTask marks the given task as deferred until the given date and returns it.
func DeferTask(ctx context.Context, pbTask *proto.Task, id string, deferUntil time.Time, identityID string) (task *Task, err error) {
	ctx, span := tracer.Start(ctx, "DeferTask")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	task, err = getTask(ctx, pbTask, id)
	if err != nil {
		return nil, err
	}

	// if task already completed, return error
	if task.isCompleted() {
		return nil, errors.New("task already completed")
	}

	// mark task as deferred
	dbase, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	err = dbase.GetDB().Transaction(func(tx *gorm.DB) error {
		errTx := tx.
			Model(&task).
			Clauses(clause.Returning{}).
			Where("name = ? AND id = ?", pbTask.GetName(), id).
			Updates(map[string]any{
				"status":         StatusDeferred,
				"deferred_until": &deferUntil,
				"assigned_to":    nil,
				"assigned_at":    nil,
			}).
			Error
		if errTx != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTaskNotFound
			} else {
				return errTx
			}
		}

		// we now save the status in the log
		return tx.Save(TaskStatus{
			TaskID: task.ID,
			Status: StatusDeferred,
			SetBy:  identityID,
		}).Error
	})
	if err != nil {
		return nil, err
	}

	return
}

// AssignTask marks the given task as assigned to the given identity and returns it.
func AssignTask(ctx context.Context, pbTask *proto.Task, id string, assignedTo, identityID string) (task *Task, err error) {
	ctx, span := tracer.Start(ctx, "AssignTask")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	task, err = getTask(ctx, pbTask, id)
	if err != nil {
		return nil, err
	}

	// if task already completed, return error
	if task.isCompleted() {
		return nil, errors.New("task already completed")
	}

	// mark task as deferred
	dbase, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	err = dbase.GetDB().Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		errTx := tx.
			Model(&task).
			Clauses(clause.Returning{}).
			Where("name = ? AND id = ?", pbTask.GetName(), id).
			Updates(Task{Status: StatusAssigned, AssignedTo: &assignedTo, AssignedAt: &now}).
			Error
		if errTx != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTaskNotFound
			} else {
				return errTx
			}
		}

		// we now save the status in the log
		return tx.Save(TaskStatus{
			TaskID:     task.ID,
			Status:     StatusAssigned,
			AssignedTo: &assignedTo,
			SetBy:      identityID,
		}).Error
	})
	if err != nil {
		return nil, err
	}

	return
}
