# 项目状态

最后更新: 2026-05-06
负责人: 后端团队
版本: v2.0.0（架构重构完成，18 个领域模块全部到位）

## 1. 项目概述

基于 Go + Gin + GORM 构建的水培农业后端系统。
核心依赖：MySQL、InfluxDB、EMQX（MQTT）。

## 2. 当前交付状态

总体评估：v2.0.0 重构已完成。全量整合 18 个领域模块，对外暴露 193 个 API 端点，分布至 21 个模块路由组。API 路径采用扁平命名规范：`/api/sensor-devices`、`/api/actuator-devices`、`/api/commands`、`/api/policies` 等。

已实现的 18 个领域模块：

- **auth** (`internal/auth/`) — JWT 认证 + RBAC 角色体系（ADMIN/OPERATOR/VIEWER），含密码散列与中间件守卫。
- **overview** (`internal/overview/`) — 仪表盘聚合：设备在线率、告警统计、遥测摘要、温室概览。
- **greenhouse** (`internal/greenhouse/`) — 温室与园区管理，包含种植分区（growing zones）。
- **device** (`internal/device/`) — 传感器/执行器设备管理，含设备通道（channels）与拓扑查询。
- **metric** (`internal/metric/`) — 测点定义字典，支持通道级别的测点绑定。
- **telemetry** (`internal/telemetry/`) — 遥测数据采集、实时/历史查询、通道级别历史数据拉取。
- **command** (`internal/command/`) — 控制命令下发与回执（receipts）追踪，命令状态机统一为 `queued/sent/acked/failed/timeout/cancelled`。
- **policy** (`internal/policy/`) — 控制策略引擎：支持阈值（threshold）、定时（schedule）、持续时长（duration）三类策略，含条件（conditions）、目标（targets）与执行记录（executions）。
- **alert** (`internal/alert/`) — 告警管理与处置闭环：告警列表/统计、指派/接管/关闭动作、时间线事件追溯。
- **notification** (`internal/notification/`) — 通知渠道 CRUD（EMAIL/SMS/WEBHOOK）+ Webhook 测试发送（HMAC-SHA256 签名）。
- **audit** (`internal/audit/`) — 审计日志查询，支持 request_id / trace_id 追踪。
- **crop** (`internal/crop/`) — 作物品种、生长阶段、种植批次、阶段计划、收获记录。
- **recipe** (`internal/recipe/`) — 营养液配方管理，含配方目标值（targets）与批次配方绑定。
- **climate** (`internal/climate/`) — 气候环境配置（climate profiles）含阶段定义、控制动作与执行日志。
- **nutrient** (`internal/nutrient/`) — 营养液管理：液箱（tanks）、换液记录（solution changes）、离子检测、浓缩液库存与消耗。
- **energy** (`internal/energy/`) — 能耗记录与汇总统计。
- **pest** (`internal/pest/`) — 病虫害观察与治理记录。
- **review** (`internal/review/`) — 批次复盘快照：汇总环境趋势、告警与控制动作，写入 `batch_review_snapshots`。

数据库迁移：
- 主迁移文件：`migrations/merged/all.up.sql`（整合全部 schema 初始化、种子数据与 PRD v1 结构，可一次性离线执行）。
- 迁移包含：三层设备模型、策略引擎结构、告警处置闭环、作物批次体系、配方与测点字典、采集辅助表、审计增强及 Phase 0–5 演示种子数据。

## 3. 已知缺口 / 风险（按优先级）

P0：

- 自动化测试覆盖仍然偏少（设备、遥测、控制、告警、批次模块已补核心路径覆盖，其余模块覆盖不足）。回归风险仍然较高。
- 多个更新/删除 handler 未检查 `RowsAffected`，可能对不存在的记录返回成功。

P1：

