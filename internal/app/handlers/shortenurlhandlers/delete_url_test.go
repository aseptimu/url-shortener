package shortenurlhandlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestDeleteUserURLs_Success(t *testing.T) {
	DeleteTaskCh = make(chan DeleteTask, 100)

	handler := NewDeleteURLHandler(&config.ConfigType{}, nil, zap.NewNop().Sugar())

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	payload := `["a","b","c"]`
	c.Request = httptest.NewRequest(http.MethodDelete, "/urls", bytes.NewBufferString(payload))
	c.Set("userID", "user123")

	handler.DeleteUserURLs(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	select {
	case task := <-DeleteTaskCh:
		if !reflect.DeepEqual(task.URLs, []string{"a", "b", "c"}) {
			t.Errorf("expected URLs %v, got %v", []string{"a", "b", "c"}, task.URLs)
		}
		if task.UserID != "user123" {
			t.Errorf("expected UserID %q, got %q", "user123", task.UserID)
		}
	default:
		t.Fatal("expected a task to be enqueued, but channel was empty")
	}
}

func TestDeleteUserURLs_BadJSON(t *testing.T) {
	handler := NewDeleteURLHandler(&config.ConfigType{}, nil, zap.NewNop().Sugar())

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/urls", bytes.NewBufferString("not a valid json"))
	c.Set("userID", "user123")

	handler.DeleteUserURLs(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for bad JSON, got %d", http.StatusBadRequest, w.Code)
	}
}
