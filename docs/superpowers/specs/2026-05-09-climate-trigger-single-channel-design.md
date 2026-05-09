# 气候联动触发源：固定单一采集通道（设计）

## 背景与问题

当前“气候联动（Climate Profiles）”自动调度器按 `trigger_metric_code` 匹配配置，并使用任意一条 `telemetry:received` 事件中的 `value` 进行阶段评估。由于不区分温室归属与具体采集通道：

- 不同温室上报相同 `metric_code`（例如 TEMP）会互相影响触发
- 同一温室多个温度采集器读数不同会互相“抢触发”，行为不可预期

## 目标

- 气候联动的触发依据固定为“某温室内某设备的某采集通道”的遥测值
- 用户在创建/编辑气候配置时必须指定：温室、设备、指标、采集通道
- 自动调度仅消费该采集通道的遥测事件，不受其他温室/其他探头影响
- 允许同一个采集通道被多个气候配置使用；同一事件命中多个配置时，全部独立评估并执行

## 非目标

- 不引入多探头聚合（最大/均值/中位数）触发口径
- 不做跨 Profile 的动作冲突合并/互斥（可能出现“动作打架”，由业务自行配置规避）
- 暂不做高吞吐缓存优化（可作为后续增强）

## 术语

- Profile：气候配置（climate_profiles）
- Stage：阶段（climate_stages）
- Action：阶段动作（climate_stage_actions）
- Trigger Channel：触发采集通道（sensor_channels.id）

## 数据模型变更（后端）

### 1) climate_profiles 新增字段

- `trigger_sensor_channel_id` BIGINT NOT NULL
- 索引：`idx_climate_profiles_trigger_sc(enabled, trigger_sensor_channel_id)` 或至少 `idx_climate_profiles_trigger_sc(trigger_sensor_channel_id)`

保留既有字段：

- `greenhouse_id`
- `trigger_metric_code`

说明：

- 设备不必落库：通过 `sensor_channels.sensor_device_id -> sensor_devices` 可反查设备与温室归属
- 保留 `greenhouse_id` 与 `trigger_metric_code` 是为了可读性与便于列表筛选；创建/更新时需与通道信息保持一致

### 2) 数据一致性校验（业务规则）

创建/更新 Profile 时，后端进行一致性校验：

- `sensor_channels.id == trigger_sensor_channel_id` 存在
- `sensor_channels.sensor_device_id` 对应的 `sensor_devices` 存在
- `sensor_devices.greenhouse_id == greenhouse_id`
- `sensor_channels.metric_code == trigger_metric_code`
- 若请求体携带 device 选择（可选）：校验 `sensor_channels.sensor_device_id == device_id`

校验失败返回 400（validation）。

## API 变更（后端）

### 请求/响应字段

对以下接口的请求/响应结构补充字段：

- Profile 创建/更新：新增 `trigger_sensor_channel_id`
- Profile 响应：返回 `trigger_sensor_channel_id`（建议同时返回便于展示的只读信息，例如 `trigger_device_code`、`trigger_channel_code`，是否增加由实现阶段再定）

### 向后兼容策略

两种选择（实现阶段二选一）：

- 严格升级（推荐）：迁移后所有 profile 必须补齐 `trigger_sensor_channel_id`，旧数据无法自动推导则需手工修复
- 过渡兼容：允许 `trigger_sensor_channel_id` 为空；调度器优先用通道匹配，空则回退 metric 逻辑（会继续存在跨温室影响，且复杂）

本设计默认采用“严格升级”。

## 自动调度器改造（后端）

### 触发匹配

- 订阅 `telemetry:received`
- 从事件数据获取 `sensor_channel_id`、`metric_code`、`value`
- 查询 `enabled=true AND trigger_sensor_channel_id = :sensor_channel_id` 的 profiles
- 对查询结果逐个执行：加载 stage/actions、计算最高命中阶段、切换阶段并执行 actions、写入执行日志

### 多 Profile 行为

- 允许同一 `sensor_channel_id` 命中多个 profiles：逐个执行，不做互斥
- cooldown 仍以 profile 为维度（同一 profile 的 60s 冷却不影响其它 profile）

## 采集通道禁用/删除语义（后端）

约定：不对 `trigger_sensor_channel_id` 建立数据库外键约束（便于独立运维通道数据），但在业务逻辑上维持“安全停用”。

- 当触发采集通道被设置为 `enabled=false`：系统自动将所有引用该通道的气候 Profile 置为 `enabled=false`
- 当触发采集通道被删除：系统自动将所有引用该通道的气候 Profile 置为 `enabled=false`，保留 `trigger_sensor_channel_id` 值不变
- 重要约束：如果采集通道是“硬删除并重建”（新的 `sensor_channels.id`），则旧 Profile 将永远无法再次触发；除非系统支持“软删除/恢复”或“禁止删除仅禁用”策略

## 前端交互设计

### Profile 表单：触发来源四项必填

创建/编辑时用户必须选择：

1. 温室（greenhouse）
2. 设备（sensor device，按温室过滤）
3. 指标（metric_code，按设备现有通道过滤或全量指标列表）
4. 采集通道（sensor channel，按设备 + 指标过滤）

提交核心字段：

- `greenhouse_id`
- `trigger_metric_code`
- `trigger_sensor_channel_id`

实现备注（选项加载策略）：

- 不新增后端“按 metric_code 过滤通道”的查询参数
- 前端按设备加载该设备的全部采集通道后，在本地按 `metric_code` 过滤候选项

### 列表展示建议

在 Profile 列表中展示触发来源（便于解释与排障）：

- 温室
- 设备编码/名称
- 通道编码
- 指标

## 迁移与数据处理

- 增加新列 `trigger_sensor_channel_id`（NOT NULL）
- 对现存数据：
  - 若历史 profile 无法映射到单一通道（例如之前仅配置 metric），需要业务选择一个默认通道并补齐
  - 若可推导（例如仅有一个通道满足 greenhouse+metric），可提供一次性迁移脚本自动填充（可选）

## 测试与验收

### 后端

- 创建/更新 Profile 校验：
  - 温室与通道归属不一致 → 400
  - 指标与通道 metric_code 不一致 → 400
- 调度器触发隔离：
  - A 温室通道上报 TEMP → 只触发绑定该通道的 profiles
  - B 温室上报同 TEMP → 不影响 A 温室 profiles
- 多 Profile 同通道：
  - 同一事件触发 2 个 profiles → 都写日志、都下发动作（可用 mock MQTT 或断开 MQTT 仅验证命令记录落库）

### 前端

- 表单联动过滤正确（温室→设备→指标→通道）
- 编辑时能回显已选通道并保持一致性

## 风险与约束

- “全部执行”在动作层面可能产生冲突，需要业务通过配置避免（例如不同 profile 控制同一执行器）
- 如果温室内通道数很多，前端下拉需要分页/搜索优化（可后续增强）

## 执行日志增强（后端）

为提升可观测性，在写入 `climate_execution_logs` 时补充记录触发来源信息（至少）：

- `trigger_sensor_channel_id`
- `trigger_metric_code`
- `collected_at`（来自遥测事件；若不可得则记录服务端接收时间）
