package http

import (
	"log/slog"
	"net/http"

	"hydroponic-backend/internal/alert"
	"hydroponic-backend/internal/audit"
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/control"
	"hydroponic-backend/internal/device"
	"hydroponic-backend/internal/overview"
	"hydroponic-backend/internal/platform/config"
	"hydroponic-backend/internal/platform/di"
	"hydroponic-backend/internal/platform/response"
	"hydroponic-backend/internal/telemetry"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func NewRouter(cfg config.Config, log *slog.Logger, mysql *gorm.DB, influx influxdb2.Client, mqttClient mqtt.Client) *gin.Engine {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestID())
	r.Use(RequestLogger(log))
	r.Use(CORS())

	deps := di.Deps{
		Config: cfg,
		Log:    log,
		MySQL:  mysql,
		Influx: influx,
		MQTT:   mqttClient,
	}

	r.GET("/healthz", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ready"})
	})
	r.StaticFile("/openapi.yaml", "docs/specs/openapi.yaml")
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi.yaml")))

	api := r.Group("/api")
	overview.RegisterRoutes(api, deps)
	device.RegisterRoutes(api, deps)
	telemetry.RegisterRoutes(api, deps)
	control.RegisterRoutes(api, deps)
	alert.RegisterRoutes(api, deps)
	auth.RegisterRoutes(api, deps)
	audit.RegisterRoutes(api, deps)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"code": 10004, "message": "not_found"})
	})

	return r
}
