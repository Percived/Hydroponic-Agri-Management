package policy

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/command"
	"hydroponic-backend/internal/device"
	"hydroponic-backend/internal/platform/event"

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

type fakeMQTTClient struct {
	connected   bool
	publishErrs []error
	publishes   []struct {
		topic   string
		payload interface{}
	}
}

func (c *fakeMQTTClient) IsConnected() bool      { return c.connected }
func (c *fakeMQTTClient) IsConnectionOpen() bool { return c.connected }
func (c *fakeMQTTClient) Connect() mqttlib.Token { return &fakeToken{} }
func (c *fakeMQTTClient) Disconnect(_ uint)      {}
func (c *fakeMQTTClient) Subscribe(string, byte, mqttlib.MessageHandler) mqttlib.Token {
	return &fakeToken{}
}
func (c *fakeMQTTClient) SubscribeMultiple(map[string]byte, mqttlib.MessageHandler) mqttlib.Token {
	return &fakeToken{}
}
func (c *fakeMQTTClient) Unsubscribe(...string) mqttlib.Token     { return &fakeToken{} }
func (c *fakeMQTTClient) AddRoute(string, mqttlib.MessageHandler) {}
func (c *fakeMQTTClient) OptionsReader() mqttlib.ClientOptionsReader {
	return mqttlib.ClientOptionsReader{}
}
func (c *fakeMQTTClient) Publish(topic string, _ byte, _ bool, payload interface{}) mqttlib.Token {
	c.publishes = append(c.publishes, struct {
		topic   string
		payload interface{}
	}{topic: topic, payload: payload})
	idx := len(c.publishes) - 1
	if idx < len(c.publishErrs) && c.publishErrs[idx] != nil {
		return &fakeToken{err: c.publishErrs[idx]}
	}
	return &fakeToken{}
}

func TestScheduleOnceExecutesOnlyOneSlot(t *testing.T) {
	db := openPolicySchedulerTestDB(t)

	policyID, channelID := seedPublishedSchedulePolicy(t, db, func(p *ControlPolicy) {
		now := time.Now().UTC()
		p.PolicyType = "SCHEDULE"
		p.ScheduleMode = stringPtr("ONCE")
		p.RunOnceAt = timePtr(now.Add(-10 * time.Second))
		p.Timezone = "Asia/Shanghai"
		p.PublishedAt = timePtr(now.Add(-1 * time.Minute))
	}, true)

	client := &fakeMQTTClient{connected: true}
	s := NewScheduler(db, event.NewHub(), client, testLogger())

	s.evaluateScheduledPolicies()
	s.evaluateScheduledPolicies()

	var commands []command.ControlCommand
	if err := db.Order("id asc").Find(&commands).Error; err != nil {
		t.Fatalf("load commands: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}

	var executions []PolicyExecution
	if err := db.Where("policy_id = ?", policyID).Order("id asc").Find(&executions).Error; err != nil {
		t.Fatalf("load executions: %v", err)
	}
	if len(executions) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(executions))
	}
	if executions[0].Decision != "EXECUTED" {
		t.Fatalf("expected execution decision EXECUTED, got %s", executions[0].Decision)
	}

	var policy ControlPolicy
	if err := db.First(&policy, policyID).Error; err != nil {
		t.Fatalf("reload policy: %v", err)
	}
	if policy.LastScheduledFor == nil || !policy.LastScheduledFor.Equal(*policy.RunOnceAt) {
		t.Fatalf("expected last_scheduled_for to equal run_once_at")
	}

	if len(client.publishes) != 1 {
		t.Fatalf("expected 1 mqtt publish, got %d", len(client.publishes))
	}

	if channelID == 0 {
		t.Fatalf("expected seeded actuator channel")
	}
}

