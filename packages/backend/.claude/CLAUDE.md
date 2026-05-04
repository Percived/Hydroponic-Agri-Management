# CLAUDE.md

This file provides guidance to Claude Code when working with the backend package.

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
│   ├── alert/                # Alert module (dto, handler, model, routes)
│   ├── audit/                # Audit log module
│   ├── auth/                 # Auth (JWT, middleware, password)
│   ├── control/              # Control commands & rules
│   ├── device/               # Devices, greenhouses, device groups
│   ├── overview/             # Dashboard overview
│   ├── platform/             # Infrastructure (config, db, di, errors, http, influx, logger, mqtt, response)
│   └── telemetry/            # Telemetry ingestion & query
├── docs/                     # Backend-specific docs (analysis, architecture, testing)
├── migrations/               # SQL migrations
└── scripts/                  # Dev scripts
```

### Module Pattern

Each domain module follows this structure:
- `model.go` - GORM data model
- `dto.go` - Request/response DTOs
- `handler.go` - HTTP handlers
- `routes.go` - Route registration via `RegisterRoutes(deps *di.Deps)`

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

## Context Hygiene

- Prefer incremental reads over full reads
- After making meaningful code changes, update `docs/HANDOFF.md` (required) and `docs/PROJECT_STATUS.md` (if scope/status changed)

## Shared Resources

- `../../shared/docs/API_SPEC.md` - Canonical API specification
- `../../shared/docs/openapi.yaml` - OpenAPI 3.0.3 spec
- `../../CLAUDE.md` - Root monorepo rules and cross-package conventions

## Adding a New Module

1. Create directory `internal/<module>/`
2. Add `model.go`, `dto.go`, `handler.go`, `routes.go`
3. Register routes in `internal/platform/http/router.go`
4. Document endpoints in `shared/docs/API_SPEC.md`
