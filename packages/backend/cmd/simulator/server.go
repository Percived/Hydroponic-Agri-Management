package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ──────────────────── 模拟实例 ────────────────────

// simulation wraps a running simulation.
type simulation struct {
	ctx    context.Context
	cancel context.CancelFunc

	sensor   *sensorSim
	actuator *actuatorSim
	env      *Environment
	mqtt     *mqttManager
	rng      *rand.Rand

	startTime      time.Time
	telemetryCount int64
	cmdACKCount    int64
}

// ──────────────────── HTTP 服务 ────────────────────

// SimServer is the HTTP server for the simulator control panel.
type SimServer struct {
	engine *gin.Engine
	hub    *SSEHub

	mu  sync.Mutex
	sim *simulation // nil when not running
}

// NewSimServer creates a new SimServer.
func NewSimServer() *SimServer {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	s := &SimServer{
		engine: engine,
		hub:    NewSSEHub(),
	}

	// CORS middleware — echo Origin to allow file:// usage
	engine.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Routes
	engine.GET("/status", s.handleStatus)
	engine.POST("/start", s.handleStart)
	engine.POST("/stop", s.handleStop)
	engine.POST("/trigger-telemetry", s.handleTriggerTelemetry)
	engine.GET("/events", s.handleEvents)

	return s
}

// Run starts the HTTP server on the given port.
func (s *SimServer) Run(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("🌐 模拟器 HTTP 服务启动: http://127.0.0.1%s", addr)
	return s.engine.Run(addr)
}

// ──────────────────── GET /status ────────────────────

func (s *SimServer) handleStatus(c *gin.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sim == nil {
		c.JSON(http.StatusOK, statusResponse{Running: false})
		return
	}

	sim := s.sim
	resp := statusResponse{
		Running:        true,
		UptimeSec:      int64(time.Since(sim.startTime).Seconds()),
		TelemetryCount: sim.sensor.totalTelemetry,
	}
	if sim.sensor != nil {
		resp.SensorDevice = sim.sensor.deviceCode
		for _, ch := range sim.sensor.channels {
			resp.SensorChannels = append(resp.SensorChannels, sensorChannelDTO{
				ChannelID:  ch.ID,
				MetricCode: ch.MetricCode,
				Unit:       ch.Unit,
			})
		}
	}
	if sim.actuator != nil {
		resp.ActuatorDevice = sim.actuator.deviceCode
	}

	c.JSON(http.StatusOK, resp)
}

// ──────────────────── POST /start ────────────────────

