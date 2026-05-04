# 水培农植信息管理后端系统 需求分析（基础版）

本文档为毕业设计的第一版需求分析，重点覆盖核心功能与约束条件，后续可持续补充与细化。

## 1. 项目定位
系统面向温室/大棚等水培环境，通过后端统一接入传感器与执行器，采集环境数据并支持自动或手动控制，为前端展示与管理提供稳定接口。

## 2. 角色与使用者
- 系统管理员：设备与用户管理、阈值与策略配置、日志与告警管理
- 农户/操作员：查看环境数据、手动控制设备、查看告警
- 只读用户（可选）：查看授权数据

## 3. 核心功能需求（细化）
本节将核心功能拆分为“需求点 + 基本验收标准”，方便后续实现与测试。

### 3.1 设备接入与管理
1. 设备注册与识别  
需求点：支持传感器与执行器登记，包含设备类型、位置、所属温室、采样频率等字段。  
验收标准：能新增、编辑、禁用设备；设备在列表中可按类型与位置筛选。
2. 设备分组与绑定  
需求点：设备可绑定到某个温室或作物区，并支持分组管理。  
验收标准：设备可以在不同分组间移动，移动后数据归属随之变化。
3. 在线状态监测  
需求点：系统基于心跳或上报间隔判断设备在线/离线。  
验收标准：超过设定阈值未上报的设备显示为离线，并触发离线告警。
4. 数据格式统一  
需求点：不同设备上报的数据统一为系统内部标准字段与单位。  
验收标准：同一指标在数据库中单位一致，前端查询不需二次转换。

接口清单（建议 RESTful）  
- `POST /api/devices` 新增设备  
- `PUT /api/devices/{deviceId}` 更新设备  
- `PATCH /api/devices/{deviceId}/status` 启用/禁用设备  
- `GET /api/devices` 设备列表（支持类型、位置、分组筛选）  
- `GET /api/devices/{deviceId}` 设备详情  
- `POST /api/device-groups` 新增分组  
- `PUT /api/device-groups/{groupId}` 更新分组  
- `GET /api/device-groups` 分组列表  
- `POST /api/device-groups/{groupId}/devices/{deviceId}` 绑定设备到分组  
- `DELETE /api/device-groups/{groupId}/devices/{deviceId}` 从分组移除设备  
- `GET /api/devices/{deviceId}/health` 设备在线状态查询  

### 3.2 数据采集与存储
1. 实时数据接入  
需求点：支持 MQTT 或 HTTP 上报数据；支持周期上报与事件上报。  
验收标准：模拟设备持续上报时，系统可稳定接收并存储。
2. 数据清洗与异常过滤  
需求点：对缺失值、超出物理范围的数据进行过滤或标记。  
验收标准：异常值不影响统计结果，且可追溯原始数据。
3. 数据存储与查询  
需求点：高频环境数据可按设备、指标、时间范围查询。  
验收标准：可查询指定设备在指定时间段内的历史曲线数据。
4. 数据保留策略（基础）  
需求点：支持设置历史数据保留周期或归档策略。  
验收标准：超过保留期的数据可被归档或删除（可选实现）。

接口清单（建议 RESTful）  
- `POST /api/telemetry` HTTP 上报数据（备用通道）  
- `GET /api/telemetry/latest` 实时数据查询（按设备/指标）  
- `GET /api/telemetry/history` 历史数据查询（按设备/指标/时间范围）  
- `GET /api/telemetry/stats` 统计数据查询（均值/最大/最小）  
- `POST /api/telemetry/retention` 设置数据保留策略（可选）  

### 3.3 环境控制
1. 手动控制  
需求点：用户可对执行器下发开关或参数调整指令。  
验收标准：指令下发后状态可回显，且记录控制日志。
2. 自动控制（阈值规则）  
需求点：支持设置指标阈值与触发动作，如温度高于阈值开启风机。  
验收标准：模拟数据触发阈值后，系统自动下发控制指令。
3. 控制策略模板（可选）  
需求点：可按作物或生长阶段保存控制策略模板。  
验收标准：模板可被复用并快速应用到温室/分组。

