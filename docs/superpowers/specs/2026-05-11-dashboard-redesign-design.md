# Dashboard Redesign Specification

**Date:** 2026-05-11
**Topic:** Dashboard Redesign for Hydroponic Agriculture Management System
**Status:** Draft

## 1. Overview
The current dashboard provides a purely IT-focused view (device online/offline status, simple charts, recent commands). This redesign shifts the focus to a **Production & Monitoring Command Center** that highlights agronomic metrics (EC, pH, DO, Batches, Strategies) alongside critical IT alerts, giving operators a comprehensive view of greenhouse health and production status.

## 2. Goals
- Provide real-time visibility into critical hydroponic parameters (EC, pH, DO, Temperature, Humidity).
- Integrate production business logic (Active Batches, Active Strategies).
- Display historical trends (24h) for key metrics to help predict anomalies.
- Enhance the UI for quick operational actions (e.g., alert acknowledgment).

## 3. Architecture & Data Flow
The redesign touches both backend aggregation APIs and frontend visualization.

### Backend Changes
Modify `GET /api/overview/dashboard` (handled in `internal/overview/handler.go`) to aggregate data from multiple domains:
- **Device & Telemetry:** Latest readings for Temp, Hum, EC, pH, DO per greenhouse.
- **Batches:** Active crop batches mapped to greenhouses.
- **Strategies:** Currently running climate and nutrient strategies.
- **Alerts:** Unresolved alerts (for quick actions).
- **History:** 24-hour aggregated telemetry data for trend charts.

### Frontend Changes
Update `packages/frontend/src/views/dashboard/index.vue`:
- **Top Stats:** Active Batches, Unresolved Alerts, Device Status, Daily Energy (if available).
- **Left Column (Greenhouse Cards):** Display detailed agronomic metrics (EC, pH, etc.) and Health Score.
- **Right Column (Trends & Production):** 24h Trend Chart component and Active Batch/Strategy list.
- **Bottom Section:** Quick Actions for Alerts and Recent Commands.
- **SSE Integration:** Subscribe to `GET /api/devices/subscribe` and `GET /api/commands/subscribe` to reflect real-time telemetry and status changes without polling.

## 4. API Contract (`GET /api/overview/dashboard`)

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "stats": {
      "active_batches_count": 3,
      "unresolved_alerts": 2,
      "devices_online": 42,
      "devices_offline": 3,
      "energy_kwh_today": 145,
      "water_l_today": 2000
    },
    "greenhouses": [
      {
        "id": "gh-1",
        "name": "Nursery Zone A",
        "health_score": "good", // good, warning, critical
        "metrics": {
          "temperature": 24.5,
          "humidity": 65,
          "ec": 1.8,
          "ph": 6.0,
          "do": 8.0,
          "co2": 450,
          "lux": 12000
        },
        "active_strategies": ["Seedling Nutrient V1"]
      }
    ],
    "trends": {
      "timestamps": ["10:00", "11:00", "..."],
      "ec_avg": [1.7, 1.8, "..."],
      "ph_avg": [6.1, 6.0, "..."]
    },
    "active_batches": [
      {
        "batch_id": "b-01",
        "crop_name": "Butterhead Lettuce",
        "stage": "Vegetative",
        "day": 14,
        "greenhouse_id": "gh-1"
      }
    ],
    "recent_alerts": [
      {
        "alert_id": "a-123",
        "severity": "CRITICAL",
        "message": "Pump P1 Offline",
        "timestamp": "2026-05-11T10:05:00Z",
        "greenhouse_name": "Nursery Zone A"
      }
    ],
    "recent_commands": [
       // Existing structure
    ]
  }
}
```

## 5. Implementation Steps
1. **Backend:** 
   - Define new DTOs in `internal/overview/dto.go`.
   - Update `overview_handler.go` to fetch cross-domain data (Batches, Telemetry history, Strategies).
   - Ensure performance by using efficient queries or caching for 24h trends.
2. **Frontend:**
   - Update `dashboard.ts` API types to match the new contract.
   - Refactor `index.vue` layout into Left/Right/Bottom panels.
   - Implement ECharts for the 24h trend.
   - Integrate SSE listeners to update `greenhouses.metrics` and `stats` dynamically.

## 6. Trade-offs & Risks
- **Performance:** Aggregating 24h data and cross-module queries might slow down the dashboard API. *Mitigation:* Consider downsampling telemetry data in InfluxDB or caching the dashboard response briefly.
- **SSE Overlap:** If SSE pushes data faster than the UI can render ECharts, it might cause stuttering. *Mitigation:* Debounce chart updates in Vue.
