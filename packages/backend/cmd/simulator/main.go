package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ──────────────────────────── 命令行参数 ────────────────────────────

var (
	apiBaseURL   = flag.String("url", "http://127.0.0.1:3000", "后端 API 地址")
	username     = flag.String("user", "admin", "登录用户名")
	password     = flag.String("pass", "admin123", "登录密码")
	mqttBroker   = flag.String("broker", "tcp://127.0.0.1:1883", "MQTT Broker 地址")
	mqttUser     = flag.String("mqtt-user", "", "MQTT 用户名")
	mqttPass     = flag.String("mqtt-pass", "", "MQTT 密码")
	deviceCode   = flag.String("device", "", "设备编码 (为空则自动注册新设备)")
	deviceName   = flag.String("name", "模拟采集器", "设备名称（注册时使用）")
	greenhouseID = flag.Uint64("gh", 1, "温室 ID")
	telemetrySec = flag.Int("interval", 10, "遥测上报间隔（秒）")
	heartbeatSec = flag.Int("heartbeat", 30, "心跳间隔（秒）")
	anomalyRate  = flag.Float64("anomaly", 0.03, "异常数据注入概率 0~1")
	durationMin  = flag.Int("duration", 0, "运行时长（分钟，0 = 永久）")
)

// MQTT 主题常量（与后端 internal/platform/mqtt/topics.go 一致）
const (
	topicPrefix    = "hydroponic"
	topicTelemetry = "telemetry"
	topicHeartbeat = "heartbeat"
	topicStatus    = "status"
	topicErrors    = "errors"
	topicAck       = "ack"
	topicCmdPrefix = "cmd"
)

// ──────────────────── API 请求/响应结构体 ────────────────────

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type loginData struct {
	Token string `json:"token"`
}

type registerReq struct {
	DeviceCode      string            `json:"device_code"`
	Name            string            `json:"name"`
	Model           string            `json:"model"`
	FirmwareVersion string            `json:"firmware_version"`
	GreenhouseID    uint64            `json:"greenhouse_id"`
	Protocol        string            `json:"protocol"`
	DeviceType      string            `json:"device_type"`
	Channels        []registerChannel `json:"channels"`
}

type registerChannel struct {
	ChannelCode         string   `json:"channel_code"`
	MetricCode          string   `json:"metric_code"`
	Unit                string   `json:"unit"`
	RangeMin            *float64 `json:"range_min"`
	RangeMax            *float64 `json:"range_max"`
	SamplingIntervalSec uint     `json:"sampling_interval_sec"`
}

type registerResp struct {
	DeviceID   uint64   `json:"device_id"`
	ChannelIDs []uint64 `json:"channel_ids"`
}

type deviceSelfResp struct {
	DeviceType string                `json:"device_type"`
	Device     sensorDeviceDetail    `json:"device"`
	Channels   []sensorChannelDetail `json:"channels"`
}

type sensorDeviceDetail struct {
	ID         uint64 `json:"id"`
	DeviceCode string `json:"device_code"`
	Name       string `json:"name"`
	Status     string `json:"status"`
}

type sensorChannelDetail struct {
	ID         uint64   `json:"id"`
	MetricCode string   `json:"metric_code"`
	Unit       string   `json:"unit"`
	RangeMin   *float64 `json:"range_min"`
	RangeMax   *float64 `json:"range_max"`
}

// ──────────────── MQTT 遥测 / 心跳 / ACK 结构体 ────────────────

// 与后端 internal/telemetry/dto.go IngestTelemetryRequest 保持一致
type telemetryItem struct {
	SensorChannelID uint64   `json:"sensor_channel_id"`
	MetricCode      string   `json:"metric_code"`
	Value           float64  `json:"value"`
	RawValue        *float64 `json:"raw_value,omitempty"`
	QualityFlag     string   `json:"quality_flag,omitempty"`
	CollectedAt     string   `json:"collected_at"`
}

// 心跳 payload（可选，后端不解析 payload）
type heartbeatPayload struct {
	TS string `json:"ts"`
}

// 状态上报
type statusPayload struct {
	Status string `json:"status"`
}

// 命令 ACK
type ackPayload struct {
	CommandID  uint64 `json:"command_id"`
	AckCode    string `json:"ack_code"`
	AckMessage string `json:"ack_message"`
}

