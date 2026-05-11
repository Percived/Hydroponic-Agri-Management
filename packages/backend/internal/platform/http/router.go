package http

import (
	"log/slog"
	"net/http"

	"hydroponic-backend/internal/alert"
	"hydroponic-backend/internal/audit"
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/climate"
	"hydroponic-backend/internal/command"
	"hydroponic-backend/internal/crop"
	"hydroponic-backend/internal/device"
	"hydroponic-backend/internal/energy"
	"hydroponic-backend/internal/greenhouse"
	"hydroponic-backend/internal/metric"
	"hydroponic-backend/internal/notification"
	"hydroponic-backend/internal/nutrient"
	"hydroponic-backend/internal/overview"
	"hydroponic-backend/internal/pest"
	"hydroponic-backend/internal/platform/config"
	"hydroponic-backend/internal/platform/di"
	"hydroponic-backend/internal/platform/event"
	"hydroponic-backend/internal/platform/mqtt"
	"hydroponic-backend/internal/platform/response"
	"hydroponic-backend/internal/policy"
	"hydroponic-backend/internal/recipe"
	"hydroponic-backend/internal/review"
	"hydroponic-backend/internal/telemetry"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func NewRouter(cfg config.Config, log *slog.Logger, mysql *gorm.DB, influx influxdb2.Client, mqttClient mqttlib.Client) *gin.Engine {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestID())
	r.Use(RequestLogger(log))
	r.Use(CORS())

	hub := event.NewHub()
	cache := telemetry.NewSensorStatusCache()

	deps := di.Deps{
		Config:   cfg,
		Log:      log,
		MySQL:    mysql,
		Influx:   influx,
		MQTT:     mqttClient,
		EventHub: hub,
	}

	// Start MQTT Ingress Service
	ingress := mqtt.NewIngressService(mysql, influx, cfg.Influx, hub, cache, mqttClient, log)
	if err := ingress.Start(); err != nil {
		log.Warn("mqtt ingress start failed", "error", err)
	}
	configRetryWorker := mqtt.NewConfigRetryWorker(mysql, mqttClient, log)
	configRetryWorker.Start()

	r.GET("/healthz", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ready"})
	})
	r.StaticFile("/openapi.yaml", "docs/specs/openapi.yaml")
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi.yaml")))

	api := r.Group("/api")

	// Core modules
	auth.RegisterRoutes(api, deps)
	overview.RegisterRoutes(api, deps)

	// Organization & facilities
	greenhouse.RegisterRoutes(api, deps)

	// Device management
	device.RegisterRoutes(api, deps)

	// Metric definitions
	metric.RegisterRoutes(api, deps)

	// Telemetry (pass cache separately to avoid import cycle)
	telemetry.RegisterRoutesWithCache(api, deps, cache)

	// Nutrient management (DWC core)
	nutrient.RegisterRoutes(api, deps)

	// Crop & batch management
	crop.RegisterRoutes(api, deps)

	// Recipe management
	recipe.RegisterRoutes(api, deps)

	// Climate control (multi-stage)
	climate.RegisterRoutes(api, deps)

	// Control policies
	policy.RegisterRoutes(api, deps)

	// Command dispatch
	command.RegisterRoutes(api, deps)

	// Alerts
	alert.RegisterRoutes(api, deps)

	// SSE real-time subscriptions
	api.GET("/alerts/subscribe", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), event.SSEHandler(deps.EventHub, "alert:created"))
	api.GET("/telemetry/subscribe", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), event.SSEHandler(deps.EventHub, "telemetry:received"))
	api.GET("/devices/subscribe", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), event.SSEHandler(deps.EventHub, "device:status"))
	api.GET("/commands/subscribe", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), event.SSEHandlerMulti(deps.EventHub, []string{"command:dispatched", "command:acked"}))

	configDeliveryHandler := mqtt.NewConfigDeliveryHandler(deps.MySQL)
	api.GET("/config-deliveries", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), configDeliveryHandler.List)
	api.GET("/config-deliveries/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), configDeliveryHandler.Get)

	// Energy consumption
	energy.RegisterRoutes(api, deps)

	// Pest & disease
	pest.RegisterRoutes(api, deps)

	// Batch review
	review.RegisterRoutes(api, deps)

	// Audit & notification
	audit.RegisterRoutes(api, deps)
	notification.RegisterRoutes(api, deps)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"code": 10004, "message": "not_found"})
	})

	return r
}