- v2.0.0 重构后新增模块（climate、command、crop、energy、nutrient、pest、policy、recipe、review）的测试覆盖几乎为零。
- 策略引擎缺少异步调度器与真实设备 ACK 消费器，`acked/timeout/cancelled` 仍需事件驱动链路补齐。
- 告警处置闭环缺少独立的 outbox 投递 worker 与重试调度器。
- 复盘聚合目前按批次时间窗做在线查询，大数据量场景下缺少离线预聚合与分页优化。
- 遥测功能依赖预填充的 `metrics` 数据。

P2：

- 启动时对依赖较为严格（Influx/MQTT 初始化失败会退出进程）。
- 默认本地 MQTT 凭据需与 compose 默认值保持对齐；部署前需验证运行时配置。

## 4. 后续步骤（按顺序）

1. 添加正确性保障：

- 对剩余的更新/删除端点强制执行 `RowsAffected` 检查，当目标不存在时返回 404。

2. 构建最小自动化测试套件：

- 认证登录 / RBAC
- 设备 CRUD + 健康检查
- 遥测数据采集 + latest/history/stats
- 策略触发路径（数据采集 -> 命令/告警）
- 新增模块核心路径（crop、recipe、climate、nutrient、energy、pest、review）

3. 补齐异步基础设施：

- 实现策略调度器与设备 ACK 消费器
- 实现 outbox 投递 worker 与重试调度器
- 引入离线预聚合任务用于复盘与能耗汇总

4. 按环境改进弹性：

- 在 `dev` 环境：如果 Influx/MQTT 不可用，允许优雅降级（配置开关）。
- 在 `prod` 环境：如果 SLO 要求，保持严格检查。

## 5. 运维命令

本地依赖：

```bash
docker compose up -d
```

初始化数据库：

```bash
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/merged/all.up.sql
```

运行后端：

```bash
go run cmd/api/main.go
```

手动冒烟测试：

```bash
scripts/dev-login-smoke.sh
```

## 6. 模块索引（快速导航）

| 模块 | 路径 | 说明 |
|------|------|------|
| 入口 | `cmd/api/main.go` | 应用启动入口，依赖注入组装 |
| 路由 | `internal/platform/http/router.go` | 全局路由注册中心 |
| 平台基础 | `internal/platform/` | 配置、数据库、DI、错误码、HTTP、InfluxDB、日志、MQTT、响应封装 |
| alert | `internal/alert/` | 告警管理与处置闭环 |
| audit | `internal/audit/` | 审计日志 |
| auth | `internal/auth/` | JWT 认证 + RBAC |
| climate | `internal/climate/` | 气候环境配置与执行日志 |
| command | `internal/command/` | 控制命令下发与回执 |
| crop | `internal/crop/` | 作物品种、生长阶段、种植批次、收获 |
| device | `internal/device/` | 传感器/执行器设备与通道 |
| energy | `internal/energy/` | 能耗记录与汇总 |
| greenhouse | `internal/greenhouse/` | 温室、园区、种植分区 |
| metric | `internal/metric/` | 测点定义与通道绑定 |
| notification | `internal/notification/` | 通知渠道 |
| nutrient | `internal/nutrient/` | 营养液管理 |
| overview | `internal/overview/` | 仪表盘聚合 |
| pest | `internal/pest/` | 病虫害观察与治理 |
| policy | `internal/policy/` | 控制策略引擎 |
| recipe | `internal/recipe/` | 营养液配方 |
| review | `internal/review/` | 批次复盘快照 |
| telemetry | `internal/telemetry/` | 遥测采集与查询 |
| 迁移 | `migrations/` | 数据库迁移脚本（主文件：`merged/all.up.sql`） |
| API 文档 | `shared/docs/API_SPEC.md` | 共享 API 规范 |
| OpenAPI | `shared/docs/openapi.yaml` | OpenAPI 3.0.3 规范 |

## 7. 新会话上下文包

在开启新模型/会话时，首先分享以下文件：

1. `docs/PROJECT_STATUS.md`
2. `docs/HANDOFF.md`
3. 你的直接目标（一句话）

## 8. 更新规则

- 内容变更时始终更新 `最后更新` 日期。
- 保持本文件稳定且处于摘要级别。
- 将短期详情放入 `docs/HANDOFF.md`。
