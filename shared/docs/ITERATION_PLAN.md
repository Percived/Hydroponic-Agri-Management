# 水培农业管理系统 — 迭代方案

> 本文档为前后端开发团队的同步依据。每个迭代阶段包含：需求背景、后端任务、前端任务、API 契约、数据模型变更。
> 约定：迭代按 Phase 顺序执行，Phase 1 完成并验证后再进入 Phase 2，以此类推。

---

## Phase 1：实时推送 + InfluxDB 查询 + 仪表盘增强

**目标**：修复系统三大断层中的前两个——数据"采而不用"、告警"告而不达"的基础设施建设。

### 1.1 实时告警推送（SSE）

**背景**：当前告警订阅 API 为占位实现，前端靠轮询获取告警数据，延迟高且浪费资源。

#### 后端任务
- 实现 `GET /api/alerts/subscribe` 的 SSE 推送能力
  - 使用 `text/event-stream` 响应头
  - 每 5 秒检查一次是否有新告警（或通过 channel 在告警创建时主动通知 SSE 连接）
  - 推送数据格式：
    ```json
    {
      "type": "new_alert",
      "data": { "id": 1, "level": "CRITICAL", "message": "温度过高: 35.2°C", ... }
    }
    ```
  - 支持 `?device_id=&level=` 过滤参数
  - 心跳保活：每 30 秒发一次 `:keepalive` 注释

#### 前端任务
- 新建 `composables/useAlertSSE.ts`，封装 EventSource 连接/重连逻辑
- 在 `AppLayout.vue` 的 header 中显示"实时告警提醒"——收到 CRITICAL 告警时闪烁红点 + 浏览器通知（Notification API）
- SSE 断开时自动重连，最长重连间隔 30 秒

---

### 1.2 遥测实时推送（SSE）

**背景**：实时遥测页面目前靠定时轮询 `GET /api/telemetry/latest`，数据新鲜度取决于轮询间隔。

#### 后端任务
- 新增 `GET /api/telemetry/subscribe` SSE 端点
  - 参数：`?device_code=D1,D2,D3`（可选，不传则推送所有设备）
  - 参数：`?metric_code=TEMP,HUMIDITY`（可选，不传则推送所有指标）
  - 遥测写入 InfluxDB/MySQL 后，通过内部 channel 广播给所有 SSE 连接
  - 推送格式：
    ```json
    {
      "device_code": "DEV-001",
      "collected_at": "2026-05-04T10:30:00Z",
      "metrics": [
        { "code": "TEMP", "value": 25.3, "unit": "°C" }
      ]
    }
    ```

#### 前端任务
- 新建 `composables/useTelemetrySSE.ts`
- 重构 `views/telemetry/realtime.vue`：页面打开时建立 SSE 连接，收到数据后实时更新图表和卡片，不再依赖定时轮询
- 当用户选择了特定设备/指标时，SSE 连接参数同步变更

---

### 1.3 InfluxDB 查询迁移

**背景**：InfluxDB 仅做写入，所有遥测查询走 MySQL。时序数据库的高性能聚合能力完全浪费。

#### 后端任务
- 新增 `GET /api/telemetry/history` 的 InfluxDB 查询路径
  - 当 `raw_value=false`（默认）时使用 InfluxDB 查询
  - 当 `raw_value=true` 或 InfluxDB 不可用时回退 MySQL
  - 利用 Flux 的 `range() + filter() + limit()` 实现分页
- 新增 `GET /api/telemetry/stats` 的 InfluxDB 查询路径
  - 使用 `aggregateWindow()` 实现高效聚合（avg/max/min）
  - 支持 `?window=1h|6h|1d` 参数控制聚合窗口
- `POST /api/telemetry` 双写保持不变

#### 前端任务
- 无需改动，API 契约不变，透明替换

---

### 1.4 仪表盘增强

**背景**：当前仪表盘只显示 3 个数字，信息密度太低。

#### 后端任务
- 增强 `GET /api/overview/dashboard` 返回数据结构：

