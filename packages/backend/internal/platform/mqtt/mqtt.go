package mqtt

import (
	"errors"
	"fmt"

	"hydroponic-backend/internal/platform/config"

	"github.com/eclipse/paho.mqtt.golang"
)

func NewClient(cfg config.MQTTConfig) (mqtt.Client, error) {
	if cfg.Broker == "" {
		return nil, errors.New("mqtt broker required")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
		opts.SetPassword(cfg.Password)
	}
	if cfg.ClientID != "" {
		opts.SetClientID(cfg.ClientID)
	}
	opts.SetAutoReconnect(true)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}
	return client, nil
}
