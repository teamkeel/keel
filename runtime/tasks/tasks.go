package tasks

import (
	"context"
	"errors"
	"time"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"go.opentelemetry.io/otel"
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
	CreatedAt     time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at" json:"updatedAt"`
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

var ErrTaskNotFound = errors.New("task not found")

// GetTask returns the task with the given ID and topic
func GetTask(ctx context.Context, pbTask *proto.Task, id string) (*Task, error) {
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

// CompleteTask marks the given task as completed and returns it. If task is not found, nil, nil is returned
func CompleteTask(ctx context.Context, pbTask *proto.Task, id string) (*Task, error) {
	t, err := GetTask(ctx, pbTask, id)
	if err != nil {
		return nil, err
	}

	// if task already completed, return it
	if t.isCompleted() {
		return t, nil
	}

	// mark task as completed
	dbase, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	err = dbase.GetDB().
		Model(&t).
		Clauses(clause.Returning{}).
		Where("name = ? AND id = ?", pbTask.GetName(), id).
		Updates(Task{Status: StatusCompleted, ResolvedAt: &now}).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		} else {
			return nil, err
		}
	}

	return t, nil
}
