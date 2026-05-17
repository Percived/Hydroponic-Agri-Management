# 水培农植信息管理后端系统（中文说明）

基于 Go 的水培温室管理后端服务，提供设备管理、遥测采集、告警与控制等能力。

## 技术栈
- Go（`gin`, `gorm`）
- MySQL（关系数据）
- InfluxDB（时序遥测）
- EMQX（MQTT Broker）

## 当前能力
- 认证与 RBAC（`ADMIN` / `OPERATOR` / `VIEWER`）
- 设备与设备分组管理
- 遥测数据采集与查询（`latest` / `history` / `stats`）
- 控制命令 / 规则 / 模板
- 告警查询与状态流转
- 审计日志查询
- 概览看板聚合（`devices_online` / `devices_offline` / `alerts_open`）

## 模块与功能
- `cmd/api`：应用入口，启动配置、依赖与 HTTP 服务。
- `internal/platform`：基础设施与公共能力（配置加载、数据库客户端、中间件、统一响应等）。
- `internal/auth`：认证与授权（JWT、用户/角色 API、RBAC 校验）。
- `internal/device`：设备与设备分组管理（CRUD、关联关系）。
- `internal/telemetry`：遥测采集与查询（最新/历史/统计），并管理保留策略配置。
- `internal/control`：控制命令、规则与模板相关 API（包含规则触发与命令下发路径）。
- `internal/alert`：告警管理 API（列表/状态/统计）。
- `internal/audit`：审计日志写入与查询。
- `internal/overview`：概览数据聚合 API（设备在线/离线、告警统计等）。
- `migrations`：数据库初始化与种子数据 SQL。
- `scripts`：本地开发辅助脚本。

## 系统架构说明（用于架构图生成）
### 架构总览
- 系统目标：提供水培温室后台服务，覆盖设备管理、遥测采集、告警、控制与审计。
- 系统边界：单体后端服务（Go/Gin），对外提供 HTTP API；对内连接 MySQL、InfluxDB、EMQX（MQTT）。
- 主要参与者：管理员、操作员、查看者（RBAC）。
- 核心能力：认证授权、设备与分组、遥测采集与查询、控制命令与规则、告警管理、审计日志、概览聚合。

### 模块与依赖
- 入口层：`cmd/api` 负责配置加载、依赖初始化与 HTTP 服务启动。
- 平台基础层：`internal/platform` 提供配置、数据库/时序库客户端、中间件、统一响应与路由装配。
- 业务模块层：
  - `internal/auth`：JWT 认证、用户与角色、RBAC 权限校验。
  - `internal/device`：设备与分组、温室（greenhouse）管理。
  - `internal/telemetry`：遥测采集与查询（最新/历史/统计）。
  - `internal/control`：控制命令、规则与模板。
  - `internal/alert`：告警管理与状态流转。
  - `internal/audit`：审计日志写入与查询。
  - `internal/overview`：概览聚合统计。
- 外部依赖：
  - MySQL：业务关系数据存储（用户/角色/设备/告警/审计等）。
  - InfluxDB：遥测时序数据存储。
  - EMQX（MQTT Broker）：设备上报与控制指令通道。

### 关键数据流与交互
- 登录与鉴权：用户 -> HTTP API -> Auth 模块 -> MySQL 校验 -> 返回 JWT；后续请求由中间件做 RBAC 校验。
- 设备与分组：用户 -> HTTP API -> Device 模块 -> MySQL（设备、分组、温室）读写。
- 遥测采集：设备/网关 -> MQTT(EMQX) -> Telemetry 摄取 -> InfluxDB 写入；同时可触发规则引擎。
- 遥测查询：用户 -> HTTP API -> Telemetry 模块 -> InfluxDB（latest/history/stats） -> 返回统计。
- 规则与控制：规则触发（遥测事件）-> Control 模块 -> 控制命令下发（MQTT）-> 设备执行。
- 告警与审计：规则或人工触发 -> Alert 模块 -> MySQL；关键操作写入 Audit。