```json
{
  "code": 0,
  "data": {
    "devices_online": 12,
    "devices_offline": 3,
    "devices_total": 15,
    "alerts_open": 5,
    "alerts_critical": 1,
    "alerts_today": 8,
    "device_type_distribution": [
      { "type": "SENSOR", "count": 10 },
      { "type": "ACTUATOR", "count": 5 }
    ],
    "greenhouse_summary": [
      {
        "greenhouse_id": 1,
        "name": "A区大棚",
        "device_count": 6,
        "avg_temp": 24.5,
        "avg_humidity": 62.3
      }
    ],
    "recent_commands": [
      { "id": 1, "command_type": "IRRIGATE", "device_name": "水泵1", "status": "EXECUTED", "created_at": "..." }
    ],
    "online_rate_trend": [
      { "hour": "08:00", "rate": 0.93 },
      { "hour": "09:00", "rate": 0.87 }
    ]
  }
}
```

#### 前端任务
- 重构 `views/dashboard/index.vue`：
  - 顶部 4 个统计卡片：设备在线/总数、未处理告警、今日告警、在线率
  - 左侧：设备类型分布饼图（ECharts）
  - 中间：告警趋势折线图（近 24 小时，按小时聚合）
  - 右侧：温室环境概览卡片列表
  - 底部：最近控制命令时间线
- 新建共享组件 `components/charts/PieChart.vue`、`components/charts/LineChart.vue`

---

## Phase 2：设备深度看板 + 批量操作 + 通知渠道

**目标**：提升设备管理深度、操作效率、告警触达能力。

### 2.1 设备深度看板

**背景**：设备详情页目前只有基础信息 + 健康状态 + 最新遥测值，缺乏趋势分析和异常标记。

#### 后端任务
- 新增 `GET /api/devices/:deviceId/telemetry-summary` 端点
  - 返回该设备所有指标在选定时间范围内的统计摘要和小时级数据
  ```json
  {
    "code": 0,
    "data": {
      "device_id": 1,
      "from": "2026-05-03T00:00:00Z",
      "to": "2026-05-04T00:00:00Z",
      "metrics": {
        "TEMP": {
          "avg": 24.5, "max": 28.2, "min": 20.1,
          "alerts": 3,
          "hourly": [
            { "hour": "00:00", "avg": 22.1 },
            { "hour": "01:00", "avg": 21.8 }
          ]
        }
      },
      "online_rate": 0.95,
      "total_online_hours": 22.8,
      "alert_events": [
        { "id": 10, "level": "WARN", "message": "...", "triggered_at": "..." }
      ]
    }
  }
  ```

#### 前端任务
- 重构 `views/devices/detail.vue`：
  - 新增指标选择器（多选），默认全选
  - 新增时间范围选择器（今天/昨天/近7天/自定义）
  - 每个选中指标的 24 小时趋势图（小图）
  - 异常时段用红色竖线/区域标记告警发生时间
  - 底部：设备运行统计（在线率、累计在线时长）

---

### 2.2 批量操作

**场景**：运维人员需要对整个温室或设备组执行统一操作。

#### 后端任务
- 新增 `POST /api/controls/batch-commands` 端点
  - 请求体：
    ```json
    {
      "target_type": "greenhouse",     // 或 "device_group" 或 "devices"
      "target_ids": [1],
      "command_type": "SWITCH",
      "payload": { "state": "ON" },
      "remark": "批量开启A区风机"
    }
    ```
  - 后端校验目标设备是否在线，逐个下发命令
  - 返回每个设备的执行结果摘要：
    ```json
    {
      "code": 0,
      "data": {
        "total": 6,
        "success": 5,
        "failed": 1,
        "results": [
          { "device_id": 1, "device_name": "风机1", "command_id": 100, "status": "SENT" },
          { "device_id": 2, "device_name": "风机2", "command_id": 101, "status": "FAILED", "reason": "设备离线" }
        ]
      }
    }
    ```
- 新增 `POST /api/devices/batch-update` 端点
  - 批量修改采样间隔、分组归属等
- 新增 `DELETE /api/devices/batch` 端点
  - 批量删除设备（仅 ADMIN）
  - 要求提供删除原因 `reason`

