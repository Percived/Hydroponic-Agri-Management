package influx

import (
	"errors"

	"hydroponic-backend/internal/platform/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func NewClient(cfg config.InfluxConfig) (influxdb2.Client, error) {
	if cfg.URL == "" || cfg.Token == "" {
		return nil, errors.New("influx url/token required")
	}
	return influxdb2.NewClient(cfg.URL, cfg.Token), nil
}
