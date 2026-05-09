package event

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSSEHandler_AlertCreated_LevelFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hub := NewHub()
	r := gin.New()
	r.GET("/alerts/subscribe", SSEHandler(hub, "alert:created"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/alerts/subscribe?level=WARN", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		r.ServeHTTP(rec, req)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)

	hub.Publish(SSEEvent{Type: "alert:created", Data: map[string]interface{}{"id": 1, "level": "CRITICAL"}})
	hub.Publish(SSEEvent{Type: "alert:created", Data: map[string]interface{}{"id": 2, "level": "WARN"}})

	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	out := rec.Body.String()
	if strings.Contains(out, `"id":1`) || strings.Contains(out, `"id": 1`) {
		t.Fatalf("unexpected CRITICAL alert delivered, out=%q", out)
	}
	if !strings.Contains(out, `"id":2`) && !strings.Contains(out, `"id": 2`) {
		t.Fatalf("expected WARN alert delivered, out=%q", out)
	}
	if !strings.Contains(out, `"type":"new_alert"`) && !strings.Contains(out, `"type": "new_alert"`) {
		t.Fatalf("expected new_alert mapping, out=%q", out)
	}
}
