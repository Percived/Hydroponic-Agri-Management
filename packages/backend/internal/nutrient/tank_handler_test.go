package nutrient

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpdateTank_AllowsUnbindingSensorChannelWithNull(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&NutrientTank{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	channelID := uint64(9)
	tank := NutrientTank{
		GrowingZoneID:       1,
		Code:                "TANK-01",
		TotalVolumeLiter:    100,
		Status:              TankStatusActive,
		TempSensorChannelID: &channelID,
	}
	if err := db.Create(&tank).Error; err != nil {
		t.Fatalf("create tank: %v", err)
	}

	h := NewHandler(db)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPut, "/nutrient-tanks/1", bytes.NewBufferString(`{"temp_sensor_channel_id":null}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.UpdateTank(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var reloaded NutrientTank
	if err := db.First(&reloaded, tank.ID).Error; err != nil {
		t.Fatalf("reload tank: %v", err)
	}
	if reloaded.TempSensorChannelID != nil {
		t.Fatalf("expected temp_sensor_channel_id to be nil, got %d", *reloaded.TempSensorChannelID)
	}
}
