package policy

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestUpdatePolicyResetsLastScheduledForWhenScheduleChanges(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openPolicySchedulerTestDB(t)
	lastScheduledFor := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)

	policy := ControlPolicy{
		PolicyCode:       "POL-SCHED-UPD-001",
		Name:             "daily schedule",
		PolicyType:       "SCHEDULE",
		GreenhouseID:     1,
		Enabled:          true,
		ScheduleMode:     stringPtr("DAILY"),
		TimeOfDay:        stringPtr("08:00:00"),
		Timezone:         "UTC",
		LastScheduledFor: &lastScheduledFor,
	}
	if err := db.Create(&policy).Error; err != nil {
		t.Fatalf("create policy: %v", err)
	}

	h := NewHandler(db)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPut, "/policies/"+strconv.FormatUint(policy.ID, 10), strings.NewReader(`{"schedule_mode":"DAILY","time_of_day":"09:30:00","timezone":"UTC"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(policy.ID, 10)}}
	c.Set("request_id", "req-update-policy")

	h.UpdatePolicy(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var updated ControlPolicy
	if err := db.First(&updated, policy.ID).Error; err != nil {
		t.Fatalf("reload policy: %v", err)
	}
	if updated.LastScheduledFor != nil {
		t.Fatalf("expected last_scheduled_for to be reset when schedule changes, got %v", updated.LastScheduledFor)
	}
	if updated.TimeOfDay == nil || *updated.TimeOfDay != "09:30:00" {
		t.Fatalf("expected updated time_of_day 09:30:00, got %+v", updated.TimeOfDay)
	}
}
