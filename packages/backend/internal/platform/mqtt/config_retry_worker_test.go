package mqtt

import (
	"testing"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeToken struct {
	err error
}

func (t *fakeToken) Wait() bool                       { return true }
func (t *fakeToken) WaitTimeout(_ time.Duration) bool { return true }
func (t *fakeToken) Done() <-chan struct{}            { ch := make(chan struct{}); close(ch); return ch }
func (t *fakeToken) Error() error                     { return t.err }

type fakePublisher struct {
	connected bool
	publishes []struct {
		topic   string
		payload interface{}
	}
	err error
}

func (p *fakePublisher) IsConnected() bool { return p.connected }

func (p *fakePublisher) Publish(topic string, _ byte, _ bool, payload interface{}) mqttlib.Token {
	p.publishes = append(p.publishes, struct {
		topic   string
		payload interface{}
	}{topic: topic, payload: payload})
	return &fakeToken{err: p.err}
}

func TestConfigRetryWorker_Tick_ResendsFailedDue(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&ConfigDelivery{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Now().UTC()
	nextRetry := now.Add(-1 * time.Second)
	d := ConfigDelivery{
		MsgID:          "m1",
		TraceID:        "t1",
		DeviceCode:     "DEV-1",
		ConfigType:     "climate_profile",
		Action:         "update",
		EntityID:       1,
		EntityRev:      1,
		SchemaVersion:  1,
		IssuedAtMS:     uint64(now.UnixMilli()),
		TTLsec:         600,
		RequireAck:     true,
		RequestPayload: `{"schema_version":1,"msg_id":"m1"}`,
		Status:         ConfigDeliveryStatusFailed,
		RetryCount:     0,
		NextRetryAt:    &nextRetry,
	}
	if err := db.Create(&d).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	pub := &fakePublisher{connected: true}
	w := NewConfigRetryWorker(db, pub, nil)
	w.tick()

	var got ConfigDelivery
	if err := db.First(&got, d.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Status != ConfigDeliveryStatusSent {
		t.Fatalf("expected SENT, got %s", got.Status)
	}
	if got.RetryCount != 1 {
		t.Fatalf("expected retry_count=1, got %d", got.RetryCount)
	}
	if got.SentAt == nil {
		t.Fatalf("expected sent_at set")
	}
	if len(pub.publishes) != 1 {
		t.Fatalf("expected 1 publish, got %d", len(pub.publishes))
	}
}

func TestConfigRetryWorker_Tick_MarksSentOverdueFailed(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&ConfigDelivery{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Now().UTC()
	sentAt := now.Add(-700 * time.Second)
	d := ConfigDelivery{
		MsgID:          "m1",
		TraceID:        "t1",
		DeviceCode:     "DEV-1",
		ConfigType:     "climate_profile",
		Action:         "update",
		EntityID:       1,
		EntityRev:      1,
		SchemaVersion:  1,
		IssuedAtMS:     uint64(now.UnixMilli()),
		TTLsec:         600,
		RequireAck:     true,
		RequestPayload: `{"schema_version":1,"msg_id":"m1"}`,
		Status:         ConfigDeliveryStatusSent,
		SentAt:         &sentAt,
	}
	if err := db.Create(&d).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	pub := &fakePublisher{connected: true}
	w := NewConfigRetryWorker(db, pub, nil)
	w.tick()

	var got ConfigDelivery
	if err := db.First(&got, d.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Status != ConfigDeliveryStatusFailed {
		t.Fatalf("expected FAILED, got %s", got.Status)
	}
	if got.NextRetryAt == nil {
		t.Fatalf("expected next_retry_at set")
	}
}
