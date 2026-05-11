package crop

import (
	"log/slog"
	"time"

	"hydroponic-backend/internal/climate"
	"hydroponic-backend/internal/platform/mqtt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type stageConfigPusher interface {
	PushToDevice(deviceCode string, cfgType string, action string, entityID uint64, payload interface{}) error
}

type BatchStageScheduler struct {
	db       *gorm.DB
	log      *slog.Logger
	pusher   stageConfigPusher
	interval time.Duration
	stopCh   chan struct{}
}

func NewBatchStageScheduler(db *gorm.DB, log *slog.Logger, pusher stageConfigPusher) *BatchStageScheduler {
	return &BatchStageScheduler{
		db:       db,
		log:      log,
		pusher:   pusher,
		interval: 30 * time.Second,
		stopCh:   make(chan struct{}),
	}
}

func (s *BatchStageScheduler) Start() {
	go s.loop()
}

func (s *BatchStageScheduler) Stop() {
	close(s.stopCh)
}

func (s *BatchStageScheduler) loop() {
	t := time.NewTicker(s.interval)
	defer t.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-t.C:
			s.tick()
		}
	}
}

func (s *BatchStageScheduler) tick() {
	now := time.Now().UTC()

	var batchIDs []uint64
	if err := s.db.Model(&CropBatch{}).Select("id").Where("status = ?", BatchStatusRunning).Find(&batchIDs).Error; err != nil {
		if s.log != nil {
			s.log.Warn("batch stage scheduler: list running batches failed", "error", err)
		}
		return
	}

	for _, id := range batchIDs {
		s.processBatch(now, id)
	}
}

type CropBatchStageConfigPayloadV1 struct {
	SchemaVersion    int      `json:"schema_version"`
	BatchID          uint64   `json:"batch_id"`
	StagePlanID      uint64   `json:"stage_plan_id"`
	GrowthStageID    uint64   `json:"growth_stage_id"`
	StageStartAt     string   `json:"stage_start_at"`
	StageEndAt       string   `json:"stage_end_at"`
	ECMin            *float64 `json:"ec_min"`
	ECMax            *float64 `json:"ec_max"`
	PHMin            *float64 `json:"ph_min"`
	PHMax            *float64 `json:"ph_max"`
	RecipeID         *uint64  `json:"recipe_id"`
	PolicyID         *uint64  `json:"policy_id"`
	ClimateProfileID *uint64  `json:"climate_profile_id"`
	SwitchedAt       string   `json:"switched_at"`
	Reason           string   `json:"reason"`
}

func (s *BatchStageScheduler) processBatch(now time.Time, batchID uint64) {
	var plan BatchStagePlan
	err := s.db.
		Where("batch_id = ? AND stage_start_at <= ? AND stage_end_at > ?", batchID, now, now).
		Order("stage_start_at ASC").
		First(&plan).Error
	if err != nil {
		return
	}

	switched, err := s.applySwitch(now, batchID, plan)
	if err != nil {
		if s.log != nil {
			s.log.Warn("batch stage scheduler: apply switch failed", "batch_id", batchID, "error", err)
		}
		return
	}
	if !switched {
		return
	}

	deviceCodes, err := s.lookupActuatorDeviceCodes(batchID)
	if err != nil {
		if s.log != nil {
			s.log.Warn("batch stage scheduler: lookup actuator devices failed", "batch_id", batchID, "error", err)
		}
		return
	}
	if len(deviceCodes) == 0 {
		return
	}

	payload := CropBatchStageConfigPayloadV1{
		SchemaVersion:    1,
		BatchID:          batchID,
		StagePlanID:      plan.ID,
		GrowthStageID:    plan.GrowthStageID,
		StageStartAt:     plan.StageStartAt.UTC().Format(time.RFC3339),
		StageEndAt:       plan.StageEndAt.UTC().Format(time.RFC3339),
		ECMin:            plan.TargetECMin,
		ECMax:            plan.TargetECMax,
		PHMin:            plan.TargetPHMin,
		PHMax:            plan.TargetPHMax,
		RecipeID:         plan.RecipeID,
		PolicyID:         plan.PolicyID,
		ClimateProfileID: plan.ClimateID,
		SwitchedAt:       now.Format(time.RFC3339),
		Reason:           "auto_schedule",
	}

	for _, code := range deviceCodes {
		_ = s.pusher.PushToDevice(code, "crop_batch_stage", "update", batchID, payload)
	}

	if plan.ClimateID != nil {
		var profile climate.ClimateProfile
		if err := s.db.Preload("Stages.Actions").First(&profile, *plan.ClimateID).Error; err == nil {
			profilePayload := climate.BuildProfileConfigPayload(profile)
			for _, code := range deviceCodes {
				_ = s.pusher.PushToDevice(code, "climate_profile", "update", profile.ID, profilePayload)
			}
		}
	}
}

func (s *BatchStageScheduler) applySwitch(now time.Time, batchID uint64, plan BatchStagePlan) (bool, error) {
	var currentStagePlan interface{} = nil
	if plan.ID > 0 {
		currentStagePlan = plan.ID
	}

	var currentGrowthStage interface{} = nil
	if plan.GrowthStageID > 0 {
		currentGrowthStage = plan.GrowthStageID
	}

	recipeVal, policyVal, climateVal := asNullableUint64(plan.RecipeID), asNullableUint64(plan.PolicyID), asNullableUint64(plan.ClimateID)

	switched := false
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var rt BatchStageRuntime
		if err := tx.Where("batch_id = ?", batchID).First(&rt).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if rt.CurrentStagePlan != nil && *rt.CurrentStagePlan == plan.ID {
			return nil
		}

		updates := map[string]interface{}{
			"current_stage_plan_id":   currentStagePlan,
			"current_growth_stage_id": currentGrowthStage,
			"last_switched_at":        now,
		}

		rtUp := BatchStageRuntime{
			BatchID:            batchID,
			CurrentStagePlan:   &plan.ID,
			CurrentGrowthStage: &plan.GrowthStageID,
			LastSwitchedAt:     &now,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "batch_id"}},
			DoUpdates: clause.Assignments(updates),
		}).Create(&rtUp).Error; err != nil {
			return err
		}

		if err := tx.Model(&CropBatch{}).Where("id = ?", batchID).Updates(map[string]interface{}{
			"active_recipe_id":          recipeVal,
			"active_policy_id":          policyVal,
			"active_climate_profile_id": climateVal,
		}).Error; err != nil {
			return err
		}

		switched = true
		return nil
	})
	return switched, err
}

func (s *BatchStageScheduler) lookupActuatorDeviceCodes(batchID uint64) ([]string, error) {
	type row struct {
		DeviceCode string `gorm:"column:device_code"`
	}
	var rows []row
	err := s.db.Table("batch_devices bd").
		Select("ad.device_code").
		Joins("JOIN actuator_devices ad ON ad.id = bd.device_id").
		Where("bd.batch_id = ? AND bd.device_type = ? AND bd.is_active = 1", batchID, DeviceTypeActuator).
		Order("bd.bound_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(rows))
	seen := map[string]struct{}{}
	for _, r := range rows {
		if r.DeviceCode == "" {
			continue
		}
		if _, ok := seen[r.DeviceCode]; ok {
			continue
		}
		seen[r.DeviceCode] = struct{}{}
		out = append(out, r.DeviceCode)
	}
	return out, nil
}

func asNullableUint64(v *uint64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

var _ stageConfigPusher = (*mqtt.ConfigPusher)(nil)
