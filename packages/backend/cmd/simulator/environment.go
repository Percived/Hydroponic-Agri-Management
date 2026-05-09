package main

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// ──────────────────── 环境状态 ────────────────────

// EnvState holds the physical environment variables.
type EnvState struct {
	Temp      float64 // 空气温度 °C
	Humidity  float64 // 空气湿度 %
	PH        float64 // pH 值
	EC        float64 // 电导率 mS/cm
	CO2       float64 // CO2 浓度 ppm
	Light     float64 // 光照强度 lx
	WaterTemp float64 // 水温 °C
	DO        float64 // 溶解氧 mg/L
	Level     float64 // 液位 cm
	ORP       float64 // 氧化还原电位 mV
	TDS       float64 // 总溶解固体 ppm
	O3        float64 // 臭氧浓度 ppb
	Turbidity float64 // 浊度 NTU
	FlowRate  float64 // 流量 L/min
}

// actuatorRuntimeState holds the current state of an actuator channel.
type actuatorRuntimeState struct {
	ChannelID    uint64
	ActuatorType string
	State        string  // ON / OFF
	Value        float64 // 功率百分比 0-100
}

// actuatorEffect defines per-second effect coefficients at 100% power.
type actuatorEffect struct {
	Temp, Humidity, CO2, Light, WaterTemp, DO, PH, EC float64
	Level, ORP, TDS, O3, Turbidity, FlowRate          float64
}

// actuatorEffects maps each actuator type to its per-second effect coefficients at 100% power.
var actuatorEffects = map[string]actuatorEffect{
	ActuatorHEATER:        {Temp: 0.05, Humidity: -0.3, WaterTemp: 0.03},
	ActuatorFAN:           {Temp: -0.1, Humidity: -0.5, CO2: -2, WaterTemp: -0.02},
	ActuatorFOGGER:        {Temp: -0.02, Humidity: 0.8},
	ActuatorCO2Gen:        {CO2: 5},
	ActuatorLED:           {Temp: 0.02, Light: 200},
	ActuatorSHADE:         {Temp: -0.01, Light: -300},
	ActuatorPUMP:          {WaterTemp: 0.01, DO: 0.02, FlowRate: 0.3},
	ActuatorAERATOR:       {WaterTemp: -0.01, DO: 0.05},
	ActuatorVALVE:         {PH: 0.01, EC: 0.01},
	ActuatorDOSING_PUMP:   {PH: 0.005, EC: 0.01, FlowRate: 0.5, TDS: 2},
	ActuatorCHILLER:       {Temp: -0.02, WaterTemp: -0.1},
	ActuatorSTIRRER:       {WaterTemp: 0.005, DO: 0.01},
	ActuatorDEHUMIDIFIER:  {Temp: 0.03, Humidity: -0.8},
	ActuatorDAMPER:        {Temp: -0.05, Humidity: -0.3},
	ActuatorUV_STERILIZER: {WaterTemp: 0.01, ORP: 0.3, O3: 0.1},
	ActuatorOZONE_GEN:     {Temp: 0.01, ORP: 0.5, O3: 0.5},
	ActuatorFILTER:        {DO: 0.01, Turbidity: -0.05},
	ActuatorRO_SYSTEM:     {PH: -0.01, EC: -0.05, TDS: -2},
	ActuatorTOP_UP_VALVE:  {PH: -0.005, EC: -0.02, Level: 0.05, TDS: -1},
	ActuatorALARM:         {},
	ActuatorCALIB_VALVE:   {},
}

// ──────────────────── 环境物理模型 ────────────────────

// Environment models the shared physical environment.
// Actuator effects accumulate in it, sensors read from it.
type Environment struct {
	mu             sync.RWMutex
	State          EnvState
	actuatorStates map[uint64]actuatorRuntimeState
	rng            *rand.Rand

	// Natural equilibrium points
	equilibrium EnvState
	// Noise intensity per tick
	noiseScale float64
}

