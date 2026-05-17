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

	sensors     []*sensorSim
	actuators   []*actuatorSim
	env         *Environment
	mqtt        *mqttManager
	rng         *rand.Rand
	fixedMu     sync.RWMutex
	fixedValues map[uint64]float64
	anomalyMu   sync.RWMutex
	anomalyRate float64
	batchID     *uint64

	startTime time.Time
}

func (sim *simulation) BuildStatusResponse() statusResponse {
	resp := statusResponse{
		Running:        true,
		UptimeSec:      int64(time.Since(sim.startTime).Seconds()),
		TelemetryCount: sim.totalTelemetryCount(),
	}
	for _, sensor := range sim.sensors {
		device := sensorDeviceStatusDTO{DeviceCode: sensor.deviceCode}
		for _, ch := range sensor.channels {
			channel := sensorChannelDTO{
				ChannelID:  ch.ID,
				MetricCode: ch.MetricCode,
				Unit:       ch.Unit,
			}
			if fixed, ok := sim.lookupFixedOverride(ch.ID); ok {
				fixedCopy := fixed
				channel.FixedValue = &fixedCopy
			}
			device.Channels = append(device.Channels, channel)
			resp.SensorChannels = append(resp.SensorChannels, channel)
		}
		resp.SensorDevices = append(resp.SensorDevices, device)
	}
	for _, actuator := range sim.actuators {
		device := actuatorDeviceStatusDTO{
			DeviceCode:   actuator.deviceCode,
			ChannelCount: len(actuator.channels),
		}
		for _, ch := range actuator.channels {
			device.Channels = append(device.Channels, actuatorChannelDTO{
				ChannelID:    ch.ID,
				ChannelCode:  ch.ChannelCode,
				ActuatorType: ch.ActuatorType,
				State:        ch.CurrentState,
				Level:        ch.CurrentLevel,
			})
		}
		resp.ActuatorDevices = append(resp.ActuatorDevices, device)
	}
	if len(resp.SensorDevices) > 0 {
		resp.SensorDevice = resp.SensorDevices[0].DeviceCode
	}
	if len(resp.ActuatorDevices) > 0 {
		resp.ActuatorDevice = resp.ActuatorDevices[0].DeviceCode
	}
	return resp
}

func (sim *simulation) TriggerTelemetry(anomalyRate float64, overrides map[uint64]float64) int64 {
	var count int64
	for _, sensor := range sim.sensors {
		count += sensor.sendTelemetryWithOverrides(anomalyRate, overrides)
	}
	return count
}

func (sim *simulation) publishStatusAll(status string) {
	for _, actuator := range sim.actuators {
		actuator.publishStatus(status)
	}
	for _, sensor := range sim.sensors {
		sensor.publishStatus(status)
	}
}

func (sim *simulation) publishHeartbeatAll() {
	for _, sensor := range sim.sensors {
		sensor.publishHeartbeat()
	}
	for _, actuator := range sim.actuators {
		actuator.publishHeartbeat()
	}
}

func (sim *simulation) subscribeAllCommands() error {
	for _, actuator := range sim.actuators {
		if err := sim.mqtt.subscribe(cmdTopic(actuator.deviceCode), actuator.onCommand); err != nil {
			return fmt.Errorf("执行器 %s: %w", actuator.deviceCode, err)
		}
	}
	for _, sensor := range sim.sensors {
		if err := sim.mqtt.subscribe(cmdTopic(sensor.deviceCode), sensor.onCommand); err != nil {
			return fmt.Errorf("传感器 %s: %w", sensor.deviceCode, err)
		}
	}
	return nil
}

func (sim *simulation) totalTelemetryCount() int64 {
	var total int64
	for _, sensor := range sim.sensors {
		total += sensor.totalTelemetry
	}
	return total
}

func (sim *simulation) totalCmdACKCount() int64 {
	var total int64
	for _, actuator := range sim.actuators {
		total += actuator.totalCmdACK
	}
	return total
}

func (sim *simulation) SetFixedOverride(channelID uint64, value float64) {
	sim.fixedMu.Lock()
	defer sim.fixedMu.Unlock()
	if sim.fixedValues == nil {
		sim.fixedValues = make(map[uint64]float64)
	}
	sim.fixedValues[channelID] = value
}

func (sim *simulation) ClearFixedOverride(channelID uint64) {
	sim.fixedMu.Lock()
	defer sim.fixedMu.Unlock()
	delete(sim.fixedValues, channelID)
}

func (sim *simulation) lookupFixedOverride(channelID uint64) (float64, bool) {
	sim.fixedMu.RLock()
	defer sim.fixedMu.RUnlock()
	value, ok := sim.fixedValues[channelID]
	return value, ok
}

func (sim *simulation) SetAnomalyRate(rate float64) {
	sim.anomalyMu.Lock()
	defer sim.anomalyMu.Unlock()
	sim.anomalyRate = rate
}