接口清单（建议 RESTful）  
- `POST /api/controls/commands` 下发控制指令  
- `GET /api/controls/commands/{commandId}` 指令执行状态查询  
- `GET /api/controls/commands` 指令历史查询  
- `POST /api/controls/rules` 新增阈值规则  
- `PUT /api/controls/rules/{ruleId}` 更新阈值规则  
- `DELETE /api/controls/rules/{ruleId}` 删除阈值规则  
- `GET /api/controls/rules` 规则列表  
- `POST /api/controls/templates` 新增策略模板（可选）  
- `POST /api/controls/templates/{templateId}/apply` 应用策略模板（可选）  
- `GET /api/controls/templates` 模板列表（可选）  

### 3.4 告警与通知
1. 阈值告警  
需求点：环境指标超出范围时自动生成告警。  
验收标准：告警包含指标、触发值、时间与设备信息。
2. 设备故障告警  
需求点：设备离线或数据采集失败时触发告警。  
验收标准：离线后在可配置时间内触发告警且不会重复刷屏。
3. 告警处理闭环  
需求点：告警支持确认、忽略、关闭等处理状态。  
验收标准：告警状态可更新，历史记录可查询。

接口清单（建议 RESTful）  
- `GET /api/alerts` 告警列表（支持类型、级别、状态筛选）  
- `GET /api/alerts/{alertId}` 告警详情  
- `PATCH /api/alerts/{alertId}/status` 更新告警状态  
- `GET /api/alerts/stats` 告警统计  
- `GET /api/alerts/subscribe` 告警订阅信息（WebSocket/SSE 入口，可选）  

### 3.5 用户与权限管理
1. 账号登录与鉴权  
需求点：支持用户名/密码登录并返回 Token。  
验收标准：未授权请求被拒绝，已授权请求可访问接口。
2. 角色权限控制  
需求点：管理员、操作员、只读用户具备不同权限。  
验收标准：只读用户无法执行控制与配置操作。
3. 操作日志审计  
需求点：记录关键操作（登录、控制、配置变更）。  
验收标准：日志可按用户与时间筛选查询。

接口清单（建议 RESTful）  
- `POST /api/auth/login` 登录  
- `POST /api/auth/logout` 退出（可选）  
- `GET /api/users` 用户列表  
- `POST /api/users` 新增用户  
- `PUT /api/users/{userId}` 更新用户  
- `PATCH /api/users/{userId}/status` 启用/禁用用户  
- `GET /api/roles` 角色列表  
- `POST /api/roles` 新增角色  
- `PUT /api/roles/{roleId}` 更新角色  
- `GET /api/audit-logs` 操作日志查询  

### 3.6 前端接口支持
1. 数据查询接口  
需求点：提供实时数据与历史数据查询接口。  
验收标准：接口返回结构统一，支持分页与时间范围查询。
2. 控制指令接口  
需求点：提供设备控制指令下发接口。  
验收标准：接口返回指令受理状态与执行结果。
3. 告警与状态接口  
需求点：提供告警列表、设备状态与系统配置查询。  
验收标准：前端可完整展示设备与告警状态。

接口清单（前端视角汇总）  
- `GET /api/overview/dashboard` 系统概览（关键指标汇总）  
- `GET /api/devices` 设备与分组信息  
- `GET /api/telemetry/latest` 实时数据  
- `GET /api/telemetry/history` 历史数据  
- `POST /api/controls/commands` 控制指令  
- `GET /api/alerts` 告警列表  
- `GET /api/system/config` 系统配置（可选）  

## 4. 非功能需求（基础）
- 性能：支持分钟级采样与多设备并发写入
- 可靠性：断线重连、数据补传（可选）
- 安全性：JWT 认证、接口参数校验
- 可扩展性：新增设备类型与新指标可配置