#### 前端任务
- 设备列表页新增批量选择模式（复选框 + 全选）
- 工具栏新增"批量操作"下拉按钮：批量启用/禁用、批量修改采样间隔、批量下发命令、批量删除
- 批量命令下发弹窗：确认设备列表、命令类型选择、参数填写、备注（必填）
- 操作结果弹窗：展示每个设备的执行结果，失败项标红并显示原因

---

### 2.3 通知渠道

**背景**：告警只在系统内可见，用户离线时无法及时获知。

#### 后端任务
- 新增数据库表 `notification_channels`：
  ```sql
  CREATE TABLE notification_channels (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    channel_type ENUM('EMAIL','SMS','WEBHOOK') NOT NULL,
    channel_config JSON NOT NULL COMMENT '{"email":"","phone":"","webhook_url":"","secret":""}',
    enabled TINYINT(1) DEFAULT 1,
    min_alert_level ENUM('INFO','WARN','CRITICAL') DEFAULT 'WARN',
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
  );
  ```
- 实现 `POST/GET/PUT/DELETE /api/notification-channels` CRUD
- 告警触发后（`telemetry/handler.go` 的 `evaluateAndTrigger()`）新增通知流程：
  1. 查询所有启用的通知渠道
  2. 过滤：告警级别 >= 渠道 `min_alert_level`
  3. 异步发送通知（goroutine + channel 队列，避免阻塞）
  4. Webhook 格式：POST JSON 到配置的 URL，包含 `secret` 签名头
- Webhook 推送格式：
  ```json
  {
    "type": "alert",
    "alert_id": 1,
    "level": "CRITICAL",
    "device_name": "温湿度传感器1",
    "greenhouse_name": "A区大棚",
    "metric_name": "温度",
    "value": 35.2,
    "threshold": 30.0,
    "message": "温度过高: 35.2°C > 30°C",
    "triggered_at": "2026-05-04T10:30:00Z"
  }
  ```

#### 前端任务
- 新增 `views/settings/notification-channels.vue`（ADMIN 菜单下）
  - 用户通知渠道列表
  - 新增渠道弹窗：选择类型（邮件/短信/Webhook），填写配置，选择最低告警级别
  - 支持测试发送（"发送测试通知"按钮，仅 Webhook 类型）
- 新建 `api/notification.ts`、`types/notification.ts`
- 侧边栏菜单新增"通知设置"项

---

### 2.4 系统配置页面（补全 PRD 阶段 3）

**背景**：`system_configs` 表已存在，但无前端管理页面。

#### 后端任务
- 新增 `GET/PUT /api/system-configs` 端点
  - GET 返回所有配置项（敏感字段如 `jwt_secret` 脱敏）
  - PUT 批量更新配置（仅 ADMIN）
  - 记录审计日志

#### 前端任务
- 新增 `views/settings/system-config.vue`（ADMIN）
  - 表格展示所有配置项（key、value、description）
  - 点击编辑修改单个配置
  - 数据保留天数、MQTT broker 地址等运维配置

---

## Phase 3：统计报表 + 移动端适配 + 操作安全

**目标**：支撑管理决策、移动场景覆盖、降低误操作风险。

### 3.1 统计报表

**背景**：管理者需要周期性回顾生产数据，当前只能逐条查看，无法生成报告。

#### 后端任务
- 新增 `GET /api/reports/overview` 端点
  - 参数：`?greenhouse_id=&from=&to=&granularity=hour|day|week`
  - 返回所有指标在该范围内的统计 + 时序数据
- 新增 `POST /api/reports/export` 端点
  - 参数：`format=csv|pdf`，`from`，`to`，`greenhouse_id`（可选）
  - CSV：直接流式返回文件
  - PDF：使用 `go-wkhtmltopdf` 或生成 HTML 后转 PDF
- 新增 `GET /api/reports/greenhouse-comparison` 端点
  - 参数：`?greenhouse_ids=1,2,3&metric_codes=TEMP,HUMIDITY&from=&to=`
  - 返回每个温室在该指标上的 avg/max/min，供横向对比

