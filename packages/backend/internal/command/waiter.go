package command

import (
	"fmt"
	"sync"
	"time"

	"hydroponic-backend/internal/platform/event"
)

type CommandReceipt struct {
	CommandID  uint64
	AckCode    string
	AckMessage string
	Timestamp  time.Time
}

type CommandWaiter struct {
	mu      sync.Mutex
	waiters map[uint64]chan CommandReceipt
	hub     *event.Hub
	timeout time.Duration
	stopCh  chan struct{}
}

func NewCommandWaiter(hub *event.Hub) *CommandWaiter {
	w := &CommandWaiter{
		waiters: make(map[uint64]chan CommandReceipt),
		hub:     hub,
		timeout: 10 * time.Second,
		stopCh:  make(chan struct{}),
	}
	go w.listenAcks()
	return w
}

func (w *CommandWaiter) listenAcks() {
	sub := w.hub.Subscribe(func(e event.SSEEvent) bool {
		return e.Type == "command:acked"
	})
	defer w.hub.Unsubscribe(sub)

	for {
		select {
		case <-w.stopCh:
			return
		case evt, ok := <-sub.Events:
			if !ok {
				return
			}
			switch data := evt.Data.(type) {
			case event.CommandAckData:
				if data.CommandID == 0 {
					continue
				}
				w.Notify(data.CommandID, CommandReceipt{
					CommandID:  data.CommandID,
					AckCode:    data.AckCode,
					AckMessage: data.AckMessage,
					Timestamp:  time.Now(),
				})
			case *event.CommandAckData:
				if data == nil || data.CommandID == 0 {
					continue
				}
				w.Notify(data.CommandID, CommandReceipt{
					CommandID:  data.CommandID,
					AckCode:    data.AckCode,
					AckMessage: data.AckMessage,
					Timestamp:  time.Now(),
				})
			case map[string]interface{}:
				cmdID := uint64(0)
				switch v := data["command_id"].(type) {
				case float64:
					cmdID = uint64(v)
				case uint64:
					cmdID = v
				case int:
					if v > 0 {
						cmdID = uint64(v)
					}
				case int64:
					if v > 0 {
						cmdID = uint64(v)
					}
				}
				if cmdID == 0 {
					continue
				}
				ackCode, _ := data["ack_code"].(string)
				ackMsg, _ := data["ack_message"].(string)
				w.Notify(cmdID, CommandReceipt{
					CommandID:  cmdID,
					AckCode:    ackCode,
					AckMessage: ackMsg,
					Timestamp:  time.Now(),
				})
			}
		}
	}
}

func (w *CommandWaiter) Register(commandID uint64) chan CommandReceipt {
	w.mu.Lock()
	defer w.mu.Unlock()
	ch := make(chan CommandReceipt, 1)
	w.waiters[commandID] = ch
	return ch
}

func (w *CommandWaiter) Notify(commandID uint64, receipt CommandReceipt) {
	w.mu.Lock()
	ch, ok := w.waiters[commandID]
	if ok {
		delete(w.waiters, commandID)
	}
	w.mu.Unlock()
	if ok {
		select {
		case ch <- receipt:
		default:
		}
	}
}

func (w *CommandWaiter) Wait(commandID uint64, timeout time.Duration) (*CommandReceipt, error) {
	w.mu.Lock()
	ch, ok := w.waiters[commandID]
	w.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("no waiter registered for command %d", commandID)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case receipt := <-ch:
		return &receipt, nil
	case <-timer.C:
		w.mu.Lock()
		delete(w.waiters, commandID)
		w.mu.Unlock()
		return nil, fmt.Errorf("command %d timed out", commandID)
	}
}

func (w *CommandWaiter) Close() {
	close(w.stopCh)
}
