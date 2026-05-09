package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ──────────────────── API 客户端 ────────────────────

type apiClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func newAPIClient(baseURL string) *apiClient {
	return &apiClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// setToken sets the JWT token for subsequent requests.
func (a *apiClient) setToken(token string) {
	a.token = token
}

// ──────────────── 登录 ────────────────

func (a *apiClient) login(username, password string) error {
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	resp, err := a.httpClient.Post(
		a.baseURL+"/api/auth/login",
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
	a.token = ld.Token
	return nil
}

// ──────────────── 设备注册 ────────────────

// registerSensorDevice registers a new sensor device with channels.
func (a *apiClient) registerSensorDevice(code, name string, greenhouseID uint64, telemetrySec int) (*registerDeviceResp, error) {
	channels := make([]registerChannelItem, len(defaultMetrics))
	for i, m := range defaultMetrics {
		rMin := m.Base - m.Range*2
		rMax := m.Base + m.Range*2
		if rMin < 0 {
			rMin = 0
		}
		channels[i] = registerChannelItem{
			ChannelCode:         fmt.Sprintf("ch_%s", m.Code),
			MetricCode:          m.Code,
			Unit:                m.Unit,
			RangeMin:            &rMin,
			RangeMax:            &rMax,
			SamplingIntervalSec: uint(telemetrySec),
		}
	}

	return a.registerDevice(code, name, "SIMULATOR-SENSOR-V2", "sensor", greenhouseID, channels)
}

// registerActuatorDevice registers a new actuator device with channels.
func (a *apiClient) registerActuatorDevice(code, name string, greenhouseID uint64, actuatorTypes []string) (*registerDeviceResp, error) {
	channels := make([]registerChannelItem, len(actuatorTypes))
	for i, at := range actuatorTypes {
		channels[i] = registerChannelItem{
			ChannelCode:  fmt.Sprintf("act_%s", at),
			ActuatorType: at,
		}
	}

	return a.registerDevice(code, name, "SIMULATOR-ACTUATOR-V2", "actuator", greenhouseID, channels)
}

func (a *apiClient) registerDevice(code, name, model, deviceType string, greenhouseID uint64, channels []registerChannelItem) (*registerDeviceResp, error) {
	req := registerDeviceReq{
		DeviceCode:      code,
		Name:            name,
		Model:           model,
		FirmwareVersion: "2.0.0",
		GreenhouseID:    greenhouseID,
		Protocol:        "MQTT",
		DeviceType:      deviceType,
		Channels:        channels,
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST",
		a.baseURL+"/api/devices/register",
		bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer "+a.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var r apiResponse
	if err := json.Unmarshal(respBody, &r); err != nil {
		return nil, fmt.Errorf("解析失败(%d): %s", resp.StatusCode, string(respBody))
	}
	if r.Code != 0 {
		return nil, fmt.Errorf("code=%d, msg=%s, response=%s", r.Code, r.Message, string(respBody))
	}

	var reg registerDeviceResp
	if err := json.Unmarshal(r.Data, &reg); err != nil {
		return nil, fmt.Errorf("解析注册结果失败: %w", err)
	}

	log.Printf("✅ 设备已注册: id=%d, type=%s, channels=%d", reg.DeviceID, deviceType, len(reg.ChannelIDs))
	return &reg, nil
}

// ──────────────── 设备自发现 ────────────────

// discoverSensorDevice looks up a sensor device's channels by device_code.
func (a *apiClient) discoverSensorDevice(deviceCode string) (*sensorDeviceSelfResp, error) {
	resp, err := a.doGet(fmt.Sprintf("/api/devices/self?device_code=%s", deviceCode))
	if err != nil {
		return nil, err
	}

	var ds sensorDeviceSelfResp
	if err := json.Unmarshal(resp.Data, &ds); err != nil {
		return nil, fmt.Errorf("解析设备信息失败: %w", err)
	}
	if ds.DeviceType != "sensor" {
		return nil, fmt.Errorf("设备类型为 %s，期望 sensor", ds.DeviceType)
	}
	return &ds, nil
}

// discoverActuatorDevice looks up an actuator device's channels by device_code.
func (a *apiClient) discoverActuatorDevice(deviceCode string) (*actuatorDeviceSelfResp, error) {
	resp, err := a.doGet(fmt.Sprintf("/api/devices/self?device_code=%s", deviceCode))
	if err != nil {
		return nil, err
	}

	var ds actuatorDeviceSelfResp
	if err := json.Unmarshal(resp.Data, &ds); err != nil {
		return nil, fmt.Errorf("解析执行器信息失败: %w", err)
	}
	if ds.DeviceType != "actuator" {
		return nil, fmt.Errorf("设备类型为 %s，期望 actuator", ds.DeviceType)
	}
	return &ds, nil
}

func (a *apiClient) doGet(path string) (*apiResponse, error) {
	httpReq, err := http.NewRequest("GET", a.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("设备未在系统中注册 (404): %s", string(body))
	}

	var r apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("解析失败: %w", err)
	}
	if r.Code != 0 {
		return nil, fmt.Errorf("code=%d, msg=%s", r.Code, r.Message)
	}
	return &r, nil
}
