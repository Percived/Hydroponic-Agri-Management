package mqtt

import (
	"encoding/json"
	"fmt"
	"log/slog"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
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

// PushToDevice sends a config update to a specific device via MQTT.
func (p *ConfigPusher) PushToDevice(deviceCode string, cfgType string, action string, entityID uint64, payload interface{}) error {
	if p.client == nil || !p.client.IsConnected() {
		p.log.Warn("config pusher: mqtt not connected, skipping push",
			"device", deviceCode, "type", cfgType, "action", action)
		return nil
	}

	msg := ConfigPushPayload{
		ConfigType: cfgType,
		Action:     action,
		EntityID:   entityID,
		Payload:    payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal config push: %w", err)
	}

	topic := fmt.Sprintf("%s/%s/%s/%s", TopicPrefix, deviceCode, TopicCmdPrefix, ConfigPushTopic)
	token := p.client.Publish(topic, 1, false, data)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("publish config push: %w", token.Error())
	}

	p.log.Info("config pusher: pushed to device",
		"device", deviceCode, "type", cfgType, "action", action, "entity_id", entityID)
	return nil
}

// PushToActuatorChannel sends a config update to the device owning an actuator channel.
func (p *ConfigPusher) PushToActuatorChannel(actuatorChannelID uint64, cfgType string, action string, entityID uint64, payload interface{}) error {
	deviceCode, err := p.lookupActuatorDeviceCode(actuatorChannelID)
	if err != nil {
		p.log.Warn("config pusher: cannot find device for actuator channel",
			"channel_id", actuatorChannelID, "error", err)
		return nil
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
