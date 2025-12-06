package tasks

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/flows"
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
	StatusCancelled Status = "CANCELLED"
	StatusStarted   Status = "STARTED"
)

// EntityFieldNameTaskID is the field name used in the entity table to link to a keel task.
const EntityFieldNameTaskID string = "keel_task_id"

type Task struct {
	ID            string     `gorm:"column:id;primaryKey;->" json:"id"`
	Name          string     `gorm:"column:name"             json:"name"`
	Status        Status     `gorm:"column:status"           json:"status"`
	FlowRunID     *string    `gorm:"column:flow_run_id"      json:"flowRunId,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at;->"    json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;->"    json:"updatedAt"`
	AssignedTo    *string    `gorm:"column:assigned_to"      json:"assignedTo,omitempty"`
	AssignedAt    *time.Time `gorm:"column:assigned_at"      json:"assignedAt,omitempty"`
	ResolvedAt    *time.Time `gorm:"column:resolved_at"      json:"resolvedAt,omitempty"`
	DeferredUntil *time.Time `gorm:"column:deferred_until"   json:"deferredUntil,omitempty"`
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
	TaskID     string    `gorm:"column:keel_task_id"     json:"taskId"`
	Status     Status    `gorm:"column:status"           json:"status"`
	FlowRunID  *string   `gorm:"column:flow_run_id"      json:"flowRunId,omitempty"`
	AssignedTo *string   `gorm:"column:assigned_to"      json:"assignedTo,omitempty"`
	SetBy      string    `gorm:"column:set_by"           json:"setBy"`
	CreatedAt  time.Time `gorm:"column:created_at;->"    json:"createdAt"`
}

func (TaskStatus) TableName() string {
	return "keel.task_status"
}