#### 前端任务
- 新增 `views/reports/index.vue`（所有角色可访问）
  - 选择温室（可选全部）、时间范围、指标
  - 指标趋势图（多线对比）
  - 温室对比表格
  - "导出 CSV"按钮
- 新增 `views/reports/greenhouse-comparison.vue`
  - 柱状图展示多温室同指标对比
  - 表格展示详细对比数据

---

### 3.2 移动端适配

**背景**：管理者需要在手机上快速查看大棚状态。

#### 前端任务
- 核心页面移动端布局适配（采用响应式断点 `768px`）：
  - `dashboard`：卡片单列排列，图表全宽
  - `alerts`：列表项简化，突出告警级别颜色标签
  - `devices/detail`：标签页切换替代左右分栏
  - `telemetry/realtime`：一次只显示一个指标，左右滑动切换
- 新建 `composables/useMobile.ts`，提供 `isMobile` 响应式变量
- 底部导航栏（移动端替代侧边栏）：仪表盘、设备、告警、我的
- Element Plus 组件已在 PRD 中指定，其内置响应式能力可充分利用

#### 后端任务
- 无需改动

---

### 3.3 操作安全机制

**背景**：关键误操作（如错误关闭灌溉）可能造成生产损失。

#### 后端任务
- `POST /api/controls/commands` 新增 `remark` 字段（必填）
- 新增两步验证接口（可选）：
  - `POST /api/controls/commands/confirm/:commandId` —— 关键命令需二次确认才下发
  - 命令状态新增 `PENDING_CONFIRM`，需另一 OPERATOR 角色用户确认后方可执行
- 命令超时处理：
  - 后台 goroutine 每分钟扫描状态为 `SENT` 超过 N 分钟未变为 `EXECUTED` 的命令
  - 自动标记为 `FAILED`，生成一条告警

#### 前端任务
- 命令下发弹窗新增"操作备注"输入框（必填，最少 5 字）
- 新增命令确认弹窗：创建命令后弹窗显示"命令已创建，等待确认执行"，显示确认状态
- 设备删除按钮新增二次确认弹窗：输入设备名称以确认删除

---

## Phase 4：智能预警 + 自动化编排 + 能耗管理

**目标**：从"被动响应"升级为"主动预防"，降低人工干预频率。

### 4.1 智能预警（动态基线）

**背景**：固定阈值无法适应季节变化和作物生长阶段。

#### 后端任务
- 新增 `POST /api/alerts/baselines/calculate` 端点
  - 根据过去 30 天数据计算每个设备/指标的正常范围（均值 ± 3σ）
  - 结果存储到新表 `metric_baselines`：
    ```sql
    CREATE TABLE metric_baselines (
      id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
      device_id BIGINT UNSIGNED NOT NULL,
      metric_id BIGINT UNSIGNED NOT NULL,
      baseline_avg DOUBLE NOT NULL,
      baseline_stddev DOUBLE NOT NULL,
      upper_bound DOUBLE NOT NULL,
      lower_bound DOUBLE NOT NULL,
      calculated_at DATETIME(3) NOT NULL,
      UNIQUE KEY uk_device_metric (device_id, metric_id)
    );
    ```
- 规则引擎新增 `operator = DEVIATION` 类型
  - 当 `abs(value - baseline_avg) / baseline_stddev > threshold` 时触发
  - 与固定阈值规则并行，互不替代
- 新增定时任务（cron）：每天凌晨 2 点自动重新计算所有基线

#### 前端任务
- 规则创建/编辑表单新增"动态基线"选项
  - 选择 `DEVIATION` 操作符后，阈值输入变为"偏离标准差倍数"
  - 显示当前设备/指标的参考基线值（avg ± Nσ）
- 告警详情中标注"动态基线告警" vs "固定阈值告警"

---

### 4.2 自动化场景编排

**背景**：单条件单动作规则无法满足复杂场景需求。

#### 后端任务
- 新增 `control_scenes` 表：
  ```sql
  CREATE TABLE control_scenes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    enabled TINYINT(1) DEFAULT 1,
    steps JSON NOT NULL COMMENT '[{"order":1,"type":"condition","metric_id":1,"operator":">","threshold":30},{"order":2,"type":"action","device_id":2,"command_type":"SWITCH","payload":{"state":"ON"},"delay_sec":0},{"order":3,"type":"wait","duration_sec":300},{"order":4,"type":"condition"...}]',
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
  );
  ```
