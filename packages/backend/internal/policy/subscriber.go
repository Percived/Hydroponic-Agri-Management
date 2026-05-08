package policy

import (
	"log/slog"

	"hydroponic-backend/internal/platform/event"

	"gorm.io/gorm"
)

type TelemetrySubscriber struct {
	db  *gorm.DB
	hub *event.Hub
	log *slog.Logger
}

func NewTelemetrySubscriber(db *gorm.DB, hub *event.Hub, log *slog.Logger) *TelemetrySubscriber {
	s := &TelemetrySubscriber{db: db, hub: hub, log: log}
	go s.run()
	return s
}

func (s *TelemetrySubscriber) run() {
	sub := s.hub.Subscribe(func(e event.SSEEvent) bool {
		return e.Type == "telemetry:received"
	})
	defer s.hub.Unsubscribe(sub)

	for evt := range sub.Events {
		data, ok := evt.Data.(map[string]interface{})
		if !ok {
			continue
		}

		s.log.Debug("policy: received telemetry event", "data", data)

		// Phase 3 will implement full policy evaluation here
		// - Match enabled THRESHOLD policies
		// - Evaluate conditions against telemetry values
		// - Execute targets (command dispatch) or create alerts
	}
}
