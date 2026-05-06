/# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Monorepo Overview

Hydroponic Agriculture Management System (水培农业管理系统) - a full-stack management platform for greenhouse/hydroponic environments with device management, real-time MQTT telemetry, InfluxDB time-series storage, control command dispatching, and role-based access control.

```
hydroponic-agri-management/
├── packages/
│   ├── frontend/    # Vue 3 + TypeScript SPA (Element Plus, Vite)
│   └── backend/     # Go + Gin HTTP API (GORM, InfluxDB, MQTT)
├── shared/
│   └── docs/        # Shared API specification (canonical source)
│       ├── API_SPEC.md
│       └── openapi.yaml
└── docker-compose.yml   # MySQL + InfluxDB + EMQX infrastructure
```

## Development Environment

- **OS**: Windows 11
- **HTTP Proxy**: `http://127.0.0.1:7897` (for git, npm, go, etc.)

### Service Ports

| Service | Port | Notes |
|---------|------|-------|
| Backend API | `3000` | Gin HTTP server |
| Frontend Dev | `8082` | Vite dev server |
| MySQL | `13307` | Docker mapped |
| InfluxDB | `9087` | Docker mapped |
| MQTT (EMQX) | `1883` | Docker mapped |
| EMQX Dashboard | `18083` | Web UI |

### Proxy Configuration

```bash
# Git
git config http.proxy http://127.0.0.1:7897
git config https.proxy http://127.0.0.1:7897

# npm
npm config set proxy http://127.0.0.1:7897
npm config set https-proxy http://127.0.0.1:7897

# Go modules
set HTTP_PROXY=http://127.0.0.1:7897
set HTTPS_PROXY=http://127.0.0.1:7897
```

## Shared Contracts

### API Response Format

Both frontend and backend agree on a unified JSON envelope:

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "request_id": "req_xxxx"
}
```

### Business Error Codes

| Code | Meaning | HTTP Status |
|------|---------|-------------|
| 0 | Success | 200/201 |
| 10001 | Validation error | 400 |
| 10002 | Unauthorized / Token invalid | 401 |
| 10003 | Forbidden / Insufficient permissions | 403 |
| 10004 | Resource not found | 404 |
| 10005 | Resource conflict | 409 |
| 10006 | Rate limit | 429 |
| 10007 | Device offline | 409 |
| 10008 | Rule conflict | 409 |
| 10009 | Data out of range | 422 |
| 10010 | Device code duplicate | 409 |

### RBAC Roles

| Role | Permissions |
|------|-------------|
| ADMIN | Full access (user management, device editing, control, system config) |
| OPERATOR | Query + device control + alert handling |
| VIEWER | Query only |

### Authentication

- JWT Bearer token via `Authorization` header
- Token obtained via `POST /api/auth/login`
- Token storage key (frontend localStorage): `hydroponic_token`

### Default Credentials

| Service | Username | Password | Notes |
|---------|----------|----------|-------|
| Web UI (Admin) | `admin` | `admin123` | Seeded by `0002_seed_auth.up.sql` |
| MySQL | `root` | `root` | Docker compose |
| InfluxDB | `admin` | `admin123` | Docker compose |
| EMQX Dashboard | `admin` | `public` | Docker compose |

### Config Secrets to Change

Before deploying, update these values in `packages/backend/configs/config.yaml`:

| Key | Default | |
|-----|---------|---------|
| `auth.jwt_secret` | `change-me` | Generate a strong random secret |
| `influx.token` | `your-token` | Match InfluxDB admin token |

## Development Commands

### Infrastructure

```bash
# From repository root
docker compose up -d        # Start MySQL + InfluxDB + EMQX
docker compose down         # Stop all services
```

### Frontend

```bash
cd packages/frontend
npm install                 # Install dependencies
npm run dev                 # Dev server on port 8082
npm run build               # Type-check + production build
npm run type-check          # TypeScript check only (vue-tsc --noEmit)
```

### Backend

```bash
cd packages/backend

# Run database migrations
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/merged/all.up.sql

