package notification

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"hydroponic-backend/internal/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTestChannel_EnforcesUserOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&NotificationChannel{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ch := NotificationChannel{
		UserID:        2,
		ChannelType:   ChannelWebhook,
		Name:          "other-user",
		Config:        []byte(`{"url":"http://example.com"}`),
		MinAlertLevel: "WARN",
		Enabled:       true,
	}
	if err := db.Create(&ch).Error; err != nil {
		t.Fatalf("create channel: %v", err)
	}

	h := NewHandler(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/notification-channels/1/test", bytes.NewBufferString(`{}`))
	c.Request = req
	c.Params = gin.Params{{Key: "channelId", Value: "1"}}
	c.Set(auth.CtxUserID, uint64(1))

	h.TestChannel(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d, body=%s", w.Code, w.Body.String())
	}
}
