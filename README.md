# 水培农业管理系统

Hydroponic Agriculture Management System — 面向温室/水培环境的全栈管理平台，支持设备管理、MQTT 实时遥测、InfluxDB 时序存储、控制指令下发及基于角色的访问控制。

## 技术栈

| 层级 | 技术 |
|------|------|
| **前端** | Vue 3 + TypeScript + Element Plus + Vite |
| **后端** | Go + Gin + GORM |
| **数据库** | MySQL 8.0（元数据）+ InfluxDB 2.7（时序数据） |
| **消息队列** | EMQX 5.6（MQTT Broker） |
| **认证** | JWT Bearer Token |
| **容器化** | Docker Compose |

## 项目结构

```
hydroponic-agri-management/
├── packages/
│   ├── frontend/          # Vue 3 + TypeScript SPA
│   └── backend/           # Go + Gin HTTP API
├── shared/
│   └── docs/              # API 规范（API_SPEC.md + OpenAPI）
├── data/                  # 持久化数据（MySQL / InfluxDB）
└── docker-compose.yml     # MySQL + InfluxDB + EMQX
```

## 快速开始

### 1. 启动基础设施

```bash
docker compose up -d
```

### 2. 运行数据库迁移

```bash
docker compose exec -T mysql mysql -uroot -proot hydroponic < packages/backend/migrations/0001_init.up.sql
docker compose exec -T mysql mysql -uroot -proot hydroponic < packages/backend/migrations/0002_seed_auth.up.sql
```

### 3. 启动后端

```bash
cd packages/backend
go run cmd/api/main.go
```

### 4. 启动前端

```bash
cd packages/frontend
npm install
npm run dev
```

前端开发服务器默认运行在 `http://localhost:8082`，后端 API 默认运行在 `http://localhost:8080`。

## 数据流架构

```
设备 (MQTT) → EMQX Broker → 后端 (MQTT Client) → InfluxDB（时序数据）
                                                  → MySQL（元数据）
浏览器 ← 前端 (Vue SPA) → 后端 API (Gin) → MySQL + InfluxDB
```

## RBAC 角色

| 角色 | 权限 |
|------|------|
| **ADMIN** | 全部权限（用户管理、设备编辑、控制、系统配置） |
| **OPERATOR** | 查询 + 设备控制 + 告警处理 |
| **VIEWER** | 仅查询 |

## 环境变量

| 变量 | 使用方 | 默认值 | 说明 |
|------|--------|--------|------|
| `VITE_API_BASE_URL` | 前端 | `/api` | 后端 API 地址 |
| `HAMB_*` 前缀 | 后端 | `configs/config.yaml` | 覆盖配置项 |

## 文档

| 文档 | 路径 |
|------|------|
| API 规范 | `shared/docs/API_SPEC.md` |
| OpenAPI 3.0 | `shared/docs/openapi.yaml` |
| 后端架构 | `packages/backend/docs/` |
| 前端 PRD | `packages/frontend/docs/FRONTEND_PRD.md` |
| 设备协议 | `packages/backend/docs/specs/DEVICE_PROTOCOL.md` |
