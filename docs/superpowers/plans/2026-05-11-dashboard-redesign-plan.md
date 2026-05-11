# Dashboard Redesign Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform the frontend dashboard into a Production & Monitoring Command Center by aggregating real-time agronomic metrics (EC, pH, DO), crop batches, active strategies, and 24-hour historical trends, supported by a refactored backend overview API.

**Architecture:** 
- The backend `internal/overview/handler.go` will be updated to aggregate data from multiple modules (`device`, `telemetry`, `crop`, `climate`, `nutrient`, `alert`). 
- The frontend dashboard view will be restructured into a grid layout featuring key stats, detailed greenhouse metric cards (with health scores), a 24-hour ECharts trend graph, active batch/strategy tables, and an actionable alert list. 
- Real-time updates will leverage the existing SSE channels (`/api/devices/subscribe` and `/api/commands/subscribe`).

**Tech Stack:** Go (Gin, GORM), TypeScript, Vue 3 (Composition API), Element Plus, ECharts.

---

## Chunk 1: Backend DTO and Data Aggregation

### Task 1: Define Backend DTOs

**Files:**
- Modify: `packages/backend/internal/overview/dto.go`

- [ ] **Step 1: Update DTO structures**
  Add structures matching the new API contract defined in the spec:
  ```go
  // Add to packages/backend/internal/overview/dto.go
  type DashboardStats struct {
      ActiveBatchesCount int     `json:"active_batches_count"`
      UnresolvedAlerts   int     `json:"unresolved_alerts"`
      DevicesOnline      int     `json:"devices_online"`
      DevicesOffline     int     `json:"devices_offline"`
      EnergyKwhToday     float64 `json:"energy_kwh_today"`
      WaterLToday        float64 `json:"water_l_today"`
  }

  type GreenhouseMetrics struct {
      Temperature float64 `json:"temperature"`
      Humidity    float64 `json:"humidity"`
      EC          float64 `json:"ec"`
      PH          float64 `json:"ph"`
      DO          float64 `json:"do"`
      CO2         float64 `json:"co2"`
      Lux         float64 `json:"lux"`
  }

  type DashboardGreenhouse struct {
      ID               string            `json:"id"`
      Name             string            `json:"name"`
      HealthScore      string            `json:"health_score"`
      Metrics          GreenhouseMetrics `json:"metrics"`
      ActiveStrategies []string          `json:"active_strategies"`
  }

  type DashboardTrends struct {
      Timestamps []string  `json:"timestamps"`
      ECAvg      []float64 `json:"ec_avg"`
      PHAvg      []float64 `json:"ph_avg"`
  }

  type DashboardActiveBatch struct {
      BatchID        string `json:"batch_id"`
      CropName       string `json:"crop_name"`
      Stage          string `json:"stage"`
      Day            int    `json:"day"`
      GreenhouseID   string `json:"greenhouse_id"`
  }

  type DashboardRecentAlert struct {
      AlertID        string    `json:"alert_id"`
      Severity       string    `json:"severity"`
      Message        string    `json:"message"`
      Timestamp      time.Time `json:"timestamp"`
      GreenhouseName string    `json:"greenhouse_name"`
  }

  // Update existing DashboardResponse to include these new fields
  ```

- [ ] **Step 2: Commit changes**
  ```bash
  git add packages/backend/internal/overview/dto.go
  git commit -m "feat(backend): update dashboard DTOs for redesign"
  ```

### Task 2: Implement Backend Aggregation Logic

**Files:**
- Modify: `packages/backend/internal/overview/handler.go`

- [ ] **Step 1: Write test for new handler logic (Optional but recommended if tests exist)**
  Check if `handler_test.go` exists. If so, add a test checking that the new fields are returned. If not, proceed to step 2.

