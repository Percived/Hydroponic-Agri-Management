# 项目状态

最后更新: 2026-05-04
负责人: 后端团队
版本: v0.2（Phase 2 完成）

## 1. 项目概述

基于 Go + Gin + GORM 构建的水培农业后端系统。
核心依赖：MySQL、InfluxDB、EMQX（MQTT）。

## 2. 当前交付状态

总体评估：MVP 模块覆盖已基本完成，可进行手动集成测试。
前端演示：交互式单页 UI 现已覆盖登录、管理、控制、遥测查询、告警和数据采集，使用真实 API 调用。
指标种子数据：已为 TEMP/HUMIDITY/PH/EC/CO2/LIGHT 添加基线指标字典迁移。

已实现模块：

- 认证 + JWT + RBAC（`ADMIN/OPERATOR/VIEWER`）
- 温室、设备及设备分组管理
- 温室/设备分组删除 API，支持事务级联解绑行为
- 遥测数据采集/查询（`latest/history/stats`）
- 基于规则的自动触发（遥测 -> 命令 + 告警）
- 控制命令/规则/模板 API
- 告警列表/状态/统计 API
- 审计日志查询 API
- 概览仪表盘聚合
- 设备遥测概览（每小时聚合 + 在线率 + 告警事件）
- 批量设备操作（更新 / 删除）
- 批量控制命令下发（按温室 / 分组 / 指定设备）
- 通知渠道 CRUD（EMAIL / SMS / WEBHOOK）+ Webhook 测试发送（HMAC-SHA256 签名）
- 系统配置管理（GET / PUT，敏感值自动脱敏）
- CORS 中间件（宽松模式），用于基于浏览器的 API 演示请求
- 浏览器 API 文档：Swagger UI（`/docs/index.html`）+ OpenAPI 规范（`/openapi.yaml`）

## 3. 已知缺口 / 风险（按优先级）

P0：

- 自动化测试仍然很少（目前仅包含初始设备模块测试覆盖）。回归风险仍然很高。
- 多个更新/删除 handler 未检查 `RowsAffected`，可能对不存在的记录返回成功。

P1：

- 部分 API 处于占位行为级别：
  - `POST /api/controls/templates/{templateId}/apply` 目前仅做验证和日志记录，不执行实际的模板应用逻辑。
  - `GET /api/alerts/subscribe` 目前仅返回 URL 元数据。
- 遥测功能依赖预填充的 `metrics` 数据。

P2：

- 启动时对依赖较为严格（Influx/MQTT 初始化失败会退出进程）。
- 默认本地 MQTT 凭据似乎与 compose 默认值不一致；部署前需验证运行时配置。

## 4. 后续步骤（按顺序）

1. 添加正确性保障：

- 对剩余的更新/删除端点强制执行 `RowsAffected` 检查，当目标不存在时返回 404。

2. 构建最小自动化测试套件：

- 认证登录 / RBAC
- 设备 CRUD + 健康检查
- 遥测数据采集 + latest/history/stats
- 规则触发路径（数据采集 -> 命令/告警）

3. 明确占位 API：

- 要么实现真正的模板应用/订阅流式传输，
- 要么在 API 文档中明确标记为未实现，并暂时隐藏路由。

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
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0001_init.up.sql
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0002_seed_auth.up.sql
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0003_seed_metrics.up.sql
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0004_notification_channels.up.sql
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

- 入口：`cmd/api/main.go`
- 路由：`internal/platform/http/router.go`
- 认证：`internal/auth/`
- 设备：`internal/device/`
- 遥测：`internal/telemetry/`
- 控制：`internal/control/`
- 告警：`internal/alert/`
- 审计：`internal/audit/`
- 概览：`internal/overview/`
- 通知：`internal/notification/`
- 迁移：`migrations/`
- API 文档：`docs/specs/API_SPEC.md`、`docs/specs/openapi.yaml`

## 7. 新会话上下文包

在开启新模型/会话时，首先分享以下文件：

1. `docs/PROJECT_STATUS.md`
2. `docs/HANDOFF.md`
3. 你的直接目标（一句话）

## 8. 更新规则

- 内容变更时始终更新 `最后更新` 日期。
- 保持本文件稳定且处于摘要级别。
- 将短期详情放入 `docs/HANDOFF.md`。
