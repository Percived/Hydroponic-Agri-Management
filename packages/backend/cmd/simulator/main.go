package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ──────────────────────────── 命令行参数 ────────────────────────────

var (
	apiBaseURL   = flag.String("url", "http://127.0.0.1:3000", "后端 API 地址")
	username     = flag.String("user", "admin", "登录用户名")
	password     = flag.String("pass", "admin123", "登录密码")
	mqttBroker   = flag.String("broker", "tcp://127.0.0.1:1883", "MQTT Broker 地址")
	mqttUser     = flag.String("mqtt-user", "", "MQTT 用户名")
	mqttPass     = flag.String("mqtt-pass", "", "MQTT 密码")
	telemetrySec = flag.Int("interval", 10, "遥测上报间隔（秒）")
	heartbeatSec = flag.Int("heartbeat", 30, "心跳间隔（秒）")
	anomalyRate  = flag.Float64("anomaly", 0.03, "异常数据注入概率 0~1")
	durationMin  = flag.Int("duration", 0, "运行时长（分钟，0 = 永久）")

	// 新增参数
	profile        = flag.String("profile", "both", "模拟器类型: sensor / actuator / both")
	sensorDevice   = flag.String("sensor-device", "", "已有 sensor 设备编码（必填）")
	actuatorDevice = flag.String("actuator-device", "", "已有 actuator 设备编码（profile=actuator/both 时必填）")
	envTickMs      = flag.Int("env-tick-ms", 1000, "环境模型 tick 间隔(ms)")

	// Server 模式
	serverMode = flag.Bool("server", false, "启动 HTTP Server 模式（控制面板）")
	serverPort = flag.Int("port", 3001, "HTTP Server 端口")
)

// ──────────────────── 入口 ────────────────────

func main() {
	flag.Parse()

	if *serverMode {
		runServerMode()
		return
	}

	runCLIMode()
}

// runServerMode starts the HTTP server for the control panel.
func runServerMode() {
	log.Println("╔══════════════════════════════════════════════╗")
	log.Println("║   水培农业 - 模拟器 HTTP 控制面板           ║")
	log.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()

	server := NewSimServer()
	if err := server.Run(*serverPort); err != nil {
		log.Fatalf("❌ HTTP 服务启动失败: %v", err)
	}
}

