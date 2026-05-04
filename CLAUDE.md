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
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0001_init.up.sql
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0002_seed_auth.up.sql

# Start API server
go run cmd/api/main.go      # Dev server on configured port (default :8080)
```

### Quick Start (All-in-One)

```bash
# 1. Start infrastructure
docker compose up -d

# 2. Run migrations (see above)

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

### When to Update Shared Docs

- If you change an API endpoint in the backend, update `shared/docs/API_SPEC.md` and `shared/docs/openapi.yaml`
- If you add a new TypeScript type that mirrors a backend struct, align field names with the API spec
- The `shared/docs/` directory is the **canonical source** for the API contract

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
| Device Protocol | `packages/backend/docs/specs/DEVICE_PROTOCOL.md` | Device MQTT protocol spec |

## Package-Specific Guidance

For package-specific rules, conventions, and implementation details, see:
- `packages/frontend/.claude/CLAUDE.md` - Frontend conventions and patterns
- `packages/backend/.claude/CLAUDE.md` - Backend conventions and patterns
- `packages/backend/AGENTS.md` - Backend session startup rules