# Start API server
go run cmd/api/main.go      # Dev server on configured port (default :3000)
```

### Quick Start (All-in-One)

```bash
# 1. Start infrastructure
docker compose up -d

# 2. Run migrations
docker compose exec -T mysql mysql -uroot -proot hydroponic < packages/backend/migrations/merged/all.up.sql

# 3. Start backend (terminal 1)
cd packages/backend && go run cmd/api/main.go

# 4. Start frontend (terminal 2)
cd packages/frontend && npm install && npm run dev
```

## Architecture

### Data Flow

```
Device (MQTT) → EMQX Broker → Backend (MQTT Client) → InfluxDB (time-series)
                                                      → MySQL (metadata)
Browser ← Frontend (Vue SPA) → Backend API (Gin) → MySQL + InfluxDB
```

### Backend Module Structure

Each domain module under `packages/backend/internal/<module>/` follows a consistent pattern:
- `model.go` - GORM data model
- `dto.go` - Request/response DTOs
- `handler.go` - HTTP handlers
- `routes.go` - Route registration (`RegisterRoutes(deps)`)

### Frontend Module Structure

Domain-aligned structure under `packages/frontend/src/`:
- `api/<module>.ts` - API calls via centralized Axios
- `types/<module>.ts` - TypeScript interfaces
- `views/<module>/` - Page components
- `stores/<module>.ts` - Pinia state (if global state needed)

## Cross-Package Conventions

### Documentation Update Rule (MANDATORY)

After ANY code change, you MUST update the corresponding documentation. Never skip this step.

| Change Type | Documents to Update |
|-------------|-------------------|
| Backend API change (handler/dto/route) | `shared/docs/API_SPEC.md`, `packages/backend/docs/HANDOFF.md` |
| Backend model/schema/migration change | `packages/backend/docs/HANDOFF.md`, `packages/backend/docs/PROJECT_STATUS.md` |
| Frontend API/type/view change | `packages/frontend/docs/HANDOFF.md` |
| Frontend scope/architecture change | `packages/frontend/docs/PROJECT_STATUS.md` |
| Shared contract change | `shared/docs/API_SPEC.md` + both HANDOFF.md |

Project status documents to keep in sync:
- `shared/docs/API_SPEC.md` — canonical API reference (when endpoints/fields change)
- `shared/docs/openapi.yaml` — OpenAPI spec (when endpoints/fields change)
- `packages/backend/docs/HANDOFF.md` — latest changes, session context (after every backend change)
- `packages/backend/docs/PROJECT_STATUS.md` — summary, version, migration commands (when scope changes)
- `packages/frontend/docs/HANDOFF.md` — latest changes, session context (after every frontend change)
- `packages/frontend/docs/PROJECT_STATUS.md` — summary, version (when scope changes)

### Adding a New API Endpoint

1. Add backend handler + route in `packages/backend/internal/<module>/`
2. Document the endpoint in `shared/docs/API_SPEC.md` and `shared/docs/openapi.yaml`
3. Add frontend API function in `packages/frontend/src/api/<module>.ts`
4. Add frontend types in `packages/frontend/src/types/<module>.ts`

### Environment Variables

| Variable | Used By | Default | Description |
|----------|---------|---------|-------------|
| `VITE_API_BASE_URL` | Frontend | `/api` | Backend API base URL |
| `HAMB_*` prefix | Backend | `configs/config.yaml` | Override config values |

## Documentation Index

| Document | Location | Description |
|----------|----------|-------------|
| API Specification | `shared/docs/API_SPEC.md` | Complete API reference (Chinese) |
| OpenAPI Spec | `shared/docs/openapi.yaml` | OpenAPI 3.0.3 machine-readable spec |
| Backend Docs | `packages/backend/docs/` | Architecture, requirements, testing |
| Frontend PRD | `packages/frontend/docs/FRONTEND_PRD.md` | Frontend product requirements |

## Package-Specific Guidance

For package-specific rules, conventions, and implementation details, see:
- `packages/frontend/.claude/CLAUDE.md` - Frontend conventions and patterns
- `packages/backend/.claude/CLAUDE.md` - Backend conventions and patterns
- `packages/backend/AGENTS.md` - Backend session startup rules