func (sim *simulation) GetAnomalyRate() float64 {
	sim.anomalyMu.RLock()
	defer sim.anomalyMu.RUnlock()
	return sim.anomalyRate
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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
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
	engine.POST("/anomaly-rate", s.handleUpdateAnomalyRate)
	engine.POST("/fixed-overrides", s.handleSetFixedOverride)
	engine.DELETE("/fixed-overrides/:channelId", s.handleDeleteFixedOverride)
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

	c.JSON(http.StatusOK, s.sim.BuildStatusResponse())
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

	sensorCodes, actuatorCodes, err := cfg.NormalizeDeviceCodes()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": err.Error()})
		return
	}

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
		cfg.MqttBroker = "tcp://127.0.0.1:18830"
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

	s.hub.PublishLog("info", "正在登录后端 API...")
	api := newAPIClient(cfg.APIBaseURL)
	if err := api.login(cfg.Username, cfg.Password); err != nil {
		s.hub.PublishLog("error", fmt.Sprintf("登录失败: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 10002, "message": "登录失败: " + err.Error()})
		return
	}
	s.hub.PublishLog("info", fmt.Sprintf("登录成功: %s", cfg.Username))

	sharedRNG := rand.New(rand.NewSource(time.Now().UnixNano()))
	env := NewEnvironment(sharedRNG)

	s.hub.PublishLog("info", fmt.Sprintf("正在连接 MQTT: %s", cfg.MqttBroker))
	mqttClientID := fmt.Sprintf("sim-srv-%d", time.Now().UnixNano())
	mqttMgr, err := newMQTTManager(cfg.MqttBroker, cfg.MqttUser, cfg.MqttPass, mqttClientID)
	if err != nil {
		s.hub.PublishLog("error", fmt.Sprintf("MQTT 连接失败: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10006, "message": "MQTT 连接失败: " + err.Error()})
		return
	}
	s.hub.PublishLog("info", "MQTT 已连接")

	sensors := make([]*sensorSim, 0, len(sensorCodes))
	for _, sensorCode := range sensorCodes {
		s.hub.PublishLog("info", fmt.Sprintf("正在发现传感器设备: %s", sensorCode))
		ds, discoverErr := api.discoverSensorDevice(sensorCode)
		if discoverErr != nil {
			s.hub.PublishLog("error", fmt.Sprintf("发现传感器失败: %v", discoverErr))
			mqttMgr.disconnect(500)
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": fmt.Sprintf("发现传感器失败(%s): %v", sensorCode, discoverErr)})
			return
		}
		sensors = append(sensors, newSensorSim(sensorCode, ds.Channels, buildMetricConfigMap(ds.Channels), mqttMgr, env, sharedRNG, s.hub, cfg.BatchID))
		s.hub.PublishLog("info", fmt.Sprintf("传感器通道: %s (%d)", sensorCode, len(ds.Channels)))
	}

	actuators := make([]*actuatorSim, 0, len(actuatorCodes))
	for _, actuatorCode := range actuatorCodes {
		s.hub.PublishLog("info", fmt.Sprintf("正在发现执行器设备: %s", actuatorCode))
		da, discoverErr := api.discoverActuatorDevice(actuatorCode)
		if discoverErr != nil {
			s.hub.PublishLog("error", fmt.Sprintf("发现执行器失败: %v", discoverErr))
			mqttMgr.disconnect(500)
			c.JSON(http.StatusBadRequest, gin.H{"code": 10004, "message": fmt.Sprintf("发现执行器失败(%s): %v", actuatorCode, discoverErr)})
			return
		}
		actuators = append(actuators, newActuatorSim(actuatorCode, da.Channels, mqttMgr, env, s.hub))
		initializeActuatorStates(env, da.Channels)
		s.hub.PublishLog("info", fmt.Sprintf("执行器通道: %s (%d)", actuatorCode, len(da.Channels)))
	}

	ctx, cancel := context.WithCancel(context.Background())
	sim := &simulation{
		ctx:         ctx,
		cancel:      cancel,
		sensors:     sensors,
		actuators:   actuators,
		env:         env,
		mqtt:        mqttMgr,
		rng:         sharedRNG,
		fixedValues: make(map[uint64]float64),
		anomalyRate: cfg.AnomalyRate,
		startTime:   time.Now(),
	}
	for _, sensor := range sim.sensors {
		sensor.fixedValueProvider = sim.lookupFixedOverride
	}

	if err := sim.subscribeAllCommands(); err != nil {
		s.hub.PublishLog("error", fmt.Sprintf("命令订阅失败: %v", err))
		mqttMgr.disconnect(500)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10006, "message": "命令订阅失败: " + err.Error()})
		return
	}
	for _, actuator := range sim.actuators {
		s.hub.PublishLog("info", fmt.Sprintf("已订阅执行器命令: %s", cmdTopic(actuator.deviceCode)))
	}
	for _, sensor := range sim.sensors {
		s.hub.PublishLog("info", fmt.Sprintf("已订阅传感器命令: %s", cmdTopic(sensor.deviceCode)))
	}

	sim.publishStatusAll("ONLINE")

	go sim.runEnvTicker(cfg.EnvTickMs)
	go sim.runTelemetryTicker(cfg.TelemetrySec)
	go sim.runHeartbeatTicker(cfg.HeartbeatSec)

	go func() {
		time.Sleep(200 * time.Millisecond)
		sim.TriggerTelemetry(sim.GetAnomalyRate(), nil)
		sim.publishHeartbeatAll()
	}()

	s.mu.Lock()
	s.sim = sim
	s.mu.Unlock()

	s.hub.PublishStatus("running")
	s.hub.PublishLog("info", "模拟器已启动")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "模拟器已启动",
		"data":    sim.BuildStatusResponse(),
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
	sim.cancel()
	sim.publishStatusAll("OFFLINE")
	sim.mqtt.disconnect(500)

	uptime := int64(time.Since(sim.startTime).Seconds())
	telemetryCount := sim.totalTelemetryCount()
	cmdACKCount := sim.totalCmdACKCount()

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

	if sim == nil || len(sim.sensors) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "模拟器未运行"})
		return
	}

	var req TriggerTelemetryReq
	_ = c.ShouldBindJSON(&req)

	overrides := make(map[uint64]float64, len(req.Overrides))
	for _, ov := range req.Overrides {
		overrides[ov.ChannelID] = ov.Value
	}

	count := sim.TriggerTelemetry(req.AnomalyRate, overrides)

	s.hub.PublishLog("info", fmt.Sprintf("手动遥测触发 (%d 通道)", count))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    triggerTelemetryResp{ChannelCount: count},
	})
}