// 接收到的命令
type incomingCmd struct {
	CommandType string          `json:"command_type"`
	Payload     json.RawMessage `json:"payload"`
	CommandID   uint64          `json:"command_id,omitempty"`
}

// ──────────────────── 传感器度量配置 ────────────────────

type metricConfig struct {
	Code    string
	Unit    string
	Base    float64
	Range   float64
	Anomaly float64
}

// 与数据库 metric_definitions 表种子的 code 保持一致
var defaultMetrics = []metricConfig{
	{Code: "TEMP", Unit: "°C", Base: 25, Range: 5, Anomaly: 40},
	{Code: "HUMIDITY", Unit: "%", Base: 70, Range: 15, Anomaly: 20},
	{Code: "PH", Unit: "pH", Base: 6.0, Range: 0.8, Anomaly: 3.5},
	{Code: "EC", Unit: "mS/cm", Base: 2.0, Range: 0.6, Anomaly: 5.5},
	{Code: "CO2", Unit: "ppm", Base: 800, Range: 300, Anomaly: 150},
	{Code: "LIGHT", Unit: "lx", Base: 35000, Range: 35000, Anomaly: 500},
	{Code: "WATER_TEMP", Unit: "°C", Base: 22, Range: 3, Anomaly: 35},
	{Code: "DO", Unit: "mg/L", Base: 6.5, Range: 1.5, Anomaly: 2.0},
}

// ──────────────────── 模拟器核心 ────────────────────

type Simulator struct {
	deviceCode string
	channels   []sensorChannelDetail
	cfgByChan  map[uint64]metricConfig
	token      string
	httpClient *http.Client
	mqttClient mqtt.Client
	rng        *rand.Rand
	mu         sync.Mutex

	// 统计
	totalTelemetry int64
	totalHeartbeat int64
	totalCmdACK    int64
	startTime      time.Time
}