func TestScheduleCommandUsesPolicyPublisherAsCreator(t *testing.T) {
	db := openPolicySchedulerTestDB(t)
	creator := auth.User{
		Username:     "creator",
		PasswordHash: "test-hash",
		Status:       auth.UserStatusEnabled,
	}
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create creator: %v", err)
	}
	publisher := auth.User{
		Username:     "publisher",
		PasswordHash: "test-hash",
		Status:       auth.UserStatusEnabled,
	}
	if err := db.Create(&publisher).Error; err != nil {
		t.Fatalf("create publisher: %v", err)
	}

	policyID, _ := seedPublishedSchedulePolicy(t, db, func(p *ControlPolicy) {
		now := time.Now().UTC()
		p.PolicyType = "SCHEDULE"
		p.ScheduleMode = stringPtr("ONCE")
		p.RunOnceAt = timePtr(now.Add(-10 * time.Second))
		p.Timezone = "UTC"
		p.CreatedBy = uint64Ptr(creator.ID)
		p.PublishedBy = uint64Ptr(publisher.ID)
		p.PublishedAt = timePtr(now.Add(-1 * time.Minute))
	}, true)

	s := NewScheduler(db, event.NewHub(), &fakeMQTTClient{connected: true}, testLogger())
	s.evaluateScheduledPolicies()

	var commands []command.ControlCommand
	if err := db.Order("id asc").Find(&commands).Error; err != nil {
		t.Fatalf("load commands: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}
	if commands[0].CreatedBy != publisher.ID {
		t.Fatalf("expected command created_by to use policy publisher, got %d", commands[0].CreatedBy)
	}

	var executions []PolicyExecution
	if err := db.Where("policy_id = ?", policyID).Find(&executions).Error; err != nil {
		t.Fatalf("load executions: %v", err)
	}
	if len(executions) != 1 || executions[0].Decision != "EXECUTED" {
		t.Fatalf("expected schedule execution to succeed, got %+v", executions)
	}
}

func TestSchedulePolicyWithoutPlanWritesSkippedExecution(t *testing.T) {
	db := openPolicySchedulerTestDB(t)

	policyID, _ := seedPublishedSchedulePolicy(t, db, func(p *ControlPolicy) {
		now := time.Now().UTC()
		p.PolicyType = "SCHEDULE"
		p.ScheduleMode = nil
		p.Timezone = "Asia/Shanghai"
		p.PublishedAt = timePtr(now.Add(-1 * time.Minute))
	}, true)

	s := NewScheduler(db, event.NewHub(), &fakeMQTTClient{connected: true}, testLogger())
	s.evaluateScheduledPolicies()

	var executions []PolicyExecution
	if err := db.Where("policy_id = ?", policyID).Find(&executions).Error; err != nil {
		t.Fatalf("load executions: %v", err)
	}
	if len(executions) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(executions))
	}
	if executions[0].Decision != "SKIPPED" || executions[0].DecisionReason != "schedule_not_configured" {
		t.Fatalf("expected schedule_not_configured skip, got %s/%s", executions[0].Decision, executions[0].DecisionReason)
	}
}

func TestScheduleDailyRequiresPublishedPolicy(t *testing.T) {
	db := openPolicySchedulerTestDB(t)

	_, _ = seedPublishedSchedulePolicy(t, db, func(p *ControlPolicy) {
		now := time.Now().UTC()
		p.PolicyType = "SCHEDULE"
		p.ScheduleMode = stringPtr("DAILY")
		p.TimeOfDay = stringPtr(now.Format("15:04:05"))
		p.Timezone = "Asia/Shanghai"
		p.PublishedAt = nil
	}, true)

	s := NewScheduler(db, event.NewHub(), &fakeMQTTClient{connected: true}, testLogger())
	s.evaluateScheduledPolicies()

	var execCount int64
	if err := db.Model(&PolicyExecution{}).Count(&execCount).Error; err != nil {
		t.Fatalf("count executions: %v", err)
	}
	if execCount != 0 {
		t.Fatalf("expected unpublished schedule policy to be ignored, got %d executions", execCount)
	}
}

