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

func TestSSEHandler_DeviceStatus_DeviceCodesFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hub := NewHub()
	r := gin.New()
	r.GET("/devices/subscribe", SSEHandler(hub, "device:status"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/devices/subscribe?device_codes=DEV-1", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		r.ServeHTTP(rec, req)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)

	hub.Publish(SSEEvent{Type: "device:status", Data: DeviceStatusSSEDataV1{SchemaVersion: 1, DeviceCode: "DEV-2", Status: "ONLINE", ReportedAt: time.Now().UTC().Format(time.RFC3339)}})
	hub.Publish(SSEEvent{Type: "device:status", Data: DeviceStatusSSEDataV1{SchemaVersion: 1, DeviceCode: "DEV-1", Status: "OFFLINE", ReportedAt: time.Now().UTC().Format(time.RFC3339)}})

	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	out := rec.Body.String()
	if strings.Contains(out, `"DEV-2"`) {
		t.Fatalf("unexpected DEV-2 delivered, out=%q", out)
	}
	if !strings.Contains(out, `"DEV-1"`) {
		t.Fatalf("expected DEV-1 delivered, out=%q", out)
	}
	if !strings.Contains(out, `"type":"device_status"`) && !strings.Contains(out, `"type": "device_status"`) {
		t.Fatalf("expected device_status mapping, out=%q", out)
	}
}

func TestSSEHandlerMulti_Commands_DeviceCodesFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hub := NewHub()
	r := gin.New()
	r.GET("/commands/subscribe", SSEHandlerMulti(hub, []string{"command:dispatched", "command:acked"}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/commands/subscribe?device_codes=ACT-1", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		r.ServeHTTP(rec, req)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)

	now := time.Now().UTC().Format(time.RFC3339)
	hub.Publish(SSEEvent{Type: "command:dispatched", Data: CommandDispatchedSSEDataV1{SchemaVersion: 1, CommandID: 11, DeviceCode: "ACT-2", Status: "SENT", DispatchedAt: now, SourceType: "MANUAL"}})
	hub.Publish(SSEEvent{Type: "command:dispatched", Data: CommandDispatchedSSEDataV1{SchemaVersion: 1, CommandID: 10, DeviceCode: "ACT-1", Status: "SENT", DispatchedAt: now, SourceType: "MANUAL"}})
	hub.Publish(SSEEvent{Type: "command:acked", Data: CommandAckData{SchemaVersion: 1, CommandID: 10, DeviceCode: "ACT-1", AckCode: "OK", AckMessage: "ok", AckPayload: map[string]interface{}{}, AckedAt: now}})

	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	out := rec.Body.String()
	if strings.Contains(out, `"ACT-2"`) {
		t.Fatalf("unexpected ACT-2 delivered, out=%q", out)
	}
	if !strings.Contains(out, `"command_id":10`) && !strings.Contains(out, `"command_id": 10`) {
		t.Fatalf("expected command_id=10 delivered, out=%q", out)
	}
	if !strings.Contains(out, `"type":"command_dispatched"`) && !strings.Contains(out, `"type": "command_dispatched"`) {
		t.Fatalf("expected command_dispatched mapping, out=%q", out)
	}
	if !strings.Contains(out, `"type":"command_acked"`) && !strings.Contains(out, `"type": "command_acked"`) {
		t.Fatalf("expected command_acked mapping, out=%q", out)
	}
}
