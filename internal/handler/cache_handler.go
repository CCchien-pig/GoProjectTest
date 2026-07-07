package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"GoProject/udm/internal/cache"
	"GoProject/udm/pkg/response"
)

// CacheHandler 快取管理 Handler
type CacheHandler struct {
	cache cache.Service
}

// NewCacheHandler 建立 CacheHandler
func NewCacheHandler(cacheService cache.Service) *CacheHandler {
	return &CacheHandler{cache: cacheService}
}

// InvalidateReq 清除快取請求
type InvalidateReq struct {
	Pattern string `json:"pattern" binding:"required"`
}

// Invalidate 按 Pattern 清除快取
// POST /api/v1/cache/invalidate
// Finding #2: 加入 admin 角色驗證，防止未授權使用者清空快取造成 Cache Avalanche
func (h *CacheHandler) Invalidate(c *gin.Context) {
	// 權限控制：只有 admin 角色才能執行快取清除
	// 從 Gin context 取得 role（由 JWT/Auth middleware 在驗證後注入）
	role, _ := c.Get("user_role")
	if role != "admin" {
		response.Forbidden(c, "only admin role can invalidate cache")
		return
	}

	var req InvalidateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid pattern format")
		return
	}

	deleted, err := h.cache.InvalidateByPattern(c.Request.Context(), req.Pattern)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "ok",
		"data": gin.H{
			"deleted_keys_count": deleted,
		},
	})
}
