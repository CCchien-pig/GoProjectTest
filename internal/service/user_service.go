package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/model"
	"GoProject/udm/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrUserNotFound 找不到使用者
	ErrUserNotFound = errors.New("user not found")
	// ErrUsernameDuplicate 使用者名稱重複
	ErrUsernameDuplicate = errors.New("username already exists")
	// ErrEmailDuplicate Email 重複
	ErrEmailDuplicate = errors.New("email already exists")
)

// UserService 定義使用者業務邏輯介面
type UserService interface {
	Create(ctx context.Context, req *dto.CreateUserReq) (*dto.UserResp, error)
	FindByID(ctx context.Context, id uuid.UUID) (*dto.UserResp, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserReq) (*dto.UserResp, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService 建立 UserService 實體
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Create(ctx context.Context, req *dto.CreateUserReq) (*dto.UserResp, error) {
	// 檢查 Username
	existingUsername, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("find by username: %w", err)
	}
	if existingUsername != nil {
		return nil, ErrUsernameDuplicate
	}

	// 檢查 Email
	existingEmail, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("find by email: %w", err)
	}
	if existingEmail != nil {
		return nil, ErrEmailDuplicate
	}

	// 密碼雜湊
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashed),
		RoleID:       req.RoleID,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// 重新讀取以取得關聯的 Role 與 Permissions
	reloaded, err := s.repo.FindByID(ctx, user.ID)
	if err == nil && reloaded != nil {
		user = reloaded
	}

	return dto.ToUserResp(user), nil
}

func (s *userService) FindByID(ctx context.Context, id uuid.UUID) (*dto.UserResp, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return dto.ToUserResp(user), nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserReq) (*dto.UserResp, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if req.Username != nil {
		existing, err := s.repo.FindByUsername(ctx, *req.Username)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrUsernameDuplicate
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		existing, err := s.repo.FindByEmail(ctx, *req.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrEmailDuplicate
		}
		user.Email = *req.Email
	}

	if req.RoleID != nil {
		user.RoleID = *req.RoleID
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	// 重新讀取以取得關聯的 Role 與 Permissions
	reloaded, err := s.repo.FindByID(ctx, user.ID)
	if err == nil && reloaded != nil {
		user = reloaded
	}

	return dto.ToUserResp(user), nil
}

func (s *userService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find user for delete: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	return nil
}