// NewEnvironment creates a new environment model with sensible defaults.
func NewEnvironment(rng *rand.Rand) *Environment {
	return &Environment{
		State: EnvState{
			Temp:      25.0,
			Humidity:  70.0,
			PH:        6.0,
			EC:        2.0,
			CO2:       800.0,
			Light:     35000.0,
			WaterTemp: 22.0,
			DO:        6.5,
			Level:     50.0,
			ORP:       350.0,
			TDS:       800.0,
			O3:        20.0,
			Turbidity: 5.0,
			FlowRate:  10.0,
		},
		actuatorStates: make(map[uint64]actuatorRuntimeState),
		equilibrium: EnvState{
			Temp:      25.0,
			Humidity:  70.0,
			PH:        6.0,
			EC:        2.0,
			CO2:       800.0,
			Light:     0.0, // Light decays to 0 at night
			WaterTemp: 22.0,
			DO:        6.5,
			Level:     50.0,
			ORP:       350.0,
			TDS:       800.0,
			O3:        20.0,
			Turbidity: 5.0,
			FlowRate:  10.0,
		},
		noiseScale: 0.05,
		rng:        rng,
	}
}

// Tick runs one simulation step (1 second of physical time elapsed).
// dt is the elapsed time in seconds (typically 1.0).
func (env *Environment) Tick(dt float64) {
	env.mu.Lock()
	defer env.mu.Unlock()

	hour := float64(time.Now().Hour())

	// ── 昼夜因子 ──
	dayFactor := sinFactor(hour, 12, 1.0) // 0 at night, 1 at noon

	// ── 1. 执行器效果叠加 ──
	for _, act := range env.actuatorStates {
		if act.State != "ON" {
			continue
		}
		eff, ok := actuatorEffects[act.ActuatorType]
		if !ok {
			continue
		}
		powerRatio := 1.0
		if act.Value > 0 && act.Value <= 100 {
			powerRatio = act.Value / 100.0
		}
		scale := powerRatio * dt

		env.State.Temp += eff.Temp * scale
		env.State.Humidity += eff.Humidity * scale
		env.State.CO2 += eff.CO2 * scale
		env.State.Light += eff.Light * scale
		env.State.WaterTemp += eff.WaterTemp * scale
		env.State.DO += eff.DO * scale
		env.State.PH += eff.PH * scale
		env.State.EC += eff.EC * scale
		env.State.Level += eff.Level * scale
		env.State.ORP += eff.ORP * scale
		env.State.TDS += eff.TDS * scale
		env.State.O3 += eff.O3 * scale
		env.State.Turbidity += eff.Turbidity * scale
		env.State.FlowRate += eff.FlowRate * scale
	}

	// ── 2. 自然衰减到平衡值（指数衰减） ──
	decayRate := 0.001 * dt
	env.State.Temp += (env.equilibrium.Temp - env.State.Temp) * decayRate
	env.State.Humidity += (env.equilibrium.Humidity - env.State.Humidity) * decayRate
	env.State.PH += (env.equilibrium.PH - env.State.PH) * decayRate
	env.State.EC += (env.equilibrium.EC - env.State.EC) * decayRate
	env.State.CO2 += (env.equilibrium.CO2 - env.State.CO2) * decayRate
	env.State.WaterTemp += (env.equilibrium.WaterTemp - env.State.WaterTemp) * decayRate
	env.State.DO += (env.equilibrium.DO - env.State.DO) * decayRate
	env.State.Level += (env.equilibrium.Level - env.State.Level) * decayRate
	env.State.ORP += (env.equilibrium.ORP - env.State.ORP) * decayRate
	env.State.TDS += (env.equilibrium.TDS - env.State.TDS) * decayRate
	env.State.O3 += (env.equilibrium.O3 - env.State.O3) * decayRate
	env.State.Turbidity += (env.equilibrium.Turbidity - env.State.Turbidity) * decayRate
	env.State.FlowRate += (env.equilibrium.FlowRate - env.State.FlowRate) * decayRate
	// Light decays to dayFactor-based equilibrium
	lightEquilibrium := 35000.0 * dayFactor
	env.State.Light += (lightEquilibrium - env.State.Light) * decayRate

	// ── 3. 昼夜对 TEMP/CO2 的直接影响 ──
	env.State.Temp += (dayFactor - 0.5) * 0.02 * dt
	env.State.CO2 += (0.5 - dayFactor) * 1.0 * dt

	// ── 3b. 液位自然蒸发/消耗 ──
	env.State.Level -= 0.001 * dt

	// ── 4. 随机噪声 ──
	env.State.Temp += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * dt
	env.State.Humidity += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 3 * dt
	env.State.PH += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 0.01 * dt
	env.State.EC += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 0.01 * dt
	env.State.CO2 += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 10 * dt
	env.State.Light += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 50 * dt
	env.State.WaterTemp += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 0.02 * dt
	env.State.DO += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 0.02 * dt
	env.State.Level += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 0.1 * dt
	env.State.ORP += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 0.3 * dt
	env.State.TDS += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 2 * dt
	env.State.O3 += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 0.2 * dt
	env.State.Turbidity += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 0.05 * dt
	env.State.FlowRate += (env.rng.Float64() - 0.5) * 2 * env.noiseScale * 0.05 * dt

	// ── 5. 值域裁剪 ──
	env.State.Temp = clamp(env.State.Temp, -10, 60)
	env.State.Humidity = clamp(env.State.Humidity, 0, 100)
	env.State.PH = clamp(env.State.PH, 0, 14)
	env.State.EC = clamp(env.State.EC, 0, 10)
	env.State.CO2 = clamp(env.State.CO2, 100, 5000)
	env.State.Light = clamp(env.State.Light, 0, 200000)
	env.State.WaterTemp = clamp(env.State.WaterTemp, 0, 50)
	env.State.DO = clamp(env.State.DO, 0, 20)
	env.State.Level = clamp(env.State.Level, 0, 200)
	env.State.ORP = clamp(env.State.ORP, 0, 800)
	env.State.TDS = clamp(env.State.TDS, 0, 5000)
	env.State.O3 = clamp(env.State.O3, 0, 200)
	env.State.Turbidity = clamp(env.State.Turbidity, 0, 100)
	env.State.FlowRate = clamp(env.State.FlowRate, 0, 50)
}

