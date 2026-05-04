package main

import (
	"fmt"
	"os"

	"hydroponic-backend/internal/platform/config"
	"hydroponic-backend/internal/platform/db"
	"hydroponic-backend/internal/platform/http"
	"hydroponic-backend/internal/platform/influx"
	"hydroponic-backend/internal/platform/logger"
	"hydroponic-backend/internal/platform/mqtt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load failed: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log)
	log.Info("config loaded")

	mysqlDB, err := db.NewMySQL(cfg.MySQL)
	if err != nil {
		log.Error("mysql init failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.CloseMySQL(mysqlDB); err != nil {
			log.Warn("mysql close failed", "error", err)
		}
	}()

	influxClient, err := influx.NewClient(cfg.Influx)
	if err != nil {
		log.Error("influx init failed", "error", err)
		os.Exit(1)
	}
	defer influxClient.Close()

	mqttClient, err := mqtt.NewClient(cfg.MQTT)
	if err != nil {
		log.Error("mqtt init failed", "error", err)
		os.Exit(1)
	}
	defer mqttClient.Disconnect(250)

	router := http.NewRouter(cfg, log, mysqlDB, influxClient, mqttClient)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("server starting", "addr", addr)

	if err := router.Run(addr); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
