package tasks

import (
	"context"
	"time"

	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
)

type Topic struct {
	Name    string   `json:"name"`
	Metrics *Metrics `json:"metrics,omitempty"`
	Stats   *Stats   `json:"stats,omitempty"`
}

type Metrics struct {
	TotalCount       int        `json:"totalCount"`
	CompletedToday   int        `json:"completedToday"`
	OldestUnresolved *time.Time `json:"oldestUnresolved,omitempty"`
}

type Stats struct {
	OpenCount            int            `json:"openCount"`
	AssignedCount        int            `json:"assignedCount"`
	DeferredCount        int            `json:"deferredCount"`
	CompletionRate       float32        `json:"completionRate"`
	CompletionTimeMedian *time.Duration `json:"completionTimeMedian,omitempty"`
	CompletionTime90     *time.Duration `json:"completionTime90P,omitempty"`
	CompletionTime99     *time.Duration `json:"completionTime99P,omitempty"`
}

// GetTopic returns the topic data (with metrics)
func GetTopic(ctx context.Context, pbTask *proto.Task, withStats bool) (*Topic, error) {
	topic := &Topic{
		Name: pbTask.GetName(),
	}

	metrics, err := getMetrics(ctx, topic)
	if err != nil {
		return nil, err
	}

	topic.Metrics = metrics

	if withStats {
		stats, err := getStats(ctx, topic)
		if err != nil {
			return nil, err
		}

		topic.Stats = stats
	}

	return topic, nil
}

// getMetrics will retrieve the metrics for the given topic
func getMetrics(ctx context.Context, topic *Topic) (*Metrics, error) {
	if topic == nil {
		return nil, nil
	}

	dbase, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var metrics Metrics

	if err := dbase.GetDB().Model(&Metrics{}).Table("keel.task").
		Select(`
			COUNT(*) AS total_count,
			COUNT(*) FILTER (
				WHERE status = ? 
				AND resolved_at::date = CURRENT_DATE
			) AS completed_today,
			MIN(created_at) FILTER (
				WHERE status <> ?
			) AS oldest_unresolved
		`, StatusCompleted, StatusCompleted).
		Where("name = ?", topic.Name).
		Scan(&metrics).Error; err != nil {
		return nil, err
	}

	return &metrics, nil
}

// getStats will retrieve the stats for the given topic
func getStats(ctx context.Context, topic *Topic) (*Stats, error) {
	if topic == nil {
		return nil, nil
	}

	dbase, err := db.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	var stats Stats

	if err := dbase.GetDB().Model(&Stats{}).Table("keel.task").
		Select(`
			COUNT(*) FILTER (WHERE status != @completed) AS open_count,
			COUNT(*) FILTER (WHERE status = @assigned) AS assigned_count,
			COUNT(*) FILTER (WHERE status = @deferred) AS deferred_count,
			COALESCE(
				COUNT(*) FILTER (WHERE status = @completed)::float / NULLIF(COUNT(*), 0), 
				0
			) AS completion_rate,
			(EXTRACT(EPOCH FROM PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY resolved_at - created_at) FILTER (WHERE status = @completed)))::bigint * 1e6 AS completion_time_median,
			(EXTRACT(EPOCH FROM PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY resolved_at - created_at) FILTER (WHERE status = @completed)))::bigint * 1e6 AS completion_time90,
			(EXTRACT(EPOCH FROM PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY resolved_at - created_at) FILTER (WHERE status = @completed)))::bigint * 1e6 AS completion_time99
		`,
			map[string]any{
				"completed": StatusCompleted,
				"assigned":  StatusAssigned,
				"deferred":  StatusDeferred,
			},
		).
		Where("name = ?", topic.Name).
		Scan(&stats).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}
