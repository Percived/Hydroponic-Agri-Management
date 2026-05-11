package mqtt

import (
	"fmt"
	"log/slog"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
	"gorm.io/gorm"
)

type mqttPublisher interface {
	IsConnected() bool
	Publish(topic string, qos byte, retained bool, payload interface{}) mqttlib.Token
}

type ConfigRetryWorker struct {
	db       *gorm.DB
	repo     *ConfigDeliveryRepo
	client   mqttPublisher
	log      *slog.Logger
	interval time.Duration
	stopCh   chan struct{}
}

func NewConfigRetryWorker(db *gorm.DB, client mqttPublisher, log *slog.Logger) *ConfigRetryWorker {
	return &ConfigRetryWorker{
		db:       db,
		repo:     NewConfigDeliveryRepo(db),
		client:   client,
		log:      log,
		interval: 5 * time.Second,
		stopCh:   make(chan struct{}),
	}
}

func (w *ConfigRetryWorker) Start() {
	go w.loop()
}

func (w *ConfigRetryWorker) Stop() {
	close(w.stopCh)
}

func (w *ConfigRetryWorker) loop() {
	t := time.NewTicker(w.interval)
	defer t.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-t.C:
			w.tick()
		}
	}
}

func (w *ConfigRetryWorker) tick() {
	if w.client == nil || !w.client.IsConnected() {
		return
	}

	now := time.Now().UTC()

	sent, err := w.repo.ListSentCandidates(now.Add(-5*time.Second), 200)
	if err == nil {
		for _, d := range sent {
			if d.SentAt == nil {
				continue
			}
			deadline := d.SentAt.Add(time.Duration(d.TTLsec) * time.Second)
			if !now.After(deadline) {
				continue
			}
			next := now.Add(backoffDuration(d.RetryCount))
			_ = w.repo.MarkFailed(d.ID, "ACK_TIMEOUT", "ack timeout", &next)
		}
	}

	due, err := w.repo.ListFailedDue(now, 100)
	if err != nil {
		return
	}

	for _, d := range due {
		if d.RequestPayload == "" {
			next := now.Add(backoffDuration(d.RetryCount))
			_ = w.repo.MarkFailed(d.ID, "BAD_PAYLOAD", "empty request_payload", &next)
			continue
		}

		topic := fmt.Sprintf("%s/%s/%s/%s", TopicPrefix, d.DeviceCode, TopicCmdPrefix, ConfigPushTopic)
		token := w.client.Publish(topic, 1, false, d.RequestPayload)
		if token.Wait() && token.Error() != nil {
			next := now.Add(backoffDuration(d.RetryCount + 1))
			_ = w.db.Model(&ConfigDelivery{}).Where("id = ?", d.ID).Updates(map[string]interface{}{
				"status":             ConfigDeliveryStatusFailed,
				"retry_count":        d.RetryCount + 1,
				"next_retry_at":      next,
				"last_error_code":    "MQTT_PUBLISH_FAILED",
				"last_error_message": token.Error().Error(),
			}).Error
			continue
		}

		_ = w.db.Model(&ConfigDelivery{}).Where("id = ?", d.ID).Updates(map[string]interface{}{
			"status":        ConfigDeliveryStatusSent,
			"retry_count":   d.RetryCount + 1,
			"next_retry_at": nil,
			"sent_at":       now,
		}).Error
	}
}

func backoffDuration(retryCount int) time.Duration {
	if retryCount < 0 {
		retryCount = 0
	}
	if retryCount > 6 {
		retryCount = 6
	}
	d := 5 * time.Second
	for i := 0; i < retryCount; i++ {
		d *= 2
	}
	if d > 5*time.Minute {
		return 5 * time.Minute
	}
	return d
}
