package mqtt

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConfigPushTopic is the topic suffix for config push commands.
const ConfigPushTopic = "config"

// ConfigPusher handles pushing configuration to devices via MQTT.
type ConfigPusher struct {
	db     *gorm.DB
	client mqttlib.Client
	log    *slog.Logger
}

// NewConfigPusher creates a new ConfigPusher.
func NewConfigPusher(db *gorm.DB, client mqttlib.Client, log *slog.Logger) *ConfigPusher {
	return &ConfigPusher{db: db, client: client, log: log}
}

// ConfigPushPayload is the payload format for config push commands.
// Fields can be extended as firmware protocol is defined.
type ConfigPushPayload struct {
	ConfigType string      `json:"config_type"` // climate_profile, control_policy, nutrient_target, crop_batch
	Action     string      `json:"action"`      // create, update, delete
	EntityID   uint64      `json:"entity_id"`
	Payload    interface{} `json:"payload"`
}

type ConfigPushPayloadV1 struct {
	SchemaVersion int         `json:"schema_version"`
	MsgID         string      `json:"msg_id"`
	TraceID       string      `json:"trace_id"`
	ConfigType    string      `json:"config_type"`
	Action        string      `json:"action"`
	EntityID      uint64      `json:"entity_id"`
	EntityRev     uint64      `json:"entity_rev"`
	IssuedAtMS    uint64      `json:"issued_at_ms"`
	TTLsec        int         `json:"ttl_sec"`
	RequireAck    bool        `json:"require_ack"`
	Payload       interface{} `json:"payload"`
}

// PushToDevice sends a config update to a specific device via MQTT.
func (p *ConfigPusher) PushToDevice(deviceCode string, cfgType string, action string, entityID uint64, payload interface{}) error {
	now := time.Now().UTC()
	issuedAtMS := uint64(now.UnixMilli())
	msgID := uuid.NewString()
	traceID := uuid.NewString()
	ttlSec := 600

	repo := NewConfigDeliveryRepo(p.db)
	var delivery ConfigDelivery
	var data []byte
	if err := p.db.Transaction(func(tx *gorm.DB) error {
		rev, err := repo.AllocateNextRev(tx, deviceCode, cfgType, entityID)
		if err != nil {
			return err
		}

		msg := ConfigPushPayloadV1{
			SchemaVersion: 1,
			MsgID:         msgID,
			TraceID:       traceID,
			ConfigType:    cfgType,
			Action:        action,
			EntityID:      entityID,
			EntityRev:     rev,
			IssuedAtMS:    issuedAtMS,
			TTLsec:        ttlSec,
			RequireAck:    true,
			Payload:       payload,
		}

		out, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("marshal config push v1: %w", err)
		}
		data = out

		delivery = ConfigDelivery{
			MsgID:          msgID,
			TraceID:        traceID,
			DeviceCode:     deviceCode,
			ConfigType:     cfgType,
			Action:         action,
			EntityID:       entityID,
			EntityRev:      rev,
			SchemaVersion:  1,
			IssuedAtMS:     issuedAtMS,
			TTLsec:         ttlSec,
			RequireAck:     true,
			RequestPayload: string(out),
			Status:         ConfigDeliveryStatusPending,
		}
		return repo.Create(tx, &delivery)
	}); err != nil {
		return err
	}

	topic := fmt.Sprintf("%s/%s/%s/%s", TopicPrefix, deviceCode, TopicCmdPrefix, ConfigPushTopic)
	if p.client == nil || !p.client.IsConnected() {
		next := now.Add(5 * time.Second)
		_ = repo.MarkFailed(delivery.ID, "MQTT_NOT_CONNECTED", "mqtt not connected", &next)
		return fmt.Errorf("mqtt not connected")
	}

	token := p.client.Publish(topic, 1, false, data)
	if token.Wait() && token.Error() != nil {
		next := now.Add(5 * time.Second)
		_ = repo.MarkFailed(delivery.ID, "MQTT_PUBLISH_FAILED", token.Error().Error(), &next)
		return fmt.Errorf("publish config push: %w", token.Error())
	}

	_ = repo.MarkSent(delivery.ID, now)

	p.log.Info("config pusher: pushed to device",
		"device", deviceCode, "type", cfgType, "action", action, "entity_id", entityID, "msg_id", msgID)
	return nil
}

// PushToActuatorChannel sends a config update to the device owning an actuator channel.
func (p *ConfigPusher) PushToActuatorChannel(actuatorChannelID uint64, cfgType string, action string, entityID uint64, payload interface{}) error {
	deviceCode, err := p.lookupActuatorDeviceCode(actuatorChannelID)
	if err != nil {
		return fmt.Errorf("lookup actuator device: %w", err)
	}
	return p.PushToDevice(deviceCode, cfgType, action, entityID, payload)
}

func (p *ConfigPusher) lookupActuatorDeviceCode(actuatorChannelID uint64) (string, error) {
	var result struct {
		DeviceCode string
	}
	err := p.db.Table("actuator_channels").
		Select("actuator_devices.device_code").
		Joins("JOIN actuator_devices ON actuator_devices.id = actuator_channels.actuator_device_id").
		Where("actuator_channels.id = ?", actuatorChannelID).
		Scan(&result).Error
	if err != nil {
		return "", err
	}
	if result.DeviceCode == "" {
		return "", fmt.Errorf("device not found for channel %d", actuatorChannelID)
	}
	return result.DeviceCode, nil
}
