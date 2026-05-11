package mqtt

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConfigDeliveryRepo_AllocateNextRev_Increments(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&ConfigDelivery{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewConfigDeliveryRepo(db)
	rev1, err := repo.AllocateNextRev(db, "DEV-1", "climate_profile", 1)
	if err != nil {
		t.Fatalf("allocate1: %v", err)
	}
	if rev1 != 1 {
		t.Fatalf("expected rev=1, got %d", rev1)
	}

	d := &ConfigDelivery{
		MsgID:          "m1",
		TraceID:        "t1",
		DeviceCode:     "DEV-1",
		ConfigType:     "climate_profile",
		Action:         "update",
		EntityID:       1,
		EntityRev:      rev1,
		SchemaVersion:  1,
		IssuedAtMS:     uint64(time.Now().UnixMilli()),
		TTLsec:         600,
		RequireAck:     true,
		RequestPayload: `{"schema_version":1}`,
		Status:         ConfigDeliveryStatusPending,
	}
	if err := repo.Create(db, d); err != nil {
		t.Fatalf("create: %v", err)
	}

	rev2, err := repo.AllocateNextRev(db, "DEV-1", "climate_profile", 1)
	if err != nil {
		t.Fatalf("allocate2: %v", err)
	}
	if rev2 != 2 {
		t.Fatalf("expected rev=2, got %d", rev2)
	}
}

func TestConfigDeliveryRepo_MarkAckedByMsgID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&ConfigDelivery{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewConfigDeliveryRepo(db)
	d := &ConfigDelivery{
		MsgID:          "m1",
		TraceID:        "t1",
		DeviceCode:     "DEV-1",
		ConfigType:     "climate_profile",
		Action:         "update",
		EntityID:       1,
		EntityRev:      1,
		SchemaVersion:  1,
		IssuedAtMS:     uint64(time.Now().UnixMilli()),
		TTLsec:         600,
		RequireAck:     true,
		RequestPayload: `{"schema_version":1}`,
		Status:         ConfigDeliveryStatusSent,
	}
	if err := repo.Create(db, d); err != nil {
		t.Fatalf("create: %v", err)
	}

	ackedAt := time.Now().UTC()
	rows, err := repo.MarkAckedByMsgID("m1", ackedAt, `{}`, "fw1", "sha256:x")
	if err != nil {
		t.Fatalf("mark acked: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected 1 row affected, got %d", rows)
	}

	var got ConfigDelivery
	if err := db.First(&got, d.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Status != ConfigDeliveryStatusAcked {
		t.Fatalf("expected ACKED, got %s", got.Status)
	}
	if got.AckedAt == nil {
		t.Fatalf("expected acked_at set")
	}
}
