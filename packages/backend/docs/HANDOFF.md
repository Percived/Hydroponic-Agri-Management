# 交接文档

最后更新: 2026-04-23
当前分支: (填写)
当前重点: 为温室和设备组添加删除 API，并实现级联解绑行为。

## 1. 近期变更

- 在设备模块中新增删除 API：
  - `DELETE /api/devices/greenhouses/:greenhouseId`
  - `DELETE /api/device-groups/:groupId`
- 更新 `internal/device/handler.go`：
  - 新增 `DeleteGreenhouse`：事务中解绑设备的 `greenhouse_id`，解绑已分组设备的 `group_id`，删除关联分组，然后删除温室。
  - 新增 `DeleteGroup`：事务中解绑设备的 `group_id`，然后删除分组。
  - 当删除目标不存在时返回 404（`not_found`）。
- 更新 `internal/device/routes.go`：
  - 注册温室/分组删除路由，权限为 ADMIN。
- 扩展 `internal/device/handler_test.go` 中的设备 handler 测试：
  - `TestDeleteGroupUnbindsDevices`
  - `TestDeleteGroupReturnsNotFound`
  - `TestDeleteGreenhouseCascadesAndUnbinds`
  - `TestDeleteGreenhouseReturnsNotFound`
- 更新文档：
  - `docs/specs/API_SPEC.md` 现已记录两个删除端点及级联解绑行为。
  - `docs/specs/openapi.yaml` 现已包含温室/分组资源的删除操作。
- 新增规划/设计产物：
  - `docs/superpowers/specs/2026-04-23-delete-greenhouse-devicegroup-design.md`
  - `docs/superpowers/plans/2026-04-23-delete-greenhouse-devicegroup.md`
- 验证通过：
  - `go test ./internal/device -run TestDelete`
  - `go test ./internal/device`
  - `go test ./...`

- 更新 `GET /api/device-groups` 的 API 文档：
  - `docs/specs/API_SPEC.md` 现已在 DeviceGroup 字段和响应示例中记录 `device_count`。
  - `docs/specs/openapi.yaml` 现已在分组列表的 200 响应 schema 中包含 `device_count`。
- 更新 `internal/device/handler.go` 中的 `GET /api/device-groups`：
  - 在每个分组项响应中新增 `device_count`。
  - 设备数量通过对 `devices.group_id` 进行分组查询聚合得出。
- 新增首个设备模块自动化测试：`internal/device/handler_test.go`：
  - 新测试 `TestListGroupsReturnsDeviceCount` 验证每个分组返回正确的 `device_count`。
- 验证通过：
  - `go test ./internal/device -run TestListGroupsReturnsDeviceCount`
  - `go test ./internal/device`
- 修复 `internal/telemetry/handler.go` 中遥测 Influx 写入上下文处理：
  - 将 Gin 请求上下文传入 `writeInflux`。
  - 停止向 Influx `WritePoint` 传递 `nil` 上下文，当请求不可用时回退到 `context.Background()`。
- 验证修复：`go test ./...` 和 `go vet ./...`（均通过）。
- 在 `/docs/index.html` 添加 Swagger UI 路由。
- 新增 OpenAPI 规范文件：`docs/specs/openapi.yaml`。
- 在 HTTP 路由中注册静态规范路由 `/openapi.yaml`。
- 修复 `/docs/*any` 与静态 OpenAPI 路由之间的 Gin 路由冲突。
- 在 `go.mod` 中添加 Swagger 依赖（`swaggo/files`、`swaggo/gin-swagger`）。
- 在 `README_CN.md` 中添加架构描述章节，用于图表生成输入。
- 新增指标种子数据迁移（`0003_seed_metrics.up.sql`），包含 TEMP/HUMIDITY/PH/EC/CO2/LIGHT。
- 将 `frontend_demo/index.html` 重构为单页交互式演示（登录、仪表盘、温室/设备管理、控制、遥测查询、告警、模拟数据采集），调用真实后端 API。
- 新增上下文管理基线：
  - `AGENTS.md`
  - `docs/PROJECT_STATUS.md`
  - `docs/HANDOFF.md`
  - `scripts/status-snapshot.ps1`
- 新增中文 README：`README_CN.md`
- 新增 API 演示页面：`frontend_demo/index.html`
- 在演示页面中添加中文端点描述。
- 为演示请求添加变量模板和更丰富的默认参数。
- 新增温室 CRUD API 和用于创建/列出温室的演示端点。
- 更新 API 规范，添加温室模型和端点（追加）。
- 将温室端点移至 API 规范的设备模块章节。
- 将温室路由更新为 `/api/devices/greenhouses` 并同步演示/文档。
- 通过显式的演示顺序对演示 UI 重新排序。
- 调整演示 UI 分组顺序：认证、用户/角色、设备（含温室）、遥测，然后是其他。
- 更新 API 规范，添加温室模型和端点（追加）。
- 新增 API 演示页面：`frontend_demo/index.html`
- 添加宽松的 CORS 中间件以允许浏览器演示请求。
- 将所有演示端点描述替换为空字符串，以避免演示页面中出现 JS 解析错误。
- 在演示发送按钮上添加直接点击处理程序，确保请求正常发送。
- 从 git 恢复 `docs/specs/API_SPEC.md` 内容，并以 UTF-8 BOM 重新保存以修复中文乱码。

## 2. 待办事项（前 5 项）

1. 继续为剩余的更新/删除 handler 添加 `RowsAffected` 检查，对不存在的资源返回 404。
2. 为 auth/device/telemetry/control 关键路径创建首批自动化测试。
3. 添加迁移脚本以填充 `metrics` 基线字典。
4. 确定并实现模板应用 + 告警订阅的策略。
5. 对齐并验证 MQTT/Influx 在开发/生产环境中的启动行为。

## 3. 阻碍 / 风险

- 仓库中目前没有自动化测试。
- 遥测数据采集依赖数据库中已存在的 `metrics` 行。
- 部分行为级 API 缺口目前被成功响应所掩盖。

## 4. 验证说明

- 如果本地 shell 中 Go 工具链不可用，则无法从终端验证编译/测试状态。
- 使用 `scripts/status-snapshot.ps1` 可快速刷新结构快照。

## 5. 下个会话如何继续

1. 阅读 `docs/PROJECT_STATUS.md`。
2. 阅读本文件（`docs/HANDOFF.md`）。
3. 从待办事项 #1 开始，除非优先级发生变化。

## 6. 快速填写模板（每次交接使用）

- 日期：
- 分支：
- 已完成范围：
- 待完成范围：
- 风险：
- 下个首要命令：
