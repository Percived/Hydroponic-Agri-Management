# CLAUDE.md

This file provides guidance to Claude Code when working with the backend package.

**Version**: v2.3.1 | **Last updated**: 2026-05-09

## Package Overview

Go + Gin HTTP API server for the Hydroponic Agri Management System. Provides RESTful endpoints for device management, telemetry ingestion/query, MQTT integration, alert handling, and role-based access control.

## Tech Stack

| Category | Technology |
|----------|------------|
| Language | Go 1.24 |
| HTTP Framework | Gin 1.11 |
| ORM | GORM 1.31 (MySQL driver) |
| Time-Series DB | InfluxDB 2.7 |
| Message Broker | EMQX 5.6 (MQTT) |
| Auth | JWT (golang-jwt/v5) |
| Validation | go-playground/validator/v10 |
| Config | Viper |
| API Docs | Swagger (swaggo) |

## Development Commands

```bash
# From packages/backend/
go run cmd/api/main.go       # Start API server
go test ./...                # Run all tests
go build cmd/api/main.go     # Build binary

# From repository root
docker compose up -d          # Start infrastructure
docker compose down           # Stop infrastructure
```

## Architecture

### Directory Structure

```
packages/backend/
├── cmd/api/main.go           # Entry point
├── configs/config.yaml       # Default configuration
├── internal/
│   ├── alert/                # Alert & alert workflow
│   ├── audit/                # Audit log
│   ├── auth/                 # Auth (JWT, middleware, password)
│   ├── climate/              # Climate profiles & execution logs
│   ├── command/              # Command dispatch & receipts
│   ├── crop/                 # Crop varieties, growth stages, batches
│   ├── device/               # Sensor/actuator devices, channels, topology
│   ├── energy/               # Energy consumption records
│   ├── greenhouse/           # Greenhouses, parks, growing zones
│   ├── metric/               # Metric definitions & channel bindings
│   ├── notification/         # Notification channels
│   ├── nutrient/             # Nutrient tanks, solution changes, ion tests
│   ├── overview/             # Dashboard overview
│   ├── pest/                 # Pest observations & treatment records
│   ├── platform/             # Infrastructure (config, db, di, errors, http, influx, logger, mqtt, response, event)
│   ├── policy/               # Control policies (conditions, schedules, targets)
│   ├── recipe/               # Nutrient recipes & targets
│   ├── review/               # Batch review snapshots
│   └── telemetry/            # Telemetry ingestion & query
├── docs/                     # Backend-specific docs
├── migrations/               # SQL migrations
└── scripts/                  # Dev scripts
```

### Module Pattern

Each domain module follows this structure:
- `model.go` - GORM data model
- `dto.go` - Request/response DTOs
- `handler.go` - HTTP handlers (main Handler struct + shared helpers)
- `routes.go` - Route registration via `RegisterRoutes(deps *di.Deps)`
- `*_handler.go` - Split sub-handlers for modules >400 lines (climate, policy, nutrient, crop)
- `scheduler.go` - Auto-scheduler (policy, climate: event-driven + timer-based)
- `cache.go` - In-memory cache (telemetry, device)

### Dependency Injection

`internal/platform/di/deps.go` defines the `Deps` struct holding all shared dependencies (Config, Logger, MySQL, InfluxDB, MQTT client). This struct is built in `main.go` and passed to each module's `RegisterRoutes()`.

### API Response Format

Uses unified JSON envelope via `internal/platform/response/response.go`:
```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_xxxx" }
```

## Session Startup Rule

Before doing full repository scans, read these files first:
1. `docs/PROJECT_STATUS.md`
2. `docs/HANDOFF.md`

If the user request can be answered using these files, avoid full-codebase traversal.

## Documentation Update Rule (MANDATORY)

After ANY code change, update the corresponding documentation. Never skip this step.

| Change Type | Documents to Update |
|-------------|-------------------|
| Handler/DTO/route change | `docs/HANDOFF.md` + `../../shared/docs/API_SPEC.md` |
| Model/schema change | `docs/HANDOFF.md` + `docs/PROJECT_STATUS.md` |
| Migration change | `docs/HANDOFF.md` + `docs/PROJECT_STATUS.md` |
| New feature or scope change | `docs/HANDOFF.md` + `docs/PROJECT_STATUS.md` |
| Bug fix affecting API response | `docs/HANDOFF.md` + `../../shared/docs/API_SPEC.md` |

## Context Hygiene

- Prefer incremental reads over full reads

## Shared Resources

- `../../shared/docs/API_SPEC.md` - Canonical API specification
- `../../shared/docs/openapi.yaml` - OpenAPI 3.0.3 spec
- `../../CLAUDE.md` - Root monorepo rules and cross-package conventions

## Adding a New Module

1. Create directory `internal/<module>/`
2. Add `model.go`, `dto.go`, `handler.go`, `routes.go`
3. Register routes in `internal/platform/http/router.go`
4. Document endpoints in `shared/docs/API_SPEC.md`