- 场景引擎：后台 goroutine 轮询已启用场景，遇到条件满足时按 step 顺序执行
  - 支持 `condition`（检测条件）、`action`（执行命令）、`wait`（等待 N 秒）
  - 支持 `goto`（条件不满足时跳转到其他 step）
- 新增 `POST/PUT/DELETE/GET /api/controls/scenes` CRUD
- 新增 `POST /api/controls/scenes/:sceneId/toggle` 启用/禁用场景

#### 前端任务
- 新增 `views/controls/scenes.vue`
  - 场景列表（名称、步骤数、启用状态）
  - 场景编辑器：步骤拖拽排序，每步可选择类型（条件检测 / 执行动作 / 等待），参数配置
  - 场景执行日志：最近 20 次执行记录
- 预制场景模板（种子数据）：
  - "高温降温"：TEMP > 32 → 开风机 → 等 5 分钟 → TEMP > 32 → 开湿帘 → TEMP < 28 → 关湿帘 → 关风机
  - "低 CO2 补气"：CO2 < 400 → 开 CO2 发生器 → 等 10 分钟 → CO2 > 600 → 关闭

---

### 4.3 能耗管理

**背景**：电力是大棚运营的主要成本之一，缺乏能耗数据无法优化。

#### 后端任务
- 新增 `device_power_records` 表：
  ```sql
  CREATE TABLE device_power_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    device_id BIGINT UNSIGNED NOT NULL,
    power_watts DOUBLE NOT NULL COMMENT '额定功率(W)',
    running_hours DOUBLE NOT NULL COMMENT '运行时长(h)',
    energy_kwh DOUBLE NOT NULL COMMENT '耗电量(kWh)',
    record_date DATE NOT NULL,
    UNIQUE KEY uk_device_date (device_id, record_date)
  );
  ```
- `devices` 表新增 `power_watts` 字段（额定功率）
- 新增 `POST /api/devices/:deviceId/power-config` 设置额定功率
- `control_commands` 的 EXECUTED 事件记录设备运行时间
- 新增 `GET /api/energy/summary` 端点
  - 参数：`?greenhouse_id=&from=&to=&granularity=day|week|month`
  - 返回：总耗电量、各温室耗电占比、各设备类型耗电占比、每日耗电趋势

#### 前端任务
- 新增 `views/energy/index.vue`
  - 时间范围选择器
  - 总耗电量卡片（含环比增减百分比）
  - 温室耗电占比饼图
  - 设备类型耗电柱状图
  - 每日耗电趋势折线图
  - 用电异常标记（当日耗电超过平均值 2 倍时高亮）
- 新增 `views/energy/cost-analysis.vue`
  - 配置电价（峰/谷/平），计算电费
  - 谷电时段用电建议
- 侧边栏新增"能耗管理"菜单项

---

## Phase 5：视频集成 + 种植方案管理 + 数据预测

**目标**：系统从"环境管理"扩展到"生产管理"，形成完整闭环。

### 5.1 视频监控集成

#### 后端任务
- 新增 `cameras` 表：
  ```sql
  CREATE TABLE cameras (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    greenhouse_id BIGINT UNSIGNED,
    stream_url VARCHAR(500) NOT NULL COMMENT 'RTSP/HLS 地址',
    snapshot_url VARCHAR(500) COMMENT '截图 API 地址',
    enabled TINYINT(1) DEFAULT 1,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3)
  );
  ```
- 新增 `POST/GET/PUT/DELETE /api/cameras` CRUD
- 告警触发时异步调用摄像头截图 API，截图存储路径关联到告警记录
- `alerts` 表新增 `snapshot_url` 字段

#### 前端任务
- 新增 `views/cameras/index.vue`
  - 摄像头卡片网格布局
  - 使用 HLS.js 播放实时视频流
  - 点击卡片打开全屏播放模态框
- 告警详情中展示告警触发时的截图（如有）
- 仪表盘新增"实时监控"卡片（可选展示）

