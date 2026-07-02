package repository

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/your-name/udm/internal/model"
	"gorm.io/gorm"
)

// DeviceRepository 定義對 devices 資料表的資料存取介面
type DeviceRepository interface {
	Create(ctx context.Context, device *model.Device) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Device, error)
	FindByDeviceCode(ctx context.Context, code string) (*model.Device, error)
	Update(ctx context.Context, device *model.Device) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, cursor string, limit int, deviceType, status, location, search string) ([]*model.Device, string, error)
}

type gormDeviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository 建立 GORM 的 DeviceRepository 實作
func NewDeviceRepository(db *gorm.DB) DeviceRepository {
	return &gormDeviceRepository{db: db}
}

func (r *gormDeviceRepository) Create(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *gormDeviceRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Device, error) {
	var device model.Device
	if err := r.db.WithContext(ctx).Preload("Owner").First(&device, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (r *gormDeviceRepository) FindByDeviceCode(ctx context.Context, code string) (*model.Device, error) {
	var device model.Device
	if err := r.db.WithContext(ctx).First(&device, "device_code = ?", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (r *gormDeviceRepository) Update(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Save(device).Error
}

func (r *gormDeviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Device{}, "id = ?", id).Error
}

func (r *gormDeviceRepository) List(ctx context.Context, cursor string, limit int, deviceType, status, location, search string) ([]*model.Device, string, error) {
	query := r.db.WithContext(ctx).Model(&model.Device{}).Preload("Owner")

	// 篩選過濾
	if deviceType != "" {
		query = query.Where("device_type = ?", deviceType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if location != "" {
		query = query.Where("location = ?", location)
	}

	// pg_trgm 模糊搜尋
	if search != "" {
		query = query.Where("device_code ILIKE ? OR name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 套用 Cursor-based 分頁 (預設降序排序：最新創建的在前面)
	if cursor != "" {
		cursorTime, cursorID, err := decodeCursor(cursor)
		if err == nil {
			// 在 PostgreSQL 中，可以使用 Tuple 比較：(created_at, id) < (cursorTime, cursorID)
			query = query.Where("(created_at, id) < (?, ?)", cursorTime, cursorID)
		}
	}

	// 排序並限制回傳筆數 (多查一筆以確認有無下一頁)
	query = query.Order("created_at DESC, id DESC").Limit(limit + 1)

	var devices []*model.Device
	if err := query.Find(&devices).Error; err != nil {
		return nil, "", err
	}

	nextCursor := ""
	if len(devices) > limit {
		// 有下一頁，生成 next_cursor 並移除最後一筆
		nextCursor = encodeCursor(devices[limit-1].CreatedAt, devices[limit-1].ID)
		devices = devices[:limit]
	}

	return devices, nextCursor, nil
}

// Cursor 編解碼輔助函式
func encodeCursor(t time.Time, id uuid.UUID) string {
	str := fmt.Sprintf("%s,%s", t.Format(time.RFC3339Nano), id.String())
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func decodeCursor(cursor string) (time.Time, uuid.UUID, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	parts := strings.Split(string(decoded), ",")
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor format")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	return t, id, nil
}
