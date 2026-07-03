package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"GoProject/udm/internal/model"
	"gorm.io/gorm"
)

// UserRepository 定義了 users 資料表的資料存取介面
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type gormUserRepository struct {
	db *gorm.DB
}

// NewUserRepository 建立 GORM 的 UserRepository 實體
func NewUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *gormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Role.Permissions").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 取得該使用者綁定的設備數量
	var count int64
	if err := r.db.WithContext(ctx).Table("user_devices").Where("user_id = ?", id).Count(&count).Error; err != nil {
		return nil, err
	}
	user.DeviceCount = count

	return &user, nil
}

func (r *gormUserRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.User, error) {
	var users []*model.User
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *gormUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Role.Permissions").First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Role.Permissions").First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *gormUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("is_active", false).Error
}
