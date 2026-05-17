package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStartConfig_NormalizeDeviceCodes(t *testing.T) {
	cfg := StartConfig{
		SensorDeviceCode:    "SENSOR-LEGACY",
		ActuatorDeviceCode:  "ACT-LEGACY",
		SensorDeviceCodes:   []string{"SENSOR-A", "SENSOR-B", "SENSOR-A"},
		ActuatorDeviceCodes: []string{"ACT-A", "", "ACT-A"},
	}

	sensors, actuators, err := cfg.NormalizeDeviceCodes()
	if err != nil {
		t.Fatalf("NormalizeDeviceCodes returned error: %v", err)
	}

	if len(sensors) != 3 {
		t.Fatalf("expected 3 unique sensor codes, got %d (%v)", len(sensors), sensors)
	}
	if sensors[0] != "SENSOR-LEGACY" || sensors[1] != "SENSOR-A" || sensors[2] != "SENSOR-B" {
		t.Fatalf("unexpected normalized sensor codes: %v", sensors)
	}

	if len(actuators) != 2 {
		t.Fatalf("expected 2 unique actuator codes, got %d (%v)", len(actuators), actuators)
	}
	if actuators[0] != "ACT-LEGACY" || actuators[1] != "ACT-A" {
		t.Fatalf("unexpected normalized actuator codes: %v", actuators)
	}
}

func TestSimulation_BuildStatusResponseIncludesMultipleDevices(t *testing.T) {
	env := NewEnvironment(rand.New(rand.NewSource(1)))
	sim := &simulation{
		env: env,
		sensors: []*sensorSim{
			newSensorSim("SENSOR-A", []sensorChannelDetail{
				{ID: 11, MetricCode: "TEMP", Unit: "C"},
			}, map[uint64]metricConfig{
				11: {Code: "TEMP", Unit: "C"},
			}, &mqttManager{client: &fakeMQTTClient{connected: true}}, env, rand.New(rand.NewSource(2)), nil, nil),
			newSensorSim("SENSOR-B", []sensorChannelDetail{
				{ID: 21, MetricCode: "PH", Unit: "pH"},
			}, map[uint64]metricConfig{
				21: {Code: "PH", Unit: "pH"},
			}, &mqttManager{client: &fakeMQTTClient{connected: true}}, env, rand.New(rand.NewSource(3)), nil, nil),
		},
		actuators: []*actuatorSim{
			newActuatorSim("ACT-A", []actuatorChannelDetail{
				{ID: 101, ChannelCode: "fan-a", ActuatorType: ActuatorFAN},
			}, &mqttManager{client: &fakeMQTTClient{connected: true}}, env, nil),
			newActuatorSim("ACT-B", []actuatorChannelDetail{
				{ID: 201, ChannelCode: "pump-a", ActuatorType: ActuatorPUMP},
			}, &mqttManager{client: &fakeMQTTClient{connected: true}}, env, nil),
		},
	}

	resp := sim.BuildStatusResponse()

	if !resp.Running {
		t.Fatalf("expected running=true")
	}
	if len(resp.SensorDevices) != 2 {
		t.Fatalf("expected 2 sensor devices, got %d", len(resp.SensorDevices))
	}
	if len(resp.ActuatorDevices) != 2 {
		t.Fatalf("expected 2 actuator devices, got %d", len(resp.ActuatorDevices))
	}
	if resp.SensorDevices[0].DeviceCode != "SENSOR-A" || resp.SensorDevices[1].DeviceCode != "SENSOR-B" {
		t.Fatalf("unexpected sensor devices: %+v", resp.SensorDevices)
	}
}

func TestSimulation_TriggerTelemetryAcrossMultipleSensors(t *testing.T) {
	env := NewEnvironment(rand.New(rand.NewSource(1)))
	mqttClient := &fakeMQTTClient{connected: true}

	sim := &simulation{
		env: env,
		sensors: []*sensorSim{
			newSensorSim("SENSOR-A", []sensorChannelDetail{
				{ID: 11, MetricCode: "TEMP", Unit: "C"},
				{ID: 12, MetricCode: "HUMIDITY", Unit: "%"},
			}, map[uint64]metricConfig{
				11: {Code: "TEMP", Unit: "C"},
				12: {Code: "HUMIDITY", Unit: "%"},
			}, &mqttManager{client: mqttClient}, env, rand.New(rand.NewSource(2)), nil, nil),
			newSensorSim("SENSOR-B", []sensorChannelDetail{
				{ID: 21, MetricCode: "PH", Unit: "pH"},
			}, map[uint64]metricConfig{
				21: {Code: "PH", Unit: "pH"},
			}, &mqttManager{client: mqttClient}, env, rand.New(rand.NewSource(3)), nil, nil),
		},
	}

	count := sim.TriggerTelemetry(0, map[uint64]float64{
		12: 88,
		21: 6.2,
	})

	if count != 3 {
		t.Fatalf("expected 3 telemetry items sent, got %d", count)
	}
	if len(mqttClient.publishes) != 2 {
		t.Fatalf("expected 2 telemetry publishes, got %d", len(mqttClient.publishes))
	}
}