func TestScheduleDailyExecutesDueSlot(t *testing.T) {
	db := openPolicySchedulerTestDB(t)

	policyID, _ := seedPublishedSchedulePolicy(t, db, func(p *ControlPolicy) {
		now := time.Now().UTC()
		p.PolicyType = "SCHEDULE"
		p.ScheduleMode = stringPtr("DAILY")
		p.TimeOfDay = stringPtr(now.Format("15:04:05"))
		p.Timezone = "UTC"
		p.PublishedAt = timePtr(now.Add(-1 * time.Minute))
	}, true)

	s := NewScheduler(db, event.NewHub(), &fakeMQTTClient{connected: true}, testLogger())
	s.evaluateScheduledPolicies()

	var execs []PolicyExecution
	if err := db.Where("policy_id = ?", policyID).Find(&execs).Error; err != nil {
		t.Fatalf("load executions: %v", err)
	}
	if len(execs) != 1 || execs[0].Decision != "EXECUTED" {
		t.Fatalf("expected daily schedule to execute once, got %+v", execs)
	}
}

func TestScheduleOnceStillExecutesAfterLongerDelay(t *testing.T) {
	db := openPolicySchedulerTestDB(t)

	policyID, _ := seedPublishedSchedulePolicy(t, db, func(p *ControlPolicy) {
		now := time.Now().UTC()
		p.PolicyType = "SCHEDULE"
		p.ScheduleMode = stringPtr("ONCE")
		p.RunOnceAt = timePtr(now.Add(-5 * time.Minute))
		p.Timezone = "UTC"
		p.PublishedAt = timePtr(now.Add(-10 * time.Minute))
	}, true)

	s := NewScheduler(db, event.NewHub(), &fakeMQTTClient{connected: true}, testLogger())
	s.evaluateScheduledPolicies()

	var execs []PolicyExecution
	if err := db.Where("policy_id = ?", policyID).Find(&execs).Error; err != nil {
		t.Fatalf("load executions: %v", err)
	}
	if len(execs) != 1 || execs[0].Decision != "EXECUTED" {
		t.Fatalf("expected delayed ONCE schedule to still execute, got %+v", execs)
	}
}

func TestScheduleWeeklyExecutesLatestMatchedSlot(t *testing.T) {
	db := openPolicySchedulerTestDB(t)

	policyID, _ := seedPublishedSchedulePolicy(t, db, func(p *ControlPolicy) {
		now := time.Now().UTC()
		p.PolicyType = "SCHEDULE"
		p.ScheduleMode = stringPtr("WEEKLY")
		p.TimeOfDay = stringPtr(now.Format("15:04:05"))
		p.Timezone = "UTC"
		mask := weeklyMaskForNonMatchingWeekday(now.Weekday())
		p.WeekdaysMask = &mask
		p.PublishedAt = timePtr(now.Add(-1 * time.Minute))
	}, true)

	s := NewScheduler(db, event.NewHub(), &fakeMQTTClient{connected: true}, testLogger())
	s.evaluateScheduledPolicies()

	var execs []PolicyExecution
	if err := db.Where("policy_id = ?", policyID).Find(&execs).Error; err != nil {
		t.Fatalf("load executions: %v", err)
	}
	if len(execs) != 1 || execs[0].Decision != "EXECUTED" {
		t.Fatalf("expected weekly schedule to execute latest matched slot, got %+v", execs)
	}
}