---

### 5.2 种植方案管理

**背景**：不同作物需要不同环境参数，需要结构化存储和复用。

#### 后端任务
- 新增 `crop_varieties` 表：
  ```sql
  CREATE TABLE crop_varieties (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '品种名称',
    description TEXT,
    growing_period_days INT COMMENT '生长周期(天)',
    ideal_params JSON NOT NULL COMMENT '[{"metric_code":"TEMP","min":20,"max":28},{"metric_code":"HUMIDITY","min":60,"max":80}]',
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3)
  );
  ```
- 新增 `planting_batches` 表：
  ```sql
  CREATE TABLE planting_batches (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '批次名称',
    greenhouse_id BIGINT UNSIGNED NOT NULL,
    variety_id BIGINT UNSIGNED NOT NULL,
    planted_at DATE NOT NULL,
    expected_harvest_at DATE,
    harvested_at DATE,
    status ENUM('GROWING','HARVESTED','FAILED') DEFAULT 'GROWING',
    notes TEXT,
    created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3)
  );
  ```
- 新增 `POST/GET/PUT/DELETE /api/crops/varieties` CRUD
- 新增 `POST/GET/PUT/DELETE /api/crops/batches` CRUD
- 新增 `GET /api/crops/batches/:batchId/environment-analysis`
  - 将种植期间的实际环境数据与 `ideal_params` 对比
  - 返回偏差分析：哪些指标超出理想范围、超出的时长比例

#### 前端任务
- 新增 `views/crops/varieties.vue`
  - 品种列表 + 新增/编辑弹窗
  - 理想参数配置 UI（指标多选 + min/max 范围滑块）
- 新增 `views/crops/batches.vue`
  - 批次列表，按状态筛选
  - 新增批次：选择温室 → 选择品种 → 种植日期 → 预计收获日期
  - 批次详情：环境偏差分析图表（实际值 vs 理想范围对比）
  - 批次状态流转（生长中 → 已收获/失败）
- 侧边栏新增"种植管理"分组，包含品种管理和批次管理

---

### 5.3 趋势预测

**背景**：基于历史数据预测未来趋势，提前采取预防措施。

#### 后端任务
- 新增 `GET /api/telemetry/predict` 端点
  - 参数：`?device_id=&metric_code=&horizon_hours=6`
  - 算法：简单移动平均 + 线性回归外推（Phase 5 使用简单算法，后续可替换为 ML 模型）
  - 返回：
    ```json
    {
      "code": 0,
      "data": {
        "device_id": 1,
        "metric_code": "TEMP",
        "predictions": [
          { "timestamp": "2026-05-04T11:00:00Z", "value": 25.1, "lower": 24.5, "upper": 25.7 },
          { "timestamp": "2026-05-04T12:00:00Z", "value": 25.8, "lower": 24.9, "upper": 26.7 }
        ],
        "confidence": 0.85,
        "trend": "RISING"
      }
    }
    ```
- 可选：集成外部天气 API 数据作为预测辅助因子（温度、湿度、光照）

#### 前端任务
- 在设备详情趋势图中叠加预测数据（虚线）
- 预测置信区间用半透明色带填充
- 趋势判断标签：上升↑ / 下降↓ / 平稳→

---

## Phase 6：多租户 + 开放 API 生态（长期规划）

**目标**：支持集团化运营，对外提供能力输出。

### 6.1 多基地管理

#### 后端任务
- 新增 `organizations`（基地/组织）表
- 所有资源表（devices, greenhouses, users 等）新增 `org_id` 外键
- 中间件新增组织隔离逻辑：用户只能查看所属组织的数据
- 新增 `SUPER_ADMIN` 角色，可跨组织查看
- 新增 `POST/GET /api/organizations` CRUD，SUPER_ADMIN 管理

#### 前端任务
- SUPER_ADMIN 登录后展示组织选择器
- 跨组织总览仪表盘（各基地对比）

### 6.2 开放 API

- API Key 管理（`POST /api/api-keys`，生成用于第三方调用的 API Key）
- 速率限制按 API Key 配置
- 开放 API 文档站点（基于 OpenAPI spec 自动生成）