## 5. 数据实体（初步）
- 用户、角色、权限
- 设备、传感器、执行器
- 环境数据（温湿度、光照、pH、EC、CO2）
- 控制指令
- 告警记录
- 系统配置

## 6. 约束与假设
- 当前阶段缺少真实设备，使用模拟数据验证功能
- 后端优先实现核心闭环：采集 -> 处理 -> 控制 -> 反馈
- 前端展示不作为本阶段重点，主要提供接口支持

## 7. 后续完善方向
- 数据分析与预测模型
- 多温室/多场地管理
- 权限细分与多租户支持
- 设备协议适配（MQTT/HTTP/Modbus）

## 8. 数据库设计（初稿）
本节给出关系型数据库的核心表设计，时序数据存储可选用 InfluxDB；如仅使用 MySQL，可在 `telemetry_data` 表中保存高频数据。

### 8.1 关系型核心表（MySQL）
1. 用户与权限
- `users`：用户基本信息  
字段：`id`, `username`, `password_hash`, `nickname`, `phone`, `email`, `status`, `created_at`, `updated_at`
- `roles`：角色定义  
字段：`id`, `name`, `description`, `created_at`, `updated_at`
- `user_roles`：用户角色关联  
字段：`id`, `user_id`, `role_id`
- `permissions`（可选）：权限点  
字段：`id`, `code`, `name`, `description`
- `role_permissions`（可选）：角色权限关联  
字段：`id`, `role_id`, `permission_id`

2. 设备与分组
- `greenhouses`：温室/大棚  
字段：`id`, `name`, `location`, `description`, `created_at`, `updated_at`
- `device_groups`：设备分组  
字段：`id`, `greenhouse_id`, `name`, `description`, `created_at`, `updated_at`
- `devices`：设备信息  
字段：`id`, `device_code`, `name`, `type`, `category`, `greenhouse_id`, `group_id`, `status`, `protocol`, `last_seen_at`, `created_at`, `updated_at`  
说明：`type` 如 `SENSOR`/`ACTUATOR`，`category` 如 `TEMP`/`HUMIDITY`/`PUMP` 等。

3. 数据与指标
- `metrics`：指标字典  
字段：`id`, `code`, `name`, `unit`, `min_value`, `max_value`, `created_at`, `updated_at`
- `telemetry_data`：采集数据（如不使用 InfluxDB）  
字段：`id`, `device_id`, `metric_id`, `value`, `collected_at`, `created_at`
- `device_metrics`（可选）：设备支持的指标  
字段：`id`, `device_id`, `metric_id`

4. 控制与规则
- `control_commands`：控制指令  
字段：`id`, `device_id`, `command_type`, `payload`, `status`, `sent_at`, `executed_at`, `created_by`, `created_at`
- `control_rules`：阈值规则  
字段：`id`, `name`, `metric_id`, `operator`, `threshold`, `action`, `target_device_id`, `enabled`, `created_by`, `created_at`, `updated_at`
- `control_templates`（可选）：策略模板  
字段：`id`, `name`, `description`, `content`, `created_by`, `created_at`, `updated_at`

5. 告警与日志
- `alerts`：告警记录  
字段：`id`, `type`, `level`, `metric_id`, `device_id`, `value`, `message`, `status`, `triggered_at`, `resolved_at`
- `audit_logs`：操作日志  
字段：`id`, `user_id`, `action`, `target_type`, `target_id`, `detail`, `created_at`
- `system_configs`：系统配置  
字段：`id`, `config_key`, `config_value`, `description`, `updated_at`

### 8.2 关键约束与索引（建议）
- `devices.device_code` 唯一索引  
- `telemetry_data`：`(device_id, metric_id, collected_at)` 复合索引  
- `alerts`：`(status, triggered_at)` 复合索引  
- `control_commands`：`(device_id, created_at)` 复合索引  

### 8.3 时序数据库（InfluxDB，可选）
- 测量表：`telemetry`  
标签：`device_id`, `metric_code`, `greenhouse_id`  
字段：`value`  
时间：`collected_at`
