package command

import (
	"testing"
	"time"

	"hydroponic-backend/internal/platform/event"
)

func TestCommandWaiter_ReceivesTypedAckData(t *testing.T) {
	hub := event.NewHub()
	w := NewCommandWaiter(hub)
	defer w.Close()

	const cmdID uint64 = 123
	w.Register(cmdID)

	time.Sleep(20 * time.Millisecond)

	hub.Publish(event.SSEEvent{
		Type: "command:acked",
		Data: event.CommandAckData{
			CommandID:  cmdID,
			AckCode:    "OK",
			AckMessage: "ok",
		},
	})

	receipt, err := w.Wait(cmdID, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if receipt.CommandID != cmdID {
		t.Fatalf("expected command_id %d, got %d", cmdID, receipt.CommandID)
	}
	if receipt.AckCode != "OK" {
		t.Fatalf("expected ack_code OK, got %q", receipt.AckCode)
	}
}
