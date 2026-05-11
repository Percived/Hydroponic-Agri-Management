package mqtt

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConfigDeliveryRepo struct {
	db *gorm.DB
}

func NewConfigDeliveryRepo(db *gorm.DB) *ConfigDeliveryRepo {
	return &ConfigDeliveryRepo{db: db}
}

func (r *ConfigDeliveryRepo) AllocateNextRev(tx *gorm.DB, deviceCode string, configType string, entityID uint64) (uint64, error) {
	if tx == nil {
		tx = r.db
	}

	var row struct {
		EntityRev uint64 `gorm:"column:entity_rev"`
	}
	err := tx.
		Model(&ConfigDelivery{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("entity_rev").
		Where("device_code = ? AND config_type = ? AND entity_id = ?", deviceCode, configType, entityID).
		Order("entity_rev DESC").
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return 0, err
	}
	return row.EntityRev + 1, nil
}

func (r *ConfigDeliveryRepo) Create(tx *gorm.DB, d *ConfigDelivery) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(d).Error
}

func (r *ConfigDeliveryRepo) MarkSent(id uint64, sentAt time.Time) error {
	return r.db.Model(&ConfigDelivery{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":   ConfigDeliveryStatusSent,
		"sent_at":  sentAt,
		"acked_at": nil,
	}).Error
}

func (r *ConfigDeliveryRepo) MarkFailed(id uint64, code string, msg string, nextRetryAt *time.Time) error {
	return r.db.Model(&ConfigDelivery{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":             ConfigDeliveryStatusFailed,
		"last_error_code":    code,
		"last_error_message": msg,
		"next_retry_at":      nextRetryAt,
	}).Error
}

func (r *ConfigDeliveryRepo) MarkRejectedByMsgID(msgID string, ackedAt time.Time, ackPayload string, errorCode string, errorMsg string, fwVersion string, appliedHash string) (int64, error) {
	res := r.db.Model(&ConfigDelivery{}).Where("msg_id = ?", msgID).Updates(map[string]interface{}{
		"status":             ConfigDeliveryStatusRejected,
		"acked_at":           ackedAt,
		"ack_payload":        ackPayload,
		"last_error_code":    errorCode,
		"last_error_message": errorMsg,
		"device_fw_version":  fwVersion,
		"applied_hash":       appliedHash,
	})
	return res.RowsAffected, res.Error
}

func (r *ConfigDeliveryRepo) MarkAckedByMsgID(msgID string, ackedAt time.Time, ackPayload string, fwVersion string, appliedHash string) (int64, error) {
	res := r.db.Model(&ConfigDelivery{}).Where("msg_id = ?", msgID).Updates(map[string]interface{}{
		"status":            ConfigDeliveryStatusAcked,
		"acked_at":          ackedAt,
		"ack_payload":       ackPayload,
		"device_fw_version": fwVersion,
		"applied_hash":      appliedHash,
	})
	return res.RowsAffected, res.Error
}

func (r *ConfigDeliveryRepo) MarkAckFailedByMsgID(msgID string, ackedAt time.Time, ackPayload string, errorCode string, errorMsg string, fwVersion string, appliedHash string) (int64, error) {
	res := r.db.Model(&ConfigDelivery{}).Where("msg_id = ?", msgID).Updates(map[string]interface{}{
		"status":             ConfigDeliveryStatusFailed,
		"acked_at":           ackedAt,
		"ack_payload":        ackPayload,
		"last_error_code":    errorCode,
		"last_error_message": errorMsg,
		"device_fw_version":  fwVersion,
		"applied_hash":       appliedHash,
	})
	return res.RowsAffected, res.Error
}

func (r *ConfigDeliveryRepo) ListFailedDue(now time.Time, limit int) ([]ConfigDelivery, error) {
	if limit <= 0 {
		limit = 100
	}
	var items []ConfigDelivery
	err := r.db.
		Where("status = ? AND next_retry_at IS NOT NULL AND next_retry_at <= ?", ConfigDeliveryStatusFailed, now).
		Order("next_retry_at ASC, id ASC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *ConfigDeliveryRepo) ListSentCandidates(before time.Time, limit int) ([]ConfigDelivery, error) {
	if limit <= 0 {
		limit = 200
	}
	var items []ConfigDelivery
	err := r.db.
		Where("status = ? AND sent_at IS NOT NULL AND sent_at <= ?", ConfigDeliveryStatusSent, before).
		Order("sent_at ASC, id ASC").
		Limit(limit).
		Find(&items).Error
	return items, err
}
