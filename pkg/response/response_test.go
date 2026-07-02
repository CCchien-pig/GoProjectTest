package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestResponseHelpers(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		OK(c, "test_data")

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var resp Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Code != http.StatusOK || resp.Message != "ok" || resp.Data.(string) != "test_data" {
			t.Errorf("unexpected response content: %+v", resp)
		}
	})

	t.Run("OKWithPagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		pg := &Pagination{
			NextCursor: "next_123",
			HasMore:    true,
			Limit:      10,
		}

		OKWithPagination(c, []string{"a", "b"}, pg)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var resp Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Pagination == nil || resp.Pagination.NextCursor != "next_123" || !resp.Pagination.HasMore || resp.Pagination.Limit != 10 {
			t.Errorf("unexpected pagination content: %+v", resp.Pagination)
		}
	})

	t.Run("BadRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		BadRequest(c, "invalid_parameter")

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var resp Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Code != http.StatusBadRequest || resp.Message != "invalid_parameter" || resp.Data != nil {
			t.Errorf("unexpected response content: %+v", resp)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		NotFound(c, "resource_not_found")

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}

		var resp Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Code != http.StatusNotFound || resp.Message != "resource_not_found" {
			t.Errorf("unexpected response content: %+v", resp)
		}
	})

	t.Run("InternalError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		InternalError(c, "internal_server_error")

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}

		var resp Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Code != http.StatusInternalServerError || resp.Message != "internal_server_error" {
			t.Errorf("unexpected response content: %+v", resp)
		}
	})

	t.Run("ServiceUnavailable", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ServiceUnavailable(c, "service_unavailable")

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
		}

		var resp Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Code != http.StatusServiceUnavailable || resp.Message != "service_unavailable" {
			t.Errorf("unexpected response content: %+v", resp)
		}
	})
}