func main() {
	flag.Parse()

	sim := &Simulator{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	log.Println("╔══════════════════════════════════════════════╗")
	log.Println("║     水培农业 - 采集器 MQTT 模拟器           ║")
	log.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()

	// ──── Phase 1: 登录 ────
	log.Println("─── Phase 1: 登录后端 API ───")
	if err := sim.login(); err != nil {
		log.Fatalf("❌ 登录失败: %v", err)
	}
	log.Printf("✅ 登录成功，用户: %s", *username)

	// ──── Phase 2: 设备注册 / 查询 ────
	log.Println("\n─── Phase 2: 设备注册 / 自发现 ───")
	if *deviceCode == "" {
		if err := sim.registerDevice(); err != nil {
			log.Fatalf("❌ 设备注册失败: %v", err)
		}
	} else {
		if err := sim.discoverDevice(); err != nil {
			log.Fatalf("❌ 设备查询失败: %v", err)
		}
	}
	log.Printf("✅ 设备编码: %s", sim.deviceCode)
	log.Printf("✅ 通道数量: %d", len(sim.channels))
	for _, ch := range sim.channels {
		cfg := sim.cfgByChan[ch.ID]
		log.Printf("   - [%d] %s (%s) 范围 %.1f~%.1f %s",
			ch.ID, ch.MetricCode, cfg.Unit,
			coalesce(ch.RangeMin, 0), coalesce(ch.RangeMax, 9999), cfg.Unit)
	}

	// ──── Phase 3: 连接 MQTT ────
	log.Println("\n─── Phase 3: 连接 MQTT Broker ───")
	if err := sim.connectMQTT(); err != nil {
		log.Fatalf("❌ MQTT 连接失败: %v", err)
	}
	log.Printf("✅ MQTT 已连接: %s", *mqttBroker)

	// ──── Phase 4: 订阅命令主题 ────
	log.Println("\n─── Phase 4: 订阅命令主题 ───")
	if err := sim.subscribeCommands(); err != nil {
		log.Fatalf("❌ 命令订阅失败: %v", err)
	}
	log.Printf("✅ 已订阅: %s/%s/%s/#", topicPrefix, sim.deviceCode, topicCmdPrefix)

	// ──── Phase 5: 上报在线状态 ────
	sim.publishStatus("ONLINE")
	log.Println("✅ 已上报状态: ONLINE")

	// ──── Phase 6: 开始模拟循环 ────
	log.Println("\n─── Phase 5: 开始模拟运行 ───")
	sim.startTime = time.Now()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	telemetryTicker := time.NewTicker(time.Duration(*telemetrySec) * time.Second)
	defer telemetryTicker.Stop()

	heartbeatTicker := time.NewTicker(time.Duration(*heartbeatSec) * time.Second)
	defer heartbeatTicker.Stop()

	var deadline time.Time
	if *durationMin > 0 {
		deadline = time.Now().Add(time.Duration(*durationMin) * time.Minute)
	}

	// 立刻发送首条遥测和心跳
	sim.sendTelemetry()
	sim.sendHeartbeat()

	log.Printf("▶ 遥测间隔=%ds | 心跳间隔=%ds | 异常率=%.0f%% | 持续=%s\n",
		*telemetrySec, *heartbeatSec, *anomalyRate*100,
		map[bool]string{true: fmt.Sprintf("%dmin", *durationMin), false: "永久"}[*durationMin > 0])
	log.Println("────────────────────────────────────────────")

loop:
	for {
		select {
		case <-telemetryTicker.C:
			sim.sendTelemetry()

		case <-heartbeatTicker.C:
			sim.sendHeartbeat()

		case <-sigCh:
			log.Println("\n⚠ 收到中断信号，正在退出...")
			break loop
		}

		if *durationMin > 0 && time.Now().After(deadline) {
			log.Println("\n⏰ 运行时长已到，正在退出...")
			break loop
		}
	}

	// ──── Phase 7: 优雅退出 ────
	log.Println("\n─── 优雅退出 ───")
	sim.publishStatus("OFFLINE")
	log.Println("✅ 已上报状态: OFFLINE")

	sim.mqttClient.Disconnect(500)
	log.Println("✅ MQTT 已断开")

	elapsed := time.Since(sim.startTime).Round(time.Second)
	log.Println("────────────────────────────────────────────")
	log.Printf("■ 模拟结束  运行: %v", elapsed)
	log.Printf("  遥测上报: %d 次  |  心跳: %d 次  |  命令ACK: %d 次",
		sim.totalTelemetry, sim.totalHeartbeat, sim.totalCmdACK)
}

// ──────────────────── Phase 1: 登录 ────────────────────

func (s *Simulator) login() error {
	body, _ := json.Marshal(map[string]string{
		"username": *username,
		"password": *password,
	})
	resp, err := s.httpClient.Post(
		strings.TrimRight(*apiBaseURL, "/")+"/api/auth/login",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	var r apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("解析失败: %w", err)
	}
	if r.Code != 0 {
		return fmt.Errorf("code=%d, msg=%s", r.Code, r.Message)
	}

	var ld loginData
	if err := json.Unmarshal(r.Data, &ld); err != nil {
		return fmt.Errorf("解析 token 失败: %w", err)
	}
	s.token = ld.Token
	return nil
}

// ──────────────────── Phase 2a: 注册新设备 ────────────────────

func (s *Simulator) registerDevice() error {
	code := fmt.Sprintf("SIM-%s-%04d", time.Now().Format("0102"), s.rng.Intn(9999))
	s.deviceCode = code

	channels := make([]registerChannel, len(defaultMetrics))
	for i, m := range defaultMetrics {
		rMin := m.Base - m.Range*2
		rMax := m.Base + m.Range*2
		if rMin < 0 {
			rMin = 0
		}
		channels[i] = registerChannel{
			ChannelCode:         fmt.Sprintf("ch_%s", m.Code),
			MetricCode:          m.Code,
			Unit:                m.Unit,
			RangeMin:            &rMin,
			RangeMax:            &rMax,
			SamplingIntervalSec: uint(*telemetrySec),
		}
	}

	req := registerReq{
		DeviceCode:      code,
		Name:            *deviceName,
		Model:           "SIMULATOR-V1",
		FirmwareVersion: "1.0.0",
		GreenhouseID:    *greenhouseID,
		Protocol:        "MQTT",
		DeviceType:      "sensor",
		Channels:        channels,
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST",
		strings.TrimRight(*apiBaseURL, "/")+"/api/devices/register",
		bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer "+s.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	var r apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("解析失败: %w", err)
	}
	if r.Code != 0 {
		return fmt.Errorf("code=%d, msg=%s", r.Code, r.Message)
	}

	var reg registerResp
	if err := json.Unmarshal(r.Data, &reg); err != nil {
		return fmt.Errorf("解析注册结果失败: %w", err)
	}

	log.Printf("✅ 设备已注册: id=%d, channels=%d", reg.DeviceID, len(reg.ChannelIDs))

	// 通过自发现获取通道详情
	return s.discoverDevice()
}

// ──────────────────── Phase 2b: 自发现已有设备 ────────────────────

func (s *Simulator) discoverDevice() error {
	if s.deviceCode == "" {
		s.deviceCode = *deviceCode
	}

	httpReq, _ := http.NewRequest("GET",
		fmt.Sprintf("%s/api/devices/self?device_code=%s",
			strings.TrimRight(*apiBaseURL, "/"), s.deviceCode),
		nil)
	httpReq.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 404 处理——提示设备未注册
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("设备 %s 未在系统中注册，请先注册或使用 --device=\"\" 自动注册", s.deviceCode)
	}

	var r apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("解析失败: %w", err)
	}
	if r.Code != 0 {
		return fmt.Errorf("code=%d, msg=%s", r.Code, r.Message)
	}

	var ds deviceSelfResp
	if err := json.Unmarshal(r.Data, &ds); err != nil {
		return fmt.Errorf("解析设备信息失败: %w", err)
	}
	if ds.DeviceType != "sensor" {
		return fmt.Errorf("设备类型为 %s，模拟器仅支持 sensor 类型", ds.DeviceType)
	}

	s.deviceCode = ds.Device.DeviceCode
	s.channels = ds.Channels
	s.cfgByChan = make(map[uint64]metricConfig, len(ds.Channels))

	// 为每个通道匹配 metricConfig（用于生成模拟数据）
	for _, ch := range ds.Channels {
		cfg := metricConfig{
			Code:  ch.MetricCode,
			Unit:  ch.Unit,
			Base:  25,
			Range: 5,
		}
		// 匹配已知度量
		for _, def := range defaultMetrics {
			if def.Code == ch.MetricCode {
				cfg = def
				break
			}
		}
		s.cfgByChan[ch.ID] = cfg
	}
	return nil
}

