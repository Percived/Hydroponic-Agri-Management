# API 规格说明（v2.0.0）

适用范围：后端 v2.0.0 重构后完整接口文档（更新日期 2026-05-06）

## 1. 通用约定

Base URL：`/api`

内容类型：`application/json; charset=utf-8`

时间格式：ISO 8601 UTC，例如 `2026-02-11T08:30:00Z`

ID 类型：`BIGINT`，响应中以数字返回

### 1.1 鉴权

鉴权方式：`Authorization: Bearer <JWT>`

角色：`ADMIN`、`OPERATOR`、`VIEWER`

权限规则：

- ADMIN 可访问所有接口
- OPERATOR 可访问查询与控制接口
- VIEWER 仅查询类接口

### 1.2 统一响应格式

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "request_id": "req_7f9c2b"
}
```

错误响应：

```json
{
  "code": 10001,
  "message": "validation_error",
  "data": {
    "errors": [{ "field": "device_code", "reason": "required" }]
  },
  "request_id": "req_7f9c2b"
}
```

### 1.3 业务错误码

| code  | 含义                | 对应 HTTP |
| ----- | ------------------- | --------- |
| 0     | 成功                | 200/201   |
| 10001 | 参数校验失败        | 400       |
| 10002 | 未登录或 Token 无效 | 401       |
| 10003 | 权限不足            | 403       |
| 10004 | 资源不存在          | 404       |
| 10005 | 资源冲突或重复      | 409       |
| 10006 | 频率限制            | 429       |
| 10007 | 设备离线            | 409       |
| 10008 | 规则冲突            | 409       |
| 10009 | 数据超出物理范围    | 422       |
| 10010 | 设备编码重复        | 409       |

### 1.4 分页格式

列表响应 `data`：

```json
{
  "page": 1,
  "page_size": 20,
  "total": 120,
  "items": []
}
```

## 2. 数据模型（响应结构）

### 2.1 用户与角色

#### User

| 字段       | 类型   | 说明                   | 示例                      |
| ---------- | ------ | ---------------------- | ------------------------- |
| id         | number | 用户 ID                | 1                         |
| username   | string | 用户名                 | "admin"                   |
| nickname   | string | 昵称                   | "管理员"                  |
| phone      | string | 电话                   | "13800138000"             |
| email      | string | 邮箱                   | "admin@example.com"       |
| status     | string | 状态：ENABLED/DISABLED | "ENABLED"                 |
| roles      | Role[] | 角色列表               | [{"id":1,"name":"ADMIN"}] |
| created_at | string | 创建时间               | "2026-01-01T00:00:00Z"    |
| updated_at | string | 更新时间               | "2026-01-01T00:00:00Z"    |

#### Role

| 字段        | 类型   | 说明                          | 示例                   |
| ----------- | ------ | ----------------------------- | ---------------------- |
| id          | number | 角色 ID                       | 1                      |
| name        | string | 角色名：ADMIN/OPERATOR/VIEWER | "ADMIN"                |
| description | string | 描述                          | "系统管理员"           |
| created_at  | string | 创建时间                      | "2026-01-01T00:00:00Z" |
| updated_at  | string | 更新时间                      | "2026-01-01T00:00:00Z" |

### 2.2 温室与种植区

#### Greenhouse

| 字段        | 类型   | 说明                   | 示例                   |
| ----------- | ------ | ---------------------- | ---------------------- |
| id          | number | 温室 ID                | 1                      |
| code        | string | 温室编码               | "GH-001"               |
| name        | string | 名称                   | "一号温室"             |
| location    | string | 位置描述               | "A区"                  |
| area_sqm    | number | 面积（平方米）         | 500.00                 |
| description | string | 描述                   | "叶菜专用温室"         |
| status      | string | 状态：ENABLED/DISABLED | "ENABLED"              |
| created_at  | string | 创建时间               | "2026-01-01T00:00:00Z" |
| updated_at  | string | 更新时间               | "2026-01-01T00:00:00Z" |
| zone_count  | number | 种植区数量（含关联时） | 4                      |

#### GrowingZone

| 字段                     | 类型   | 说明                   | 示例                   |
| ------------------------ | ------ | ---------------------- | ---------------------- |
| id                       | number | 种植区 ID              | 1                      |
| greenhouse_id            | number | 所属温室 ID            | 1                      |
| code                     | string | 种植区编码             | "ZONE-A1"              |
| name                     | string | 名称                   | "A1区"                 |
| system_type              | string | 种植系统：DWC/NFT      | "DWC"                  |
| tank_volume_liter        | number | 水箱容积（升）         | 200.00                 |
| planting_density_per_sqm | number | 种植密度（株/平方米）  | 25.00                  |
| status                   | string | 状态：ENABLED/DISABLED | "ENABLED"              |
| created_at               | string | 创建时间               | "2026-01-01T00:00:00Z" |
| updated_at               | string | 更新时间               | "2026-01-01T00:00:00Z" |

### 2.3 传感器设备与通道

#### SensorDevice

| 字段             | 类型         | 说明                       | 示例                   |
| ---------------- | ------------ | -------------------------- | ---------------------- |
| id               | number       | 设备 ID                    | 1                      |
| greenhouse_id    | number       | 所属温室 ID                | 1                      |
| growing_zone_id  | number\|null | 所属种植区 ID              | 1                      |
| device_code      | string       | 设备编码（唯一）           | "SENSOR-001"           |
| name             | string       | 名称                       | "温室温度传感器"       |
| model            | string       | 型号                       | "DHT22"                |
| firmware_version | string       | 固件版本                   | "v1.2.0"               |
| status           | string       | 状态：ONLINE/OFFLINE/FAULT | "ONLINE"               |
| last_seen_at     | string\|null | 最后在线时间               | "2026-01-01T12:00:00Z" |
| protocol         | string       | 通信协议                   | "MQTT"                 |
| metadata         | string       | 元数据（JSON）             | "{}"                   |
| created_at       | string       | 创建时间                   | "2026-01-01T00:00:00Z" |
| updated_at       | string       | 更新时间                   | "2026-01-01T00:00:00Z" |

#### SensorChannel

| 字段                  | 类型         | 说明              | 示例                   |
| --------------------- | ------------ | ----------------- | ---------------------- |
| id                    | number       | 通道 ID           | 1                      |
| sensor_device_id      | number       | 所属传感器设备 ID | 1                      |
| channel_code          | string       | 通道编码          | "CH-TEMP"              |
| metric_code           | string       | 指标编码          | "TEMP"                 |
| unit                  | string       | 单位              | "°C"                   |
| precision_digits      | number       | 精度位数          | 2                      |
| range_min             | number\|null | 量程下限          | -40.0                  |
| range_max             | number\|null | 量程上限          | 80.0                   |
| sampling_interval_sec | number       | 采样间隔（秒）    | 60                     |
| enabled               | number       | 启用：1/0         | 1                      |
| last_reported_at      | string\|null | 最后上报时间      | "2026-01-01T12:00:00Z" |
| metadata              | string       | 元数据（JSON）    | "{}"                   |
| created_at            | string       | 创建时间          | "2026-01-01T00:00:00Z" |
| updated_at            | string       | 更新时间          | "2026-01-01T00:00:00Z" |

### 2.4 执行器设备与通道

#### ActuatorDevice

| 字段             | 类型         | 说明                       | 示例                   |
| ---------------- | ------------ | -------------------------- | ---------------------- |
| id               | number       | 设备 ID                    | 1                      |
| greenhouse_id    | number       | 所属温室 ID                | 1                      |
| growing_zone_id  | number\|null | 所属种植区 ID              | 1                      |
| device_code      | string       | 设备编码（唯一）           | "ACTUATOR-001"         |
| name             | string       | 名称                       | "循环水泵"             |
| model            | string       | 型号                       | "PUMP-100"             |
| firmware_version | string       | 固件版本                   | "v1.0.0"               |
| status           | string       | 状态：ONLINE/OFFLINE/FAULT | "ONLINE"               |
| last_seen_at     | string\|null | 最后在线时间               | "2026-01-01T12:00:00Z" |
| protocol         | string       | 通信协议                   | "MQTT"                 |
| metadata         | string       | 元数据（JSON）             | "{}"                   |
| created_at       | string       | 创建时间                   | "2026-01-01T00:00:00Z" |
| updated_at       | string       | 更新时间                   | "2026-01-01T00:00:00Z" |

#### ActuatorChannel

| 字段               | 类型         | 说明                                                                                                                                                                                                   | 示例                   |
| ------------------ | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------- |
| id                 | number       | 通道 ID                                                                                                                                                                                                | 1                      |
| actuator_device_id | number       | 所属执行器设备 ID                                                                                                                                                                                      | 1                      |
| channel_code       | string       | 通道编码                                                                                                                                                                                               | "CH-PUMP"              |
| actuator_type      | string       | 执行器类型：PUMP/AERATOR/FAN/VALVE/SHADE/LED/HEATER/CO2_GEN/FOGGER/DOSING_PUMP/CHILLER/STIRRER/DEHUMIDIFIER/DAMPER/UV_STERILIZER/OZONE_GENERATOR/FILTER/RO_SYSTEM/TOP_UP_VALVE/ALARM/CALIBRATION_VALVE | "PUMP"                 |
| current_state      | string       | 当前状态：ON/OFF                                                                                                                                                                                       | "OFF"                  |
| rated_power_watt   | number\|null | 额定功率（瓦）                                                                                                                                                                                         | 100.00                 |
| enabled            | number       | 启用：1/0                                                                                                                                                                                              | 1                      |
| metadata           | string       | 元数据（JSON）                                                                                                                                                                                         | "{}"                   |
| created_at         | string       | 创建时间                                                                                                                                                                                               | "2026-01-01T00:00:00Z" |
| updated_at         | string       | 更新时间                                                                                                                                                                                               | "2026-01-01T00:00:00Z" |

### 2.5 遥测记录

#### TelemetryRecord

| 字段              | 类型         | 说明                                                 | 示例                   |
| ----------------- | ------------ | ---------------------------------------------------- | ---------------------- |
| id                | number       | 记录 ID                                              | 1                      |
| sensor_channel_id | number       | 传感器通道 ID                                        | 1                      |
| metric_code       | string       | 指标编码                                             | "TEMP"                 |
| value             | number       | 处理后的值                                           | 25.50                  |
| raw_value         | number\|null | 原始值                                               | 25.50                  |
| quality_flag      | string       | 质量标志：normal/missing/out_of_range/device_offline | "normal"               |
| collected_at      | string       | 采集时间                                             | "2026-01-01T12:00:00Z" |
| ingested_at       | string       | 入库时间                                             | "2026-01-01T12:00:01Z" |
| batch_id          | number\|null | 关联批次 ID                                          | 1                      |
| created_at        | string       | 创建时间                                             | "2026-01-01T12:00:01Z" |

### 2.6 指标定义

#### MetricDefinition

| 字段             | 类型         | 说明          | 示例                   |
| ---------------- | ------------ | ------------- | ---------------------- |
| id               | number       | ID            | 1                      |
| code             | string       | 指标编码      | "TEMP"                 |
| name             | string       | 名称          | "温度"                 |
| unit             | string       | 单位          | "°C"                   |
| precision_digits | number       | 精度位数      | 2                      |
| normal_range_min | number\|null | 正常范围下限  | 18.0                   |
| normal_range_max | number\|null | 正常范围上限  | 30.0                   |
| is_core          | number       | 核心指标：1/0 | 1                      |
| status           | string       | 状态          | "ENABLED"              |
| created_at       | string       | 创建时间      | "2026-01-01T00:00:00Z" |
| updated_at       | string       | 更新时间      | "2026-01-01T00:00:00Z" |

当前系统指标（共 14 个）：

| 编码       | 名称         | 单位  | 精度 | 正常范围    | 核心指标 |
| ---------- | ------------ | ----- | ---- | ----------- | -------- |
| TEMP       | 温度         | °C    | 1    | 18.0–26.0   | 是       |
| HUMIDITY   | 湿度         | %     | 1    | 50.0–80.0   | 是       |
| PH         | 酸碱度       | pH    | 1    | 5.5–6.5     | 是       |
| EC         | 电导率       | mS/cm | 1    | 1.2–2.0     | 是       |
| CO2        | 二氧化碳     | ppm   | 0    | 400–1200    | 是       |
| LIGHT      | 光照         | lx    | 0    | 10000–60000 | 是       |
| DO         | 溶解氧       | mg/L  | 1    | 5.0–8.0     | 是       |
| WATER_TEMP | 水温         | °C    | 1    | 18.0–24.0   | 否       |
| LEVEL      | 液位         | cm    | 1    | 30.0–70.0   | 否       |
| ORP        | 氧化还原电位 | mV    | 0    | 200–500     | 否       |
| TDS        | 总溶解固体   | ppm   | 0    | 400–1200    | 否       |
| O3         | 臭氧浓度     | ppb   | 1    | 5.0–35.0    | 否       |
| TURBIDITY  | 浊度         | NTU   | 1    | 0.0–15.0    | 否       |
| FLOW_RATE  | 流量         | L/min | 1    | 5.0–15.0    | 否       |

### 2.7 控制命令

#### ControlCommand

| 字段                | 类型             | 说明                                           | 示例                   |
| ------------------- | ---------------- | ---------------------------------------------- | ---------------------- |
| id                  | number           | 命令 ID                                        | 1                      |
| actuator_channel_id | number           | 目标执行器通道 ID                              | 1                      |
| command_type        | string           | 命令类型                                       | "SWITCH"               |
| payload             | string           | 命令负载（JSON）                               | "{\"state\":\"ON\"}"   |
| status              | string           | 状态：PENDING/QUEUED/SENT/ACKED/TIMEOUT/FAILED | "PENDING"              |
| sent_at             | string\|null     | 发送时间                                       | "2026-01-01T12:00:00Z" |
| acked_at            | string\|null     | 确认时间                                       | "2026-01-01T12:00:01Z" |
| request_id          | string           | 请求追踪 ID                                    | "req_abc123"           |
| created_by          | number           | 创建者用户 ID                                  | 1                      |
| created_at          | string           | 创建时间                                       | "2026-01-01T00:00:00Z" |
| receipts            | CommandReceipt[] | 确认回执列表（含关联时）                       | []                     |

#### CommandReceipt

| 字段           | 类型         | 说明                                         | 示例                   |
| -------------- | ------------ | -------------------------------------------- | ---------------------- |
| id             | number       | 回执 ID                                      | 1                      |
| command_id     | number       | 命令 ID                                      | 1                      |
| receipt_seq    | number       | 回执序号                                     | 1                      |
| receipt_status | string       | 回执状态：ACCEPTED/REJECTED/PROCESSED/FAILED | "ACCEPTED"             |
| ack_code       | string       | 确认码                                       | "OK"                   |
| ack_message    | string       | 确认消息                                     | "Command executed"     |
| ack_payload    | string       | 确认负载（JSON）                             | "{}"                   |
| ack_at         | string\|null | 确认时间                                     | "2026-01-01T12:00:01Z" |
| created_at     | string       | 创建时间                                     | "2026-01-01T12:00:01Z" |

### 2.8 控制策略

#### ControlPolicy

| 字段            | 类型              | 说明                                  | 示例                   |
| --------------- | ----------------- | ------------------------------------- | ---------------------- |
| id              | number            | 策略 ID                               | 1                      |
| policy_code     | string            | 策略编码（唯一）                      | "POL-TEMP-CTRL"        |
| name            | string            | 名称                                  | "温度自动控制"         |
| policy_type     | string            | 策略类型：THRESHOLD/SCHEDULE/DURATION | "THRESHOLD"            |
| greenhouse_id   | number            | 所属温室 ID                           | 1                      |
| growing_zone_id | number\|null      | 所属种植区 ID                         | 1                      |
| priority        | number            | 优先级                                | 100                    |
| retry_limit     | number            | 重试次数上限                          | 3                      |
| timeout_sec     | number            | 超时秒数                              | 30                     |
| enabled         | number            | 启用：1/0                             | 1                      |
| version         | string            | 版本                                  | "v1"                   |
| effective_from  | string\|null      | 生效开始时间                          | "2026-01-01T00:00:00Z" |
| effective_to    | string\|null      | 生效结束时间                          | null                   |
| created_by      | number\|null      | 创建者                                | 1                      |
| published_by    | number\|null      | 发布者                                | 1                      |
| published_at    | string\|null      | 发布时间                              | "2026-01-01T00:00:00Z" |
| created_at      | string            | 创建时间                              | "2026-01-01T00:00:00Z" |
| updated_at      | string            | 更新时间                              | "2026-01-01T00:00:00Z" |
| conditions      | PolicyCondition[] | 条件列表（含关联时）                  | []                     |
| targets         | PolicyTarget[]    | 目标列表（含关联时）                  | []                     |

#### PolicyCondition

| 字段                  | 类型         | 说明                       | 示例                   |
| --------------------- | ------------ | -------------------------- | ---------------------- |
| id                    | number       | 条件 ID                    | 1                      |
| policy_id             | number       | 策略 ID                    | 1                      |
| metric_code           | string       | 指标编码                   | "TEMP"                 |
| operator              | string       | 运算符：>/>=/</<=/==/!=    | ">"                    |
| threshold_value       | number       | 阈值                       | 30.0                   |
| hysteresis            | number\|null | 滞后值                     | 1.0                    |
| window_sec            | number\|null | 滑动窗口（秒）             | 60                     |
| required_duration_sec | number\|null | 持续时间要求（秒）         | 30                     |
| aggregation           | string       | 聚合方式：avg/max/min/last | "avg"                  |
| enabled               | number       | 启用：1/0                  | 1                      |
| created_at            | string       | 创建时间                   | "2026-01-01T00:00:00Z" |
| updated_at            | string       | 更新时间                   | "2026-01-01T00:00:00Z" |

#### PolicyTarget

| 字段                | 类型   | 说明             | 示例                   |
| ------------------- | ------ | ---------------- | ---------------------- |
| id                  | number | 目标 ID          | 1                      |
| policy_id           | number | 策略 ID          | 1                      |
| actuator_channel_id | number | 执行器通道 ID    | 1                      |
| command_type        | string | 命令类型         | "SWITCH"               |
| command_payload     | string | 命令负载（JSON） | "{\"state\":\"ON\"}"   |
| execution_order     | number | 执行顺序         | 1                      |
| enabled             | number | 启用：1/0        | 1                      |
| created_at          | string | 创建时间         | "2026-01-01T00:00:00Z" |
| updated_at          | string | 更新时间         | "2026-01-01T00:00:00Z" |

#### PolicyExecution

| 字段                | 类型         | 说明                                   | 示例                   |
| ------------------- | ------------ | -------------------------------------- | ---------------------- |
| id                  | number       | 执行记录 ID                            | 1                      |
| policy_id           | number       | 策略 ID                                | 1                      |
| policy_name         | string       | 策略名称                               | "温度自动控制"         |
| trigger_source      | string       | 触发来源：TELEMETRY/SCHEDULE/MANUAL    | "TELEMETRY"            |
| trigger_metric_code | string       | 触发指标                               | "TEMP"                 |
| trigger_value       | number\|null | 触发值                                 | 30.50                  |
| decision            | string       | 决策：EXECUTED/SKIPPED/FAILED/CONFLICT | "EXECUTED"             |
| decision_reason     | string       | 决策原因                               | "Threshold exceeded"   |
| command_id          | number\|null | 关联命令 ID                            | 10                     |
| batch_id            | number\|null | 关联批次 ID                            | 1                      |
| executed_at         | string\|null | 执行时间                               | "2026-01-01T12:00:00Z" |
| created_at          | string       | 创建时间                               | "2026-01-01T12:00:00Z" |

### 2.9 告警

#### Alert

| 字段                | 类型         | 说明                                     | 示例                   |
| ------------------- | ------------ | ---------------------------------------- | ---------------------- |
| id                  | number       | 告警 ID                                  | 1                      |
| type                | string       | 类型：THRESHOLD/DEVICE_OFFLINE/SYSTEM    | "THRESHOLD"            |
| level               | string       | 级别：INFO/WARN/CRITICAL                 | "WARN"                 |
| metric_code         | string       | 指标编码                                 | "TEMP"                 |
| sensor_channel_id   | number\|null | 传感器通道 ID                            | 1                      |
| actuator_channel_id | number\|null | 执行器通道 ID                            | null                   |
| trigger_value       | number\|null | 触发值                                   | 35.0                   |
| message             | string       | 告警消息                                 | "温度超出阈值"         |
| status              | string       | 状态：OPEN/ACKNOWLEDGED/RESOLVED/IGNORED | "OPEN"                 |
| triggered_at        | string       | 触发时间                                 | "2026-01-01T12:00:00Z" |
| resolved_at         | string\|null | 解决时间                                 | null                   |
| resolved_by         | number\|null | 解决者用户 ID                            | null                   |
| timeline_count      | number       | 时间线事件数                             | 3                      |
| created_at          | string       | 创建时间                                 | "2026-01-01T12:00:00Z" |
| updated_at          | string       | 更新时间                                 | "2026-01-01T12:00:00Z" |

#### AlertTimelineEvent

| 字段          | 类型         | 说明                                                                        | 示例                   |
| ------------- | ------------ | --------------------------------------------------------------------------- | ---------------------- |
| id            | number       | 事件 ID                                                                     | 1                      |
| alert_id      | number       | 告警 ID                                                                     | 1                      |
| event_type    | string       | 事件类型：TRIGGERED/AUTO_ACTION/MANUAL_ACTION/ACKNOWLEDGED/RESOLVED/COMMENT | "TRIGGERED"            |
| event_source  | string       | 来源：SYSTEM/MANUAL                                                         | "SYSTEM"               |
| operator_id   | number\|null | 操作人用户 ID                                                               | null                   |
| comment       | string       | 备注                                                                        | ""                     |
| event_payload | string       | 事件负载（JSON）                                                            | "{}"                   |
| event_time    | string       | 事件时间                                                                    | "2026-01-01T12:00:00Z" |
| created_at    | string       | 创建时间                                                                    | "2026-01-01T12:00:00Z" |

### 2.10 告警统计

#### AlertStats

| 字段               | 类型   | 说明              |
| ------------------ | ------ | ----------------- |
| open_count         | number | 未处理数量        |
| acknowledged_count | number | 已确认数量        |
| resolved_count     | number | 已解决数量        |
| ignored_count      | number | 已忽略数量        |
| info_count         | number | INFO 级别数量     |
| warn_count         | number | WARN 级别数量     |
| critical_count     | number | CRITICAL 级别数量 |

### 2.11 通知渠道

#### NotificationChannel

| 字段            | 类型    | 说明                             | 示例                            |
| --------------- | ------- | -------------------------------- | ------------------------------- |
| id              | number  | 渠道 ID                          | 1                               |
| user_id         | number  | 用户 ID                          | 1                               |
| channel_type    | string  | 渠道类型：EMAIL/SMS/WEBHOOK      | "EMAIL"                         |
| name            | string  | 名称                             | "邮件通知"                      |
| config          | object  | 渠道配置（JSON）                 | {"address":"admin@example.com"} |
| min_alert_level | string  | 最低告警级别：INFO/WARN/CRITICAL | "WARN"                          |
| enabled         | boolean | 是否启用                         | true                            |
| created_at      | string  | 创建时间                         | "2026-01-01T00:00:00Z"          |
| updated_at      | string  | 更新时间                         | "2026-01-01T00:00:00Z"          |

### 2.12 审计日志

#### AuditLog

| 字段        | 类型         | 说明               | 示例                   |
| ----------- | ------------ | ------------------ | ---------------------- |
| id          | number       | 日志 ID            | 1                      |
| user_id     | number       | 操作用户 ID        | 1                      |
| action      | string       | 操作动作           | "CREATE"               |
| target_type | string       | 目标类型           | "GREENHOUSE"           |
| target_id   | number\|null | 目标 ID            | 1                      |
| detail      | object       | 操作详情（JSON）   | {}                     |
| request_id  | string       | 请求 ID            | "req_abc123"           |
| before_data | object       | 变更前数据（JSON） | {}                     |
| after_data  | object       | 变更后数据（JSON） | {}                     |
| created_at  | string       | 创建时间           | "2026-01-01T12:00:00Z" |

### 2.13 品种与生长阶段

#### CropVariety

| 字段               | 类型         | 说明             | 示例                   |
| ------------------ | ------------ | ---------------- | ---------------------- |
| id                 | number       | 品种 ID          | 1                      |
| code               | string       | 品种编码         | "LETTUCE-ICE"          |
| name               | string       | 品种名称         | "冰菜"                 |
| description        | string       | 描述             | "水培专用品种"         |
| default_cycle_days | number\|null | 默认生长周期天数 | 45                     |
| created_at         | string       | 创建时间         | "2026-01-01T00:00:00Z" |
| updated_at         | string       | 更新时间         | "2026-01-01T00:00:00Z" |

#### GrowthStage

| 字段                  | 类型         | 说明         | 示例                   |
| --------------------- | ------------ | ------------ | ---------------------- |
| id                    | number       | 阶段 ID      | 1                      |
| code                  | string       | 阶段编码     | "SEEDLING"             |
| name                  | string       | 阶段名称     | "育苗期"               |
| sort_order            | number       | 排序         | 1                      |
| default_duration_days | number\|null | 默认持续天数 | 14                     |
| created_at            | string       | 创建时间     | "2026-01-01T00:00:00Z" |
| updated_at            | string       | 更新时间     | "2026-01-01T00:00:00Z" |

### 2.14 种植批次

#### CropBatch

| 字段                      | 类型         | 说明                                               | 示例                   |
| ------------------------- | ------------ | -------------------------------------------------- | ---------------------- |
| id                        | number       | 批次 ID                                            | 1                      |
| batch_no                  | string       | 批次号（唯一）                                     | "BATCH-2026-001"       |
| greenhouse_id             | number       | 温室 ID                                            | 1                      |
| growing_zone_id           | number\|null | 种植区 ID                                          | 1                      |
| crop_variety_id           | number       | 品种 ID                                            | 1                      |
| variety_code              | string       | 品种编码（关联）                                   | "LETTUCE-ICE"          |
| variety_name              | string       | 品种名称（关联）                                   | "冰菜"                 |
| status                    | string       | 状态：PLANNED/RUNNING/HARVESTING/COMPLETED/ABORTED | "RUNNING"              |
| planting_density          | number\|null | 种植密度                                           | 25.00                  |
| total_plants              | number\|null | 总株数                                             | 1000                   |
| started_at                | string\|null | 开始时间                                           | "2026-01-01T00:00:00Z" |
| ended_at                  | string\|null | 结束时间                                           | null                   |
| expected_harvest_at       | string\|null | 预计收获时间                                       | "2026-02-15T00:00:00Z" |
| recipe_version            | string       | 配方版本                                           | "v1"                   |
| policy_version            | string       | 策略版本                                           | "v1"                   |
| active_recipe_id          | number\|null | 当前生效配方 ID（自动联动写入）                    | 10                     |
| active_policy_id          | number\|null | 当前生效策略 ID（自动联动写入）                    | 20                     |
| active_climate_profile_id | number\|null | 当前生效气候 Profile ID（自动联动写入）            | 30                     |
| note                      | string       | 备注                                               | ""                     |
| created_by                | number\|null | 创建者                                             | 1                      |
| created_at                | string       | 创建时间                                           | "2026-01-01T00:00:00Z" |
| updated_at                | string       | 更新时间                                           | "2026-01-01T00:00:00Z" |

#### BatchStagePlan

| 字段               | 类型         | 说明                | 示例                   |
| ------------------ | ------------ | ------------------- | ---------------------- |
| id                 | number       | 计划 ID             | 1                      |
| batch_id           | number       | 批次 ID             | 1                      |
| growth_stage_id    | number       | 生长阶段 ID         | 1                      |
| recipe_id          | number\|null | 阶段配方 ID         | 10                     |
| policy_id          | number\|null | 阶段策略 ID         | 20                     |
| climate_profile_id | number\|null | 阶段气候 Profile ID | 30                     |
| stage_start_at     | string       | 阶段开始时间        | "2026-01-01T00:00:00Z" |
| stage_end_at       | string       | 阶段结束时间        | "2026-01-14T00:00:00Z" |
| target_ec_min      | number\|null | 目标 EC 下限        | 1.2                    |
| target_ec_max      | number\|null | 目标 EC 上限        | 1.8                    |
| target_ph_min      | number\|null | 目标 pH 下限        | 5.5                    |
| target_ph_max      | number\|null | 目标 pH 上限        | 6.5                    |
| created_at         | string       | 创建时间            | "2026-01-01T00:00:00Z" |
| updated_at         | string       | 更新时间            | "2026-01-01T00:00:00Z" |

#### HarvestRecord

| 字段              | 类型         | 说明              | 示例                   |
| ----------------- | ------------ | ----------------- | ---------------------- |
| id                | number       | 收获记录 ID       | 1                      |
| batch_id          | number       | 批次 ID           | 1                      |
| harvested_at      | string       | 收获时间          | "2026-02-15T00:00:00Z" |
| harvest_weight_kg | number       | 收获重量（kg）    | 50.500                 |
| grade             | string       | 等级：A/B/C/Waste | "A"                    |
| grade_weight_kg   | number       | 该等级重量（kg）  | 30.000                 |
| note              | string       | 备注              | ""                     |
| harvested_by      | number\|null | 收获人            | 1                      |
| created_at        | string       | 创建时间          | "2026-02-15T00:00:00Z" |

### 2.15 营养液配方

#### NutrientRecipe

| 字段            | 类型         | 说明                        | 示例                   |
| --------------- | ------------ | --------------------------- | ---------------------- |
| id              | number       | 配方 ID                     | 1                      |
| recipe_code     | string       | 配方编码（唯一）            | "REC-LETTUCE"          |
| name            | string       | 名称                        | "生菜标准配方"         |
| crop_variety_id | number\|null | 关联品种 ID                 | 1                      |
| description     | string       | 描述                        | ""                     |
| version         | string       | 版本                        | "v1"                   |
| status          | string       | 状态：DRAFT/ACTIVE/ARCHIVED | "ACTIVE"               |
| effective_from  | string\|null | 生效时间                    | "2026-01-01T00:00:00Z" |
| effective_to    | string\|null | 失效时间                    | null                   |
| created_by      | number\|null | 创建者                      | 1                      |
| published_by    | number\|null | 发布者                      | 1                      |
| published_at    | string\|null | 发布时间                    | "2026-01-01T00:00:00Z" |
| created_at      | string       | 创建时间                    | "2026-01-01T00:00:00Z" |
| updated_at      | string       | 更新时间                    | "2026-01-01T00:00:00Z" |

#### RecipeStageTarget

| 字段            | 类型         | 说明        | 示例                   |
| --------------- | ------------ | ----------- | ---------------------- |
| id              | number       | 目标 ID     | 1                      |
| recipe_id       | number       | 配方 ID     | 1                      |
| growth_stage_id | number\|null | 生长阶段 ID | 1                      |
| metric_code     | string       | 指标编码    | "EC"                   |
| target_min      | number\|null | 目标下限    | 1.2                    |
| target_max      | number\|null | 目标上限    | 1.8                    |
| tolerance       | number\|null | 容差        | 0.1                    |
| unit            | string       | 单位        | "mS/cm"                |
| enabled         | number       | 启用：1/0   | 1                      |
| created_at      | string       | 创建时间    | "2026-01-01T00:00:00Z" |
| updated_at      | string       | 更新时间    | "2026-01-01T00:00:00Z" |

#### RecipeIonTarget

| 字段            | 类型         | 说明             | 示例                   |
| --------------- | ------------ | ---------------- | ---------------------- |
| id              | number       | 目标 ID          | 1                      |
| recipe_id       | number       | 配方 ID          | 1                      |
| growth_stage_id | number\|null | 生长阶段 ID      | 1                      |
| ion_code        | string       | 离子编码         | "NO3"                  |
| target_min_mg_l | number\|null | 目标下限（mg/L） | 150.0                  |
| target_max_mg_l | number\|null | 目标上限（mg/L） | 200.0                  |
| enabled         | number       | 启用：1/0        | 1                      |
| created_at      | string       | 创建时间         | "2026-01-01T00:00:00Z" |
| updated_at      | string       | 更新时间         | "2026-01-01T00:00:00Z" |

#### BatchRecipeBinding

| 字段           | 类型         | 说明                        | 示例                   |
| -------------- | ------------ | --------------------------- | ---------------------- |
| id             | number       | 绑定 ID                     | 1                      |
| batch_id       | number       | 批次 ID                     | 1                      |
| recipe_id      | number       | 配方 ID                     | 1                      |
| binding_type   | string       | 绑定类型：PRIMARY/SECONDARY | "PRIMARY"              |
| version        | string       | 版本                        | "v1"                   |
| effective_from | string       | 生效时间                    | "2026-01-01T00:00:00Z" |
| effective_to   | string\|null | 失效时间                    | null                   |
| status         | string       | 状态：ACTIVE/INACTIVE       | "ACTIVE"               |
| created_by     | number\|null | 创建者                      | 1                      |
| created_at     | string       | 创建时间                    | "2026-01-01T00:00:00Z" |
| updated_at     | string       | 更新时间                    | "2026-01-01T00:00:00Z" |

### 2.16 气候控制

#### ClimateProfile

| 字段                      | 类型           | 说明                              | 示例                   |
| ------------------------- | -------------- | --------------------------------- | ---------------------- |
| id                        | number         | 配置 ID                           | 1                      |
| greenhouse_id             | number         | 温室 ID                           | 1                      |
| code                      | string         | 配置编码                          | "CLIMATE-DEFAULT"      |
| name                      | string         | 名称                              | "默认气候配置"         |
| description               | string         | 描述                              | ""                     |
| trigger_metric_code       | string         | 触发指标编码                      | "TEMP"                 |
| trigger_sensor_channel_id | number\|null   | 触发采集通道 ID（固定单通道触发） | 123                    |
| enabled                   | number         | 启用：1/0                         | 1                      |
| stages_count              | number         | 阶段数                            | 3                      |
| created_at                | string         | 创建时间                          | "2026-01-01T00:00:00Z" |
| updated_at                | string         | 更新时间                          | "2026-01-01T00:00:00Z" |
| stages                    | ClimateStage[] | 阶段列表（含关联时）              | []                     |

#### ClimateStage

| 字段              | 类型                 | 说明                  | 示例                   |
| ----------------- | -------------------- | --------------------- | ---------------------- |
| id                | number               | 阶段 ID               | 1                      |
| profile_id        | number               | 配置 ID               | 1                      |
| stage_level       | number               | 阶段序号              | 1                      |
| name              | string               | 阶段名称              | "低温阶段"             |
| trigger_operator  | string               | 触发运算符：>/>=/</<= | "<"                    |
| trigger_threshold | number               | 触发阈值              | 18.0                   |
| hysteresis        | number               | 滞后值                | 1.0                    |
| action_count      | number               | 动作数                | 2                      |
| created_at        | string               | 创建时间              | "2026-01-01T00:00:00Z" |
| updated_at        | string               | 更新时间              | "2026-01-01T00:00:00Z" |
| actions           | ClimateStageAction[] | 动作列表（含关联时）  | []                     |

#### ClimateStageAction

| 字段                | 类型   | 说明             | 示例                   |
| ------------------- | ------ | ---------------- | ---------------------- |
| id                  | number | 动作 ID          | 1                      |
| stage_id            | number | 阶段 ID          | 1                      |
| actuator_channel_id | number | 执行器通道 ID    | 1                      |
| command_type        | string | 命令类型         | "SWITCH"               |
| command_payload     | string | 命令负载（JSON） | "{\"state\":\"ON\"}"   |
| execution_order     | number | 执行顺序         | 1                      |
| enabled             | number | 启用：1/0        | 1                      |
| created_at          | string | 创建时间         | "2026-01-01T00:00:00Z" |
| updated_at          | string | 更新时间         | "2026-01-01T00:00:00Z" |

#### ClimateExecutionLog

| 字段                      | 类型         | 说明             | 示例                   |
| ------------------------- | ------------ | ---------------- | ---------------------- |
| id                        | number       | 日志 ID          | 1                      |
| profile_id                | number       | 配置 ID          | 1                      |
| profile_name              | string       | 配置名称（关联） | "默认气候配置"         |
| from_stage_level          | number\|null | 来源阶段序号     | null                   |
| to_stage_level            | number       | 目标阶段序号     | 2                      |
| trigger_value             | number       | 触发值           | 15.0                   |
| trigger_sensor_channel_id | number\|null | 触发采集通道 ID  | 123                    |
| trigger_metric_code       | string\|null | 触发指标         | "TEMP"                 |
| collected_at              | string\|null | 遥测采集时间     | "2026-01-01T12:00:00Z" |
| executed_actions_count    | number       | 执行动作数       | 2                      |
| executed_at               | string       | 执行时间         | "2026-01-01T12:00:00Z" |
| created_at                | string       | 创建时间         | "2026-01-01T12:00:00Z" |

### 2.17 营养液管理

#### NutrientTank

| 字段                    | 类型         | 说明                        | 示例                   |
| ----------------------- | ------------ | --------------------------- | ---------------------- |
| id                      | number       | 水箱 ID                     | 1                      |
| growing_zone_id         | number       | 种植区 ID                   | 1                      |
| code                    | string       | 水箱编码                    | "TANK-A1"              |
| total_volume_liter      | number       | 总容积（升）                | 200.00                 |
| current_volume_liter    | number\|null | 当前容积（升）              | 180.00                 |
| status                  | string       | 状态：ACTIVE/INACTIVE/EMPTY | "ACTIVE"               |
| ec_sensor_channel_id    | number\|null | EC 传感器通道 ID            | 1                      |
| ph_sensor_channel_id    | number\|null | pH 传感器通道 ID            | 2                      |
| level_sensor_channel_id | number\|null | 液位传感器通道 ID           | 3                      |
| temp_sensor_channel_id  | number\|null | 水温传感器通道 ID           | 4                      |
| created_at              | string       | 创建时间                    | "2026-01-01T00:00:00Z" |
| updated_at              | string       | 更新时间                    | "2026-01-01T00:00:00Z" |

#### SolutionChangeEvent

| 字段                  | 类型         | 说明                                          | 示例                   |
| --------------------- | ------------ | --------------------------------------------- | ---------------------- |
| id                    | number       | 事件 ID                                       | 1                      |
| tank_id               | number       | 水箱 ID                                       | 1                      |
| change_type           | string       | 换液类型：FULL_REPLACE/PARTIAL_REFRESH/TOP_UP | "FULL_REPLACE"         |
| volume_replaced_liter | number       | 更换容积（升）                                | 200.00                 |
| source_water_ec       | number\|null | 水源 EC                                       | 0.3                    |
| source_water_ph       | number\|null | 水源 pH                                       | 7.0                    |
| before_ec             | number\|null | 更换前 EC                                     | 2.5                    |
| before_ph             | number\|null | 更换前 pH                                     | 6.8                    |
| after_ec              | number\|null | 更换后 EC                                     | 1.5                    |
| after_ph              | number\|null | 更换后 pH                                     | 6.0                    |
| nutrient_a_added_ml   | number\|null | A 液添加量（ml）                              | 500.00                 |
| nutrient_b_added_ml   | number\|null | B 液添加量（ml）                              | 500.00                 |
| acid_added_ml         | number\|null | 酸液添加量（ml）                              | 10.00                  |
| alkali_added_ml       | number\|null | 碱液添加量（ml）                              | 0                      |
| note                  | string       | 备注                                          | ""                     |
| operated_by           | number\|null | 操作人                                        | 1                      |
| operated_at           | string       | 操作时间                                      | "2026-01-15T00:00:00Z" |
| created_at            | string       | 创建时间                                      | "2026-01-15T00:00:00Z" |

#### IonTestRecord

| 字段         | 类型         | 说明                      | 示例                   |
| ------------ | ------------ | ------------------------- | ---------------------- |
| id           | number       | 检测记录 ID               | 1                      |
| tank_id      | number       | 水箱 ID                   | 1                      |
| batch_id     | number\|null | 批次 ID                   | 1                      |
| sample_code  | string       | 样本编码（唯一）          | "SAMPLE-001"           |
| sampled_at   | string       | 采样时间                  | "2026-01-10T00:00:00Z" |
| tested_at    | string\|null | 检测时间                  | "2026-01-11T00:00:00Z" |
| test_method  | string       | 检测方法：LAB/STRIP/METER | "LAB"                  |
| no3_n        | number\|null | 硝态氮（mg/L）            | 150.00                 |
| nh4_n        | number\|null | 铵态氮（mg/L）            | 10.00                  |
| p            | number\|null | 磷（mg/L）                | 40.00                  |
| k            | number\|null | 钾（mg/L）                | 200.00                 |
| ca           | number\|null | 钙（mg/L）                | 150.00                 |
| mg           | number\|null | 镁（mg/L）                | 50.00                  |
| s            | number\|null | 硫（mg/L）                | 30.00                  |
| fe           | number\|null | 铁（mg/L）                | 2.0000                 |
| mn           | number\|null | 锰（mg/L）                | 0.5000                 |
| zn           | number\|null | 锌（mg/L）                | 0.0500                 |
| b            | number\|null | 硼（mg/L）                | 0.5000                 |
| cu           | number\|null | 铜（mg/L）                | 0.0200                 |
| mo           | number\|null | 钼（mg/L）                | 0.0100                 |
| ec_at_sample | number\|null | 采样时 EC                 | 1.5000                 |
| ph_at_sample | number\|null | 采样时 pH                 | 6.0000                 |
| lab_name     | string       | 实验室名称                | ""                     |
| report_url   | string       | 报告 URL                  | ""                     |
| note         | string       | 备注                      | ""                     |
| created_by   | number\|null | 创建者                    | 1                      |
| created_at   | string       | 创建时间                  | "2026-01-11T00:00:00Z" |

#### NutrientConcentrateInventory

| 字段                | 类型         | 说明                        | 示例                   |
| ------------------- | ------------ | --------------------------- | ---------------------- |
| id                  | number       | 库存 ID                     | 1                      |
| greenhouse_id       | number       | 温室 ID                     | 1                      |
| concentrate_type    | string       | 浓缩液类型：A/B/ACID/ALKALI | "A"                    |
| brand               | string       | 品牌                        | ""                     |
| product_name        | string       | 产品名称                    | "A 液浓缩液"           |
| total_volume_ml     | number       | 总容量（ml）                | 5000.00                |
| remaining_volume_ml | number       | 剩余容量（ml）              | 3000.00                |
| unit_price          | number\|null | 单价                        | 50.00                  |
| batch_no            | string       | 批号                        | ""                     |
| expired_at          | string\|null | 过期日期                    | "2026-12-31"           |
| status              | string       | 状态：IN_USE/EMPTY/EXPIRED  | "IN_USE"               |
| created_at          | string       | 创建时间                    | "2026-01-01T00:00:00Z" |
| updated_at          | string       | 更新时间                    | "2026-01-01T00:00:00Z" |

#### ConcentrateUsageLog

| 字段               | 类型         | 说明         | 示例                   |
| ------------------ | ------------ | ------------ | ---------------------- |
| id                 | number       | 使用记录 ID  | 1                      |
| inventory_id       | number       | 库存 ID      | 1                      |
| solution_change_id | number\|null | 换液事件 ID  | 1                      |
| tank_id            | number\|null | 水箱 ID      | 1                      |
| volume_used_ml     | number       | 使用量（ml） | 500.00                 |
| used_by            | number\|null | 使用人       | 1                      |
| used_at            | string       | 使用时间     | "2026-01-15T00:00:00Z" |
| created_at         | string       | 创建时间     | "2026-01-15T00:00:00Z" |

### 2.18 能耗记录

#### EnergyConsumptionRecord

| 字段                | 类型         | 说明                            | 示例                   |
| ------------------- | ------------ | ------------------------------- | ---------------------- |
| id                  | number       | 记录 ID                         | 1                      |
| greenhouse_id       | number       | 温室 ID                         | 1                      |
| record_type         | string       | 类型：ELECTRICITY/WATER/CO2_GAS | "ELECTRICITY"          |
| consumption_value   | number       | 消耗量                          | 150.5000               |
| unit                | string       | 单位                            | "kWh"                  |
| record_period_start | string       | 记录周期开始                    | "2026-01-01T00:00:00Z" |
| record_period_end   | string       | 记录周期结束                    | "2026-01-02T00:00:00Z" |
| meter_reading_start | number\|null | 表头起始读数                    | 1000.0000              |
| meter_reading_end   | number\|null | 表头结束读数                    | 1150.5000              |
| batch_id            | number\|null | 批次 ID                         | 1                      |
| recorded_by         | number\|null | 记录人                          | 1                      |
| created_at          | string       | 创建时间                        | "2026-01-02T00:00:00Z" |

### 2.19 病虫害

#### PestDiseaseObservation

| 字段                 | 类型         | 说明                            | 示例                   |
| -------------------- | ------------ | ------------------------------- | ---------------------- |
| id                   | number       | 观察记录 ID                     | 1                      |
| greenhouse_id        | number       | 温室 ID                         | 1                      |
| growing_zone_id      | number\|null | 种植区 ID                       | 1                      |
| batch_id             | number\|null | 批次 ID                         | 1                      |
| observed_at          | string       | 观察时间                        | "2026-01-10T00:00:00Z" |
| pest_or_disease      | string       | 病虫名称                        | "蚜虫"                 |
| severity             | string       | 严重程度：LIGHT/MODERATE/SEVERE | "MODERATE"             |
| affected_area_pct    | number\|null | 受影响面积百分比                | 15.00                  |
| affected_plant_count | number\|null | 受影响株数                      | 50                     |
| symptoms             | string       | 症状描述                        | "叶片有黄斑"           |
| photo_urls           | string       | 照片 URL（JSON 数组）           | "[]"                   |
| observed_by          | number\|null | 观察人                          | 1                      |
| created_at           | string       | 创建时间                        | "2026-01-10T00:00:00Z" |

#### TreatmentRecord

| 字段                   | 类型         | 说明                                   | 示例                   |
| ---------------------- | ------------ | -------------------------------------- | ---------------------- |
| id                     | number       | 治理记录 ID                            | 1                      |
| observation_id         | number\|null | 关联观察 ID                            | 1                      |
| greenhouse_id          | number       | 温室 ID                                | 1                      |
| growing_zone_id        | number\|null | 种植区 ID                              | 1                      |
| batch_id               | number\|null | 批次 ID                                | 1                      |
| treatment_type         | string       | 治理类型：CHEMICAL/BIOLOGICAL/PHYSICAL | "BIOLOGICAL"           |
| product_name           | string       | 药品名称                               | "瓢虫"                 |
| active_ingredient      | string       | 有效成分                               | ""                     |
| dosage                 | string       | 用量                                   | "100只/亩"             |
| application_method     | string       | 施用方式：SPRAY/DRENCH/FOG/RELEASE     | "RELEASE"              |
| safety_interval_days   | number\|null | 安全间隔天数                           | 7                      |
| reentry_interval_hours | number\|null | 再进入间隔小时数                       | 24                     |
| treated_at             | string       | 治理时间                               | "2026-01-10T00:00:00Z" |
| treated_by             | number\|null | 治理人                                 | 1                      |
| note                   | string       | 备注                                   | ""                     |
| created_at             | string       | 创建时间                               | "2026-01-10T00:00:00Z" |

### 2.20 批次审查快照

#### BatchReviewSnapshot

| 字段          | 类型   | 说明                                 | 示例                   |
| ------------- | ------ | ------------------------------------ | ---------------------- |
| id            | number | 快照 ID                              | 1                      |
| batch_id      | number | 批次 ID                              | 1                      |
| snapshot_type | string | 快照类型：DAILY/WEEKLY/STAGE_SUMMARY | "DAILY"                |
| window_start  | string | 窗口开始时间                         | "2026-01-01T00:00:00Z" |
| window_end    | string | 窗口结束时间                         | "2026-01-02T00:00:00Z" |
| summary       | object | 摘要数据（JSON）                     | {}                     |
| generated_at  | string | 生成时间                             | "2026-01-02T00:00:00Z" |
| created_at    | string | 创建时间                             | "2026-01-02T00:00:00Z" |

---

## 3. API 端点

### 3.1 认证 (Auth)

**POST /api/auth/login**

鉴权：无需认证

请求体：

| 字段     | 类型   | 必填 | 规则      | 示例       |
| -------- | ------ | ---- | --------- | ---------- |
| username | string | 是   | 3-32 字符 | "admin"    |
| password | string | 是   | 6-64 字符 | "admin123" |

响应：

```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOi...",
    "user": {
      "id": 1,
      "username": "admin",
      "nickname": "管理员",
      "roles": [{ "id": 1, "name": "ADMIN" }]
    }
  }
}
```

**POST /api/auth/logout**

鉴权：任意已认证用户

请求体：无

响应：

```json
{ "code": 0, "data": {} }
```

### 3.2 用户管理 (Users)

**GET /api/users**

鉴权：ADMIN

查询参数：

| 字段      | 类型   | 必填 | 说明              |
| --------- | ------ | ---- | ----------------- |
| page      | number | 否   | 页码，默认 1      |
| page_size | number | 否   | 每页条数，默认 20 |

响应：分页列表，items 为 User[]

**POST /api/users**

鉴权：ADMIN

请求体：

| 字段     | 类型     | 必填 | 规则          | 示例         |
| -------- | -------- | ---- | ------------- | ------------ |
| username | string   | 是   | 3-32 字符     | "operator1"  |
| password | string   | 是   | 6-64 字符     | "pass1234"   |
| nickname | string   | 否   | 最多 64 字符  | "操作员"     |
| roles    | string[] | 是   | 至少 1 个角色 | ["OPERATOR"] |

响应：

```json
{ "code": 0, "data": { "id": 2 } }
```

**PUT /api/users/:userId**

鉴权：ADMIN

请求体：

| 字段     | 类型     | 必填 | 规则         |
| -------- | -------- | ---- | ------------ |
| nickname | string   | 否   | 最多 64 字符 |
| phone    | string   | 否   | 最多 32 字符 |
| email    | string   | 否   | 最多 64 字符 |
| roles    | string[] | 否   | 角色列表     |

**PATCH /api/users/:userId/status**

鉴权：ADMIN

请求体：

| 字段   | 类型   | 必填 | 规则             | 示例       |
| ------ | ------ | ---- | ---------------- | ---------- |
| status | string | 是   | ENABLED/DISABLED | "DISABLED" |

### 3.3 角色管理 (Roles)

**GET /api/roles**

鉴权：ADMIN

响应：分页列表，items 为 Role[]

**POST /api/roles**

鉴权：ADMIN

请求体：

| 字段        | 类型   | 必填 | 规则         | 示例          |
| ----------- | ------ | ---- | ------------ | ------------- |
| name        | string | 是   | 1-32 字符    | "CUSTOM_ROLE" |
| description | string | 否   | 最多 64 字符 | "自定义角色"  |

响应：

```json
{ "code": 0, "data": { "id": 4 } }
```

**PUT /api/roles/:roleId**

鉴权：ADMIN

请求体：

| 字段        | 类型   | 必填 | 规则         |
| ----------- | ------ | ---- | ------------ |
| description | string | 否   | 最多 64 字符 |

### 3.4 温室 (Greenhouses)

**POST /api/greenhouses**

鉴权：ADMIN

请求体：

| 字段        | 类型   | 必填 | 规则           | 示例       |
| ----------- | ------ | ---- | -------------- | ---------- |
| code        | string | 是   | 最多 32 字符   | "GH-001"   |
| name        | string | 是   | 最多 64 字符   | "一号温室" |
| location    | string | 否   | 最多 128 字符  | "A区"      |
| area_sqm    | number | 否   | 面积（平方米） | 500.00     |
| description | string | 否   | 最多 255 字符  | "叶菜专用" |

**PUT /api/greenhouses/:id**

鉴权：ADMIN

请求体：同创建，所有字段可选；额外支持 status 字段（ENABLED/DISABLED）

**GET /api/greenhouses**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size

响应：分页列表，items 为 GreenhouseResponse[]

**GET /api/greenhouses/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：GreenhouseResponse（含 zone_count）

**DELETE /api/greenhouses/:id**

鉴权：ADMIN

响应：

```json
{ "code": 0, "data": {} }
```

**GET /api/greenhouses/:id/zones**

鉴权：ADMIN / OPERATOR / VIEWER

说明：获取指定温室下的所有种植区

响应：分页列表，items 为 GrowingZoneResponse[]

### 3.5 种植区 (GrowingZones)

**POST /api/growing-zones**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                     | 类型   | 必填 | 规则         | 示例      |
| ------------------------ | ------ | ---- | ------------ | --------- |
| greenhouse_id            | number | 是   | 温室 ID      | 1         |
| code                     | string | 是   | 最多 32 字符 | "ZONE-A1" |
| name                     | string | 是   | 最多 64 字符 | "A1区"    |
| system_type              | string | 否   | DWC/NFT      | "DWC"     |
| tank_volume_liter        | number | 否   | 水箱容积     | 200.00    |
| planting_density_per_sqm | number | 否   | 种植密度     | 25.00     |

**PUT /api/growing-zones/:id**

鉴权：ADMIN / OPERATOR

请求体：同创建，所有字段可选；额外支持 status 字段（ENABLED/DISABLED）

**GET /api/growing-zones**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, greenhouse_id

响应：分页列表，items 为 GrowingZoneResponse[]

**GET /api/growing-zones/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：GrowingZoneResponse

**DELETE /api/growing-zones/:id**

鉴权：ADMIN

响应：

```json
{ "code": 0, "data": {} }
```

### 3.6 传感器设备 (SensorDevices)

**POST /api/sensor-devices**

鉴权：ADMIN / OPERATOR

请求体：

| 字段             | 类型         | 必填 | 规则                   | 示例         |
| ---------------- | ------------ | ---- | ---------------------- | ------------ |
| device_code      | string       | 是   | 最多 64 字符（唯一）   | "SENSOR-001" |
| name             | string       | 是   | 最多 64 字符           | "温度传感器" |
| model            | string       | 否   | 最多 64 字符           | "DHT22"      |
| firmware_version | string       | 否   | 最多 64 字符           | "v1.2.0"     |
| greenhouse_id    | number       | 是   | 温室 ID                | 1            |
| growing_zone_id  | number\|null | 否   | 种植区 ID              | 1            |
| protocol         | string       | 否   | 通信协议，最多 16 字符 | "MQTT"       |
| metadata         | string       | 否   | JSON 字符串            | "{}"         |

**PUT /api/sensor-devices/:id**

鉴权：ADMIN / OPERATOR

请求体：

| 字段             | 类型         | 必填 | 规则                 |
| ---------------- | ------------ | ---- | -------------------- |
| name             | string       | 否   | 最多 64 字符         |
| model            | string       | 否   | 最多 64 字符         |
| firmware_version | string       | 否   | 最多 64 字符         |
| growing_zone_id  | number\|null | 否   | 种植区 ID            |
| status           | string       | 否   | ONLINE/OFFLINE/FAULT |
| metadata         | string       | 否   | JSON 字符串          |

**GET /api/sensor-devices**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, greenhouse_id, status

响应：分页列表，items 为 SensorDeviceResponse[]

**GET /api/sensor-devices/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：SensorDeviceResponse

**DELETE /api/sensor-devices/:id**

鉴权：ADMIN

响应：

```json
{ "code": 0, "data": {} }
```

### 3.7 传感器通道 (SensorChannels)

**POST /api/sensor-channels**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                  | 类型         | 必填 | 规则          | 示例      |
| --------------------- | ------------ | ---- | ------------- | --------- |
| sensor_device_id      | number       | 是   | 传感器设备 ID | 1         |
| channel_code          | string       | 是   | 最多 64 字符  | "CH-TEMP" |
| metric_code           | string       | 是   | 最多 32 字符  | "TEMP"    |
| unit                  | string       | 是   | 最多 16 字符  | "°C"      |
| precision_digits      | number       | 否   | 默认 2        | 2         |
| range_min             | number\|null | 否   | 量程下限      | -40.0     |
| range_max             | number\|null | 否   | 量程上限      | 80.0      |
| sampling_interval_sec | number       | 否   | 默认 60       | 60        |
| metadata              | string       | 否   | JSON 字符串   | "{}"      |

**PUT /api/sensor-channels/:id**

鉴权：ADMIN / OPERATOR

请求体：同创建，所有字段可选；额外支持 enabled 字段（1/0）

**GET /api/sensor-channels**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, sensor_device_id, enabled

响应：分页列表，items 为 SensorChannelResponse[]

**GET /api/sensor-channels/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：SensorChannelResponse

**DELETE /api/sensor-channels/:id**

鉴权：ADMIN

说明：禁止删除采集通道；如需停用请使用 `PUT /api/sensor-channels/:id` 设置 enabled=0

### 3.8 执行器设备 (ActuatorDevices)

**POST /api/actuator-devices**

鉴权：ADMIN / OPERATOR

请求体：

| 字段             | 类型         | 必填 | 规则                   | 示例           |
| ---------------- | ------------ | ---- | ---------------------- | -------------- |
| device_code      | string       | 是   | 最多 64 字符（唯一）   | "ACTUATOR-001" |
| name             | string       | 是   | 最多 64 字符           | "循环水泵"     |
| model            | string       | 否   | 最多 64 字符           | "PUMP-100"     |
| firmware_version | string       | 否   | 最多 64 字符           | "v1.0.0"       |
| greenhouse_id    | number       | 是   | 温室 ID                | 1              |
| growing_zone_id  | number\|null | 否   | 种植区 ID              | 1              |
| protocol         | string       | 否   | 通信协议，最多 16 字符 | "MQTT"         |
| metadata         | string       | 否   | JSON 字符串            | "{}"           |

**PUT /api/actuator-devices/:id**

鉴权：ADMIN / OPERATOR

请求体：同 SensorDevice 更新结构

**GET /api/actuator-devices**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, greenhouse_id, status

响应：分页列表，items 为 ActuatorDeviceResponse[]

**GET /api/actuator-devices/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：ActuatorDeviceResponse

**DELETE /api/actuator-devices/:id**

鉴权：ADMIN

响应：

```json
{ "code": 0, "data": {} }
```

### 3.9 执行器通道 (ActuatorChannels)

**POST /api/actuator-channels**

鉴权：ADMIN / OPERATOR

请求体：

| 字段               | 类型         | 必填 | 规则                                                                                                                                                                                                     | 示例      |
| ------------------ | ------------ | ---- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| actuator_device_id | number       | 是   | 执行器设备 ID                                                                                                                                                                                            | 1         |
| channel_code       | string       | 是   | 最多 64 字符                                                                                                                                                                                             | "CH-PUMP" |
| actuator_type      | string       | 是   | 最多 32 字符：PUMP/AERATOR/FAN/VALVE/SHADE/LED/HEATER/CO2_GEN/FOGGER/DOSING_PUMP/CHILLER/STIRRER/DEHUMIDIFIER/DAMPER/UV_STERILIZER/OZONE_GENERATOR/FILTER/RO_SYSTEM/TOP_UP_VALVE/ALARM/CALIBRATION_VALVE | "PUMP"    |
| rated_power_watt   | number\|null | 否   | 额定功率（瓦）                                                                                                                                                                                           | 100.00    |
| metadata           | string       | 否   | JSON 字符串                                                                                                                                                                                              | "{}"      |

**PUT /api/actuator-channels/:id**

鉴权：ADMIN / OPERATOR

请求体：同创建，所有字段可选；额外支持 current_state（ON/OFF）和 enabled（1/0）

**GET /api/actuator-channels**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, actuator_device_id, actuator_type

响应：分页列表，items 为 ActuatorChannelResponse[]

**GET /api/actuator-channels/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：ActuatorChannelResponse

**DELETE /api/actuator-channels/:id**

鉴权：ADMIN

响应：

```json
{ "code": 0, "data": {} }
```

### 3.10 遥测 (Telemetry)

**POST /api/telemetry/ingest**

鉴权：ADMIN / OPERATOR

说明：支持单条或批量（最多 200 条）写入

单条请求体：

| 字段              | 类型         | 必填 | 规则                                       | 示例                   |
| ----------------- | ------------ | ---- | ------------------------------------------ | ---------------------- |
| sensor_channel_id | number       | 是   | 传感器通道 ID                              | 1                      |
| metric_code       | string       | 是   | 1-32 字符                                  | "TEMP"                 |
| value             | number       | 是   | 处理后数值                                 | 25.50                  |
| raw_value         | number\|null | 否   | 原始值                                     | 25.50                  |
| quality_flag      | string       | 否   | normal/missing/out_of_range/device_offline | "normal"               |
| collected_at      | string       | 是   | ISO 8601                                   | "2026-01-01T12:00:00Z" |
| batch_id          | number\|null | 否   | 批次 ID                                    | 1                      |

批量请求体：

```json
{
  "items": [
    {
      "sensor_channel_id": 1,
      "metric_code": "TEMP",
      "value": 25.5,
      "collected_at": "2026-01-01T12:00:00Z"
    },
    {
      "sensor_channel_id": 2,
      "metric_code": "HUMIDITY",
      "value": 65.0,
      "collected_at": "2026-01-01T12:00:00Z"
    }
  ]
}
```

**GET /api/telemetry/query**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：

| 字段              | 类型   | 必填 | 说明                                                    |
| ----------------- | ------ | ---- | ------------------------------------------------------- |
| sensor_channel_id | string | 否   | 传感器通道 ID，支持逗号分隔多 ID（上限 50），如 "1,2,3" |
| metric_code       | string | 否   | 指标编码                                                |
| start_time        | string | 否   | 开始时间                                                |
| end_time          | string | 否   | 结束时间                                                |
| batch_id          | number | 否   | 批次 ID                                                 |
| quality_flag      | string | 否   | 质量标志                                                |
| page              | number | 否   | 页码，默认 1                                            |
| page_size         | number | 否   | 每页条数，默认 20                                       |

响应：分页列表，items 为 TelemetryRecordResponse[]

**GET /api/telemetry/subscribe**

鉴权：ADMIN / OPERATOR / VIEWER

说明：SSE 订阅遥测实时事件（事件类型：`telemetry_update`），用于采集中心实时总览。支持按设备/指标过滤。

查询参数：

| 字段         | 类型   | 必填 | 说明                                                                                         |
| ------------ | ------ | ---- | -------------------------------------------------------------------------------------------- |
| token        | string | 否   | JWT（仅用于 EventSource 场景无法自定义 Header 时）；也可使用 `Authorization: Bearer <token>` |
| device_codes | string | 否   | 逗号分隔设备编码，如 `"SENSOR-001,SENSOR-002"`                                               |
| metric_codes | string | 否   | 逗号分隔指标编码，如 `"TEMP,PH,EC"`                                                          |

推送数据（SSE `data:` 行 JSON）：

```json
{
  "type": "telemetry_update",
  "data": {
    "schema_version": 1,
    "sensor_channel_id": 1,
    "metric_code": "TEMP",
    "value": 25.5,
    "quality_flag": "normal",
    "collected_at": "2026-01-01T12:00:00Z",
    "device_code": "SENSOR-001"
  }
}
```

**GET /api/telemetry/channels/latest**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：

| 字段 | 类型   | 必填 | 说明                                    |
| ---- | ------ | ---- | --------------------------------------- |
| ids  | string | 是   | 逗号分隔的通道 ID，上限 100，如 "1,2,3" |

说明：批量获取多个通道的最新遥测数据（内存缓存 + DB 回退）

响应：`{ "items": [{ "sensor_channel_id": 1, "metric_code": "TEMP", "value": 25.5, "quality_flag": "normal", "collected_at": "..." }] }`

**GET /api/telemetry/channels/:channelId/latest**

鉴权：ADMIN / OPERATOR / VIEWER

说明：获取指定传感器通道的最新一条遥测数据

响应：TelemetryRecordResponse

**GET /api/telemetry/channels/:channelId/history**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：start_time, end_time, page, page_size

说明：获取指定传感器通道的历史遥测数据

响应：分页列表，items 为 TelemetryRecordResponse[]

**DELETE /api/telemetry**

鉴权：ADMIN

查询参数：before（删除此时间之前的记录）

响应：

```json
{ "code": 0, "data": { "deleted": 1000 } }
```

### 3.11 指标定义 (Metrics)

**GET /api/metrics**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, is_core

响应：分页列表，items 为 MetricDefinitionResponse[]

**GET /api/metrics/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：MetricDefinitionResponse

### 3.12 控制命令 (Commands)

**POST /api/commands**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                | 类型   | 必填 | 规则                      | 示例           |
| ------------------- | ------ | ---- | ------------------------- | -------------- |
| actuator_channel_id | number | 是   | 执行器通道 ID             | 1              |
| command_type        | string | 是   | 1-32 字符                 | "SWITCH"       |
| payload             | object | 是   | 命令负载                  | {"state":"ON"} |
| request_id          | string | 否   | 请求追踪 ID，最多 64 字符 | "req_abc123"   |

**GET /api/commands**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, status, actuator_channel_id

响应：分页列表，items 为 CommandResponse[]

**GET /api/commands/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：CommandResponse（含 receipts）

**POST /api/commands/:id/send**

鉴权：ADMIN / OPERATOR

请求体：

| 字段       | 类型   | 必填 | 规则         |
| ---------- | ------ | ---- | ------------ |
| request_id | string | 否   | 最多 64 字符 |

**POST /api/commands/:id/ack**

鉴权：ADMIN / OPERATOR

请求体：

| 字段        | 类型   | 必填 | 规则          | 示例               |
| ----------- | ------ | ---- | ------------- | ------------------ |
| ack_code    | string | 是   | 最多 32 字符  | "OK"               |
| ack_message | string | 否   | 最多 255 字符 | "Command executed" |
| ack_payload | object | 否   | 确认负载      | {}                 |

**POST /api/commands/:id/receipts**

鉴权：ADMIN / OPERATOR

请求体：

| 字段           | 类型   | 必填 | 规则                               | 示例       |
| -------------- | ------ | ---- | ---------------------------------- | ---------- |
| receipt_seq    | number | 是   | >=1                                | 1          |
| receipt_status | string | 是   | ACCEPTED/REJECTED/PROCESSED/FAILED | "ACCEPTED" |
| ack_code       | string | 否   | 最多 32 字符                       | "OK"       |
| ack_message    | string | 否   | 最多 255 字符                      | ""         |
| ack_payload    | object | 否   | 确认负载                           | {}         |

**GET /api/commands/:id/receipts**

鉴权：ADMIN / OPERATOR / VIEWER

响应：items 为 CommandReceiptResponse[]

**GET /api/commands/:id/receipts/:receiptId**

鉴权：ADMIN / OPERATOR / VIEWER

响应：CommandReceiptResponse

### 3.12.1 配置投递 (Config Deliveries)

**GET /api/config-deliveries**

鉴权：ADMIN / OPERATOR

说明：查询设备配置投递记录（用于排障与验收可靠投递/回执/重试闭环）。

查询参数：

| 字段        | 类型   | 必填 | 说明                                         |
| ----------- | ------ | ---- | -------------------------------------------- |
| page        | number | 否   | 页码，默认 1                                 |
| page_size   | number | 否   | 每页条数，默认 20，上限 200                  |
| device_code | string | 否   | 设备编码过滤                                 |
| config_type | string | 否   | 配置类型过滤                                 |
| status      | string | 否   | 状态过滤：PENDING/SENT/ACKED/REJECTED/FAILED |
| msg_id      | string | 否   | 消息 ID 过滤                                 |
| entity_id   | string | 否   | 实体 ID 过滤                                 |

响应：分页列表，items 为 ConfigDeliveryResponse（不含 request_payload/ack_payload）

**GET /api/config-deliveries/:id**

鉴权：ADMIN / OPERATOR

说明：获取单条投递详情（含 request_payload/ack_payload）。

响应：ConfigDeliveryResponse

### 3.13 控制策略 (Policies)

#### 策略 CRUD

**POST /api/policies**

鉴权：ADMIN / OPERATOR

请求体：

| 字段            | 类型         | 必填 | 规则                        | 示例                   |
| --------------- | ------------ | ---- | --------------------------- | ---------------------- |
| policy_code     | string       | 是   | 1-64 字符（唯一）           | "POL-TEMP"             |
| name            | string       | 是   | 1-128 字符                  | "温度控制策略"         |
| policy_type     | string       | 是   | THRESHOLD/SCHEDULE/DURATION | "THRESHOLD"            |
| greenhouse_id   | number       | 是   | 温室 ID                     | 1                      |
| growing_zone_id | number\|null | 否   | 种植区 ID                   | 1                      |
| priority        | number       | 否   | 优先级，默认 100            | 100                    |
| retry_limit     | number       | 否   | 重试次数，默认 3，最多 10   | 3                      |
| timeout_sec     | number       | 否   | 超时秒数，默认 30，>=1      | 30                     |
| enabled         | number       | 否   | 1/0，默认 1                 | 1                      |
| version         | string       | 否   | 1-32 字符                   | "v1"                   |
| effective_from  | string       | 否   | ISO 8601                    | "2026-01-01T00:00:00Z" |
| effective_to    | string       | 否   | ISO 8601                    | null                   |

**POST /api/policies/full**

鉴权：ADMIN / OPERATOR

说明：创建策略并嵌套创建条件和目标

请求体：CreatePolicyWithNestedRequest，包含 policy_code, name, policy_type, greenhouse_id 以及 conditions 数组（每个 condition 可含 targets 数组）

**GET /api/policies**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, greenhouse_id, policy_type, enabled

响应：分页列表，items 为 ControlPolicyResponse[]

**GET /api/policies/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：ControlPolicyResponse（含 conditions 和 targets）

**PUT /api/policies/:id**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选

**DELETE /api/policies/:id**

鉴权：ADMIN

响应：

```json
{ "code": 0, "data": {} }
```

#### 策略操作

**POST /api/policies/:id/publish**

鉴权：ADMIN

请求体：

| 字段    | 类型   | 必填 | 规则      |
| ------- | ------ | ---- | --------- |
| version | string | 否   | 1-32 字符 |

**POST /api/policies/:id/archive**

鉴权：ADMIN

说明：归档策略（设置 enabled=0）

**POST /api/policies/:id/execute**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                | 类型         | 必填 | 规则                      | 示例     |
| ------------------- | ------------ | ---- | ------------------------- | -------- |
| trigger_source      | string       | 是   | MANUAL/TELEMETRY/SCHEDULE | "MANUAL" |
| trigger_metric_code | string       | 否   | 1-32 字符                 | "TEMP"   |
| trigger_value       | number\|null | 否   | 触发值                    | 30.50    |

**GET /api/policies/:id/executions**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size

响应：分页列表，items 为 PolicyExecutionResponse[]

#### 策略条件 (PolicyConditions)

**GET /api/policies/:id/conditions**

鉴权：ADMIN / OPERATOR / VIEWER

响应：items 为 PolicyConditionResponse[]

**POST /api/policies/:id/conditions**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                  | 类型         | 必填 | 规则                    | 示例   |
| --------------------- | ------------ | ---- | ----------------------- | ------ |
| metric_code           | string       | 是   | 1-32 字符               | "TEMP" |
| operator              | string       | 是   | >/>=/</<=/==/!=         | ">"    |
| threshold_value       | number       | 是   | 阈值                    | 30.0   |
| hysteresis            | number\|null | 否   | 滞后值                  | 1.0    |
| window_sec            | number\|null | 否   | 滑动窗口（秒），>=1     | 60     |
| required_duration_sec | number\|null | 否   | 持续时间要求（秒），>=1 | 30     |
| aggregation           | string       | 否   | avg/max/min/last        | "avg"  |
| enabled               | number       | 否   | 1/0，默认 1             | 1      |

**GET /api/policies/:id/conditions/:conditionId**

鉴权：ADMIN / OPERATOR / VIEWER

响应：PolicyConditionResponse

**PUT /api/policies/:id/conditions/:conditionId**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选

**DELETE /api/policies/:id/conditions/:conditionId**

鉴权：ADMIN / OPERATOR

#### 策略目标 (PolicyTargets)

**GET /api/policies/:id/targets**

鉴权：ADMIN / OPERATOR / VIEWER

响应：items 为 PolicyTargetResponse[]

**POST /api/policies/:id/targets**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                | 类型   | 必填 | 规则          | 示例           |
| ------------------- | ------ | ---- | ------------- | -------------- |
| actuator_channel_id | number | 是   | 执行器通道 ID | 1              |
| command_type        | string | 是   | 1-32 字符     | "SWITCH"       |
| command_payload     | object | 是   | 命令负载      | {"state":"ON"} |
| execution_order     | number | 否   | >=1，默认 1   | 1              |
| enabled             | number | 否   | 1/0，默认 1   | 1              |

**GET /api/policies/:id/targets/:targetId**

鉴权：ADMIN / OPERATOR / VIEWER

响应：PolicyTargetResponse

**PUT /api/policies/:id/targets/:targetId**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选

**DELETE /api/policies/:id/targets/:targetId**

鉴权：ADMIN / OPERATOR

#### 策略执行记录（全局）

**GET /api/policy-executions**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, policy_id

响应：分页列表，items 为 PolicyExecutionResponse[]

**GET /api/policy-executions/:executionId**

鉴权：ADMIN / OPERATOR / VIEWER

响应：PolicyExecutionResponse

### 3.14 告警 (Alerts)

**POST /api/alerts**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                | 类型         | 必填 | 规则                            | 示例                   |
| ------------------- | ------------ | ---- | ------------------------------- | ---------------------- |
| type                | string       | 是   | THRESHOLD/DEVICE_OFFLINE/SYSTEM | "THRESHOLD"            |
| level               | string       | 是   | INFO/WARN/CRITICAL              | "WARN"                 |
| metric_code         | string       | 否   | 最多 32 字符                    | "TEMP"                 |
| sensor_channel_id   | number\|null | 否   | 传感器通道 ID                   | 1                      |
| actuator_channel_id | number\|null | 否   | 执行器通道 ID                   | null                   |
| trigger_value       | number\|null | 否   | 触发值                          | 35.0                   |
| message             | string       | 是   | 最多 255 字符                   | "温度超出阈值"         |
| triggered_at        | string       | 是   | ISO 8601                        | "2026-01-01T12:00:00Z" |

**GET /api/alerts**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, status, level, type

响应：分页列表，items 为 AlertResponse[]

**GET /api/alerts/subscribe**

鉴权：ADMIN / OPERATOR / VIEWER

说明：SSE 订阅告警实时事件（事件类型：`new_alert`），用于顶部通知与告警中心实时刷新。支持按级别与设备过滤。

查询参数：

| 字段         | 类型   | 必填 | 说明                                                                                         |
| ------------ | ------ | ---- | -------------------------------------------------------------------------------------------- |
| token        | string | 否   | JWT（仅用于 EventSource 场景无法自定义 Header 时）；也可使用 `Authorization: Bearer <token>` |
| level        | string | 否   | 告警级别过滤，支持逗号分隔，如 `"CRITICAL,WARN"`                                             |
| device_codes | string | 否   | 逗号分隔设备编码，如 `"SENSOR-001,ACT-001"`（也兼容 `device_code` 单值）                     |

推送数据（SSE `data:` 行 JSON）：

```json
{
  "type": "new_alert",
  "data": {
    "schema_version": 1,
    "id": 123,
    "device_code": "SENSOR-001",
    "type": "DEVICE_OFFLINE",
    "level": "WARN",
    "metric_code": "",
    "sensor_channel_id": null,
    "actuator_channel_id": null,
    "batch_id": null,
    "trigger_value": null,
    "message": "[SENSOR-001] 设备离线: xxx",
    "status": "OPEN",
    "triggered_at": "2026-01-01T12:00:00Z",
    "resolved_at": null,
    "resolved_by": null,
    "timeline_count": 1,
    "created_at": "2026-01-01T12:00:00Z",
    "updated_at": "2026-01-01T12:00:00Z"
  }
}
```

**GET /api/devices/subscribe**

鉴权：ADMIN / OPERATOR / VIEWER

说明：SSE 订阅设备状态事件（事件类型：`device_status`），用于实时在线状态展示。支持按设备过滤。

查询参数：

| 字段         | 类型   | 必填 | 说明                                                                                         |
| ------------ | ------ | ---- | -------------------------------------------------------------------------------------------- |
| token        | string | 否   | JWT（仅用于 EventSource 场景无法自定义 Header 时）；也可使用 `Authorization: Bearer <token>` |
| device_codes | string | 否   | 逗号分隔设备编码，如 `"SENSOR-001,ACT-001"`                                                  |

推送数据（SSE `data:` 行 JSON）：

```json
{
  "type": "device_status",
  "data": {
    "schema_version": 1,
    "device_code": "SENSOR-001",
    "status": "OFFLINE",
    "reason": "heartbeat_timeout",
    "reported_at": "2026-01-01T12:00:00Z"
  }
}
```

**GET /api/commands/subscribe**

鉴权：ADMIN / OPERATOR

说明：SSE 订阅命令下发/回执事件（事件类型：`command_dispatched`、`command_acked`），用于策略控制页实时回显。支持按设备过滤。

查询参数：

| 字段         | 类型   | 必填 | 说明                                                                                         |
| ------------ | ------ | ---- | -------------------------------------------------------------------------------------------- |
| token        | string | 否   | JWT（仅用于 EventSource 场景无法自定义 Header 时）；也可使用 `Authorization: Bearer <token>` |
| device_codes | string | 否   | 逗号分隔设备编码，如 `"ACT-001,ACT-002"`                                                     |

推送数据（SSE `data:` 行 JSON）：

```json
{
  "type": "command_dispatched",
  "data": {
    "schema_version": 1,
    "command_id": 10,
    "device_code": "ACT-001",
    "status": "SENT",
    "dispatched_at": "2026-01-01T12:00:00Z",
    "source_type": "MANUAL",
    "source_id": 0
  }
}
```

```json
{
  "type": "command_acked",
  "data": {
    "schema_version": 1,
    "command_id": 10,
    "device_code": "ACT-001",
    "ack_code": "OK",
    "ack_message": "ok",
    "ack_payload": {},
    "acked_at": "2026-01-01T12:00:01Z"
  }
}
```

**GET /api/alerts/stats**

鉴权：ADMIN / OPERATOR / VIEWER

响应：AlertStatsResponse

```json
{
  "open_count": 5,
  "acknowledged_count": 3,
  "resolved_count": 10,
  "ignored_count": 1,
  "info_count": 4,
  "warn_count": 8,
  "critical_count": 7
}
```

**GET /api/alerts/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：AlertResponse（含 timeline_count）

**PATCH /api/alerts/:id/status**

鉴权：ADMIN / OPERATOR

请求体：

| 字段        | 类型         | 必填 | 规则                               | 示例                   |
| ----------- | ------------ | ---- | ---------------------------------- | ---------------------- |
| status      | string       | 是   | OPEN/ACKNOWLEDGED/RESOLVED/IGNORED | "RESOLVED"             |
| resolved_at | string\|null | 否   | 解决时间                           | "2026-01-01T13:00:00Z" |
| comment     | string       | 否   | 备注，最多 255 字符                | "已处理"               |
| resolved_by | number\|null | 否   | 解决人用户 ID                      | 1                      |

**GET /api/alerts/:id/timeline**

鉴权：ADMIN / OPERATOR / VIEWER

响应：items 为 AlertTimelineEventResponse[]

**POST /api/alerts/:id/timeline**

鉴权：ADMIN / OPERATOR

请求体：

| 字段          | 类型         | 必填 | 规则                                                              | 示例                   |
| ------------- | ------------ | ---- | ----------------------------------------------------------------- | ---------------------- |
| event_type    | string       | 是   | TRIGGERED/AUTO_ACTION/MANUAL_ACTION/ACKNOWLEDGED/RESOLVED/COMMENT | "COMMENT"              |
| event_source  | string       | 是   | SYSTEM/MANUAL                                                     | "MANUAL"               |
| operator_id   | number\|null | 否   | 操作人用户 ID                                                     | 1                      |
| comment       | string       | 否   | 备注，最多 255 字符                                               | "正在处理"             |
| event_payload | string       | 否   | JSON 字符串                                                       | "{}"                   |
| event_time    | string       | 是   | 事件时间                                                          | "2026-01-01T12:30:00Z" |

### 3.15 通知渠道 (Notification Channels)

**GET /api/notification-channels**

鉴权：ADMIN / OPERATOR / VIEWER（仅返回当前用户的渠道）

响应：

```json
{
  "code": 0,
  "data": {
    "items": [{ "id": 1, "channel_type": "EMAIL", "name": "邮件通知", "config": {...}, "min_alert_level": "WARN", "enabled": true }]
  }
}
```

**POST /api/notification-channels**

鉴权：ADMIN / OPERATOR

请求体：

| 字段            | 类型    | 必填 | 规则                          | 示例                            |
| --------------- | ------- | ---- | ----------------------------- | ------------------------------- |
| channel_type    | string  | 是   | EMAIL/SMS/WEBHOOK/IN_APP      | "EMAIL"                         |
| name            | string  | 是   | 1-64 字符                     | "邮件通知"                      |
| config          | object  | 是   | 渠道配置（JSON）              | {"address":"admin@example.com"} |
| min_alert_level | string  | 否   | INFO/WARN/CRITICAL，默认 WARN | "WARN"                          |
| enabled         | boolean | 是   | 是否启用                      | true                            |

**PUT /api/notification-channels/:channelId**

鉴权：ADMIN / OPERATOR（仅限当前用户自己的渠道）

请求体：同创建但所有字段可选

**DELETE /api/notification-channels/:channelId**

鉴权：ADMIN / OPERATOR（仅限当前用户自己的渠道）

**POST /api/notification-channels/:channelId/test**

鉴权：ADMIN / OPERATOR

说明：发送测试通知（仅 WEBHOOK 类型支持）

响应：

```json
{ "code": 0, "data": { "sent": true } }
```

### 3.16 仪表盘概览 (Overview Dashboard)

**GET /api/overview/dashboard**

鉴权：ADMIN / OPERATOR / VIEWER

响应：

```json
{
  "code": 0,
  "data": {
    "sensors_online": 10,
    "sensors_offline": 2,
    "sensors_total": 12,
    "actuators_online": 8,
    "actuators_offline": 1,
    "actuators_total": 9,
    "alerts_open": 5,
    "alerts_critical": 2,
    "alerts_today": 3,
    "greenhouse_summary": [
      {
        "greenhouse_id": 1,
        "name": "一号温室",
        "sensor_count": 6,
        "actuator_count": 4,
        "zone_count": 4,
        "avg_temp": 25.5,
        "avg_humidity": 65.0
      }
    ],
    "recent_commands": [
      {
        "id": 1,
        "command_type": "SWITCH",
        "device_name": "循环水泵",
        "status": "SENT",
        "created_at": "2026-01-01T12:00:00Z"
      }
    ]
  }
}
```

### 3.17 审计日志 (Audit Logs)

**GET /api/audit-logs**

鉴权：ADMIN

查询参数：page, page_size, user_id, action, target_type

响应：分页列表，items 为 AuditLog[]

### 3.18 品种管理 (Crop Varieties)

**GET /api/crop-varieties**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size

响应：分页列表，items 为 CropVarietyResponse[]

**POST /api/crop-varieties**

鉴权：ADMIN

请求体：

| 字段               | 类型         | 必填 | 规则              | 示例           |
| ------------------ | ------------ | ---- | ----------------- | -------------- |
| code               | string       | 是   | 1-32 字符（唯一） | "LETTUCE-ICE"  |
| name               | string       | 是   | 1-64 字符         | "冰菜"         |
| description        | string       | 否   | 最多 255 字符     | "水培专用品种" |
| default_cycle_days | number\|null | 否   | 默认生长周期天数  | 45             |

**GET /api/crop-varieties/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：CropVarietyResponse

**PUT /api/crop-varieties/:id**

鉴权：ADMIN

请求体：name, description, default_cycle_days 均可选

**DELETE /api/crop-varieties/:id**

鉴权：ADMIN

### 3.19 生长阶段 (Growth Stages)

**GET /api/growth-stages**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size

响应：分页列表，items 为 GrowthStageResponse[]

**POST /api/growth-stages**

鉴权：ADMIN

请求体：

| 字段                  | 类型         | 必填 | 规则              | 示例       |
| --------------------- | ------------ | ---- | ----------------- | ---------- |
| code                  | string       | 是   | 1-32 字符（唯一） | "SEEDLING" |
| name                  | string       | 是   | 1-64 字符         | "育苗期"   |
| sort_order            | number       | 否   | 排序              | 1          |
| default_duration_days | number\|null | 否   | 默认持续天数      | 14         |

**GET /api/growth-stages/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：GrowthStageResponse

**PUT /api/growth-stages/:id**

鉴权：ADMIN

请求体：name, sort_order, default_duration_days 均可选

**DELETE /api/growth-stages/:id**

鉴权：ADMIN

### 3.20 种植批次 (Crop Batches)

**GET /api/batches**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, status, greenhouse_id, growing_zone_id, crop_variety_id

响应：分页列表，items 为 CropBatchResponse[]

**POST /api/batches**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                | 类型         | 必填 | 规则              | 示例                   |
| ------------------- | ------------ | ---- | ----------------- | ---------------------- |
| batch_no            | string       | 是   | 1-64 字符（唯一） | "BATCH-2026-001"       |
| greenhouse_id       | number       | 是   | 温室 ID           | 1                      |
| growing_zone_id     | number\|null | 否   | 种植区 ID         | 1                      |
| crop_variety_id     | number       | 是   | 品种 ID           | 1                      |
| status              | string       | 否   | 默认 PLANNED      | "PLANNED"              |
| planting_density    | number\|null | 否   | 种植密度          | 25.00                  |
| total_plants        | number\|null | 否   | 总株数            | 1000                   |
| started_at          | string\|null | 否   | 开始时间          | "2026-01-01T00:00:00Z" |
| ended_at            | string\|null | 否   | 结束时间          | null                   |
| expected_harvest_at | string\|null | 否   | 预计收获时间      | "2026-02-15T00:00:00Z" |
| recipe_version      | string       | 否   | 配方版本          | "v1"                   |
| policy_version      | string       | 否   | 策略版本          | "v1"                   |
| note                | string       | 否   | 备注              | ""                     |

**GET /api/batches/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：CropBatchResponse

**PUT /api/batches/:id**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选

**DELETE /api/batches/:id**

鉴权：ADMIN

**POST /api/batches/:id/transition**

鉴权：ADMIN / OPERATOR

请求体：

| 字段   | 类型   | 必填 | 规则                                 | 示例      |
| ------ | ------ | ---- | ------------------------------------ | --------- |
| status | string | 是   | RUNNING/HARVESTING/COMPLETED/ABORTED | "RUNNING" |
| note   | string | 否   | 备注                                 | ""        |

### 3.21 批次阶段计划 (Batch Stage Plans)

**GET /api/batch-stage-plans**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, batch_id

响应：分页列表，items 为 BatchStagePlanResponse[]

**POST /api/batch-stage-plans**

鉴权：ADMIN / OPERATOR

请求体：

| 字段               | 类型         | 必填 | 规则                | 示例                   |
| ------------------ | ------------ | ---- | ------------------- | ---------------------- |
| batch_id           | number       | 是   | 批次 ID             | 1                      |
| growth_stage_id    | number       | 是   | 生长阶段 ID         | 1                      |
| recipe_id          | number\|null | 否   | 阶段配方 ID         | 10                     |
| policy_id          | number\|null | 否   | 阶段策略 ID         | 20                     |
| climate_profile_id | number\|null | 否   | 阶段气候 Profile ID | 30                     |
| stage_start_at     | string       | 是   | 阶段开始时间        | "2026-01-01T00:00:00Z" |
| stage_end_at       | string       | 是   | 阶段结束时间        | "2026-01-14T00:00:00Z" |
| target_ec_min      | number\|null | 否   | 目标 EC 下限        | 1.2                    |
| target_ec_max      | number\|null | 否   | 目标 EC 上限        | 1.8                    |
| target_ph_min      | number\|null | 否   | 目标 pH 下限        | 5.5                    |
| target_ph_max      | number\|null | 否   | 目标 pH 上限        | 6.5                    |

**GET /api/batch-stage-plans/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：BatchStagePlanResponse

**PUT /api/batch-stage-plans/:id**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选（`recipe_id/policy_id/climate_profile_id` 支持显式传 null 清空）

**DELETE /api/batch-stage-plans/:id**

鉴权：ADMIN

### 3.22 收获记录 (Harvest Records)

**GET /api/harvests**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, batch_id

响应：分页列表，items 为 HarvestRecordResponse[]

**POST /api/harvests**

鉴权：ADMIN / OPERATOR

请求体：

| 字段              | 类型   | 必填 | 规则     | 示例                   |
| ----------------- | ------ | ---- | -------- | ---------------------- |
| batch_id          | number | 是   | 批次 ID  | 1                      |
| harvested_at      | string | 是   | 收获时间 | "2026-02-15T00:00:00Z" |
| harvest_weight_kg | number | 是   | >0       | 50.500                 |
| grade             | string | 否   | 默认 A   | "A"                    |
| grade_weight_kg   | number | 是   | >0       | 30.000                 |
| note              | string | 否   | 备注     | ""                     |

**GET /api/harvests/summary/:batchId**

鉴权：ADMIN / OPERATOR / VIEWER

响应：HarvestSummaryResponse

```json
{
  "batch_id": 1,
  "total_weight_kg": 150.5,
  "grades": [
    { "grade": "A", "weight_kg": 100.0, "count": 2 },
    { "grade": "B", "weight_kg": 50.5, "count": 1 }
  ]
}
```

### 3.23 营养液配方 (Recipes)

**GET /api/recipes**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, status, crop_variety_id

响应：分页列表，items 为 NutrientRecipeResponse[]

**POST /api/recipes**

鉴权：ADMIN / OPERATOR

请求体：

| 字段            | 类型          | 必填 | 规则                  | 示例                   |
| --------------- | ------------- | ---- | --------------------- | ---------------------- |
| recipe_code     | string        | 是   | 1-64 字符（唯一）     | "REC-LETTUCE"          |
| name            | string        | 是   | 1-128 字符            | "生菜标准配方"         |
| crop_variety_id | number\|null  | 否   | 品种 ID               | 1                      |
| description     | string        | 否   | 描述                  | ""                     |
| version         | string        | 否   | 默认 v1               | "v1"                   |
| status          | string        | 否   | DRAFT/ACTIVE/ARCHIVED | "DRAFT"                |
| effective_from  | string\|null  | 否   | 生效时间              | "2026-01-01T00:00:00Z" |
| effective_to    | string\|null  | 否   | 失效时间              | null                   |
| stage_targets   | StageTarget[] | 否   | 阶段指标目标列表      | []                     |
| ion_targets     | IonTarget[]   | 否   | 离子目标列表          | []                     |

**GET /api/recipes/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：NutrientRecipeResponse

**PUT /api/recipes/:id**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选，status 支持 DRAFT/ACTIVE/ARCHIVED

**DELETE /api/recipes/:id**

鉴权：ADMIN

**POST /api/recipes/:id/publish**

鉴权：ADMIN

请求体：

| 字段    | 类型   | 必填 | 规则      |
| ------- | ------ | ---- | --------- |
| version | string | 是   | 1-32 字符 |

#### 配方目标

**GET /api/recipes/:id/targets**

鉴权：ADMIN / OPERATOR / VIEWER

响应：RecipeTargetsResponse（含 stage_targets 和 ion_targets）

**PUT /api/recipes/:id/targets**

鉴权：ADMIN / OPERATOR

请求体：UpdateRecipeTargetsRequest（含 stage_targets 和 ion_targets）

#### 配方绑定

**POST /api/recipes/:id/bind**

鉴权：ADMIN / OPERATOR

请求体：CreateBindingRequest（含 batch_id, recipe_id, binding_type, effective_from 等）

**GET /api/recipe-bindings**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, batch_id, recipe_id

响应：分页列表，items 为 BatchRecipeBindingResponse[]

**GET /api/recipe-bindings/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：BatchRecipeBindingResponse

**PUT /api/recipe-bindings/:id**

鉴权：ADMIN / OPERATOR

请求体：UpdateBindingRequest（binding_type, version, status 等均可选）

**DELETE /api/recipe-bindings/:id**

鉴权：ADMIN

### 3.24 气候控制 (Climate Profiles)

#### 气候配置 CRUD

**POST /api/climate-profiles**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                      | 类型   | 必填 | 规则                          | 示例              |
| ------------------------- | ------ | ---- | ----------------------------- | ----------------- |
| greenhouse_id             | number | 是   | 温室 ID                       | 1                 |
| code                      | string | 是   | 1-64 字符                     | "CLIMATE-DEFAULT" |
| name                      | string | 是   | 1-128 字符                    | "默认气候配置"    |
| description               | string | 否   | 最多 255 字符                 | ""                |
| trigger_metric_code       | string | 是   | 1-32 字符                     | "TEMP"            |
| trigger_sensor_channel_id | number | 是   | 采集通道 ID（固定单通道触发） | 123               |
| enabled                   | number | 否   | 1/0，默认 1                   | 1                 |

说明：

- 自动联动触发依据为 `trigger_sensor_channel_id` 对应采集通道的遥测值（不会被其他温室/其他采集通道影响）
- 当触发采集通道被禁用（enabled=0）时，系统会自动将引用该通道的气候配置置为 enabled=0

**POST /api/climate-profiles/full**

鉴权：ADMIN / OPERATOR

说明：创建气候配置并嵌套创建阶段和动作

请求体：CreateClimateProfileWithStagesRequest，包含温室配置信息及 stages 数组（每个 stage 可含 actions 数组）；必须提供 trigger_sensor_channel_id

**GET /api/climate-profiles**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, greenhouse_id

响应：分页列表，items 为 ClimateProfileResponse[]

**GET /api/climate-profiles/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：ClimateProfileResponse（含 stages）

**PUT /api/climate-profiles/:id**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选（支持 trigger_sensor_channel_id）

**DELETE /api/climate-profiles/:id**

鉴权：ADMIN

#### 气候阶段 (ClimateStages)

**POST /api/climate-profiles/:id/stages**

鉴权：ADMIN / OPERATOR

请求体：

| 字段              | 类型         | 必填 | 规则      | 示例       |
| ----------------- | ------------ | ---- | --------- | ---------- |
| stage_level       | number       | 是   | >=1       | 1          |
| name              | string       | 是   | 1-64 字符 | "低温阶段" |
| trigger_operator  | string       | 是   | >/>=/</<= | "<"        |
| trigger_threshold | number       | 是   | 触发阈值  | 18.0       |
| hysteresis        | number\|null | 否   | 滞后值    | 1.0        |

约束：

- 同一气候配置（profile）的所有阶段不允许混用触发方向（只能全为 `>`/`>=` 或全为 `<`/`<=`）
- 阈值需随 `stage_level` 单调变化（`>` 方向递增，`<` 方向递减）

**GET /api/climate-profiles/:id/stages**

鉴权：ADMIN / OPERATOR / VIEWER

响应：items 为 ClimateStageResponse[]

**GET /api/climate-profiles/:id/stages/:stageId**

鉴权：ADMIN / OPERATOR / VIEWER

响应：ClimateStageResponse（含 actions）

**PUT /api/climate-profiles/:id/stages/:stageId**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选

**DELETE /api/climate-profiles/:id/stages/:stageId**

鉴权：ADMIN / OPERATOR

#### 气候阶段动作 (ClimateStageActions)

**POST /api/climate-profiles/:id/stages/:stageId/actions**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                | 类型   | 必填 | 规则          | 示例           |
| ------------------- | ------ | ---- | ------------- | -------------- |
| actuator_channel_id | number | 是   | 执行器通道 ID | 1              |
| command_type        | string | 是   | 1-32 字符     | "SWITCH"       |
| command_payload     | object | 是   | 命令负载      | {"state":"ON"} |
| execution_order     | number | 否   | >=1，默认 1   | 1              |
| enabled             | number | 否   | 1/0，默认 1   | 1              |

**PUT /api/climate-profiles/:id/stages/:stageId/actions/:actionId**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选

**DELETE /api/climate-profiles/:id/stages/:stageId/actions/:actionId**

鉴权：ADMIN / OPERATOR

#### 气候执行日志

**POST /api/climate-profiles/:id/execute**

鉴权：ADMIN / OPERATOR

请求体：

| 字段             | 类型         | 必填 | 规则     | 示例 |
| ---------------- | ------------ | ---- | -------- | ---- |
| trigger_value    | number       | 是   | 触发值   | 15.0 |
| from_stage_level | number\|null | 否   | 来源阶段 | null |
| to_stage_level   | number       | 是   | >=1      | 2    |

**GET /api/climate-profiles/:id/execution-logs**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size

响应：分页列表，items 为 ClimateExecutionLogResponse[]

**GET /api/climate-execution-logs**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, profile_id

说明：全局气候执行日志列表

响应：分页列表，items 为 ClimateExecutionLogResponse[]

### 3.25 营养液管理 (Nutrient)

#### 营养液水箱

**GET /api/nutrient-tanks**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, growing_zone_id, status

响应：分页列表，items 为 NutrientTankResponse[]

**POST /api/nutrient-tanks**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                    | 类型         | 必填 | 规则                  | 示例      |
| ----------------------- | ------------ | ---- | --------------------- | --------- |
| growing_zone_id         | number       | 是   | 种植区 ID             | 1         |
| code                    | string       | 是   | 1-32 字符             | "TANK-A1" |
| total_volume_liter      | number       | 是   | >0                    | 200.00    |
| current_volume_liter    | number\|null | 否   | 当前容积              | 180.00    |
| status                  | string       | 否   | ACTIVE/INACTIVE/EMPTY | "ACTIVE"  |
| ec_sensor_channel_id    | number\|null | 否   | EC 传感器通道 ID      | 1         |
| ph_sensor_channel_id    | number\|null | 否   | pH 传感器通道 ID      | 2         |
| level_sensor_channel_id | number\|null | 否   | 液位传感器通道 ID     | 3         |
| temp_sensor_channel_id  | number\|null | 否   | 水温传感器通道 ID     | 4         |

**GET /api/nutrient-tanks/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：NutrientTankResponse

**PUT /api/nutrient-tanks/:id**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选

**DELETE /api/nutrient-tanks/:id**

鉴权：ADMIN

#### 换液事件

**GET /api/solution-changes**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, tank_id, change_type

响应：分页列表，items 为 SolutionChangeEventResponse[]

**POST /api/solution-changes**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                  | 类型         | 必填 | 规则                                | 示例                   |
| --------------------- | ------------ | ---- | ----------------------------------- | ---------------------- |
| tank_id               | number       | 是   | 水箱 ID                             | 1                      |
| change_type           | string       | 是   | FULL_REPLACE/PARTIAL_REFRESH/TOP_UP | "FULL_REPLACE"         |
| volume_replaced_liter | number       | 是   | >0                                  | 200.00                 |
| source_water_ec       | number\|null | 否   | 水源 EC                             | 0.3                    |
| source_water_ph       | number\|null | 否   | 水源 pH                             | 7.0                    |
| before_ec             | number\|null | 否   | 更换前 EC                           | 2.5                    |
| before_ph             | number\|null | 否   | 更换前 pH                           | 6.8                    |
| after_ec              | number\|null | 否   | 更换后 EC                           | 1.5                    |
| after_ph              | number\|null | 否   | 更换后 pH                           | 6.0                    |
| nutrient_a_added_ml   | number\|null | 否   | A液添加量                           | 500.00                 |
| nutrient_b_added_ml   | number\|null | 否   | B液添加量                           | 500.00                 |
| acid_added_ml         | number\|null | 否   | 酸液添加量                          | 10.00                  |
| alkali_added_ml       | number\|null | 否   | 碱液添加量                          | 0                      |
| note                  | string       | 否   | 备注                                | ""                     |
| operated_at           | string       | 是   | 操作时间                            | "2026-01-15T00:00:00Z" |

**GET /api/solution-changes/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：SolutionChangeEventResponse

#### 离子检测

**GET /api/ion-tests**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, tank_id, batch_id, test_method

响应：分页列表，items 为 IonTestRecordResponse[]

**POST /api/ion-tests**

鉴权：ADMIN / OPERATOR

请求体：

| 字段         | 类型         | 必填 | 规则                 | 示例                   |
| ------------ | ------------ | ---- | -------------------- | ---------------------- |
| tank_id      | number       | 是   | 水箱 ID              | 1                      |
| batch_id     | number\|null | 否   | 批次 ID              | 1                      |
| sample_code  | string       | 是   | 1-64 字符（唯一）    | "SAMPLE-001"           |
| sampled_at   | string       | 是   | 采样时间             | "2026-01-10T00:00:00Z" |
| tested_at    | string\|null | 否   | 检测时间             | "2026-01-11T00:00:00Z" |
| test_method  | string       | 否   | LAB/STRIP/METER      | "LAB"                  |
| no3_n ~ mo   | number\|null | 否   | 各离子浓度（见模型） | -                      |
| ec_at_sample | number\|null | 否   | 采样时 EC            | 1.5000                 |
| ph_at_sample | number\|null | 否   | 采样时 pH            | 6.0000                 |
| lab_name     | string       | 否   | 实验室名称           | ""                     |
| report_url   | string       | 否   | 报告 URL             | ""                     |
| note         | string       | 否   | 备注                 | ""                     |

**GET /api/ion-tests/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：IonTestRecordResponse

**PUT /api/ion-tests/:id**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选

**DELETE /api/ion-tests/:id**

鉴权：ADMIN

#### 浓缩液库存

**GET /api/concentrate-inventory**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, greenhouse_id, concentrate_type, status

响应：分页列表，items 为 ConcentrateInventoryResponse[]

**POST /api/concentrate-inventory**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                | 类型         | 必填 | 规则                 | 示例         |
| ------------------- | ------------ | ---- | -------------------- | ------------ |
| greenhouse_id       | number       | 是   | 温室 ID              | 1            |
| concentrate_type    | string       | 是   | A/B/ACID/ALKALI      | "A"          |
| brand               | string       | 否   | 品牌                 | ""           |
| product_name        | string       | 否   | 产品名称             | "A液浓缩液"  |
| total_volume_ml     | number       | 是   | >0                   | 5000.00      |
| remaining_volume_ml | number       | 否   | 默认 0               | 5000.00      |
| unit_price          | number\|null | 否   | 单价                 | 50.00        |
| batch_no            | string       | 否   | 批号                 | ""           |
| expired_at          | string\|null | 否   | 过期日期             | "2026-12-31" |
| status              | string       | 否   | IN_USE/EMPTY/EXPIRED | "IN_USE"     |

**GET /api/concentrate-inventory/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：ConcentrateInventoryResponse

**PUT /api/concentrate-inventory/:id**

鉴权：ADMIN / OPERATOR

请求体：同创建但所有字段可选

**DELETE /api/concentrate-inventory/:id**

鉴权：ADMIN

#### 浓缩液使用日志

**GET /api/concentrate-usage-logs**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, inventory_id, tank_id

响应：分页列表，items 为 ConcentrateUsageLogResponse[]

**POST /api/concentrate-usage-logs**

鉴权：ADMIN / OPERATOR

请求体：

| 字段               | 类型         | 必填 | 规则        | 示例                   |
| ------------------ | ------------ | ---- | ----------- | ---------------------- |
| inventory_id       | number       | 是   | 库存 ID     | 1                      |
| solution_change_id | number\|null | 否   | 换液事件 ID | 1                      |
| tank_id            | number\|null | 否   | 水箱 ID     | 1                      |
| volume_used_ml     | number       | 是   | >0          | 500.00                 |
| used_at            | string       | 是   | 使用时间    | "2026-01-15T00:00:00Z" |

### 3.26 能耗管理 (Energy)

**POST /api/energy-records**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                | 类型         | 必填 | 规则                      | 示例                   |
| ------------------- | ------------ | ---- | ------------------------- | ---------------------- |
| greenhouse_id       | number       | 是   | 温室 ID                   | 1                      |
| record_type         | string       | 是   | ELECTRICITY/WATER/CO2_GAS | "ELECTRICITY"          |
| consumption_value   | number       | 是   | 消耗量                    | 150.5000               |
| unit                | string       | 是   | 1-16 字符                 | "kWh"                  |
| record_period_start | string       | 是   | 记录周期开始              | "2026-01-01T00:00:00Z" |
| record_period_end   | string       | 是   | 记录周期结束              | "2026-01-02T00:00:00Z" |
| meter_reading_start | number\|null | 否   | 表头起始读数              | 1000.0000              |
| meter_reading_end   | number\|null | 否   | 表头结束读数              | 1150.5000              |
| batch_id            | number\|null | 否   | 批次 ID                   | 1                      |
| recorded_by         | number\|null | 否   | 记录人                    | 1                      |

**PUT /api/energy-records/:id**

鉴权：ADMIN / OPERATOR

请求体：consumption_value, unit, meter_reading_start, meter_reading_end, batch_id 均可选

**DELETE /api/energy-records/:id**

鉴权：ADMIN

**GET /api/energy-records**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, greenhouse_id, record_type, batch_id

响应：分页列表，items 为 EnergyConsumptionRecord[]

**GET /api/energy-records/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：EnergyConsumptionRecord

**GET /api/energy-records/greenhouse/:greenhouseId**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, record_type

说明：获取指定温室的能耗记录

**GET /api/energy-records/batch/:batchId**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, record_type

说明：获取指定批次的能耗记录

**GET /api/energy-records/summary**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：greenhouse_id, record_type, start_time, end_time

说明：能耗汇总统计

响应：

```json
{
  "code": 0,
  "data": {
    "items": [
      { "record_type": "ELECTRICITY", "total_value": 500.5, "unit": "kWh" },
      { "record_type": "WATER", "total_value": 2000.0, "unit": "L" }
    ]
  }
}
```

### 3.27 病虫害管理 (Pest & Disease)

#### 病虫害观察

**POST /api/pest-observations**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                 | 类型         | 必填 | 规则                    | 示例                   |
| -------------------- | ------------ | ---- | ----------------------- | ---------------------- |
| greenhouse_id        | number       | 是   | 温室 ID                 | 1                      |
| growing_zone_id      | number\|null | 否   | 种植区 ID               | 1                      |
| batch_id             | number\|null | 否   | 批次 ID                 | 1                      |
| observed_at          | string       | 是   | 观察时间                | "2026-01-10T00:00:00Z" |
| pest_or_disease      | string       | 是   | 1-64 字符               | "蚜虫"                 |
| severity             | string       | 是   | LIGHT/MODERATE/SEVERE   | "MODERATE"             |
| affected_area_pct    | number\|null | 否   | 受影响面积百分比        | 15.00                  |
| affected_plant_count | number\|null | 否   | 受影响株数              | 50                     |
| symptoms             | string       | 否   | 症状描述，最多 255 字符 | "叶片有黄斑"           |
| photo_urls           | string       | 否   | 照片 URL（JSON）        | "[]"                   |
| observed_by          | number\|null | 否   | 观察人                  | 1                      |

**PUT /api/pest-observations/:id**

鉴权：ADMIN / OPERATOR

请求体：pest_or_disease, severity, affected_area_pct, affected_plant_count, symptoms, photo_urls 均可选

**DELETE /api/pest-observations/:id**

鉴权：ADMIN

**GET /api/pest-observations**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, greenhouse_id, severity, batch_id

响应：分页列表，items 为 PestDiseaseObservation[]

**GET /api/pest-observations/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：PestDiseaseObservation

**GET /api/pest-observations/greenhouse/:greenhouseId**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, severity

**GET /api/pest-observations/batch/:batchId**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, severity

**GET /api/pest-observations/:id/treatments**

鉴权：ADMIN / OPERATOR / VIEWER

说明：获取指定观察记录关联的治理记录

响应：items 为 TreatmentRecord[]

#### 治理记录

**POST /api/treatment-records**

鉴权：ADMIN / OPERATOR

请求体：

| 字段                   | 类型         | 必填 | 规则                         | 示例                   |
| ---------------------- | ------------ | ---- | ---------------------------- | ---------------------- |
| observation_id         | number\|null | 否   | 关联观察 ID                  | 1                      |
| greenhouse_id          | number       | 是   | 温室 ID                      | 1                      |
| growing_zone_id        | number\|null | 否   | 种植区 ID                    | 1                      |
| batch_id               | number\|null | 否   | 批次 ID                      | 1                      |
| treatment_type         | string       | 是   | CHEMICAL/BIOLOGICAL/PHYSICAL | "BIOLOGICAL"           |
| product_name           | string       | 是   | 1-128 字符                   | "瓢虫"                 |
| active_ingredient      | string       | 否   | 有效成分，最多 128 字符      | ""                     |
| dosage                 | string       | 是   | 1-64 字符                    | "100只/亩"             |
| application_method     | string       | 是   | SPRAY/DRENCH/FOG/RELEASE     | "RELEASE"              |
| safety_interval_days   | number\|null | 否   | 安全间隔天数                 | 7                      |
| reentry_interval_hours | number\|null | 否   | 再进入间隔小时数             | 24                     |
| treated_at             | string       | 是   | 治理时间                     | "2026-01-10T00:00:00Z" |
| treated_by             | number\|null | 否   | 治理人                       | 1                      |
| note                   | string       | 否   | 备注，最多 255 字符          | ""                     |

**PUT /api/treatment-records/:id**

鉴权：ADMIN / OPERATOR

请求体：treatment_type, product_name, active_ingredient, dosage, application_method, safety_interval_days, reentry_interval_hours, note 均可选

**DELETE /api/treatment-records/:id**

鉴权：ADMIN

**GET /api/treatment-records**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, greenhouse_id, treatment_type, batch_id

响应：分页列表，items 为 TreatmentRecord[]

**GET /api/treatment-records/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：TreatmentRecord

**GET /api/treatment-records/greenhouse/:greenhouseId**

鉴权：ADMIN / OPERATOR / VIEWER

**GET /api/treatment-records/batch/:batchId**

鉴权：ADMIN / OPERATOR / VIEWER

### 3.28 批次审查 (Batch Review)

**POST /api/reviews**

鉴权：ADMIN / OPERATOR

请求体：

| 字段          | 类型   | 必填 | 规则                       | 示例                   |
| ------------- | ------ | ---- | -------------------------- | ---------------------- |
| batch_id      | number | 是   | 批次 ID                    | 1                      |
| snapshot_type | string | 是   | DAILY/WEEKLY/STAGE_SUMMARY | "DAILY"                |
| window_start  | string | 是   | 窗口开始                   | "2026-01-01T00:00:00Z" |
| window_end    | string | 是   | 窗口结束                   | "2026-01-02T00:00:00Z" |
| summary       | object | 是   | 摘要数据（JSON）           | {}                     |
| generated_at  | string | 是   | 生成时间                   | "2026-01-02T00:00:00Z" |

**PUT /api/reviews/:id**

鉴权：ADMIN / OPERATOR

请求体：snapshot_type, summary 均可选

**DELETE /api/reviews/:id**

鉴权：ADMIN

**POST /api/reviews/generate**

鉴权：ADMIN / OPERATOR

请求体：

| 字段          | 类型   | 必填 | 规则                       | 示例                   |
| ------------- | ------ | ---- | -------------------------- | ---------------------- |
| batch_id      | number | 是   | 批次 ID                    | 1                      |
| snapshot_type | string | 是   | DAILY/WEEKLY/STAGE_SUMMARY | "DAILY"                |
| window_start  | string | 是   | 窗口开始                   | "2026-01-01T00:00:00Z" |
| window_end    | string | 是   | 窗口结束                   | "2026-01-02T00:00:00Z" |

说明：自动生成审查快照

**GET /api/reviews**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, batch_id, snapshot_type

响应：分页列表，items 为 BatchReviewSnapshot[]

**GET /api/reviews/:id**

鉴权：ADMIN / OPERATOR / VIEWER

响应：BatchReviewSnapshot

**GET /api/reviews/batches/:batchId**

鉴权：ADMIN / OPERATOR / VIEWER

查询参数：page, page_size, snapshot_type

说明：获取指定批次的审查快照

---

## 附录 A：端点索引

| #        | 模块                  | 端点数量 | 路径前缀                                                                                                                      |
| -------- | --------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------- |
| 1        | Auth                  | 9        | `/api/auth`, `/api/users`, `/api/roles`                                                                                       |
| 2        | Greenhouses           | 11       | `/api/greenhouses`, `/api/growing-zones`                                                                                      |
| 3        | SensorDevices         | 5        | `/api/sensor-devices`                                                                                                         |
| 4        | SensorChannels        | 5        | `/api/sensor-channels`                                                                                                        |
| 5        | ActuatorDevices       | 5        | `/api/actuator-devices`                                                                                                       |
| 6        | ActuatorChannels      | 5        | `/api/actuator-channels`                                                                                                      |
| 7        | Telemetry             | 6        | `/api/telemetry`                                                                                                              |
| 8        | Metrics               | 2        | `/api/metrics`                                                                                                                |
| 9        | Commands              | 8        | `/api/commands`                                                                                                               |
| 10       | Policies              | 22       | `/api/policies`, `/api/policy-executions`                                                                                     |
| 11       | Alerts                | 7        | `/api/alerts`                                                                                                                 |
| 12       | Notification Channels | 5        | `/api/notification-channels`                                                                                                  |
| 13       | Overview Dashboard    | 1        | `/api/overview`                                                                                                               |
| 14       | Audit Logs            | 1        | `/api/audit-logs`                                                                                                             |
| 15       | Crop & Batch          | 22       | `/api/crop-varieties`, `/api/growth-stages`, `/api/batches`, `/api/batch-stage-plans`, `/api/harvests`                        |
| 16       | Recipe                | 13       | `/api/recipes`, `/api/recipe-bindings`                                                                                        |
| 17       | Climate               | 17       | `/api/climate-profiles`, `/api/climate-execution-logs`                                                                        |
| 18       | Nutrient              | 20       | `/api/nutrient-tanks`, `/api/solution-changes`, `/api/ion-tests`, `/api/concentrate-inventory`, `/api/concentrate-usage-logs` |
| 19       | Energy                | 8        | `/api/energy-records`                                                                                                         |
| 20       | Pest & Disease        | 15       | `/api/pest-observations`, `/api/treatment-records`                                                                            |
| 21       | Batch Review          | 7        | `/api/reviews`                                                                                                                |
| **合计** |                       | **193**  |                                                                                                                               |

## 附录 B：基础设施端点

以下端点不属于 `/api` 组，无需鉴权：

| 方法 | 路径            | 说明             |
| ---- | --------------- | ---------------- |
| GET  | `/healthz`      | 存活检查         |
| GET  | `/readyz`       | 就绪检查         |
| GET  | `/openapi.yaml` | OpenAPI 规范文件 |
| GET  | `/docs/*any`    | Swagger UI       |

## 附录 C：角色权限速查

| 角色     | 可访问模块 | 权限范围                     |
| -------- | ---------- | ---------------------------- |
| ADMIN    | 全部       | 读写删除（全部操作）         |
| OPERATOR | 大部分     | 读写（除用户管理、角色管理） |
| VIEWER   | 查询类     | 仅 GET 请求                  |