func TestSimulation_FixedValueOverridesEnvAndManualOverride(t *testing.T) {
	env := NewEnvironment(rand.New(rand.NewSource(1)))
	mqttClient := &fakeMQTTClient{connected: true}
	sensor := newSensorSim("SENSOR-A", []sensorChannelDetail{
		{ID: 11, MetricCode: "TEMP", Unit: "C"},
	}, map[uint64]metricConfig{
		11: {Code: "TEMP", Unit: "C"},
	}, &mqttManager{client: mqttClient}, env, rand.New(rand.NewSource(2)), nil, nil)

	sim := &simulation{
		env:     env,
		sensors: []*sensorSim{sensor},
	}
	sim.SetFixedOverride(11, 33.3)
	sensor.fixedValueProvider = sim.lookupFixedOverride

	count := sim.TriggerTelemetry(0, map[uint64]float64{11: 99.9})
	if count != 1 {
		t.Fatalf("expected 1 telemetry item sent, got %d", count)
	}
	if len(mqttClient.publishes) != 1 {
		t.Fatalf("expected 1 publish, got %d", len(mqttClient.publishes))
	}

	items := decodeTelemetryPayload(t, mqttClient.publishes[0].payload)
	if got := items[0].Value; got != 33.3 {
		t.Fatalf("expected fixed value 33.3, got %v", got)
	}
}

func TestSimulation_BuildStatusResponseIncludesFixedValues(t *testing.T) {
	env := NewEnvironment(rand.New(rand.NewSource(1)))
	sensor := newSensorSim("SENSOR-A", []sensorChannelDetail{
		{ID: 11, MetricCode: "TEMP", Unit: "C"},
	}, map[uint64]metricConfig{
		11: {Code: "TEMP", Unit: "C"},
	}, &mqttManager{client: &fakeMQTTClient{connected: true}}, env, rand.New(rand.NewSource(2)), nil, nil)

	sim := &simulation{
		env:         env,
		sensors:     []*sensorSim{sensor},
		fixedValues: map[uint64]float64{11: 28.5},
	}

	resp := sim.BuildStatusResponse()
	if len(resp.SensorDevices) != 1 || len(resp.SensorDevices[0].Channels) != 1 {
		t.Fatalf("unexpected sensor devices response: %+v", resp.SensorDevices)
	}
	if resp.SensorDevices[0].Channels[0].FixedValue == nil {
		t.Fatalf("expected fixed value in status response")
	}
	if got := *resp.SensorDevices[0].Channels[0].FixedValue; got != 28.5 {
		t.Fatalf("expected fixed value 28.5, got %v", got)
	}
}

func TestSimulation_SetAnomalyRate(t *testing.T) {
	sim := &simulation{}

	sim.SetAnomalyRate(0.75)

	if got := sim.GetAnomalyRate(); got != 0.75 {
		t.Fatalf("expected anomaly rate 0.75, got %v", got)
	}
}

func TestSimServer_SetFixedOverride(t *testing.T) {
	server := NewSimServer()
	server.sim = &simulation{fixedValues: make(map[uint64]float64)}

	req := httptest.NewRequest(http.MethodPost, "/fixed-overrides", strings.NewReader(`{"channel_id":11,"value":42.5}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if got, ok := server.sim.fixedValues[11]; !ok || got != 42.5 {
		t.Fatalf("expected fixed override stored, got map=%v", server.sim.fixedValues)
	}
}

func TestSimServer_DeleteFixedOverride(t *testing.T) {
	server := NewSimServer()
	server.sim = &simulation{fixedValues: map[uint64]float64{11: 42.5}}

	req := httptest.NewRequest(http.MethodDelete, "/fixed-overrides/11", nil)
	rec := httptest.NewRecorder()

	server.engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if _, ok := server.sim.fixedValues[11]; ok {
		t.Fatalf("expected fixed override deleted, got map=%v", server.sim.fixedValues)
	}
}

func TestSimServer_UpdateAnomalyRate(t *testing.T) {
	server := NewSimServer()
	server.sim = &simulation{}

	req := httptest.NewRequest(http.MethodPost, "/anomaly-rate", strings.NewReader(`{"anomaly_rate":0.6}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
	if got := server.sim.GetAnomalyRate(); got != 0.6 {
		t.Fatalf("expected anomaly rate 0.6, got %v", got)
	}
}

func decodeTelemetryPayload(t *testing.T, payload interface{}) []telemetryItem {
	t.Helper()
	raw, ok := payload.([]byte)
	if !ok {
		t.Fatalf("expected []byte payload, got %T", payload)
	}
	var items []telemetryItem
	if err := json.Unmarshal(raw, &items); err != nil {
		t.Fatalf("unmarshal telemetry payload: %v", err)
	}
	return items
}
