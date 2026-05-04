package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

// ---- config ----

var (
	baseURL     = flag.String("url", "http://127.0.0.1:3000", "Backend API base URL")
	username    = flag.String("user", "admin", "Login username")
	password    = flag.String("pass", "admin123", "Login password")
	intervalSec = flag.Int("interval", 10, "Telemetry report interval (seconds)")
	anomalyRate = flag.Float64("anomaly", 0.03, "Anomaly injection probability 0~1")
	durationMin = flag.Int("duration", 0, "Run duration in minutes (0 = forever)")
	batchSize   = flag.Int("batch", 5, "Max devices per request")
)

// ---- API types ----

type LoginResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

type DeviceItem struct {
	ID         uint64 `json:"id"`
	DeviceCode string `json:"device_code"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Category   string `json:"category"`
	Status     string `json:"status"`
}

type DeviceListResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []DeviceItem `json:"items"`
	} `json:"data"`
}

type TelemetryMetric struct {
	Code  string  `json:"code"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type IngestReq struct {
	DeviceCode string            `json:"device_code"`
	Metrics    []TelemetryMetric `json:"metrics"`
}

type IngestResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Accepted int `json:"accepted"`
	} `json:"data"`
}

// ---- value generators ----

type metricConfig struct {
	Code    string
	Unit    string
	Base    float64 // baseline
	Range   float64 // normal oscillation range (±)
	Anomaly float64 // anomaly target value (trigger alert)
}

var metricConfigs = map[string]metricConfig{
	"TEMP":     {"TEMP", "C", 22, 4, 33},
	"HUMIDITY": {"HUMIDITY", "%", 70, 10, 25},
	"PH":       {"PH", "pH", 6.0, 0.5, 4.5},
	"EC":       {"EC", "mS/cm", 2.0, 0.5, 5.5},
	"CO2":      {"CO2", "ppm", 800, 300, 200},
	"LIGHT":    {"LIGHT", "lx", 35000, 35000, 500},
}

func generateValue(cfg metricConfig, rng *rand.Rand) float64 {
	if rng.Float64() < *anomalyRate {
		// return an anomaly value that will exceed normal thresholds
		return cfg.Anomaly + rng.Float64()*cfg.Range*0.3
	}

	// time-of-day factor (LIGHT & CO2 have diurnal cycle)
	hour := float64(time.Now().Hour())
	todFactor := 1.0
	switch cfg.Code {
	case "LIGHT":
		// light peaks at noon, near zero at night
		todFactor = sinFactor(hour, 12, 1.0)
	case "CO2":
		// CO2 is higher at night (plants respire)
		todFactor = 1.0 - sinFactor(hour, 12, 0.3)
	}

	noise := (rng.Float64() - 0.5) * 2 * cfg.Range * 0.3 // ±30% of range as noise
	value := cfg.Base + cfg.Range*todFactor*rng.Float64() + noise

	// clamp negative values to a small positive
	if value < 0 {
		value = 0.01
	}
	return value
}

// sinFactor returns a 0~1 value based on hour of day with peak at peakHour
func sinFactor(hour, peakHour, amplitude float64) float64 {
	// map hour to radians, peak at peakHour
	rad := ((hour - peakHour + 12) / 24) * 2 * 3.1415926535
	return amplitude * (1 + cosApprox(rad)) / 2
}

// fast cosine approximation (good enough for our purposes)
func cosApprox(x float64) float64 {
	// range reduce to [-pi, pi]
	for x > 3.1415926535 {
		x -= 2 * 3.1415926535
	}
	for x < -3.1415926535 {
		x += 2 * 3.1415926535
	}
	x2 := x * x
	return 1 - x2/2 + x2*x2/24 - x2*x2*x2/720
}

// ---- API client ----

type Client struct {
	http    *http.Client
	baseURL string
	token   string
}

