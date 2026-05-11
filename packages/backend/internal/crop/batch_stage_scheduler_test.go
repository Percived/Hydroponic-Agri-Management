package crop

import (
	"testing"
	"time"

	"hydroponic-backend/internal/device"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type pushCall struct {
	deviceCode string
	cfgType    string
	action     string
	entityID   uint64
}

type fakeStagePusher struct {
	calls []pushCall
}

func (p *fakeStagePusher) PushToDevice(deviceCode string, cfgType string, action string, entityID uint64, _ interface{}) error {
	p.calls = append(p.calls, pushCall{
		deviceCode: deviceCode,
		cfgType:    cfgType,
		action:     action,
		entityID:   entityID,
	})
	return nil
}

func TestBatchStageScheduler_ProcessBatch_SwitchOnce(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&CropBatch{}, &BatchStagePlan{}, &BatchStageRuntime{}, &BatchDevice{}, &device.ActuatorDevice{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Now().UTC()
	recipeID := uint64(10)
	policyID := uint64(20)

	b := CropBatch{
		BatchNo:       "B1",
		GreenhouseID:  1,
		CropVarietyID: 1,
		Status:        BatchStatusRunning,
	}
	if err := db.Create(&b).Error; err != nil {
		t.Fatalf("create batch: %v", err)
	}

	plan := BatchStagePlan{
		BatchID:       b.ID,
		GrowthStageID: 1,
		RecipeID:      &recipeID,
		PolicyID:      &policyID,
		StageStartAt:  now.Add(-1 * time.Hour),
		StageEndAt:    now.Add(1 * time.Hour),
	}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}

	ad1 := device.ActuatorDevice{DeviceCode: "A-1", Name: "a1", GreenhouseID: 1}
	ad2 := device.ActuatorDevice{DeviceCode: "A-2", Name: "a2", GreenhouseID: 1}
	if err := db.Create(&ad1).Error; err != nil {
		t.Fatalf("create actuator1: %v", err)
	}
	if err := db.Create(&ad2).Error; err != nil {
		t.Fatalf("create actuator2: %v", err)
	}

	bd1 := BatchDevice{BatchID: b.ID, DeviceType: DeviceTypeActuator, DeviceID: ad1.ID, BoundAt: now, IsActive: true}
	bd2 := BatchDevice{BatchID: b.ID, DeviceType: DeviceTypeActuator, DeviceID: ad2.ID, BoundAt: now, IsActive: true}
	if err := db.Create(&bd1).Error; err != nil {
		t.Fatalf("bind1: %v", err)
	}
	if err := db.Create(&bd2).Error; err != nil {
		t.Fatalf("bind2: %v", err)
	}

	pusher := &fakeStagePusher{}
	s := NewBatchStageScheduler(db, nil, pusher)

	s.processBatch(now, b.ID)
	if len(pusher.calls) != 2 {
		t.Fatalf("expected 2 pushes, got %d", len(pusher.calls))
	}
	for _, c := range pusher.calls {
		if c.cfgType != "crop_batch_stage" {
			t.Fatalf("expected cfgType crop_batch_stage, got %s", c.cfgType)
		}
		if c.entityID != b.ID {
			t.Fatalf("expected entityID batch id, got %d", c.entityID)
		}
	}

	var rt BatchStageRuntime
	if err := db.Where("batch_id = ?", b.ID).First(&rt).Error; err != nil {
		t.Fatalf("runtime: %v", err)
	}
	if rt.CurrentStagePlan == nil || *rt.CurrentStagePlan != plan.ID {
		t.Fatalf("expected runtime current_stage_plan_id=%d", plan.ID)
	}

	var b2 CropBatch
	if err := db.First(&b2, b.ID).Error; err != nil {
		t.Fatalf("reload batch: %v", err)
	}
	if b2.ActiveRecipeID == nil || *b2.ActiveRecipeID != recipeID {
		t.Fatalf("expected active_recipe_id=%d", recipeID)
	}
	if b2.ActivePolicyID == nil || *b2.ActivePolicyID != policyID {
		t.Fatalf("expected active_policy_id=%d", policyID)
	}

	s.processBatch(now.Add(10*time.Second), b.ID)
	if len(pusher.calls) != 2 {
		t.Fatalf("expected no re-push, got %d calls", len(pusher.calls))
	}
}

func TestBatchStageScheduler_ProcessBatch_StageChangeTriggers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&CropBatch{}, &BatchStagePlan{}, &BatchStageRuntime{}, &BatchDevice{}, &device.ActuatorDevice{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Now().UTC()
	b := CropBatch{
		BatchNo:       "B1",
		GreenhouseID:  1,
		CropVarietyID: 1,
		Status:        BatchStatusRunning,
	}
	if err := db.Create(&b).Error; err != nil {
		t.Fatalf("create batch: %v", err)
	}

	plan1 := BatchStagePlan{
		BatchID:       b.ID,
		GrowthStageID: 1,
		StageStartAt:  now.Add(-2 * time.Hour),
		StageEndAt:    now.Add(-1 * time.Hour),
	}
	plan2 := BatchStagePlan{
		BatchID:       b.ID,
		GrowthStageID: 2,
		StageStartAt:  now.Add(-1 * time.Minute),
		StageEndAt:    now.Add(1 * time.Hour),
	}
	if err := db.Create(&plan1).Error; err != nil {
		t.Fatalf("create plan1: %v", err)
	}
	if err := db.Create(&plan2).Error; err != nil {
		t.Fatalf("create plan2: %v", err)
	}

	ad := device.ActuatorDevice{DeviceCode: "A-1", Name: "a1", GreenhouseID: 1}
	if err := db.Create(&ad).Error; err != nil {
		t.Fatalf("create actuator: %v", err)
	}
	bd := BatchDevice{BatchID: b.ID, DeviceType: DeviceTypeActuator, DeviceID: ad.ID, BoundAt: now, IsActive: true}
	if err := db.Create(&bd).Error; err != nil {
		t.Fatalf("bind: %v", err)
	}

	pusher := &fakeStagePusher{}
	s := NewBatchStageScheduler(db, nil, pusher)

	s.processBatch(now.Add(-90*time.Minute), b.ID)
	if len(pusher.calls) != 1 {
		t.Fatalf("expected 1 push for stage1, got %d", len(pusher.calls))
	}

	s.processBatch(now, b.ID)
	if len(pusher.calls) != 2 {
		t.Fatalf("expected stage change to trigger push, got %d", len(pusher.calls))
	}

	s.processBatch(now.Add(10*time.Second), b.ID)
	if len(pusher.calls) != 2 {
		t.Fatalf("expected no re-push, got %d", len(pusher.calls))
	}
}