type paginationFields struct {
	Limit int
	After *string
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

type filterFields struct {
	Statuses []Status
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

// getTask returns the task with the given ID and topic.
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

// getTaskEntityID returns the ID of the entity holding the data relating to this task.
func getTaskEntityID(ctx context.Context, pbTask *proto.Task, id string) (*string, error) {
	dbase, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var entityID *string

	err = dbase.GetDB().
		Table(strcase.ToSnake(pbTask.GetName())).
		Select("id").
		Where(fmt.Sprintf("%s = ?", EntityFieldNameTaskID), id).
		Scan(&entityID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return entityID, nil
}

// getTaskQueue will retrieve the queue of tasks for the given topic.
func getTaskQueue(ctx context.Context, pbTask *proto.Task, filters *filterFields, page *paginationFields) ([]*Task, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var tasks []*Task

	q := database.GetDB().Limit(page.GetLimit())

	q.Select("keel.task.*")
	q.Joins(fmt.Sprintf("INNER JOIN %s ON keel.task.id = %s.%s", strcase.ToSnake(pbTask.GetName()), strcase.ToSnake(pbTask.GetName()), EntityFieldNameTaskID))
	q = q.Where("name = ?", pbTask.GetName())

	if filters != nil {
		if len(filters.Statuses) > 0 {
			q = q.Where("status IN ?", filters.Statuses)
			q = q.Where("(deferred_until IS NULL OR deferred_until <= ?)", time.Now())
		}
	}

	if page.After != nil {
		q.Where("created_at < (?)", database.GetDB().Model(&Task{}).Select("created_at").Where("id = ?", *page.After))
	}

	for _, orderBy := range pbTask.GetOrderBy() {
		var direction string
		switch orderBy.GetDirection() {
		case proto.OrderDirection_ORDER_DIRECTION_ASCENDING:
			direction = "ASC"
		case proto.OrderDirection_ORDER_DIRECTION_DECENDING:
			direction = "DESC"
		}
		q = q.Order(fmt.Sprintf("%s.%s %s", strcase.ToSnake(pbTask.GetName()), strcase.ToSnake(orderBy.GetFieldName()), direction))
	}

	q = q.Order("created_at DESC")

	result := q.Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}

	return tasks, nil
}

// GetTaskQueue returns the ordered queue of tasks for a given topic.
func GetTaskQueue(ctx context.Context, pbTask *proto.Task, inputs map[string]any) (tasks []*Task, err error) {
	ctx, span := tracer.Start(ctx, "GetTaskQueue")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	pf := paginationFields{}
	pf.Parse(inputs)

	ff := filterFields{}
	ff.Parse(inputs)

	tasks, err = getTaskQueue(ctx, pbTask, &ff, &pf)

	return
}

// NewTask creates a new task and returns it.
func NewTask(ctx context.Context, pbTask *proto.Task, identityID string, deferUntil *time.Time, data map[string]any) (task *Task, err error) {
	ctx, span := tracer.Start(ctx, "NewTask")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	dbase, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	task = &Task{
		Name:   pbTask.GetName(),
		Status: StatusNew,
	}

	if deferUntil != nil {
		task.DeferredUntil = deferUntil
		task.Status = StatusDeferred
	}

	err = dbase.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Returning{}).
			Create(&task).Error; err != nil {
			return err
		}

		err = tx.Save(TaskStatus{
			TaskID: task.ID,
			Status: task.Status,
			SetBy:  identityID,
		}).Error
		if err != nil {
			return err
		}

		d := map[string]any{EntityFieldNameTaskID: task.ID}
		for key, value := range data {
			d[strcase.ToSnake(key)] = value
		}

		err = tx.Table(strcase.ToSnake(task.Name)).Create(d).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return
}

// StartTask creates and runs a flow for the given task.
func StartTask(ctx context.Context, pbTask *proto.Task, id string, identityID string) (task *Task, err error) {
	ctx, span := tracer.Start(ctx, "StartTask")
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

	if task.isCompleted() {
		return nil, errors.New("task already completed")
	}

	// if there is no flow associated, then start a new flow
	if task.FlowRunID == nil {
		return startFlow(ctx, pbTask, task, identityID)
	}

	if task.FlowRunID != nil {
		// get the current flow status
		flowRun, err := flows.GetFlowRunState(ctx, *task.FlowRunID)
		if err != nil {
			return nil, err
		}

		switch flowRun.Status {
		case flows.StatusNew, flows.StatusRunning, flows.StatusAwaitingInput, flows.StatusCompleted:
			return task, nil
		case flows.StatusFailed, flows.StatusCancelled:
			return startFlow(ctx, pbTask, task, identityID)
		}

	}

	return task, nil
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

// CancelTask marks the given task as cancelled and returns it.
func CancelTask(ctx context.Context, pbTask *proto.Task, id string, identityID string) (task *Task, err error) {
	ctx, span := tracer.Start(ctx, "CancelTask")
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

	// mark task as cancelled
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
			Updates(Task{Status: StatusCancelled, ResolvedAt: &now}).
			Error
		if errTx != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTaskNotFound
			} else {
				return errTx
			}
		}

		return tx.Save(TaskStatus{
			TaskID: task.ID,
			Status: StatusCancelled,
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

// NextTask will assign and return the next available task to the authenticated identity.
// It does not create or start a flow.
func NextTask(ctx context.Context, pbTask *proto.Task, identityID string) (task *Task, err error) {
	ctx, span := tracer.Start(ctx, "NextTask")
	defer span.End()

	defer func() {
		if err != nil {
			span.RecordError(err, trace.WithStackTrace(true))
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	dbase, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var result *Task

	err = dbase.GetDB().Transaction(func(tx *gorm.DB) error {
		// 1) Check for an existing assigned task for this identity across all topics
		var existing Task
		errTx := tx.
			Where("assigned_to = ? AND status = ?", identityID, StatusAssigned).
			Order("assigned_at DESC NULLS LAST").
			Limit(1).
			Take(&existing).Error
		if errTx == nil {
			result = &existing
			return nil
		}
		if !errors.Is(errTx, gorm.ErrRecordNotFound) {
			return errTx
		}

		// 2) Find a candidate task to assign using row-level locking with SKIP LOCKED
		// Include both NEW and DEFERRED tasks (DEFERRED tasks with future defer_until are automatically excluded)
		tasks, errTx := getTaskQueue(ctx, pbTask, &filterFields{Statuses: []Status{StatusNew, StatusDeferred}}, &paginationFields{Limit: 1})
		if errTx != nil {
			return errTx
		}

		if len(tasks) == 0 {
			return ErrTaskNotFound
		}

		candidate := tasks[0]

		// 3) Update the candidate to ASSIGNED and set assigned fields, returning the updated row
		now := time.Now()
		errTx = tx.
			Model(candidate).
			Clauses(clause.Returning{}).
			Where("id = ?", candidate.ID).
			Updates(Task{Status: StatusAssigned, AssignedTo: &identityID, AssignedAt: &now}).
			Error
		if errTx != nil {
			return errTx
		}

		// 4) Insert status log entry
		if errTx = tx.Save(TaskStatus{
			TaskID:     candidate.ID,
			Status:     StatusAssigned,
			AssignedTo: &identityID,
			SetBy:      identityID,
		}).Error; errTx != nil {
			return errTx
		}

		result = candidate
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// startFlow starts a flow for the given task
func startFlow(ctx context.Context, pbTask *proto.Task, task *Task, identityID string) (*Task, error) {
	entityID, err := getTaskEntityID(ctx, pbTask, task.ID)
	if err != nil {
		return nil, err
	}

	flowInputs := map[string]any{}
	if entityID != nil {
		flowInputs["entityId"] = *entityID
	}

	newFlowRun, err := flows.NewFlowRun(ctx, pbTask.GetFlow(), flowInputs, &identityID)
	if err != nil {
		return nil, err
	}

	dbase, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	err = dbase.GetDB().Transaction(func(tx *gorm.DB) error {
		errTx := tx.
			Model(&task).
			Clauses(clause.Returning{}).
			Where("name = ? AND id = ?", pbTask.GetName(), task.ID).
			Updates(Task{FlowRunID: &newFlowRun.ID, Status: StatusStarted}).
			Error
		if errTx != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTaskNotFound
			} else {
				return errTx
			}
		}

		return tx.Save(TaskStatus{
			TaskID:     task.ID,
			Status:     StatusStarted,
			AssignedTo: &identityID,
			FlowRunID:  &newFlowRun.ID,
			SetBy:      identityID,
		}).Error
	})
	if err != nil {
		return nil, err
	}

	// This must happen after setting the status due to possible race conditions (i.e. if the flow completes instantly)
	_, err = flows.StartFlow(ctx, pbTask.GetFlow(), newFlowRun.ID, flowInputs)
	if err != nil {
		return nil, err
	}

	return task, nil
}
