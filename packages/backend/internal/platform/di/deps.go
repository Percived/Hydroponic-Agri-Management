package di

import (
	"log/slog"

	"hydroponic-backend/internal/platform/config"
	"hydroponic-backend/internal/platform/event"

	"github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"gorm.io/gorm"
)

type Deps struct {
	Config   config.Config
	Log      *slog.Logger
	MySQL    *gorm.DB
	Influx   influxdb2.Client
	MQTT     mqtt.Client
	EventHub *event.Hub
}
