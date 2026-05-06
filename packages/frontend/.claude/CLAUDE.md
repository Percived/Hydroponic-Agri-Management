# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Hydroponic Agriculture Management System Web Frontend - a management platform for greenhouse/hydroponic environments with device management, real-time monitoring, telemetry data visualization, and control command dispatching.

## Tech Stack

| Category | Technology |
|----------|------------|
| Framework | Vue 3 + Composition API |
| UI Library | Element Plus |
| Build Tool | Vite |
| State Management | Pinia |
| Router | Vue Router |
| HTTP Client | Axios |
| Charts | ECharts |
| Styles | Sass (SCSS) |
| Language | TypeScript |

## Development Commands

```bash
npm install        # Install dependencies
npm run dev        # Start development server (default: http://localhost:5173)
npm run build      # Build for production (includes type-check)
npm run preview    # Preview production build
npm run type-check # TypeScript type check only
npm run lint       # ESLint check and fix
```

## Architecture

### Directory Structure

```
src/
├── api/           # 20 API modules (alert, audit, auth, climate, control, crop, dashboard, device, energy, greenhouse, index, metric, notification, nutrient, pest, policy, recipe, request, telemetry, user)
├── assets/        # Static assets (styles/variables.scss, styles/global.scss)
├── components/    # Shared components (batch/, charts/, control/, device/, layout/, telemetry/)
├── composables/   # Vue composables (useAuth.ts, usePermission.ts)
├── router/        # Route configuration with guards
├── stores/        # Pinia stores (auth.ts, greenhouse.ts)
├── types/         # 20 TypeScript definition files (alert, api, audit, climate, control, crop, dashboard, device, domain, energy, greenhouse, index, metric, notification, nutrient, pest, policy, recipe, telemetry, user)
├── utils/         # Utilities (storage.ts, format.ts)
└── views/         # 17 view directories (alerts/, audit-logs/, batches/, climate/, common/, controls/, dashboard/, devices/, energy/, greenhouses/, login/, nutrient/, pest/, recipes/, settings/, telemetry/, users/)
```

### API Layer

All API requests go through centralized Axios instance (`src/api/request.ts`):

**Request Flow:**
1. Request interceptor adds `Authorization: Bearer <token>` header
2. Response interceptor handles business errors (`code !== 0`)
3. HTTP errors mapped: `401 → clearAuth → /login`, `403 → permission denied`

**Response Format:**
```typescript
interface ApiResponse<T> {
  code: number      // 0 = success
  message: string
  data: T
  request_id: string
}
```

**Usage:**
```typescript
// In api/*.ts files
import { get, post } from './request'
export const getDevices = () => get<DeviceListResponse>('/devices')
```

### Authentication

- **Storage Keys**: `hydroponic_token`, `hydroponic_user`
- **Token Flow**: Login → store in localStorage → auto-attach to requests → 401 clears and redirects
- **Store**: `useAuthStore()` provides `user`, `token`, `isLoggedIn`, `roles`, `login()`, `logout()`

### Permissions

**Roles (descending authority):**
| Role | Permissions |
|------|-------------|
| ADMIN | Full access (user management, device editing, control) |
| OPERATOR | Query + device control |
| VIEWER | Query only |

**Usage in components:**
```typescript
import { usePermission } from '@/composables'
const { canEditDevice, canControlDevice, canManageUser } = usePermission()
```

**Usage in routes:**
```typescript
meta: { roles: [Role.ADMIN] }  // Only ADMIN can access
```

### State Management

- **Global state**: Pinia stores (`auth.ts`, `device.ts`)
- **Local state**: `ref()`/`reactive()` in components
- **Persisted state**: Only auth (token + user) via localStorage

## Code Conventions

### Naming

- **Files**: kebab-case for views (`device-groups/`), PascalCase for components
- **Components**: PascalCase (`AppHeader.vue`)
- **Composables**: camelCase with `use` prefix (`useAuth.ts`)
- **Stores**: camelCase (`useAuthStore`)
- **Types**: PascalCase interfaces, UPPER_CASE enums

### Vue Components

- Use `<script setup lang="ts">` syntax
- Prefer Composition API with `ref()`/`computed()`
- Import types from `@/types`

### API Module Pattern

```typescript
// src/api/example.ts
import { get, post } from './request'
import type { Example } from '@/types'

export const getExampleList = () => get<Example[]>('/examples')
export const createExample = (data: CreateExampleRequest) => post<Example>('/examples', data)
```

## Environment Variables

| Variable | Development | Production |
|----------|-------------|------------|
| `VITE_API_BASE_URL` | `http://localhost:8080/api` | Configure in `.env.production` |

## Implemented Features (v0.7.0)

- [x] Login page with JWT authentication
- [x] Dashboard with overview stats and charts
- [x] Asset Center: sensor/actuator device lists, greenhouse/zones management
- [x] Collection Center: realtime curves, history trends, batch trends
- [x] Strategy Control: control policies (threshold + schedule), climate profiles, command dispatch
- [x] Nutrient Management: tanks, ion tests, recipes
- [x] Alerts: list with workflow (assign/takeover/close), timeline view
- [x] Batch Management: ledger, harvest records, stage plans, batch review
- [x] Pest observations & energy records
- [x] User management (admin only)
- [x] Audit logs (admin only)
- [x] Notification channels (admin only)
- [x] Route guards with auth + role-based access control
- [x] 26 routes across 17 view directories

## Documentation Update Rule (MANDATORY)

After ANY code change, update the corresponding documentation:

| Change | Documents to Update |
|--------|-------------------|
| API module change (`src/api/*.ts`) | `docs/HANDOFF.md` |
| Type definition change (`src/types/*.ts`) | `docs/HANDOFF.md`、`../../shared/docs/API_SPEC.md` |
| View/page component change | `docs/HANDOFF.md` |
| New feature or scope change | `docs/HANDOFF.md` + `docs/PROJECT_STATUS.md` |
| Shared contract change | `../../shared/docs/API_SPEC.md` |

See root `CLAUDE.md` for full documentation update rules.

## Documentation Index

- `docs/FRONTEND_PRD.md` - Product requirements
- `docs/plans/2026-04-20-mvp-frontend-design.md` - MVP design decisions
- `docs/HANDOFF.md` - Session handoff (update after every change)
- `docs/PROJECT_STATUS.md` - Project status snapshot
- `../../shared/docs/API_SPEC.md` - API specification (canonical, shared with backend)
- `../../shared/docs/openapi.yaml` - OpenAPI 3.0.3 spec (shared with backend)
- `../../CLAUDE.md` - Root monorepo rules and cross-package conventions

## Common Tasks

### Add a new page

1. Create view in `src/views/<feature>/index.vue`
2. Add route in `src/router/index.ts`
3. Add API module if needed in `src/api/<feature>.ts`
4. Add types in `src/types/<feature>.ts`

### Add a new API endpoint

1. Define types in `src/types/`
2. Add function in `src/api/<module>.ts` using `get/post/put/del`
3. Export from `src/api/index.ts`

### Add permission check

```typescript
// In component
const { hasRole, canControlDevice } = usePermission()
if (canControlDevice()) { /* show control button */ }

// In route meta
meta: { roles: [Role.ADMIN, Role.OPERATOR] }
```
