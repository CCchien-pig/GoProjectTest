package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/model"
)

type mockUserRepository struct {
	users        map[uuid.UUID]*model.User
	usersByName  map[string]*model.User
	usersByEmail map[string]*model.User
	onCreate     func(user *model.User) error
	onFindByID   func(id uuid.UUID) (*model.User, error)
	onUpdate     func(user *model.User) error
	onSoftDelete func(id uuid.UUID) error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:        make(map[uuid.UUID]*model.User),
		usersByName:  make(map[string]*model.User),
		usersByEmail: make(map[string]*model.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
	if m.onCreate != nil {
		return m.onCreate(user)
	}
	user.ID = uuid.New()
	m.users[user.ID] = user
	m.usersByName[user.Username] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if m.onFindByID != nil {
		return m.onFindByID(id)
	}
	return m.users[id], nil
}

func (m *mockUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	return m.usersByName[username], nil
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.usersByEmail[email], nil
}

func (m *mockUserRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.User, error) {
	var list []*model.User
	for _, id := range ids {
		if u, ok := m.users[id]; ok {
			list = append(list, u)
		}
	}
	return list, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *model.User) error {
	if m.onUpdate != nil {
		return m.onUpdate(user)
	}
	m.users[user.ID] = user
	m.usersByName[user.Username] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.onSoftDelete != nil {
		return m.onSoftDelete(id)
	}
	if user, exists := m.users[id]; exists {
		user.IsActive = false
	}
	return nil
}

func TestUserService_Create(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	req := &dto.CreateUserReq{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		RoleID:   uuid.New(),
	}

	resp, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if resp.Username != "testuser" || resp.Email != "test@example.com" || !resp.IsActive {
		t.Errorf("unexpected user response: %+v", resp)
	}

	// 測試?��?註�? Username
	_, err = svc.Create(context.Background(), req)
	if err == nil || err != ErrUsernameDuplicate {
		t.Errorf("expected ErrUsernameDuplicate, got %v", err)
	}

	// 測試?��?註�? Email
	req2 := &dto.CreateUserReq{
		Username: "another",
		Email:    "test@example.com",
		Password: "password",
		RoleID:   uuid.New(),
	}
	_, err = svc.Create(context.Background(), req2)
	if err == nil || err != ErrEmailDuplicate {
		t.Errorf("expected ErrEmailDuplicate, got %v", err)
	}
}

func TestUserService_FindByID(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewUserService(repo)

	req := &dto.CreateUserReq{
		Username: "findme",
		Email:    "findme@example.com",
		Password: "password",
		RoleID:   uuid.New(),
	}
	created, _ := svc.Create(context.Background(), req)

	resp, err := svc.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("failed to find user: %v", err)
	}
	if resp.ID != created.ID || resp.Username != "findme" {
		t.Errorf("unexpected user found: %+v", resp)
	}

	_, err = svc.FindByID(context.Background(), uuid.New())
	if err == nil || err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound for fake ID, got %v", err)
	}
}