func (s *SimServer) handleUpdateAnomalyRate(c *gin.Context) {
	s.mu.Lock()
	sim := s.sim
	s.mu.Unlock()

	if sim == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "模拟器未运行"})
		return
	}

	var req anomalyRateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "请求参数无效"})
		return
	}

	sim.SetAnomalyRate(req.AnomalyRate)
	s.hub.PublishLog("info", fmt.Sprintf("已更新异常率: %.4f", req.AnomalyRate))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

func (s *SimServer) handleSetFixedOverride(c *gin.Context) {
	s.mu.Lock()
	sim := s.sim
	s.mu.Unlock()

	if sim == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "模拟器未运行"})
		return
	}

	var req fixedOverrideReq
	if err := c.ShouldBindJSON(&req); err != nil || req.ChannelID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "请求参数无效"})
		return
	}

	sim.SetFixedOverride(req.ChannelID, req.Value)
	s.hub.PublishLog("info", fmt.Sprintf("已设置通道固定值: channel=%d value=%.2f", req.ChannelID, req.Value))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

func (s *SimServer) handleDeleteFixedOverride(c *gin.Context) {
	s.mu.Lock()
	sim := s.sim
	s.mu.Unlock()

	if sim == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "模拟器未运行"})
		return
	}

	var channelID uint64
	if _, err := fmt.Sscan(c.Param("channelId"), &channelID); err != nil || channelID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "channel_id 无效"})
		return
	}

	sim.ClearFixedOverride(channelID)
	s.hub.PublishLog("info", fmt.Sprintf("已清除通道固定值: channel=%d", channelID))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
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

			if len(sim.sensors) > 0 && sim.sensors[0].hub != nil {
				snapshot := sim.env.GetSnapshot()
				snapshot.TS = time.Now().UTC().Format(time.RFC3339)
				sim.sensors[0].hub.PublishEnv(snapshot)
			}
		}
	}
}

func (sim *simulation) runTelemetryTicker(intervalSec int) {
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sim.ctx.Done():
			return
		case <-ticker.C:
			sim.TriggerTelemetry(sim.GetAnomalyRate(), nil)
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
			sim.publishHeartbeatAll()
		}
	}
}

func buildMetricConfigMap(channels []sensorChannelDetail) map[uint64]metricConfig {
	cfgByChan := make(map[uint64]metricConfig, len(channels))
	for _, ch := range channels {
		cfg := metricConfig{Code: ch.MetricCode, Unit: ch.Unit, Base: 25, Range: 5}
		for _, def := range defaultMetrics {
			if def.Code == ch.MetricCode {
				cfg = def
				break
			}
		}
		cfgByChan[ch.ID] = cfg
	}
	return cfgByChan
}

func initializeActuatorStates(env *Environment, channels []actuatorChannelDetail) {
	for _, ch := range channels {
		state := ch.CurrentState
		if state == "" {
			state = "OFF"
		}
		value := 0.0
		if ch.CurrentLevel != nil {
			value = *ch.CurrentLevel
		} else if state == "ON" {
			value = 100.0
		}
		env.UpdateActuatorState(ch.ID, ch.ActuatorType, state, value)
	}
}
