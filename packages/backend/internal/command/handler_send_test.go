package command

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hydroponic-backend/internal/platform/event"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSendCommand_MQTTNotConnected_MarksFailedAndCreatesReceipt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&ControlCommand{}, &ControlCommandReceipt{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cmd := ControlCommand{
		ActuatorChannelID: 1,
		CommandType:       "SWITCH",
		Payload:           `{"state":"ON"}`,
		Status:            "PENDING",
		CreatedBy:         1,
	}
	if err := db.Create(&cmd).Error; err != nil {
		t.Fatalf("create cmd: %v", err)
	}

	h := NewHandler(db, nil, event.NewHub())

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/commands/:id/send", h.SendCommand)

	req := httptest.NewRequest(http.MethodPost, "/commands/"+itoa(cmd.ID)+"/send", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: itoa(cmd.ID)}}

	h.SendCommand(c)

	var reloaded ControlCommand
	if err := db.First(&reloaded, cmd.ID).Error; err != nil {
		t.Fatalf("reload cmd: %v", err)
	}
	if reloaded.Status != "FAILED" {
		t.Fatalf("expected FAILED, got %s", reloaded.Status)
	}
	if reloaded.SentAt != nil {
		t.Fatalf("expected sent_at nil")
	}

	var rcpt ControlCommandReceipt
	if err := db.Where("command_id = ?", cmd.ID).First(&rcpt).Error; err != nil {
		t.Fatalf("receipt: %v", err)
	}
	if rcpt.ReceiptStatus != "FAILED" {
		t.Fatalf("expected receipt FAILED, got %s", rcpt.ReceiptStatus)
	}
	if rcpt.AckAt == nil || time.Since(*rcpt.AckAt) > time.Minute {
		t.Fatalf("expected ack_at set")
	}
}

func itoa(v uint64) string {
	return fmt.Sprintf("%d", v)
}