// ──────────────────── Phase 3: 连接 MQTT ────────────────────

func (s *Simulator) connectMQTT() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(*mqttBroker)
	opts.SetClientID(fmt.Sprintf("sim-%s-%d", s.deviceCode, os.Getpid()))
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(3 * time.Second)
	opts.SetMaxReconnectInterval(10 * time.Second)

	if *mqttUser != "" {
		opts.SetUsername(*mqttUser)
		opts.SetPassword(*mqttPass)
	}

	opts.OnConnect = func(_ mqtt.Client) {
		log.Println("🔗 MQTT 连接成功")
	}
	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Printf("⚠ MQTT 连接断开: %v", err)
	}
	opts.OnReconnecting = func(_ mqtt.Client, _ *mqtt.ClientOptions) {
		log.Println("🔄 MQTT 正在重连...")
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	s.mqttClient = client
	return nil
}

// ──────────────────── Phase 4: 订阅命令 ────────────────────

func (s *Simulator) subscribeCommands() error {
	topic := fmt.Sprintf("%s/%s/%s/#", topicPrefix, s.deviceCode, topicCmdPrefix)
	token := s.mqttClient.Subscribe(topic, 1, s.onCommand)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// 收到命令时的回调
func (s *Simulator) onCommand(_ mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	parts := strings.Split(topic, "/")
	cmdType := ""
	if len(parts) >= 4 {
		cmdType = parts[3]
	}

	log.Printf("📥 收到命令: topic=%s, type=%s, payload=%s", topic, cmdType, string(msg.Payload()))

	// 解析命令
	var cmd incomingCmd
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		// payload 不一定包含 command_id，可能是配置推送等
		log.Printf("   无法解析命令结构: %v", err)
		return
	}

	if cmd.CommandID == 0 {
		log.Printf("   收到配置推送 / 无 command_id 的消息，无需 ACK")
		return
	}

	// 模拟执行延迟
	time.Sleep(100 * time.Millisecond)

	// 发送 ACK
	s.mu.Lock()
	s.totalCmdACK++
	s.mu.Unlock()

	ack := ackPayload{
		CommandID:  cmd.CommandID,
		AckCode:    "ok",
		AckMessage: fmt.Sprintf("simulated %s completed", cmdType),
	}
	ackData, _ := json.Marshal(ack)
	ackTopic := fmt.Sprintf("%s/%s/%s", topicPrefix, s.deviceCode, topicAck)

	token := s.mqttClient.Publish(ackTopic, 1, false, ackData)
	if token.Wait() && token.Error() != nil {
		log.Printf("❌ ACK 发送失败: %v", token.Error())
		return
	}
	log.Printf("✅ 已发送 ACK: cmd_id=%d, code=%s", cmd.CommandID, ack.AckCode)
}