// runCLIMode runs the simulator in CLI mode (original behavior, but without auto-registration).
func runCLIMode() {
	runSensor := *profile == "sensor" || *profile == "both"
	runActuator := *profile == "actuator" || *profile == "both"

	if !runSensor && !runActuator {
		log.Fatalf("❌ 无效的 profile: %s (必须是 sensor / actuator / both)", *profile)
	}

	log.Println("╔══════════════════════════════════════════════╗")
	log.Println("║   水培农业 - 闭环模拟器 (传感器+执行器)     ║")
	log.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()

	// ──── 共享 RNG ────
	sharedRNG := rand.New(rand.NewSource(time.Now().UnixNano()))

	// ──── Phase 1: 登录 ────
	log.Println("─── Phase 1: 登录后端 API ───")
	api := newAPIClient(*apiBaseURL)
	if err := api.login(*username, *password); err != nil {
		log.Fatalf("❌ 登录失败: %v", err)
	}
	log.Printf("✅ 登录成功，用户: %s", *username)

	// ──── Phase 2: 自发现设备 ────
	log.Println("\n─── Phase 2: 自发现设备 ───")

	var sensor *sensorSim
	var actuator *actuatorSim

	if runSensor {
		s, err := setupSensor(api, sharedRNG)
		if err != nil {
			log.Fatalf("❌ 传感器初始化失败: %v", err)
		}
		sensor = s
		log.Printf("✅ 传感器设备: %s, 通道: %d", sensor.deviceCode, len(sensor.channels))
		for _, ch := range sensor.channels {
			cfg := sensor.cfgByChan[ch.ID]
			log.Printf("   - [%d] %s (%s) 范围 %.1f~%.1f %s",
				ch.ID, ch.MetricCode, cfg.Unit,
				coalesce(ch.RangeMin, 0), coalesce(ch.RangeMax, 9999), cfg.Unit)
		}
	}

	if runActuator {
		a, err := setupActuator(api)
		if err != nil {
			log.Fatalf("❌ 执行器初始化失败: %v", err)
		}
		actuator = a
		log.Printf("✅ 执行器设备: %s, 通道: %d", actuator.deviceCode, len(actuator.channels))
		for _, ch := range actuator.channels {
			log.Printf("   - [%d] %s (%s)", ch.ID, ch.ChannelCode, ch.ActuatorType)
		}
	}

	// ──── Phase 3: 初始化环境模型 ────
	log.Println("\n─── Phase 3: 初始化环境物理模型 ───")
	env := NewEnvironment(sharedRNG)
	log.Println("✅ 环境模型已就绪")

	// ──── Phase 4: 连接 MQTT ────
	log.Println("\n─── Phase 4: 连接 MQTT Broker ───")
	mqttClientID := fmt.Sprintf("sim-%d", os.Getpid())
	mqttMgr, err := newMQTTManager(*mqttBroker, *mqttUser, *mqttPass, mqttClientID)
	if err != nil {
		log.Fatalf("❌ MQTT 连接失败: %v", err)
	}
	log.Printf("✅ MQTT 已连接: %s", *mqttBroker)

	// 注入 MQTT manager 到 simulators
	if sensor != nil {
		sensor.mqtt = mqttMgr
		sensor.env = env
	}
	if actuator != nil {
		actuator.mqtt = mqttMgr
		actuator.env = env
	}

	// ──── Phase 5: 订阅命令主题 ────
	log.Println("\n─── Phase 5: 订阅命令主题 ───")
	if actuator != nil {
		topic := cmdTopic(actuator.deviceCode)
		if err := mqttMgr.subscribe(topic, actuator.onCommand); err != nil {
			log.Fatalf("❌ 执行器命令订阅失败: %v", err)
		}
		log.Printf("✅ 已订阅执行器命令: %s", topic)
	}
	if sensor != nil {
		topic := cmdTopic(sensor.deviceCode)
		if err := mqttMgr.subscribe(topic, sensor.onCommand); err != nil {
			log.Fatalf("❌ 传感器命令订阅失败: %v", err)
		}
		log.Printf("✅ 已订阅传感器命令: %s", topic)
	}

	// ──── Phase 6: 上报在线状态 ────
	if actuator != nil {
		actuator.publishStatus("ONLINE")
		log.Println("✅ 执行器已上报状态: ONLINE")
	}
	if sensor != nil {
		sensor.publishStatus("ONLINE")
		log.Println("✅ 传感器已上报状态: ONLINE")
	}

	// ──── Phase 7: 开始模拟循环 ────
	log.Println("\n─── Phase 6: 开始模拟运行 ───")
	startTime := time.Now()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	envTickInterval := time.Duration(*envTickMs) * time.Millisecond
	envTicker := time.NewTicker(envTickInterval)
	defer envTicker.Stop()

	telemetryTicker := time.NewTicker(time.Duration(*telemetrySec) * time.Second)
	defer telemetryTicker.Stop()

	heartbeatTicker := time.NewTicker(time.Duration(*heartbeatSec) * time.Second)
	defer heartbeatTicker.Stop()

	var deadline time.Time
	if *durationMin > 0 {
		deadline = time.Now().Add(time.Duration(*durationMin) * time.Minute)
	}

	// Immediately send first telemetry and heartbeat
	if sensor != nil {
		sensor.sendTelemetry(*anomalyRate)
	}
	if sensor != nil {
		sensor.publishHeartbeat()
	}
	if actuator != nil {
		actuator.publishHeartbeat()
	}

	log.Printf("▶ 模式=%s | 环境tick=%v | 遥测=%ds | 心跳=%ds | 异常率=%.0f%%",
		*profile, envTickInterval, *telemetrySec, *heartbeatSec, *anomalyRate*100)
	log.Println("────────────────────────────────────────────")

loop:
	for {
		select {
		case <-envTicker.C:
			env.Tick(envTickInterval.Seconds())

		case <-telemetryTicker.C:
			if sensor != nil {
				sensor.sendTelemetry(*anomalyRate)
			}

		case <-heartbeatTicker.C:
			if sensor != nil {
				sensor.publishHeartbeat()
			}
			if actuator != nil {
				actuator.publishHeartbeat()
			}

		case <-sigCh:
			log.Println("\n⚠ 收到中断信号，正在退出...")
			break loop
		}

		if *durationMin > 0 && time.Now().After(deadline) {
			log.Println("\n⏰ 运行时长已到，正在退出...")
			break loop
		}
	}

	// ──── Phase 8: 优雅退出 ────
	log.Println("\n─── 优雅退出 ───")
	if actuator != nil {
		actuator.publishStatus("OFFLINE")
		log.Println("✅ 执行器已上报状态: OFFLINE")
	}
	if sensor != nil {
		sensor.publishStatus("OFFLINE")
		log.Println("✅ 传感器已上报状态: OFFLINE")
	}

	mqttMgr.disconnect(500)
	log.Println("✅ MQTT 已断开")

	elapsed := time.Since(startTime).Round(time.Second)
	log.Println("────────────────────────────────────────────")
	log.Printf("■ 模拟结束  运行: %v", elapsed)
	if sensor != nil {
		log.Printf("  传感器遥测: %d 次", sensor.totalTelemetry)
	}
	if actuator != nil {
		log.Printf("  执行器 ACK: %d 次", actuator.totalCmdACK)
	}
}

// ──────────────────── 传感器初始化 ────────────────────

func setupSensor(api *apiClient, rng *rand.Rand) (*sensorSim, error) {
	if *sensorDevice == "" {
		return nil, fmt.Errorf("必须指定传感器设备编码 (-sensor-device)")
	}
	deviceCode := *sensorDevice
	ds, err := api.discoverSensorDevice(deviceCode)
	if err != nil {
		return nil, fmt.Errorf("发现传感器失败: %w", err)
	}
	channels := ds.Channels

	cfgByChan := make(map[uint64]metricConfig, len(channels))
	for _, ch := range channels {
		cfg := metricConfig{
			Code:  ch.MetricCode,
			Unit:  ch.Unit,
			Base:  25,
			Range: 5,
		}
		for _, def := range defaultMetrics {
			if def.Code == ch.MetricCode {
				cfg = def
				break
			}
		}
		cfgByChan[ch.ID] = cfg
	}

	return newSensorSim(deviceCode, channels, cfgByChan, nil, nil, rng, nil), nil
}

// ──────────────────── 执行器初始化 ────────────────────

func setupActuator(api *apiClient) (*actuatorSim, error) {
	if *actuatorDevice == "" {
		return nil, fmt.Errorf("必须指定执行器设备编码 (-actuator-device)")
	}
	deviceCode := *actuatorDevice
	ds, err := api.discoverActuatorDevice(deviceCode)
	if err != nil {
		return nil, fmt.Errorf("发现执行器失败: %w", err)
	}
	channels := ds.Channels

	return newActuatorSim(deviceCode, channels, nil, nil, nil), nil
}