func (s *SimServer) handleStart(c *gin.Context) {
	s.mu.Lock()
	if s.sim != nil {
		s.mu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"code": 10005, "message": "模拟器已在运行中"})
		return
	}
	s.mu.Unlock()

	var cfg StartConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "请求参数无效: " + err.Error()})
		return
	}

	// Validate required field
	if cfg.SensorDeviceCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "sensor_device_code 不能为空"})
		return
	}

	// Set defaults
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = "http://127.0.0.1:3000"
	}
	if cfg.Username == "" {
		cfg.Username = "admin"
	}
	if cfg.Password == "" {
		cfg.Password = "admin123"
	}
	if cfg.MqttBroker == "" {
		cfg.MqttBroker = "tcp://127.0.0.1:1883"
	}
	if cfg.GreenhouseID == 0 {
		cfg.GreenhouseID = 1
	}
	if cfg.TelemetrySec == 0 {
		cfg.TelemetrySec = 10
	}
	if cfg.HeartbeatSec == 0 {
		cfg.HeartbeatSec = 30
	}
	if cfg.EnvTickMs == 0 {
		cfg.EnvTickMs = 1000
	}

	// ── Phase 1: Login ──
	s.hub.PublishLog("info", "正在登录后端 API...")
	api := newAPIClient(cfg.APIBaseURL)
	if err := api.login(cfg.Username, cfg.Password); err != nil {
		s.hub.PublishLog("error", fmt.Sprintf("登录失败: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 10002, "message": "登录失败: " + err.Error()})
		return
	}
	s.hub.PublishLog("info", fmt.Sprintf("登录成功: %s", cfg.Username))

	// ── Phase 2: Discover sensor device ──
	s.hub.PublishLog("info", fmt.Sprintf("正在发现传感器设备: %s", cfg.SensorDeviceCode))
	ds, err := api.discoverSensorDevice(cfg.SensorDeviceCode)
	if err != nil {
		s.hub.PublishLog("error", fmt.Sprintf("发现传感器失败: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": "发现传感器失败: " + err.Error()})
		return
	}

	cfgByChan := make(map[uint64]metricConfig, len(ds.Channels))
	for _, ch := range ds.Channels {
		cfg := metricConfig{Code: ch.MetricCode, Unit: ch.Unit, Base: 25, Range: 5}
		for _, def := range defaultMetrics {
			if def.Code == ch.MetricCode {
				cfg = def
				break
			}
		}
		cfgByChan[ch.ID] = cfg
	}
	s.hub.PublishLog("info", fmt.Sprintf("传感器通道: %d", len(ds.Channels)))

	// ── Phase 2b: Discover actuator device (optional) ──
	var actuatorChannels []actuatorChannelDetail
	if cfg.ActuatorDeviceCode != "" {
		s.hub.PublishLog("info", fmt.Sprintf("正在发现执行器设备: %s", cfg.ActuatorDeviceCode))
		da, err := api.discoverActuatorDevice(cfg.ActuatorDeviceCode)
		if err != nil {
			s.hub.PublishLog("error", fmt.Sprintf("发现执行器失败: %v", err))
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": "发现执行器失败: " + err.Error()})
			return
		}
		actuatorChannels = da.Channels
		s.hub.PublishLog("info", fmt.Sprintf("执行器通道: %d", len(actuatorChannels)))
	}

	// ── Phase 3: Create Environment ──
	sharedRNG := rand.New(rand.NewSource(time.Now().UnixNano()))
	env := NewEnvironment(sharedRNG)

	// ── Phase 4: Connect MQTT ──
	s.hub.PublishLog("info", fmt.Sprintf("正在连接 MQTT: %s", cfg.MqttBroker))
	mqttClientID := fmt.Sprintf("sim-srv-%d", time.Now().UnixNano())
	mqttMgr, err := newMQTTManager(cfg.MqttBroker, cfg.MqttUser, cfg.MqttPass, mqttClientID)
	if err != nil {
		s.hub.PublishLog("error", fmt.Sprintf("MQTT 连接失败: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10006, "message": "MQTT 连接失败: " + err.Error()})
		return
	}
	s.hub.PublishLog("info", "MQTT 已连接")

	// Create simulators with hub injected
	sen := newSensorSim(cfg.SensorDeviceCode, ds.Channels, cfgByChan, mqttMgr, env, sharedRNG, s.hub)
	var act *actuatorSim
	if len(actuatorChannels) > 0 {
		act = newActuatorSim(cfg.ActuatorDeviceCode, actuatorChannels, mqttMgr, env, s.hub)

		// Initialize actuator states from backend data
		initCount := 0
		for _, ch := range actuatorChannels {
			state := ch.CurrentState
			if state == "" {
				state = "OFF"
			}
			value := 0.0
			if state == "ON" {
				value = 100.0
			}
			env.UpdateActuatorState(ch.ID, ch.ActuatorType, state, value)
			initCount++
		}
		s.hub.PublishLog("info", fmt.Sprintf("执行器状态已从后端同步: %d 个通道", initCount))
	}

	// ── Phase 5: Subscribe to commands ──
	if act != nil {
		topic := cmdTopic(act.deviceCode)
		if err := mqttMgr.subscribe(topic, act.onCommand); err != nil {
			s.hub.PublishLog("error", fmt.Sprintf("执行器命令订阅失败: %v", err))
			mqttMgr.disconnect(500)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 10006, "message": "执行器命令订阅失败: " + err.Error()})
			return
		}
		s.hub.PublishLog("info", fmt.Sprintf("已订阅执行器命令: %s", topic))
	}
	topic := cmdTopic(sen.deviceCode)
	if err := mqttMgr.subscribe(topic, sen.onCommand); err != nil {
		s.hub.PublishLog("error", fmt.Sprintf("传感器命令订阅失败: %v", err))
		mqttMgr.disconnect(500)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10006, "message": "传感器命令订阅失败: " + err.Error()})
		return
	}
	s.hub.PublishLog("info", fmt.Sprintf("已订阅传感器命令: %s", topic))

	// ── Phase 6: Publish ONLINE ──
	if act != nil {
		act.publishStatus("ONLINE")
	}
	sen.publishStatus("ONLINE")

	// ── Phase 7: Start simulation loop ──
	ctx, cancel := context.WithCancel(context.Background())

	sim := &simulation{
		ctx:       ctx,
		cancel:    cancel,
		sensor:    sen,
		actuator:  act,
		env:       env,
		mqtt:      mqttMgr,
		rng:       sharedRNG,
		startTime: time.Now(),
	}

	// Launch goroutines
	go sim.runEnvTicker(cfg.EnvTickMs)
	go sim.runTelemetryTicker(cfg.TelemetrySec, cfg.AnomalyRate)
	go sim.runHeartbeatTicker(cfg.HeartbeatSec)

	// Send initial telemetry and heartbeat immediately
	go func() {
		time.Sleep(200 * time.Millisecond)
		sen.sendTelemetry(cfg.AnomalyRate)
		sen.publishHeartbeat()
		if act != nil {
			act.publishHeartbeat()
		}
	}()

	s.mu.Lock()
	s.sim = sim
	s.mu.Unlock()

	s.hub.PublishStatus("running")
	s.hub.PublishLog("info", "模拟器已启动")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "模拟器已启动",
		"data": statusResponse{
			Running:        true,
			SensorDevice:   cfg.SensorDeviceCode,
			ActuatorDevice: cfg.ActuatorDeviceCode,
			UptimeSec:      0,
			TelemetryCount: 0,
		},
	})
}