// UpdateActuatorState updates the state of an actuator channel.
// Called when a command is received via MQTT.
func (env *Environment) UpdateActuatorState(channelID uint64, actuatorType, state string, value float64) {
	env.mu.Lock()
	defer env.mu.Unlock()

	env.actuatorStates[channelID] = actuatorRuntimeState{
		ChannelID:    channelID,
		ActuatorType: actuatorType,
		State:        state,
		Value:        value,
	}
}

// GetSensorReading returns the current environment value for a given metric code.
func (env *Environment) GetSensorReading(metricCode string) float64 {
	env.mu.RLock()
	defer env.mu.RUnlock()

	switch metricCode {
	case "TEMP":
		return env.State.Temp
	case "HUMIDITY":
		return env.State.Humidity
	case "PH":
		return env.State.PH
	case "EC":
		return env.State.EC
	case "CO2":
		return env.State.CO2
	case "LIGHT":
		return env.State.Light
	case "WATER_TEMP":
		return env.State.WaterTemp
	case "DO":
		return env.State.DO
	case "LEVEL":
		return env.State.Level
	case "ORP":
		return env.State.ORP
	case "TDS":
		return env.State.TDS
	case "O3":
		return env.State.O3
	case "TURBIDITY":
		return env.State.Turbidity
	case "FLOW_RATE":
		return env.State.FlowRate
	default:
		return 0
	}
}

// GetActuatorStates returns a copy of current actuator states (for heartbeat).
func (env *Environment) GetActuatorStates() map[uint64]actuatorRuntimeState {
	env.mu.RLock()
	defer env.mu.RUnlock()

	out := make(map[uint64]actuatorRuntimeState, len(env.actuatorStates))
	for k, v := range env.actuatorStates {
		out[k] = v
	}
	return out
}

// ──────────────────── 工具函数 ────────────────────

func sinFactor(hour, peakHour, amplitude float64) float64 {
	rad := ((hour - peakHour + 12) / 24) * 2 * math.Pi
	return amplitude * (1 + math.Cos(rad)) / 2
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
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
