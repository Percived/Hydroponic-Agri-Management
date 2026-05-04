package device

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type listGroupsResponse struct {
	Code int    `json:"code"`
	Data groups `json:"data"`
}

type groups struct {
	Items []groupItem `json:"items"`
}

type groupItem struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Greenhouse  uint64 `json:"greenhouse_id"`
	DeviceCount int64  `json:"device_count"`
}

func TestListGroupsReturnsDeviceCount(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&DeviceGroup{}, &Device{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	g1 := DeviceGroup{GreenhouseID: 1, Name: "group-1", Description: "d1"}
	g2 := DeviceGroup{GreenhouseID: 1, Name: "group-2", Description: "d2"}
	if err := db.Create(&g1).Error; err != nil {
		t.Fatalf("create g1 failed: %v", err)
	}
	if err := db.Create(&g2).Error; err != nil {
		t.Fatalf("create g2 failed: %v", err)
	}

	if err := db.Create(&[]Device{
		newTestDevice("D-001", g1.ID),
		newTestDevice("D-002", g1.ID),
		newTestDevice("D-003", g2.ID),
		newTestDeviceWithoutGroup("D-004"),
	}).Error; err != nil {
		t.Fatalf("seed devices failed: %v", err)
	}

	h := NewHandler(db)
	r := gin.New()
	r.GET("/api/device-groups", h.ListGroups)

	req := httptest.NewRequest(http.MethodGet, "/api/device-groups", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	var resp listGroupsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	deviceCountByGroup := make(map[uint64]int64, len(resp.Data.Items))
	for _, item := range resp.Data.Items {
		deviceCountByGroup[item.ID] = item.DeviceCount
	}

	if deviceCountByGroup[g1.ID] != 2 {
		t.Fatalf("group %d device_count = %d, want 2", g1.ID, deviceCountByGroup[g1.ID])
	}
	if deviceCountByGroup[g2.ID] != 1 {
		t.Fatalf("group %d device_count = %d, want 1", g2.ID, deviceCountByGroup[g2.ID])
	}
}

func TestDeleteGroupUnbindsDevices(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&DeviceGroup{}, &Device{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	group := DeviceGroup{GreenhouseID: 1, Name: "group-del", Description: "to delete"}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("create group failed: %v", err)
	}
	if err := db.Create(&[]Device{
		newTestDevice("DG-001", group.ID),
		newTestDevice("DG-002", group.ID),
		newTestDeviceWithoutGroup("DG-003"),
	}).Error; err != nil {
		t.Fatalf("seed devices failed: %v", err)
	}

	h := NewHandler(db)
	r := gin.New()
	r.DELETE("/api/device-groups/:groupId", h.DeleteGroup)

	req := httptest.NewRequest(http.MethodDelete, "/api/device-groups/"+itoa(group.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d, body: %s", w.Code, w.Body.String())
	}

	var count int64
	if err := db.Model(&DeviceGroup{}).Where("id = ?", group.ID).Count(&count).Error; err != nil {
		t.Fatalf("query group failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected group deleted, count=%d", count)
	}

	var groupedDevices int64
	if err := db.Model(&Device{}).Where("group_id = ?", group.ID).Count(&groupedDevices).Error; err != nil {
		t.Fatalf("query devices failed: %v", err)
	}
	if groupedDevices != 0 {
		t.Fatalf("expected devices unbound from group, still has %d", groupedDevices)
	}
}

func TestDeleteGroupReturnsNotFound(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&DeviceGroup{}, &Device{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	h := NewHandler(db)
	r := gin.New()
	r.DELETE("/api/device-groups/:groupId", h.DeleteGroup)

	req := httptest.NewRequest(http.MethodDelete, "/api/device-groups/999999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d, body: %s", w.Code, w.Body.String())
	}
}

func TestDeleteGreenhouseCascadesAndUnbinds(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&Greenhouse{}, &DeviceGroup{}, &Device{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	greenhouse := Greenhouse{Name: "gh-del"}
	if err := db.Create(&greenhouse).Error; err != nil {
		t.Fatalf("create greenhouse failed: %v", err)
	}
	group1 := DeviceGroup{GreenhouseID: greenhouse.ID, Name: "g-1"}
	group2 := DeviceGroup{GreenhouseID: greenhouse.ID, Name: "g-2"}
	if err := db.Create(&group1).Error; err != nil {
		t.Fatalf("create group1 failed: %v", err)
	}
	if err := db.Create(&group2).Error; err != nil {
		t.Fatalf("create group2 failed: %v", err)
	}

	if err := db.Create(&[]Device{
		newTestDeviceWithGreenhouseAndGroup("GH-001", greenhouse.ID, group1.ID),
		newTestDeviceWithGreenhouseAndGroup("GH-002", greenhouse.ID, group2.ID),
		newTestDeviceWithGreenhouse("GH-003", greenhouse.ID),
	}).Error; err != nil {
		t.Fatalf("seed greenhouse devices failed: %v", err)
	}

	h := NewHandler(db)
	r := gin.New()
	r.DELETE("/api/devices/greenhouses/:greenhouseId", h.DeleteGreenhouse)

	req := httptest.NewRequest(http.MethodDelete, "/api/devices/greenhouses/"+itoa(greenhouse.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d, body: %s", w.Code, w.Body.String())
	}

	var greenhouseCount int64
	if err := db.Model(&Greenhouse{}).Where("id = ?", greenhouse.ID).Count(&greenhouseCount).Error; err != nil {
		t.Fatalf("query greenhouse failed: %v", err)
	}
	if greenhouseCount != 0 {
		t.Fatalf("expected greenhouse deleted, count=%d", greenhouseCount)
	}

	var groupCount int64
	if err := db.Model(&DeviceGroup{}).Where("greenhouse_id = ?", greenhouse.ID).Count(&groupCount).Error; err != nil {
		t.Fatalf("query groups failed: %v", err)
	}
	if groupCount != 0 {
		t.Fatalf("expected groups deleted, count=%d", groupCount)
	}

	var boundGreenhouseDevices int64
	if err := db.Model(&Device{}).Where("greenhouse_id = ?", greenhouse.ID).Count(&boundGreenhouseDevices).Error; err != nil {
		t.Fatalf("query greenhouse devices failed: %v", err)
	}
	if boundGreenhouseDevices != 0 {
		t.Fatalf("expected devices unbound from greenhouse, count=%d", boundGreenhouseDevices)
	}

	var boundGroupDevices int64
	if err := db.Model(&Device{}).Where("group_id IN ?", []uint64{group1.ID, group2.ID}).Count(&boundGroupDevices).Error; err != nil {
		t.Fatalf("query grouped devices failed: %v", err)
	}
	if boundGroupDevices != 0 {
		t.Fatalf("expected devices unbound from groups, count=%d", boundGroupDevices)
	}
}

func TestDeleteGreenhouseReturnsNotFound(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&Greenhouse{}, &DeviceGroup{}, &Device{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	h := NewHandler(db)
	r := gin.New()
	r.DELETE("/api/devices/greenhouses/:greenhouseId", h.DeleteGreenhouse)

	req := httptest.NewRequest(http.MethodDelete, "/api/devices/greenhouses/999999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: %d, body: %s", w.Code, w.Body.String())
	}
}

func newTestDevice(code string, groupID uint64) Device {
	return Device{
		DeviceCode:          code,
		Name:                code,
		Type:                DeviceTypeSensor,
		Category:            "TEMP",
		GroupID:             &groupID,
		Status:              DeviceStatusEnabled,
		Protocol:            ProtocolMQTT,
		SamplingIntervalSec: 60,
	}
}

func newTestDeviceWithoutGroup(code string) Device {
	return Device{
		DeviceCode:          code,
		Name:                code,
		Type:                DeviceTypeSensor,
		Category:            "TEMP",
		Status:              DeviceStatusEnabled,
		Protocol:            ProtocolMQTT,
		SamplingIntervalSec: 60,
	}
}

func newTestDeviceWithGreenhouse(code string, greenhouseID uint64) Device {
	return Device{
		DeviceCode:          code,
		Name:                code,
		Type:                DeviceTypeSensor,
		Category:            "TEMP",
		GreenhouseID:        &greenhouseID,
		Status:              DeviceStatusEnabled,
		Protocol:            ProtocolMQTT,
		SamplingIntervalSec: 60,
	}
}

func newTestDeviceWithGreenhouseAndGroup(code string, greenhouseID uint64, groupID uint64) Device {
	return Device{
		DeviceCode:          code,
		Name:                code,
		Type:                DeviceTypeSensor,
		Category:            "TEMP",
		GreenhouseID:        &greenhouseID,
		GroupID:             &groupID,
		Status:              DeviceStatusEnabled,
		Protocol:            ProtocolMQTT,
		SamplingIntervalSec: 60,
	}
}

func itoa(v uint64) string {
	return fmt.Sprintf("%d", v)
}