// ──────────────────── 遥测上报 ────────────────────

func (s *Simulator) sendTelemetry() {
	if len(s.channels) == 0 {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	items := make([]telemetryItem, 0, len(s.channels))

	for _, ch := range s.channels {
		cfg, ok := s.cfgByChan[ch.ID]
		if !ok {
			cfg = metricConfig{Code: ch.MetricCode, Unit: ch.Unit, Base: 25, Range: 5}
		}
		val := generateValue(cfg, s.rng)
		qualityFlag := "normal"

		// 检测异常
		if s.rng.Float64() < *anomalyRate {
			qualityFlag = "out_of_range"
			log.Printf("⚠ 注入异常数据: %s = %.1f", cfg.Code, val)
		}

		items = append(items, telemetryItem{
			SensorChannelID: ch.ID,
			MetricCode:      cfg.Code,
			Value:           round(val, 2),
			QualityFlag:     qualityFlag,
			CollectedAt:     now,
		})
	}

	payload, err := json.Marshal(items)
	if err != nil {
		log.Printf("❌ 序列化遥测数据失败: %v", err)
		return
	}

	topic := fmt.Sprintf("%s/%s/%s", topicPrefix, s.deviceCode, topicTelemetry)
	token := s.mqttClient.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		log.Printf("❌ 遥测上报失败: %v", token.Error())
		return
	}

	s.mu.Lock()
	s.totalTelemetry++
	cnt := s.totalTelemetry
	s.mu.Unlock()

	if cnt%10 == 1 || cnt == 1 {
		log.Printf("[%s] ✅ 遥测上报 #%d (%d 通道)",
			time.Now().Format("15:04:05"), cnt, len(items))
	}
}

// ──────────────────── 心跳上报 ────────────────────

func (s *Simulator) sendHeartbeat() {
	payload, _ := json.Marshal(heartbeatPayload{
		TS: time.Now().UTC().Format(time.RFC3339),
	})

	topic := fmt.Sprintf("%s/%s/%s", topicPrefix, s.deviceCode, topicHeartbeat)
	token := s.mqttClient.Publish(topic, 0, false, payload) // QoS 0, 心跳不需要可靠投递
	if token.Wait() && token.Error() != nil {
		log.Printf("❌ 心跳发送失败: %v", token.Error())
		return
	}

	s.mu.Lock()
	s.totalHeartbeat++
	s.mu.Unlock()
}

// ──────────────────── 状态上报 ────────────────────

func (s *Simulator) publishStatus(status string) {
	payload, _ := json.Marshal(statusPayload{Status: status})

	topic := fmt.Sprintf("%s/%s/%s", topicPrefix, s.deviceCode, topicStatus)
	token := s.mqttClient.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		log.Printf("❌ 状态上报失败: %v", token.Error())
	}
}

// ──────────────────── 模拟数据生成 ────────────────────

func generateValue(cfg metricConfig, rng *rand.Rand) float64 {
	// 昼夜因子
	hour := float64(time.Now().Hour())
	todFactor := 1.0
	switch cfg.Code {
	case "LIGHT":
		todFactor = sinFactor(hour, 12, 1.0)
	case "CO2":
		todFactor = 1.0 - sinFactor(hour, 12, 0.3)
	case "TEMP":
		todFactor = 0.7 + 0.3*sinFactor(hour, 14, 1.0)
	}

	noise := (rng.Float64() - 0.5) * 2 * cfg.Range * 0.3
	value := cfg.Base + cfg.Range*todFactor*rng.Float64() + noise

	if value < 0 {
		value = 0.01
	}
	return value
}

func sinFactor(hour, peakHour, amplitude float64) float64 {
	rad := ((hour - peakHour + 12) / 24) * 2 * math.Pi
	return amplitude * (1 + math.Cos(rad)) / 2
}

func round(v float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(v*pow) / pow
}

func coalesce(p *float64, defaultVal float64) float64 {
	if p == nil {
		return defaultVal
	}
	return *p
}