// ──────────────────── POST /stop ────────────────────

func (s *SimServer) handleStop(c *gin.Context) {
	s.mu.Lock()
	sim := s.sim
	s.sim = nil
	s.mu.Unlock()

	if sim == nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "模拟器未在运行", "data": stopResponse{}})
		return
	}

	s.hub.PublishLog("info", "正在停止模拟器...")

	// Cancel context to stop goroutines
	sim.cancel()

	// Publish OFFLINE
	if sim.actuator != nil {
		sim.actuator.publishStatus("OFFLINE")
	}
	if sim.sensor != nil {
		sim.sensor.publishStatus("OFFLINE")
	}

	// Disconnect MQTT
	sim.mqtt.disconnect(500)

	// Compute stats
	uptime := int64(time.Since(sim.startTime).Seconds())
	telemetryCount := int64(0)
	cmdACKCount := int64(0)
	if sim.sensor != nil {
		telemetryCount = sim.sensor.totalTelemetry
	}
	if sim.actuator != nil {
		cmdACKCount = sim.actuator.totalCmdACK
	}

	s.hub.PublishStatus("stopped")
	s.hub.PublishLog("info", fmt.Sprintf("模拟器已停止，运行时长: %ds", uptime))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "模拟器已停止",
		"data": stopResponse{
			UptimeSec:      uptime,
			TelemetryCount: telemetryCount,
			CmdACKCount:    cmdACKCount,
		},
	})
}

// ──────────────────── POST /trigger-telemetry ────────────────────

func (s *SimServer) handleTriggerTelemetry(c *gin.Context) {
	s.mu.Lock()
	sim := s.sim
	s.mu.Unlock()

	if sim == nil || sim.sensor == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "模拟器未运行"})
		return
	}

	var req TriggerTelemetryReq
	_ = c.ShouldBindJSON(&req)

	overrides := make(map[uint64]float64, len(req.Overrides))
	for _, ov := range req.Overrides {
		overrides[ov.ChannelID] = ov.Value
	}

	count := sim.sensor.sendTelemetryWithOverrides(req.AnomalyRate, overrides)

	s.hub.PublishLog("info", fmt.Sprintf("手动遥测触发 (%d 通道)", count))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    triggerTelemetryResp{ChannelCount: count},
	})
}

// ──────────────────── GET /events (SSE) ────────────────────

func (s *SimServer) handleEvents(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	ch := s.hub.Subscribe()
	defer s.hub.Unsubscribe(ch)

	ctx := c.Request.Context()
	flusher, _ := c.Writer.(http.Flusher)

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			if flusher != nil {
				flusher.Flush()
			}
		}
	}
}

// ──────────────────── Goroutines ────────────────────

func (sim *simulation) runEnvTicker(envTickMs int) {
	ticker := time.NewTicker(time.Duration(envTickMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sim.ctx.Done():
			return
		case <-ticker.C:
			dt := float64(envTickMs) / 1000.0
			sim.env.Tick(dt)

			// Emit env snapshot via SSE (hub is injected into simulators)
			if sim.sensor != nil && sim.sensor.hub != nil {
				snapshot := sim.env.GetSnapshot()
				snapshot.TS = time.Now().UTC().Format(time.RFC3339)
				sim.sensor.hub.PublishEnv(snapshot)
			}
		}
	}
}

func (sim *simulation) runTelemetryTicker(intervalSec int, anomalyRate float64) {
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sim.ctx.Done():
			return
		case <-ticker.C:
			sim.sensor.sendTelemetry(anomalyRate)
		}
	}
}

func (sim *simulation) runHeartbeatTicker(intervalSec int) {
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sim.ctx.Done():
			return
		case <-ticker.C:
			sim.sensor.publishHeartbeat()
			if sim.actuator != nil {
				sim.actuator.publishHeartbeat()
			}
		}
	}
}