func TestScheduleFailureDoesNotReopenClaimedSlot(t *testing.T) {
	db := openPolicySchedulerTestDB(t)

	policyID, channelID := seedPublishedSchedulePolicy(t, db, func(p *ControlPolicy) {
		now := time.Now().UTC()
		p.PolicyType = "SCHEDULE"
		p.ScheduleMode = stringPtr("ONCE")
		p.RunOnceAt = timePtr(now.Add(-10 * time.Second))
		p.Timezone = "UTC"
		p.PublishedAt = timePtr(now.Add(-1 * time.Minute))
	}, true)

	secondTarget := PolicyTarget{
		PolicyID:          policyID,
		ActuatorChannelID: channelID,
		CommandType:       "SWITCH",
		CommandPayload:    `{"state":"OFF"}`,
		ExecutionOrder:    2,
		Enabled:           true,
	}
	if err := db.Create(&secondTarget).Error; err != nil {
		t.Fatalf("create second target: %v", err)
	}

	client := &fakeMQTTClient{
		connected:   true,
		publishErrs: []error{nil, io.ErrUnexpectedEOF},
	}
	s := NewScheduler(db, event.NewHub(), client, testLogger())

	s.evaluateScheduledPolicies()

	var firstRunCommands []command.ControlCommand
	if err := db.Order("id asc").Find(&firstRunCommands).Error; err != nil {
		t.Fatalf("load first run commands: %v", err)
	}
	if len(firstRunCommands) != 2 {
		t.Fatalf("expected 2 commands after partial failure, got %d", len(firstRunCommands))
	}

	var policy ControlPolicy
	if err := db.First(&policy, policyID).Error; err != nil {
		t.Fatalf("reload policy after first run: %v", err)
	}
	if policy.LastScheduledFor == nil || policy.RunOnceAt == nil || !policy.LastScheduledFor.Equal(*policy.RunOnceAt) {
		t.Fatalf("expected claimed slot to stay recorded after failure")
	}

	client.publishErrs = nil
	s.evaluateScheduledPolicies()

	var allCommands []command.ControlCommand
	if err := db.Order("id asc").Find(&allCommands).Error; err != nil {
		t.Fatalf("load all commands: %v", err)
	}
	if len(allCommands) != 2 {
		t.Fatalf("expected no duplicate commands after retry scan, got %d", len(allCommands))
	}

	var executions []PolicyExecution
	if err := db.Where("policy_id = ?", policyID).Order("id asc").Find(&executions).Error; err != nil {
		t.Fatalf("load executions: %v", err)
	}
	if len(executions) != 1 || executions[0].Decision != "FAILED" {
		t.Fatalf("expected single failed execution, got %+v", executions)
	}
}

func openPolicySchedulerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&auth.User{},
		&ControlPolicy{},
		&PolicyCondition{},
		&PolicyTarget{},
		&PolicyExecution{},
		&command.ControlCommand{},
		&device.ActuatorDevice{},
		&device.ActuatorChannel{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	admin := auth.User{
		Username:     "admin",
		PasswordHash: "test-hash",
		Status:       auth.UserStatusEnabled,
	}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return db
}

func seedPublishedSchedulePolicy(t *testing.T, db *gorm.DB, mutate func(*ControlPolicy), withTarget bool) (uint64, uint64) {
	t.Helper()

	actuator := device.ActuatorDevice{
		DeviceCode:   "ACT-001",
		Name:         "actuator",
		GreenhouseID: 1,
	}
	if err := db.Create(&actuator).Error; err != nil {
		t.Fatalf("create actuator: %v", err)
	}

	channel := device.ActuatorChannel{
		ActuatorDeviceID: actuator.ID,
		ChannelCode:      "pump-1",
		ActuatorType:     device.ActuatorTypePump,
		Enabled:          true,
	}
	if err := db.Create(&channel).Error; err != nil {
		t.Fatalf("create channel: %v", err)
	}

	policy := ControlPolicy{
		PolicyCode:   "POL-SCHED-001",
		Name:         "schedule policy",
		PolicyType:   "SCHEDULE",
		GreenhouseID: 1,
		Enabled:      true,
	}
	if mutate != nil {
		mutate(&policy)
	}
	if err := db.Create(&policy).Error; err != nil {
		t.Fatalf("create policy: %v", err)
	}

	if withTarget {
		target := PolicyTarget{
			PolicyID:          policy.ID,
			ActuatorChannelID: channel.ID,
			CommandType:       "SWITCH",
			CommandPayload:    `{"state":"ON"}`,
			ExecutionOrder:    1,
			Enabled:           true,
		}
		if err := db.Create(&target).Error; err != nil {
			t.Fatalf("create target: %v", err)
		}
	}

	return policy.ID, channel.ID
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func stringPtr(v string) *string { return &v }

func timePtr(v time.Time) *time.Time { return &v }

func uint64Ptr(v uint64) *uint64 { return &v }

func weeklyMaskForNonMatchingWeekday(weekday time.Weekday) uint8 {
	switch weekday {
	case time.Monday:
		return 1 << 1
	case time.Tuesday:
		return 1 << 2
	case time.Wednesday:
		return 1 << 3
	case time.Thursday:
		return 1 << 4
	case time.Friday:
		return 1 << 5
	case time.Saturday:
		return 1 << 6
	default:
		return 1 << 0
	}
}