---

## 数据结构变更汇总

| Phase | 新表 | 变更表 |
|-------|------|--------|
| 2 | `notification_channels` | — |
| 3 | — | `control_commands` 加 `remark`，`alerts` 加 `confirmed_by` |
| 4 | `metric_baselines`, `control_scenes`, `device_power_records` | `devices` 加 `power_watts` |
| 5 | `cameras`, `crop_varieties`, `planting_batches` | `alerts` 加 `snapshot_url` |
| 6 | `organizations`, `api_keys` | 所有资源表加 `org_id` |

---

## 前端路由 & 菜单变更汇总

| Phase | 新增路由 | 新增菜单位置 |
|-------|----------|-------------|
| 1 | —（增强现有页面） | — |
| 2 | `/settings/notification-channels`, `/settings/system-config` | 系统设置分组 |
| 3 | `/reports`, `/reports/comparison` | 数据报表 |
| 4 | `/controls/scenes`, `/energy`, `/energy/cost` | 控制场景、能耗管理 |
| 5 | `/cameras`, `/crops/varieties`, `/crops/batches` | 视频监控、种植管理分组 |
| 6 | —（增强现有路由） | 组织切换器（顶部） |

---

## API 端点变更汇总

| Phase | 方法 | 路径 | 说明 |
|-------|------|------|------|
| 1 | GET | `/api/alerts/subscribe` | 占位 → SSE 实现 |
| 1 | GET | `/api/telemetry/subscribe` | 新增，SSE 遥测推送 |
| 1 | GET | `/api/overview/dashboard` | 增强返回字段 |
| 1 | GET | `/api/telemetry/history` | 改为 InfluxDB 查询 |
| 1 | GET | `/api/telemetry/stats` | 改为 InfluxDB 查询，新增 `window` 参数 |
| 2 | GET | `/api/devices/:deviceId/telemetry-summary` | 新增 |
| 2 | POST | `/api/controls/batch-commands` | 新增 |
| 2 | POST | `/api/devices/batch-update` | 新增 |
| 2 | DELETE | `/api/devices/batch` | 新增 |
| 2 | CRUD | `/api/notification-channels` | 新增 |
| 2 | GET/PUT | `/api/system-configs` | 新增 |
| 3 | GET | `/api/reports/overview` | 新增 |
| 3 | POST | `/api/reports/export` | 新增 |
| 3 | GET | `/api/reports/greenhouse-comparison` | 新增 |
| 3 | POST | `/api/controls/commands/confirm/:commandId` | 新增 |
| 4 | POST | `/api/alerts/baselines/calculate` | 新增 |
| 4 | CRUD | `/api/controls/scenes` | 新增 |
| 4 | GET | `/api/energy/summary` | 新增 |
| 5 | CRUD | `/api/cameras` | 新增 |
| 5 | CRUD | `/api/crops/varieties` | 新增 |
| 5 | CRUD | `/api/crops/batches` | 新增 |
| 5 | GET | `/api/crops/batches/:batchId/environment-analysis` | 新增 |
| 5 | GET | `/api/telemetry/predict` | 新增 |
| 6 | CRUD | `/api/organizations` | 新增 |
| 6 | CRUD | `/api/api-keys` | 新增 |

---

## 实施约定

1. **前后端同步**：每个 Phase 开始时，前后端负责人共同阅读本方案对应章节，确认无歧义后开工。
2. **API 先行**：后端先提供 API 的 Mock 响应（或 Swagger 文档），前端可并行开发。
3. **契约更新**：API 变更后，同步更新 `shared/docs/API_SPEC.md` 和 `shared/docs/openapi.yaml`。
4. **种子数据**：新增表需要配套的迁移 SQL 文件（`migrations/` 目录下按序号命名）。
5. **审计日志**：所有 CRUD 的 C/U/D 操作必须记录审计日志。
6. **测试**：每个新端点至少覆盖 1 个正常场景和 1 个异常场景的测试。
7. **Phase 验收标准**：Phase 内所有功能开发完成 + 前后端联调通过 + 无 P0 阻塞 Bug → 进入下一 Phase。
