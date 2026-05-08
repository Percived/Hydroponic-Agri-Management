package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"hydroponic-backend/internal/platform/event"

	"gorm.io/gorm"
)

type AlertSubscriber struct {
	db  *gorm.DB
	hub *event.Hub
	log *slog.Logger
}

func NewAlertSubscriber(db *gorm.DB, hub *event.Hub, log *slog.Logger) *AlertSubscriber {
	s := &AlertSubscriber{db: db, hub: hub, log: log}
	go s.run()
	return s
}

func (s *AlertSubscriber) run() {
	sub := s.hub.Subscribe(func(e event.SSEEvent) bool {
		return e.Type == "alert:created"
	})
	defer s.hub.Unsubscribe(sub)

	for evt := range sub.Events {
		data, ok := evt.Data.(map[string]interface{})
		if !ok {
			continue
		}
		s.dispatch(data)
	}
}

func (s *AlertSubscriber) dispatch(data map[string]interface{}) {
	level, _ := data["level"].(string)
	message, _ := data["message"].(string)

	if level == "" {
		return
	}

	// Find all enabled channels whose min_alert_level allows this alert
	var channels []NotificationChannel
	s.db.Where("enabled = ?", true).
		Find(&channels)

	levelPriority := map[string]int{
		"INFO":     0,
		"WARN":     1,
		"CRITICAL": 2,
	}

	alertPriority := levelPriority[level]

	for _, ch := range channels {
		chPriority := levelPriority[ch.MinAlertLevel]
		if chPriority > alertPriority {
			continue
		}

		switch ch.ChannelType {
		case ChannelWebhook:
			s.sendWebhook(ch, data)
		case ChannelEmail, ChannelSMS, ChannelInApp:
			s.log.Info("notification dispatched",
				"channel_type", ch.ChannelType,
				"channel_id", ch.ID,
				"alert_level", level,
				"alert_message", message)
		}
	}
}

func (s *AlertSubscriber) sendWebhook(ch NotificationChannel, data map[string]interface{}) {
	var cfg struct {
		URL    string `json:"url"`
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(ch.Config, &cfg); err != nil || cfg.URL == "" {
		return
	}

	body, _ := json.Marshal(map[string]interface{}{
		"type":       "alert",
		"channel_id": ch.ID,
		"alert":      data,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})

	req, err := http.NewRequest("POST", cfg.URL, bytes.NewReader(body))
	if err != nil {
		s.log.Error("notification: webhook request failed", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	if cfg.Secret != "" {
		mac := hmac.New(sha256.New, []byte(cfg.Secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Signature", sig)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.log.Error("notification: webhook send failed", "error", err)
		return
	}
	defer resp.Body.Close()

	s.log.Info("notification: webhook sent",
		"channel_id", ch.ID,
		"status", resp.StatusCode)
}