- [ ] **Step 2: Update `GetDashboard` handler**
  Update the handler to fetch from the DB:
  - Count online/offline devices.
  - Count unresolved alerts.
  - Fetch active crop batches (`status = 'ACTIVE'`).
  - Fetch active climate/nutrient strategies.
  - For `DashboardTrends`, fetch 24h average data from InfluxDB or mock it if InfluxDB querying is too complex for now (add a TODO comment).
  - Map the results to the new `DashboardResponse`.

- [ ] **Step 3: Run backend build to verify**
  ```bash
  cd packages/backend && go build ./...
  ```
  Expected: Builds successfully.

- [ ] **Step 4: Commit changes**
  ```bash
  git add packages/backend/internal/overview/handler.go
  git commit -m "feat(backend): implement new dashboard data aggregation"
  ```

---

## Chunk 2: Frontend Types and API Client

### Task 3: Update Frontend Types & API

**Files:**
- Modify: `packages/frontend/src/types/dashboard.ts`
- Modify: `packages/frontend/src/api/dashboard.ts`

- [ ] **Step 1: Update TypeScript interfaces**
  In `packages/frontend/src/types/dashboard.ts`, export the new interfaces (`DashboardStats`, `GreenhouseMetrics`, `DashboardGreenhouse`, `DashboardTrends`, `DashboardActiveBatch`, `DashboardRecentAlert`) to match the backend DTOs.

- [ ] **Step 2: Verify `api/dashboard.ts`**
  Ensure the API call `getDashboard()` is typed to return `Promise<ApiResponse<DashboardData>>` using the updated types. (It might already be correct if it just uses the generic `DashboardData` type).

- [ ] **Step 3: Commit changes**
  ```bash
  git add packages/frontend/src/types/dashboard.ts packages/frontend/src/api/dashboard.ts
  git commit -m "feat(frontend): update dashboard types and api client"
  ```

---

## Chunk 3: Frontend View Redesign

### Task 4: Refactor Dashboard View Component

**Files:**
- Modify: `packages/frontend/src/views/dashboard/index.vue`

- [ ] **Step 1: Implement Quick Stats Top Bar**
  Replace the existing top cards with the new `stats` data (Active Batches, Unresolved Alerts, Device Status, Energy).

- [ ] **Step 2: Implement Left Column (Greenhouse Cards)**
  Create cards iterating over `greenhouses`. For each, display the name, health score indicator, and the detailed metrics (EC, pH, DO, Temp, Hum, CO2, Lux) using a dense grid or list.

- [ ] **Step 3: Implement Right Column (Trends & Production)**
  - Integrate `echarts` to display the `trends` data (EC and pH over 24h).
  - Create a list/table for `active_batches`.

- [ ] **Step 4: Implement Bottom Section (Alerts & Commands)**
  - Add the `recent_alerts` list with a quick "Acknowledge" button (mock the action if the API isn't wired yet).
  - Keep the existing `recent_commands` table.

- [ ] **Step 5: Run frontend dev server to verify layout**
  ```bash
  cd packages/frontend && npm run build -- --emptyOutDir=false
  ```
  Expected: Compiles without type errors.

- [ ] **Step 6: Commit changes**
  ```bash
  git add packages/frontend/src/views/dashboard/index.vue
  git commit -m "feat(frontend): redesign dashboard layout and components"
  ```

### Task 5: Integrate SSE Real-time Updates

**Files:**
- Modify: `packages/frontend/src/views/dashboard/index.vue`

- [ ] **Step 1: Add SSE listeners**
  Use the existing SSE utilities (e.g., `useTelemetrySSE` or direct `EventSource` to `/api/devices/subscribe` and `/api/commands/subscribe`).
  Update the local `dashboardData.value` reactively when new telemetry arrives (e.g., updating a specific greenhouse's Temp/EC) or when an alert status changes.

- [ ] **Step 2: Commit changes**
  ```bash
  git add packages/frontend/src/views/dashboard/index.vue
  git commit -m "feat(frontend): integrate sse for real-time dashboard updates"
  ```
