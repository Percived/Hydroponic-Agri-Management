package policy

import (
	"fmt"
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

	h := NewHandler(db, nil)
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

func TestUpdatePublishedThresholdPolicyClearsPublishState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openPolicySchedulerTestDB(t)
	publishedAt := time.Now().UTC().Add(-30 * time.Minute).Truncate(time.Second)
	publishedBy := uint64(1)

	policy := ControlPolicy{
		PolicyCode:   "POL-TH-UPD-001",
		Name:         "threshold policy",
		PolicyType:   "THRESHOLD",
		GreenhouseID: 1,
		Enabled:      true,
		PublishedAt:  &publishedAt,
		PublishedBy:  &publishedBy,
	}
	if err := db.Create(&policy).Error; err != nil {
		t.Fatalf("create policy: %v", err)
	}

	h := NewHandler(db, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPut, "/policies/"+strconv.FormatUint(policy.ID, 10), strings.NewReader(`{"name":"threshold policy updated"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(policy.ID, 10)}}
	c.Set("request_id", "req-update-threshold-publish-state")

	h.UpdatePolicy(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var updated ControlPolicy
	if err := db.First(&updated, policy.ID).Error; err != nil {
		t.Fatalf("reload policy: %v", err)
	}
	if updated.PublishedAt != nil {
		t.Fatalf("expected published_at to be cleared, got %v", updated.PublishedAt)
	}
	if updated.PublishedBy != nil {
		t.Fatalf("expected published_by to be cleared, got %v", updated.PublishedBy)
	}
}

func TestPublishPolicyResetsThresholdCooldown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openPolicySchedulerTestDB(t)
	policyID := seedThresholdPolicy(t, db, false)
	scheduler := NewScheduler(db, nil, nil, testLogger())
	scheduler.setCooldown(fmt.Sprintf("%d", policyID))

	h := NewHandler(db, scheduler)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/policies/"+strconv.FormatUint(policyID, 10)+"/publish", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(policyID, 10)}}
	c.Set("user_id", uint64(1))
	c.Set("request_id", "req-publish-reset-cooldown")

	h.PublishPolicy(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if scheduler.isInCooldown(fmt.Sprintf("%d", policyID)) {
		t.Fatalf("expected publish to reset threshold cooldown")
	}
}

func TestPublishPolicyResetsThresholdConditionState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openPolicySchedulerTestDB(t)
	policyID := seedThresholdPolicy(t, db, false)

	var condition PolicyCondition
	if err := db.Where("policy_id = ?", policyID).First(&condition).Error; err != nil {
		t.Fatalf("load condition: %v", err)
	}

	scheduler := NewScheduler(db, nil, nil, testLogger())
	stateKey := fmt.Sprintf("%d:%d", policyID, condition.ID)
	scheduler.condStates[stateKey] = &conditionState{FirstTrueAt: time.Now().UTC().Add(-10 * time.Second)}

	h := NewHandler(db, scheduler)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/policies/"+strconv.FormatUint(policyID, 10)+"/publish", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(policyID, 10)}}
	c.Set("user_id", uint64(1))
	c.Set("request_id", "req-publish-reset-condition-state")

	h.PublishPolicy(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if _, exists := scheduler.condStates[stateKey]; exists {
		t.Fatalf("expected publish to reset threshold condition state")
	}
}

func TestUpdatePublishedSchedulePolicyClearsPublishState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openPolicySchedulerTestDB(t)
	publishedAt := time.Now().UTC().Add(-30 * time.Minute).Truncate(time.Second)
	publishedBy := uint64(1)

	policy := ControlPolicy{
		PolicyCode:   "POL-SCH-UPD-001",
		Name:         "schedule policy",
		PolicyType:   "SCHEDULE",
		GreenhouseID: 1,
		Enabled:      true,
		ScheduleMode: stringPtr("DAILY"),
		TimeOfDay:    stringPtr("08:00:00"),
		Timezone:     "UTC",
		PublishedAt:  &publishedAt,
		PublishedBy:  &publishedBy,
	}
	if err := db.Create(&policy).Error; err != nil {
		t.Fatalf("create policy: %v", err)
	}

	h := NewHandler(db, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPut, "/policies/"+strconv.FormatUint(policy.ID, 10), strings.NewReader(`{"name":"schedule policy updated","schedule_mode":"DAILY","time_of_day":"08:00:00","timezone":"UTC"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(policy.ID, 10)}}
	c.Set("request_id", "req-update-schedule-publish-state")

	h.UpdatePolicy(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var updated ControlPolicy
	if err := db.First(&updated, policy.ID).Error; err != nil {
		t.Fatalf("reload policy: %v", err)
	}
	if updated.PublishedAt != nil {
		t.Fatalf("expected published_at to be cleared, got %v", updated.PublishedAt)
	}
	if updated.PublishedBy != nil {
		t.Fatalf("expected published_by to be cleared, got %v", updated.PublishedBy)
	}
}
