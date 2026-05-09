# CLAUDE.md

Hydroponic Agriculture Management System (水培农业管理系统) — full-stack management for greenhouse/hydroponic environments.

**Versions**: Backend v2.3.1 | Frontend v0.8.1 | Last updated: 2026-05-09

## Monorepo Structure

```
├── packages/frontend/   # Vue 3 + TS SPA (Element Plus, Vite)
├── packages/backend/    # Go + Gin HTTP API (GORM, InfluxDB, MQTT)
├── shared/docs/         # API_SPEC.md (canonical), openapi.yaml
└── docker-compose.yml   # MySQL:13307, InfluxDB:9087, EMQX:1883/18083
```

**Ports**: Backend `:3000` | Frontend `:8082` | MySQL `13307` | InfluxDB `9087` | MQTT `1883` | EMQX Dashboard `18083`
**Proxy**: `http://127.0.0.1:7897` for git, npm, go
**OS**: Windows 11

## Dev Commands

```bash
docker compose up -d                          # Infra
cd packages/frontend && npm run dev           # Frontend :8082
cd packages/backend && go run cmd/api/main.go # Backend :3000
# Migration: docker compose exec -T mysql mysql -uroot -proot hydroponic < packages/backend/migrations/merged/all.up.sql
```

## Architecture

**Data flow**: `Device (MQTT) → EMQX → Backend → InfluxDB (time-series) / MySQL (metadata)` | `Browser ← SSE ← Backend (alerts, telemetry)` | `Browser → Frontend → Backend API → MySQL + InfluxDB`

**Backend pattern** (`internal/<module>/`): `model.go` (GORM) → `dto.go` → `handler.go` → `routes.go` (`RegisterRoutes(deps)`). Large modules (>400 lines) split into `*_handler.go` sub-files; policy & climate have `scheduler.go`.

**Frontend pattern** (`src/`): `api/<module>.ts` → `types/<module>.ts` → `views/<module>/` → `stores/<module>.ts` (optional)

## API Contracts

- **Envelope**: `{"code": 0, "message": "ok", "data": {}, "request_id": "req_xxxx"}`
- **Auth**: JWT Bearer token, login via `POST /api/auth/login`, localStorage key `hydroponic_token`
- **RBAC**: `ADMIN` (full) | `OPERATOR` (query+control+alerts) | `VIEWER` (query only)
- **Error codes**: 0=success, 10001=validation, 10002=unauthorized, 10003=forbidden, 10004=not found, 10005=conflict, 10007=device offline. See source code for full list.
- **Full spec**: `shared/docs/API_SPEC.md`

## Cross-Package Rules

### Doc Update (MANDATORY after every change)

| Change | Update |
|--------|--------|
| Backend API (handler/dto/route) | `shared/docs/API_SPEC.md` + `packages/backend/docs/HANDOFF.md` |
| Backend model/schema | `packages/backend/docs/HANDOFF.md` + `PROJECT_STATUS.md` |
| Frontend API/type/view | `packages/frontend/docs/HANDOFF.md` |
| Shared contract | `shared/docs/API_SPEC.md` + both HANDOFF.md |

### Adding a New Endpoint
1. Backend: handler + route in `internal/<module>/`
2. Docs: `shared/docs/API_SPEC.md` + `openapi.yaml`
3. Frontend: `api/<module>.ts` + `types/<module>.ts`

## Automatic Task Routing

Classify every request and follow the matching workflow. Read `.claude/agents/<role>.md` before starting each phase.

| Request | Workflow |
|---------|----------|
| New feature / refactor / architecture / DB design | **PlanMode → Architect → (approve) → Developer → Review** |
| Bug fix / small change / add endpoint | **Developer → Review** |
| Code review / quality check | **Reviewer** |

**Phase actions:**
- **Architect**: Explore codebase → design with Plan agent → present plan → wait for approval. Never skip to coding for >3 files.
- **Developer**: TDD → parallel subagents for independent tasks → simplify → verify (tests+build).
- **Reviewer**: codebase-analysis for context → simplify for quality → output `🔴Blocking / 🟡Suggestions / 🟢Optimizations`.

**Hard rules:** PlanMode first for non-trivial features | Self-review after every change | Verify before claiming done.

## Package-Specific Guidance

- `packages/frontend/.claude/CLAUDE.md` — Frontend conventions
- `packages/backend/.claude/CLAUDE.md` — Backend conventions
- `packages/backend/AGENTS.md` — Backend session startup
- `.claude/agents/software-architect.md` — Architect role
- `.claude/agents/software-developer.md` — Developer role
- `.claude/agents/code-reviewer.md` — Reviewer role