func NewClient(baseURL string) *Client {
	return &Client{
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (c *Client) Login(username, password string) error {
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	resp, err := c.http.Post(c.baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	var r LoginResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("login decode: %w", err)
	}
	if r.Code != 0 {
		return fmt.Errorf("login failed: %s", r.Message)
	}
	c.token = r.Data.Token
	log.Printf("✓ 登录成功, token=%s...", c.token[:20])
	return nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.http.Do(req)
}

func (c *Client) ListDevices() ([]DeviceItem, error) {
	req, _ := http.NewRequest("GET", c.baseURL+"/api/devices", nil)
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var r DeviceListResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r.Data.Items, nil
}

func (c *Client) IngestTelemetry(deviceCode string, metrics []TelemetryMetric) (int, error) {
	body, _ := json.Marshal(IngestReq{
		DeviceCode: deviceCode,
		Metrics:    metrics,
	})
	req, _ := http.NewRequest("POST", c.baseURL+"/api/telemetry", bytes.NewReader(body))
	resp, err := c.do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var r IngestResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, err
	}
	if r.Code != 0 {
		return 0, fmt.Errorf("ingest: %s", r.Message)
	}
	return r.Data.Accepted, nil
}

// ---- main ----

func main() {
	flag.Parse()

	client := NewClient(*baseURL)

	// 1. login
	if err := client.Login(*username, *password); err != nil {
		log.Fatalf("登录失败: %v", err)
	}

	// 2. fetch devices
	allDevices, err := client.ListDevices()
	if err != nil {
		log.Fatalf("获取设备列表失败: %v", err)
	}

	// filter to only ENABLED sensors
	var sensors []DeviceItem
	for _, d := range allDevices {
		if d.Type == "SENSOR" && d.Status == "ENABLED" {
			if _, ok := metricConfigs[d.Category]; ok {
				sensors = append(sensors, d)
			}
		}
	}
	if len(sensors) == 0 {
		log.Fatal("未找到启用的传感器设备，请先执行种子数据迁移")
	}

	log.Printf("✓ 发现 %d 个传感器", len(sensors))
	for _, s := range sensors {
		cfg := metricConfigs[s.Category]
		log.Printf("  - %s (%s) [%s, 基准 %.1f%s]", s.DeviceCode, s.Name, s.Category, cfg.Base, cfg.Unit)
	}

	// 3. set up timer
	interval := time.Duration(*intervalSec) * time.Second
	var deadline time.Time
	if *durationMin > 0 {
		deadline = time.Now().Add(time.Duration(*durationMin) * time.Minute)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var totalReports, totalMetrics int64
	startTime := time.Now()

	log.Printf("▶ 开始模拟，间隔=%v，异常率=%.0f%%，持续=%s",
		interval, *anomalyRate*100,
		map[bool]string{true: fmt.Sprintf("%dmin", *durationMin), false: "永久"}[*durationMin > 0])
	log.Println("────────────────────────────────────────────")

	// run first report immediately
	reportBatch(client, sensors, rng, &totalReports, &totalMetrics)

loop:
	for {
		select {
		case <-ticker.C:
			reportBatch(client, sensors, rng, &totalReports, &totalMetrics)

			if *durationMin > 0 && time.Now().After(deadline) {
				break loop
			}

		case <-sigCh:
			log.Println("\n收到中断信号，正在退出...")
			break loop
		}
	}

	elapsed := time.Since(startTime).Round(time.Second)
	log.Println("────────────────────────────────────────────")
	log.Printf("■ 模拟结束  运行: %v | 上报: %d 次 | 数据点: %d | 设备: %d",
		elapsed, totalReports, totalMetrics, len(sensors))
}

func reportBatch(client *Client, sensors []DeviceItem, rng *rand.Rand, totalReports, totalMetrics *int64) {
	ok, fail := 0, 0
	for _, s := range sensors {
		cfg := metricConfigs[s.Category]
		value := generateValue(cfg, rng)
		metrics := []TelemetryMetric{{Code: cfg.Code, Value: round(value, 1), Unit: cfg.Unit}}

		_, err := client.IngestTelemetry(s.DeviceCode, metrics)
		if err != nil {
			fail++
			log.Printf("✗ %s: %v", s.DeviceCode, err)
		} else {
			ok++
		}
		*totalMetrics++

		// small delay between devices to avoid flooding
		time.Sleep(50 * time.Millisecond)
	}
	*totalReports++

	now := time.Now().Format("15:04:05")
	if fail > 0 {
		log.Printf("[%s] 上报 %d/%d 成功, %d 失败", now, ok, ok+fail, fail)
	} else {
		log.Printf("[%s] ✓ %d 个设备", now, ok)
	}
}

func round(v float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int(v*pow+0.5)) / pow
}
