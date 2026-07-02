package scylla

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"GoProject/udm/internal/model"
)

// AlertEventRepository 定義�?ScyllaDB alert_events 表�?資�?存�?介面
type AlertEventRepository interface {
	Insert(ctx context.Context, event *model.AlertEvent) error
	QueryByDevice(ctx context.Context, deviceID uuid.UUID, month string, severity string) ([]*model.AlertEvent, error)
	Acknowledge(ctx context.Context, deviceID uuid.UUID, month string, triggeredAt time.Time, ruleID uuid.UUID) error
}

type scyllaAlertEventRepository struct {
	client *Client
}

// NewAlertEventRepository 建�? ScyllaDB ??AlertEventRepository 實�?
func NewAlertEventRepository(client *Client) AlertEventRepository {
	return &scyllaAlertEventRepository{client: client}
}

func (r *scyllaAlertEventRepository) Insert(ctx context.Context, event *model.AlertEvent) error {
	query := `INSERT INTO alert_events (device_id, month, triggered_at, rule_id, metric_name, metric_value, threshold, severity, acknowledged) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	return r.client.Session.Query(query, event.DeviceID, event.Month, event.TriggeredAt, event.RuleID, event.MetricName, event.MetricValue, event.Threshold, event.Severity, event.Acknowledged).WithContext(ctx).Exec()
}

func (r *scyllaAlertEventRepository) QueryByDevice(ctx context.Context, deviceID uuid.UUID, month string, severity string) ([]*model.AlertEvent, error) {
	var result []*model.AlertEvent

	query := `SELECT device_id, month, triggered_at, rule_id, metric_name, metric_value, threshold, severity, acknowledged FROM alert_events WHERE device_id = ? AND month = ?`
	iter := r.client.Session.Query(query, deviceID, month).WithContext(ctx).Iter()

	var devID gocql.UUID
	var mth string
	var triggeredAt time.Time
	var ruleID gocql.UUID
	var mName string
	var mValue, threshold float64
	var sev string
	var ack bool

	for iter.Scan(&devID, &mth, &triggeredAt, &ruleID, &mName, &mValue, &threshold, &sev, &ack) {
		if severity != "" && sev != severity {
			continue
		}

		result = append(result, &model.AlertEvent{
			DeviceID:     uuid.UUID(devID),
			Month:        mth,
			TriggeredAt:  triggeredAt,
			RuleID:       uuid.UUID(ruleID),
			MetricName:   mName,
			MetricValue:  mValue,
			Threshold:    threshold,
			Severity:     sev,
			Acknowledged: ack,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("iter close: %w", err)
	}

	return result, nil
}

func (r *scyllaAlertEventRepository) Acknowledge(ctx context.Context, deviceID uuid.UUID, month string, triggeredAt time.Time, ruleID uuid.UUID) error {
	query := `UPDATE alert_events SET acknowledged = true WHERE device_id = ? AND month = ? AND triggered_at = ? AND rule_id = ?`
	return r.client.Session.Query(query, deviceID, month, triggeredAt, ruleID).WithContext(ctx).Exec()
}
