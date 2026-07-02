package scylla

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/your-name/udm/internal/dto"
	"github.com/your-name/udm/internal/model"
)

// TelemetryRepository 定義對 ScyllaDB telemetry 表的資料存取介面
type TelemetryRepository interface {
	BatchInsert(ctx context.Context, deviceID uuid.UUID, points []dto.TelemetryPoint) error
	Query(ctx context.Context, deviceID uuid.UUID, start, end time.Time, metricName string) ([]*model.TelemetryData, error)
	QueryLatest(ctx context.Context, deviceID uuid.UUID) ([]*model.TelemetryData, error)
	DeleteByRange(ctx context.Context, deviceID uuid.UUID, start, end time.Time) error
}

type scyllaTelemetryRepository struct {
	client *Client
}

// NewTelemetryRepository 建立 ScyllaDB 的 TelemetryRepository 實作
func NewTelemetryRepository(client *Client) TelemetryRepository {
	return &scyllaTelemetryRepository{client: client}
}

func (r *scyllaTelemetryRepository) BatchInsert(ctx context.Context, deviceID uuid.UUID, points []dto.TelemetryPoint) error {
	if len(points) == 0 {
		return nil
	}

	const maxBatchSize = 100
	for i := 0; i < len(points); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(points) {
			end = len(points)
		}
		batchPoints := points[i:end]

		batch := r.client.Session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)
		for _, p := range batchPoints {
			dateStr := p.RecordedAt.Format("2006-01-02")
			query := `INSERT INTO telemetry (device_id, date, recorded_at, metric_name, value, unit, tags) VALUES (?, ?, ?, ?, ?, ?, ?)`
			batch.Query(query, deviceID, dateStr, p.RecordedAt, p.MetricName, p.Value, p.Unit, p.Tags)
		}

		if err := r.client.Session.ExecuteBatch(batch); err != nil {
			return fmt.Errorf("execute batch: %w", err)
		}
	}

	return nil
}

func (r *scyllaTelemetryRepository) Query(ctx context.Context, deviceID uuid.UUID, start, end time.Time, metricName string) ([]*model.TelemetryData, error) {
	var result []*model.TelemetryData

	startDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDate := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		queryStr := `SELECT device_id, date, recorded_at, metric_name, value, unit, tags FROM telemetry WHERE device_id = ? AND date = ? AND recorded_at >= ? AND recorded_at <= ?`
		iter := r.client.Session.Query(queryStr, deviceID, dateStr, start, end).WithContext(ctx).Iter()

		var devID gocql.UUID
		var date string
		var recordedAt time.Time
		var mName, unit string
		var tags map[string]string
		var value float64

		for iter.Scan(&devID, &date, &recordedAt, &mName, &value, &unit, &tags) {
			if metricName != "" && mName != metricName {
				continue
			}

			result = append(result, &model.TelemetryData{
				DeviceID:   uuid.UUID(devID),
				Date:       date,
				RecordedAt: recordedAt,
				MetricName: mName,
				Value:      value,
				Unit:       unit,
				Tags:       tags,
			})
		}

		if err := iter.Close(); err != nil {
			return nil, fmt.Errorf("iter close: %w", err)
		}
	}

	// 降序排序 (newest first)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].RecordedAt.Before(result[j].RecordedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

func (r *scyllaTelemetryRepository) QueryLatest(ctx context.Context, deviceID uuid.UUID) ([]*model.TelemetryData, error) {
	now := time.Now()
	daysToQuery := []string{
		now.Format("2006-01-02"),
		now.AddDate(0, 0, -1).Format("2006-01-02"),
	}

	latestMetrics := make(map[string]*model.TelemetryData)

	for _, dateStr := range daysToQuery {
		queryStr := `SELECT device_id, date, recorded_at, metric_name, value, unit, tags FROM telemetry WHERE device_id = ? AND date = ? LIMIT 100`
		iter := r.client.Session.Query(queryStr, deviceID, dateStr).WithContext(ctx).Iter()

		var devID gocql.UUID
		var date string
		var recordedAt time.Time
		var mName, unit string
		var tags map[string]string
		var value float64

		for iter.Scan(&devID, &date, &recordedAt, &mName, &value, &unit, &tags) {
			if _, exists := latestMetrics[mName]; !exists {
				latestMetrics[mName] = &model.TelemetryData{
					DeviceID:   uuid.UUID(devID),
					Date:       date,
					RecordedAt: recordedAt,
					MetricName: mName,
					Value:      value,
					Unit:       unit,
					Tags:       tags,
				}
			}
		}
		_ = iter.Close()
	}

	var result []*model.TelemetryData
	for _, val := range latestMetrics {
		result = append(result, val)
	}

	return result, nil
}

func (r *scyllaTelemetryRepository) DeleteByRange(ctx context.Context, deviceID uuid.UUID, start, end time.Time) error {
	startDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDate := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		queryStr := `DELETE FROM telemetry WHERE device_id = ? AND date = ? AND recorded_at >= ? AND recorded_at <= ?`
		err := r.client.Session.Query(queryStr, deviceID, dateStr, start, end).WithContext(ctx).Exec()
		if err != nil {
			return fmt.Errorf("delete range for date %s: %w", dateStr, err)
		}
	}

	return nil
}