### 部署与运行时视图
- 运行形态：单体后端服务（Go/Gin），对外暴露 HTTP API。
- 依赖组件：MySQL、InfluxDB、EMQX（MQTT Broker）。
- 交互关系：
  - HTTP 客户端（前端/运维/第三方）-> 后端服务（API）。
  - 设备/网关 -> MQTT -> 后端采集/规则 -> InfluxDB。
  - 后端服务 -> MySQL（关系数据）与 InfluxDB（时序数据）。
  - 后端服务 -> MQTT -> 设备执行控制指令。

### 外部接口与协议
- 对外 API：HTTP REST（JSON），以 `/api/*` 为前缀。
- 设备侧协议：MQTT 进行遥测上报与控制指令下发。
- 认证方式：JWT（登录获取，后续请求在 Header 传递）。

### 约束与假设
- 遥测统计依赖 InfluxDB；业务数据依赖 MySQL。
- 遥测采集可触发规则 -> 控制命令与告警。
- 角色模型固定为 `ADMIN/OPERATOR/VIEWER`。

### 架构图生成输入（结构化）
- 系统边界：单体后端服务（Go/Gin），对外提供 HTTP API；对内连接 MySQL、InfluxDB、EMQX。
- 主要模块：auth、device、telemetry、control、alert、audit、overview、platform。
- 外部系统：MySQL、InfluxDB、EMQX(MQTT)、HTTP 客户端（前端/运维/第三方）、设备/网关。
- 关键数据流：
  - 用户登录：HTTP 客户端 -> auth -> MySQL -> JWT
  - 设备管理：HTTP 客户端 -> device -> MySQL
  - 遥测上报：设备/网关 -> MQTT(EMQX) -> telemetry -> InfluxDB
  - 遥测查询：HTTP 客户端 -> telemetry -> InfluxDB -> HTTP 客户端
  - 规则控制：telemetry 触发 -> control -> MQTT(EMQX) -> 设备
  - 告警审计：control/telemetry -> alert/audit -> MySQL
- 部署关系：后端服务与 MySQL/InfluxDB/EMQX 均为独立服务，后端通过网络访问依赖。

## 快速开始
### 1. 启动依赖
```bash
docker compose up -d
```

### 2. 初始化数据库
```bash
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0001_init.up.sql
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0002_seed_auth.up.sql
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0003_seed_metrics.up.sql
```

### 3. 运行后端
```bash
go run cmd/api/main.go
```

### 4. 登录冒烟
```bash
scripts/dev-login-smoke.sh
```

## 默认本地配置
来自 `configs/config.yaml`：
- API：`http://127.0.0.1:8080`
- MySQL：`127.0.0.1:3307`（容器内部为 `3306`）
- InfluxDB：`http://127.0.0.1:8086`
- MQTT：`tcp://127.0.0.1:18830`

默认种子管理员账号：
- 用户名：`admin`
- 密码：`admin123`

## 文档索引
- API 规范：`docs/specs/API_SPEC.md`
- 设备协议：`docs/specs/DEVICE_PROTOCOL.md`
- 需求分析：`docs/analysis/REQUIREMENTS.md`
- MVP 范围：`docs/analysis/MVP.md`
- 架构图：`docs/architecture/DFD-0-Context.md`、`docs/architecture/DFD-1-System.md`、`docs/architecture/DFD-2-Modules.md`、`docs/architecture/SEQUENCE.md`
- 手工验收：`docs/testing/MANUAL_API_ACCEPTANCE.md`

## 开发检查
```bash
GOCACHE=/tmp/gocache go test ./...
```

## 备注
- 若本机已占用 `3306`，本项目通过主机端口 `3307` 避免冲突。
- 若本机 `1883` 端口不可用，Docker 会将 EMQX 的宿主机端口调整为 `18830`，容器内部仍保持 `1883`。
- Influx 写入失败会记录为警告，不会阻断遥测采集流程。
