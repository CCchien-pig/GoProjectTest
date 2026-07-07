package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"GoProject/udm/internal/dto"
	"GoProject/udm/internal/service"
	"GoProject/udm/pkg/response"
)

type mockUserService struct {
	onCreate     func(req *dto.CreateUserReq) (*dto.UserResp, error)
	onFindByID   func(id uuid.UUID) (*dto.UserResp, error)
	onUpdate     func(id uuid.UUID, req *dto.UpdateUserReq) (*dto.UserResp, error)
	onSoftDelete func(id uuid.UUID) error
}

func (m *mockUserService) Create(ctx context.Context, req *dto.CreateUserReq) (*dto.UserResp, error) {
	if m.onCreate != nil {
		return m.onCreate(req)
	}
	return &dto.UserResp{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Role:      &dto.RoleResp{ID: req.RoleID},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockUserService) FindByID(ctx context.Context, id uuid.UUID) (*dto.UserResp, error) {
	if m.onFindByID != nil {
		return m.onFindByID(id)
	}
	return &dto.UserResp{
		ID:        id,
		Username:  "mockuser",
		Email:     "mock@example.com",
		Role:      &dto.RoleResp{Name: "viewer"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockUserService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserReq) (*dto.UserResp, error) {
	if m.onUpdate != nil {
		return m.onUpdate(id, req)
	}
	return &dto.UserResp{
		ID:        id,
		Username:  "updateduser",
		Email:     "updated@example.com",
		Role:      &dto.RoleResp{Name: "operator"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockUserService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.onSoftDelete != nil {
		return m.onSoftDelete(id)
	}
	return nil
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestUserHandler_Create(t *testing.T) {
	svc := &mockUserService{}
	h := NewUserHandler(svc)

	r := gin.New()
	r.POST("/users", h.Create)

	reqPayload := dto.CreateUserReq{
		Username: "adminuser",
		Email:    "admin@example.com",
		Password: "password",
		RoleID:   uuid.New(),
	}

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected code 200, got %d", w.Code)
	}

	var res response.Response
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res.Code != http.StatusOK {
		t.Errorf("expected 200 inside response, got %d", res.Code)
	}
}

func TestUserHandler_FindByID(t *testing.T) {
	uid := uuid.New()
	svc := &mockUserService{
		onFindByID: func(id uuid.UUID) (*dto.UserResp, error) {
			if id != uid {
				return nil, service.ErrUserNotFound
			}
			return &dto.UserResp{ID: id, Username: "found"}, nil
		},
	}
	h := NewUserHandler(svc)

	r := gin.New()
	r.GET("/users/:id", h.FindByID)

	// Test Found
	req, _ := http.NewRequest("GET", "/users/"+uid.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	// Test Not Found
	req2, _ := http.NewRequest("GET", "/users/"+uuid.New().String(), nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Errorf("expected status NotFound, got %d", w2.Code)
	}
}
