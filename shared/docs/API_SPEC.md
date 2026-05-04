# API 规格说明（定稿 v1.1）

适用范围：Phase 2（更新日期 2026-05-04）

## 1. 通用约定

Base URL：`/api`

内容类型：`application/json; charset=utf-8`

时间格式：ISO 8601 UTC，例如 `2026-02-11T08:30:00.000Z`

ID 类型：`BIGINT`，响应中以数字返回

### 1.1 鉴权

鉴权方式：`Authorization: Bearer <JWT>`

角色：`ADMIN`、`OPERATOR`、`VIEWER`

权限规则：
ADMIN 可访问所有接口；OPERATOR 可访问查询与控制接口；VIEWER 仅查询类接口。

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

### 2.1 Device

| 字段          | 类型   | 规则                    | 示例                       |
| ------------- | ------ | ----------------------- | -------------------------- |
| id            | number | 正整数                  | 101                        |
| device_code   | string | 唯一，1-64 字符         | "DEV-001"                  |
| name          | string | 1-64 字符               | "温室温度传感器"           |
| type          | string | `SENSOR` 或 `ACTUATOR`  | "SENSOR"                   |
| category      | string | 1-32 字符               | "TEMP"                     |
| greenhouse_id | number | 可空                    | 1                          |
| group_id      | number | 可空                    | 10                         |
| status        | string | `ENABLED` 或 `DISABLED` | "ENABLED"                  |
| protocol      | string | `MQTT` 或 `HTTP`        | "MQTT"                     |
| last_seen_at  | string | 可空                    | "2026-02-11T08:30:00.000Z" |
| created_at    | string | 必填                    | "2026-02-11T08:00:00.000Z" |
| updated_at    | string | 必填                    | "2026-02-11T08:10:00.000Z" |

### 2.2 Greenhouse

| 字段        | 类型   | 规则                | 示例                       |
| ----------- | ------ | ------------------- | -------------------------- |
| id          | number | 正整数              | 1                          |
| name        | string | 1-64 字符           | "一号温室"                 |
| location    | string | 可空，最大 128 字符 | "东侧园区"                 |
| description | string | 可空，最大 255 字符 | "生菜种植区"               |
| created_at  | string | 必填                | "2026-02-11T08:00:00.000Z" |
| updated_at  | string | 必填                | "2026-02-11T08:10:00.000Z" |

### 2.3 DeviceGroup

| 字段          | 类型   | 规则                                   | 示例                       |
| ------------- | ------ | -------------------------------------- | -------------------------- |
| id            | number | 正整数                                 | 10                         |
| greenhouse_id | number | 必填                                   | 1                          |
| name          | string | 1-64 字符                              | "A 区"                     |
| description   | string | 可空                                   | "生菜区"                   |
| device_count  | number | 只在分组列表接口返回，表示组内设备数量 | 12                         |
| created_at    | string | 必填                                   | "2026-02-11T08:00:00.000Z" |
| updated_at    | string | 必填                                   | "2026-02-11T08:10:00.000Z" |

### 2.4 TelemetryPoint

| 字段         | 类型   | 规则           | 示例                       |
| ------------ | ------ | -------------- | -------------------------- |
| device_id    | number | 正整数         | 101                        |
| metric_code  | string | 1-32 字符      | "TEMP"                     |
| value        | number | 清洗后数值     | 24.6                       |
| raw_value    | number | 原始数值       | 24.6                       |
| quality      | number | 0 正常，1 异常 | 0                          |
| collected_at | string | 必填           | "2026-02-11T08:30:00.000Z" |

### 2.5 ControlCommand

| 字段         | 类型   | 规则                                 | 示例                       |
| ------------ | ------ | ------------------------------------ | -------------------------- |
| id           | number | 正整数                               | 501                        |
| device_id    | number | 正整数                               | 201                        |
| command_type | string | 1-32 字符                            | "SWITCH"                   |
| payload      | object | JSON                                 | {"state":"ON"}             |
| status       | string | `PENDING`/`SENT`/`EXECUTED`/`FAILED` | "SENT"                     |
| created_by   | number | 用户 ID                              | 1                          |
| created_at   | string | 必填                                 | "2026-02-11T08:35:00.000Z" |
| sent_at      | string | 可空                                 | "2026-02-11T08:35:01.000Z" |
| executed_at  | string | 可空                                 | "2026-02-11T08:35:05.000Z" |

### 2.6 ControlRule

| 字段               | 类型    | 规则                   | 示例                                               |
| ------------------ | ------- | ---------------------- | -------------------------------------------------- |
| id                 | number  | 正整数                 | 701                                                |
| name               | string  | 1-64 字符              | "高温开风机"                                       |
| metric_code        | string  | 1-32 字符              | "TEMP"                                             |
| operator           | string  | `>` `>=` `<` `<=` `==` | ">"                                                |
| threshold          | number  | 必填                   | 30                                                 |
| action             | object  | JSON                   | {"command_type":"SWITCH","payload":{"state":"ON"}} |
| target_device_id   | number  | 必填                   | 201                                                |
| target_device_name | string  | 可空（列表接口返回）   | "通风风机"                                         |
| enabled            | boolean | 必填                   | true                                               |
| created_by         | number  | 用户 ID                | 1                                                  |
| created_at         | string  | 必填                   | "2026-02-11T08:40:00.000Z"                         |
| updated_at         | string  | 必填                   | "2026-02-11T08:41:00.000Z"                         |

### 2.7 Alert

| 字段         | 类型   | 规则                                 | 示例                       |
| ------------ | ------ | ------------------------------------ | -------------------------- |
| id           | number | 正整数                               | 901                        |
| type         | string | `THRESHOLD`/`OFFLINE`/`DEVICE_ERROR` | "THRESHOLD"                |
| level        | string | `INFO`/`WARN`/`CRITICAL`             | "WARN"                     |
| metric_id    | number | 可空                                 | 1                          |
| device_id    | number | 正整数                               | 101                        |
| device_name  | string | 可空（列表/详情接口返回）            | "环境温度传感器"           |
| value        | number | 可空                                 | 32.5                       |
| message      | string | 1-255 字符                           | "温度过高"                 |
| status       | string | `OPEN`/`ACK`/`CLOSED`                | "OPEN"                     |
| triggered_at | string | 必填                                 | "2026-02-11T08:45:00.000Z" |
| resolved_at  | string | 可空                                 | null                       |
| created_at   | string | 必填                                 | "2026-02-11T08:45:00.000Z" |

### 2.8 User

| 字段       | 类型   | 规则                 | 示例                       |
| ---------- | ------ | -------------------- | -------------------------- |
| id         | number | 正整数               | 1                          |
| username   | string | 3-32 字符            | "admin"                    |
| nickname   | string | 可空                 | "管理员"                   |
| phone      | string | 可空                 | "13800000000"              |
| email      | string | 可空                 | "admin@example.com"        |
| status     | string | `ENABLED`/`DISABLED` | "ENABLED"                  |
| roles      | array  | 角色列表             | ["ADMIN"]                  |
| created_at | string | 必填                 | "2026-02-11T07:00:00.000Z" |

### 2.9 Role

| 字段        | 类型   | 规则      | 示例         |
| ----------- | ------ | --------- | ------------ |
| id          | number | 正整数    | 1            |
| name        | string | 1-32 字符 | "ADMIN"      |
| description | string | 可空      | "系统管理员" |

### 2.10 AuditLog

| 字段        | 类型   | 规则      | 示例                       |
| ----------- | ------ | --------- | -------------------------- |
| id          | number | 正整数    | 3001                       |
| user_id     | number | 正整数    | 1                          |
| action      | string | 1-64 字符 | "CONTROL_COMMAND"          |
| target_type | string | 1-64 字符 | "DEVICE"                   |
| target_id   | number | 可空      | 201                        |
| detail      | object | JSON      | {"state":"ON"}             |
| created_at  | string | 必填      | "2026-02-11T08:35:00.000Z" |

### 2.11 NotificationChannel

| 字段            | 类型    | 规则                        | 示例                                     |
| --------------- | ------- | --------------------------- | ---------------------------------------- |
| id              | number  | 正整数                      | 1                                        |
| user_id         | number  | 正整数                      | 1                                        |
| channel_type    | string  | `EMAIL`/`SMS`/`WEBHOOK`     | "WEBHOOK"                                |
| name            | string  | 1-64 字符                   | "钉钉通知"                               |
| config          | object  | JSON                        | {"url":"https://hooks.example.com","secret":"key"} |
| min_alert_level | string  | `INFO`/`WARN`/`CRITICAL`    | "WARN"                                   |
| enabled         | boolean | true/false                  | true                                     |
| created_at      | string  | 必填                        | "2026-02-11T08:00:00.000Z"               |
| updated_at      | string  | 必填                        | "2026-02-11T08:10:00.000Z"               |

### 2.12 SystemConfig

| 字段         | 类型   | 规则        | 示例                       |
| ------------ | ------ | ----------- | -------------------------- |
| id           | number | 正整数      | 1                          |
| config_key   | string | 1-64 字符   | "retention_days"           |
| config_value | string | 0-255 字符  | "365"                      |
| description  | string | 0-255 字符  | "遥测数据保留天数"         |
| updated_at   | string | 必填        | "2026-02-11T08:00:00.000Z" |

### 2.13 TelemetrySummary

| 字段        | 类型   | 规则 | 说明 |
| ----------- | ------ | ---- | ---- |
| metrics     | object | 必填 | `{ "TEMP": AverageMetric, ... }` 以 metric_code 为 key |
| online_rate | number | 0-1  | 有数据的小时数 / 总小时数 |
| alert_events | array | 必填 | 时间范围内的告警事件简要列表 |

AverageMetric 子结构：

| 字段   | 类型   | 规则 | 说明 |
| ------ | ------ | ---- | ---- |
| code   | string | 必填 | 指标编码 |
| name   | string | 必填 | 指标名称 |
| unit   | string | 必填 | 单位 |
| avg    | number | 可空 | 平均值 |
| max    | number | 可空 | 最大值 |
| min    | number | 可空 | 最小值 |
| alerts | number | 必填 | 告警数量 |
| hourly | array  | 必填 | `[{ "hour": "2026-02-11T08:00:00Z", "avg": 24.5 }, ...]` |

## 3. 接口定义

### 3.1 鉴权

**POST /api/auth/login**

鉴权：无需

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| username | string | 是 | 3-32 字符 | "admin" |
| password | string | 是 | 6-64 字符 | "**\*\***" |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "jwt.token.here",
    "expires_in": 7200,
    "user": {
      "id": 1,
      "username": "admin",
      "roles": ["ADMIN"]
    }
  },
  "request_id": "req_1"
}
```

**POST /api/auth/logout**

鉴权：ADMIN/OPERATOR/VIEWER

说明：可选，实现为服务端拉黑或客户端丢弃 Token。

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_2" }
```

### 3.2 用户与角色

**GET /api/users**

鉴权：ADMIN

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| page | number | 否 | >=1 | 1 |
| page_size | number | 否 | 1-200 | 20 |
| status | string | 否 | ENABLED/DISABLED | "ENABLED" |
| keyword | string | 否 | 1-64 字符 | "admin" |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": [
      { "id": 1, "username": "admin", "roles": ["ADMIN"], "status": "ENABLED" }
    ]
  },
  "request_id": "req_3"
}
```

**POST /api/users**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| username | string | 是 | 3-32 字符，唯一 | "operator1" |
| password | string | 是 | 6-64 字符 | "**\*\***" |
| nickname | string | 否 | 0-64 字符 | "张三" |
| roles | array | 是 | 角色名列表 | ["OPERATOR"] |

响应示例：

```json
{ "code": 0, "message": "ok", "data": { "id": 2 }, "request_id": "req_4" }
```

**PUT /api/users/{userId}**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| nickname | string | 否 | 0-64 字符 | "李四" |
| phone | string | 否 | 0-32 字符 | "13800000000" |
| email | string | 否 | 0-64 字符 | "user@example.com" |
| roles | array | 否 | 角色名列表 | ["VIEWER"] |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_5" }
```

**PATCH /api/users/{userId}/status**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| status | string | 是 | ENABLED/DISABLED | "DISABLED" |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_6" }
```

**GET /api/roles**

鉴权：ADMIN

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "items": [{ "id": 1, "name": "ADMIN" }] },
  "request_id": "req_7"
}
```

**POST /api/roles**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| name | string | 是 | 1-32 字符，唯一 | "OPERATOR" |
| description | string | 否 | 0-64 字符 | "操作员" |

响应示例：

```json
{ "code": 0, "message": "ok", "data": { "id": 2 }, "request_id": "req_8" }
```

**PUT /api/roles/{roleId}**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| description | string | 否 | 0-64 字符 | "只读用户" |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_9" }
```

### 3.3 温室管理

**POST /api/devices/greenhouses**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| name | string | 是 | 1-64 字符 | "一号温室" |
| location | string | 否 | 0-128 字符 | "东侧园区" |
| description | string | 否 | 0-255 字符 | "生菜种植区" |

响应示例：

```json
{ "code": 0, "message": "ok", "data": { "id": 1 }, "request_id": "req_g1" }
```

**PUT /api/devices/greenhouses/{greenhouseId}**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| name | string | 否 | 1-64 字符 | "二号温室" |
| location | string | 否 | 0-128 字符 | "西侧园区" |
| description | string | 否 | 0-255 字符 | "番茄种植区" |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_g2" }
```

**GET /api/devices/greenhouses**

鉴权：ADMIN/OPERATOR/VIEWER

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| page | number | 否 | >=1 | 1 |
| page_size | number | 否 | 1-200 | 20 |
| keyword | string | 否 | 1-64 字符 | "一号" |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": [
      {
        "id": 1,
        "name": "一号温室",
        "location": "东侧园区",
        "description": "生菜种植区",
        "created_at": "2026-02-11T08:00:00.000Z",
        "updated_at": "2026-02-11T08:10:00.000Z"
      }
    ]
  },
  "request_id": "req_g3"
}
```

**GET /api/devices/greenhouses/{greenhouseId}**

鉴权：ADMIN/OPERATOR/VIEWER

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "name": "一号温室",
    "location": "东侧园区",
    "description": "生菜种植区",
    "created_at": "2026-02-11T08:00:00.000Z",
    "updated_at": "2026-02-11T08:10:00.000Z"
  },
  "request_id": "req_g4"
}
```

**DELETE /api/devices/greenhouses/{greenhouseId}**

鉴权：ADMIN

说明：删除温室时会在事务内先解绑关联数据：

- 将该温室下设备的 `greenhouse_id` 置空
- 将该温室下分组内设备的 `group_id` 置空
- 删除该温室下设备组
- 最后删除温室

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_g5" }
```

### 3.4 设备与分组

**POST /api/devices**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| device_code | string | 是 | 1-64 字符，唯一 | "DEV-001" |
| name | string | 是 | 1-64 字符 | "温度传感器" |
| type | string | 是 | SENSOR/ACTUATOR | "SENSOR" |
| category | string | 是 | 1-32 字符 | "TEMP" |
| protocol | string | 是 | MQTT/HTTP | "MQTT" |
| greenhouse_id | number | 否 | 正整数 | 1 |
| group_id | number | 否 | 正整数 | 10 |
| sampling_interval_sec | number | 否 | 5-3600 | 60 |

响应示例：

```json
{ "code": 0, "message": "ok", "data": { "id": 101 }, "request_id": "req_10" }
```

**PUT /api/devices/{deviceId}**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| name | string | 否 | 1-64 字符 | "温度传感器" |
| category | string | 否 | 1-32 字符 | "TEMP" |
| greenhouse_id | number | 否 | 正整数 | 1 |
| group_id | number | 否 | 正整数 | 10 |
| sampling_interval_sec | number | 否 | 5-3600 | 60 |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_11" }
```

**PATCH /api/devices/{deviceId}/status**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| status | string | 是 | ENABLED/DISABLED | "DISABLED" |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_12" }
```

**GET /api/devices**

鉴权：ADMIN/OPERATOR/VIEWER

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| page | number | 否 | >=1 | 1 |
| page_size | number | 否 | 1-200 | 20 |
| type | string | 否 | SENSOR/ACTUATOR | "SENSOR" |
| category | string | 否 | 1-32 字符 | "TEMP" |
| group_id | number | 否 | 正整数 | 10 |
| greenhouse_id | number | 否 | 正整数 | 1 |
| status | string | 否 | ENABLED/DISABLED | "ENABLED" |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": [{ "id": 101, "device_code": "DEV-001", "status": "ENABLED" }]
  },
  "request_id": "req_13"
}
```

**GET /api/devices/{deviceId}**

鉴权：ADMIN/OPERATOR/VIEWER

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "id": 101, "device_code": "DEV-001", "type": "SENSOR" },
  "request_id": "req_14"
}
```

**GET /api/devices/{deviceId}/health**

鉴权：ADMIN/OPERATOR/VIEWER

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "device_id": 101,
    "online": true,
    "last_seen_at": "2026-02-11T08:30:00.000Z"
  },
  "request_id": "req_15"
}
```

**GET /api/devices/{deviceId}/telemetry-summary**

鉴权：ADMIN/OPERATOR/VIEWER

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| from | string | 否 | ISO 8601，默认当前-24h | "2026-02-11T00:00:00.000Z" |
| to   | string | 否 | ISO 8601，默认当前时间 | "2026-02-11T23:59:59.000Z" |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "metrics": {
      "TEMP": {
        "code": "TEMP",
        "name": "温度",
        "unit": "°C",
        "avg": 24.5,
        "max": 26.1,
        "min": 23.2,
        "alerts": 2,
        "hourly": [
          { "hour": "2026-02-11T08:00:00Z", "avg": 24.3 },
          { "hour": "2026-02-11T09:00:00Z", "avg": 24.7 }
        ]
      }
    },
    "online_rate": 0.92,
    "alert_events": [
      {
        "id": 901,
        "type": "THRESHOLD",
        "level": "WARN",
        "message": "温度过高",
        "status": "OPEN",
        "triggered_at": "2026-02-11T08:45:00.000Z"
      }
    ]
  },
  "request_id": "req_15a"
}
```

**POST /api/devices/batch-update**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| device_ids | array | 是 | 1-100 个 | [1,2,3] |
| updates | object | 是 | 允许的字段: `group_id`, `sampling_interval_sec`, `status` | {"status":"DISABLED"} |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "affected": 3 },
  "request_id": "req_15b"
}
```

**DELETE /api/devices/batch**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| device_ids | array | 是 | 1-100 个 | [1,2,3] |
| reason | string | 否 | 0-255 字符 | "报废更换" |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "deleted": 3 },
  "request_id": "req_15c"
}
```

**POST /api/device-groups**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| greenhouse_id | number | 是 | 正整数 | 1 |
| name | string | 是 | 1-64 字符 | "A 区" |
| description | string | 否 | 0-255 字符 | "生菜区" |

响应示例：

```json
{ "code": 0, "message": "ok", "data": { "id": 10 }, "request_id": "req_16" }
```

**PUT /api/device-groups/{groupId}**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| name | string | 否 | 1-64 字符 | "B 区" |
| description | string | 否 | 0-255 字符 | "番茄区" |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_17" }
```

**DELETE /api/device-groups/{groupId}**

鉴权：ADMIN

说明：删除设备组时会在事务内先将该组下设备的 `group_id` 置空，再删除设备组。

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_17_1" }
```

**GET /api/device-groups**

鉴权：ADMIN/OPERATOR/VIEWER

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| greenhouse_id | number | 否 | 正整数 | 1 |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": 10,
        "name": "A 区",
        "description": "生菜区",
        "greenhouse_id": 1,
        "device_count": 12
      }
    ]
  },
  "request_id": "req_18"
}
```

**POST /api/device-groups/{groupId}/devices/{deviceId}**

鉴权：ADMIN

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_19" }
```

**DELETE /api/device-groups/{groupId}/devices/{deviceId}**

鉴权：ADMIN

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_20" }
```

### 3.5 数据采集与查询

**POST /api/telemetry**

鉴权：设备凭证或 ADMIN/OPERATOR

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| device_code | string | 是 | 1-64 字符 | "DEV-001" |
| collected_at | string | 否 | ISO 8601 | "2026-02-11T08:30:00.000Z" |
| metrics | array | 是 | 1-50 条 | [{"code":"TEMP","value":24.6,"unit":"C"}] |

metrics 子项字段：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| code | string | 是 | 1-32 字符 | "TEMP" |
| value | number | 是 | 合法数值 | 24.6 |
| unit | string | 否 | 0-16 字符 | "C" |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "accepted": 1 },
  "request_id": "req_21"
}
```

**GET /api/telemetry/latest**

鉴权：ADMIN/OPERATOR/VIEWER

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| device_id | number | 是 | 正整数 | 101 |
| metrics | string | 否 | 逗号分隔 | "TEMP,HUMIDITY" |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "device_id": 101,
        "metric_code": "TEMP",
        "value": 24.6,
        "quality": 0,
        "collected_at": "2026-02-11T08:30:00.000Z"
      }
    ]
  },
  "request_id": "req_22"
}
```

**GET /api/telemetry/history**

鉴权：ADMIN/OPERATOR/VIEWER

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| device_id | number | 是 | 正整数 | 101 |
| metric_code | string | 是 | 1-32 字符 | "TEMP" |
| start_time | string | 是 | ISO 8601 | "2026-02-11T08:00:00.000Z" |
| end_time | string | 是 | ISO 8601 | "2026-02-11T09:00:00.000Z" |
| include_raw | boolean | 否 | true/false | false |
| page | number | 否 | >=1 | 1 |
| page_size | number | 否 | 1-2000 | 100 |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "page": 1,
    "page_size": 100,
    "total": 1,
    "items": [
      {
        "device_id": 101,
        "metric_code": "TEMP",
        "value": 24.6,
        "raw_value": 24.6,
        "quality": 0,
        "collected_at": "2026-02-11T08:30:00.000Z"
      }
    ]
  },
  "request_id": "req_23"
}
```

**GET /api/telemetry/stats**

鉴权：ADMIN/OPERATOR/VIEWER

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| device_id | number | 是 | 正整数 | 101 |
| metric_code | string | 是 | 1-32 字符 | "TEMP" |
| start_time | string | 是 | ISO 8601 | "2026-02-11T08:00:00.000Z" |
| end_time | string | 是 | ISO 8601 | "2026-02-11T09:00:00.000Z" |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "avg": 24.2, "max": 25.1, "min": 23.8 },
  "request_id": "req_24"
}
```

**POST /api/telemetry/retention**

鉴权：ADMIN

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| keep_days | number | 是 | 7-3650 | 365 |
| archive | boolean | 否 | true/false | true |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_25" }
```

### 3.6 控制与规则

**POST /api/controls/commands**

鉴权：ADMIN/OPERATOR

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| device_id | number | 是 | 正整数 | 201 |
| command_type | string | 是 | 1-32 字符 | "SWITCH" |
| payload | object | 是 | JSON | {"state":"ON"} |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "id": 501, "status": "PENDING" },
  "request_id": "req_26"
}
```

**POST /api/controls/batch-commands**

鉴权：ADMIN/OPERATOR

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| target_type | string | 是 | `greenhouse`/`device_group`/`devices` | "devices" |
| target_ids | array | 是 | 1-100 个 | [201, 202] |
| command_type | string | 是 | 1-32 字符 | "SWITCH" |
| payload | object | 是 | JSON | {"state":"ON"} |
| remark | string | 否 | 0-255 字符 | "批量开关" |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "results": [
      { "device_id": 201, "command_id": 501, "status": "SENT" },
      { "device_id": 202, "command_id": 502, "status": "FAILED" }
    ]
  },
  "request_id": "req_26a"
}
```

**GET /api/controls/commands/{commandId}**

鉴权：ADMIN/OPERATOR/VIEWER

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "id": 501, "status": "EXECUTED" },
  "request_id": "req_27"
}
```

**GET /api/controls/commands**

鉴权：ADMIN/OPERATOR/VIEWER

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| device_id | number | 否 | 正整数 | 201 |
| status | string | 否 | PENDING/SENT/EXECUTED/FAILED | "EXECUTED" |
| page | number | 否 | >=1 | 1 |
| page_size | number | 否 | 1-200 | 20 |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": [{ "id": 501, "status": "EXECUTED" }]
  },
  "request_id": "req_28"
}
```

**POST /api/controls/rules**

鉴权：ADMIN/OPERATOR

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| name | string | 是 | 1-64 字符 | "高温开风机" |
| metric_code | string | 是 | 1-32 字符 | "TEMP" |
| operator | string | 是 | > >= < <= == | ">" |
| threshold | number | 是 | 合理范围 | 30 |
| action | object | 是 | JSON | {"command_type":"SWITCH","payload":{"state":"ON"}} |
| target_device_id | number | 是 | 正整数 | 201 |
| enabled | boolean | 是 | true/false | true |

响应示例：

```json
{ "code": 0, "message": "ok", "data": { "id": 701 }, "request_id": "req_29" }
```

**PUT /api/controls/rules/{ruleId}**

鉴权：ADMIN/OPERATOR

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| name | string | 否 | 1-64 字符 | "高温开风机" |
| operator | string | 否 | > >= < <= == | ">=" |
| threshold | number | 否 | 合理范围 | 28 |
| action | object | 否 | JSON | {"command_type":"SWITCH","payload":{"state":"ON"}} |
| target_device_id | number | 否 | 正整数 | 201 |
| enabled | boolean | 否 | true/false | true |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_30" }
```

**DELETE /api/controls/rules/{ruleId}**

鉴权：ADMIN/OPERATOR

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_31" }
```

**GET /api/controls/rules**

鉴权：ADMIN/OPERATOR/VIEWER

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| metric_code | string | 否 | 1-32 字符 | "TEMP" |
| enabled | boolean | 否 | true/false | true |
| page | number | 否 | >=1 | 1 |
| page_size | number | 否 | 1-200 | 20 |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": [{
      "id": 701,
      "name": "高温开风机",
      "metric_code": "TEMP",
      "operator": ">",
      "threshold": 30.0,
      "action": {"command_type": "SWITCH", "payload": {"state": "ON"}},
      "target_device_id": 201,
      "target_device_name": "通风风机",
      "enabled": true,
      "created_at": "2026-02-11T08:40:00.000Z",
      "updated_at": "2026-02-11T08:41:00.000Z"
    }]
  },
  "request_id": "req_32"
}
```

**POST /api/controls/templates**

鉴权：ADMIN

说明：可选

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| name | string | 是 | 1-64 字符 | "生菜模板" |
| description | string | 否 | 0-255 字符 | "春季" |
| content | object | 是 | JSON | {"rules":[...] } |

响应示例：

```json
{ "code": 0, "message": "ok", "data": { "id": 801 }, "request_id": "req_33" }
```

**POST /api/controls/templates/{templateId}/apply**

鉴权：ADMIN

说明：可选

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| target_group_id | number | 是 | 正整数 | 10 |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_34" }
```

**GET /api/controls/templates**

鉴权：ADMIN/OPERATOR/VIEWER

说明：可选

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "items": [{ "id": 801, "name": "生菜模板" }] },
  "request_id": "req_35"
}
```

### 3.7 告警

**GET /api/alerts**

鉴权：ADMIN/OPERATOR/VIEWER

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| type | string | 否 | THRESHOLD/OFFLINE/DEVICE_ERROR | "THRESHOLD" |
| level | string | 否 | INFO/WARN/CRITICAL | "WARN" |
| status | string | 否 | OPEN/ACK/CLOSED | "OPEN" |
| page | number | 否 | >=1 | 1 |
| page_size | number | 否 | 1-200 | 20 |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": [{
      "id": 901,
      "type": "THRESHOLD",
      "level": "WARN",
      "metric_id": 1,
      "device_id": 101,
      "device_name": "环境温度传感器",
      "value": 32.5,
      "message": "温度过高",
      "status": "OPEN",
      "triggered_at": "2026-02-11T08:45:00.000Z",
      "resolved_at": null,
      "created_at": "2026-02-11T08:45:00.000Z"
    }]
  },
  "request_id": "req_36"
}
```

**GET /api/alerts/{alertId}**

鉴权：ADMIN/OPERATOR/VIEWER

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 901,
    "type": "THRESHOLD",
    "level": "WARN",
    "metric_id": 1,
    "device_id": 101,
    "device_name": "环境温度传感器",
    "value": 32.5,
    "message": "温度过高",
    "status": "OPEN",
    "triggered_at": "2026-02-11T08:45:00.000Z",
    "resolved_at": null,
    "created_at": "2026-02-11T08:45:00.000Z"
  },
  "request_id": "req_37"
}
```

**PATCH /api/alerts/{alertId}/status**

鉴权：ADMIN/OPERATOR

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| status | string | 是 | ACK/CLOSED | "ACK" |
| comment | string | 否 | 0-255 字符 | "已处理" |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_38" }
```

**GET /api/alerts/stats**

鉴权：ADMIN/OPERATOR/VIEWER

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "open_count": 5, "ack_count": 2, "closed_count": 10 },
  "request_id": "req_39"
}
```

**GET /api/alerts/subscribe**

鉴权：ADMIN/OPERATOR/VIEWER

说明：可选，SSE 或 WebSocket 入口。

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "url": "/api/alerts/subscribe" },
  "request_id": "req_40"
}
```

### 3.8 通知渠道

**GET /api/notification-channels**

鉴权：ADMIN/OPERATOR/VIEWER

说明：返回当前用户的全部通知渠道。

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "user_id": 1,
        "channel_type": "WEBHOOK",
        "name": "钉钉通知",
        "config": {"url": "https://hooks.example.com", "secret": "***"},
        "min_alert_level": "WARN",
        "enabled": true,
        "created_at": "2026-02-11T08:00:00.000Z",
        "updated_at": "2026-02-11T08:10:00.000Z"
      }
    ]
  },
  "request_id": "req_n1"
}
```

**POST /api/notification-channels**

鉴权：ADMIN/OPERATOR

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| channel_type | string | 是 | EMAIL/SMS/WEBHOOK | "WEBHOOK" |
| name | string | 是 | 1-64 字符 | "钉钉通知" |
| config | object | 是 | JSON | {"url":"https://hooks.example.com","secret":"key"} |
| min_alert_level | string | 否 | INFO/WARN/CRITICAL，默认 WARN | "WARN" |
| enabled | boolean | 是 | true/false | true |

响应示例：

```json
{ "code": 0, "message": "ok", "data": { "id": 1 }, "request_id": "req_n2" }
```

**PUT /api/notification-channels/{channelId}**

鉴权：ADMIN/OPERATOR

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| name | string | 否 | 1-64 字符 | "钉钉通知" |
| config | object | 否 | JSON | {"url":"https://hooks.example.com","secret":"key"} |
| min_alert_level | string | 否 | INFO/WARN/CRITICAL | "CRITICAL" |
| enabled | boolean | 否 | true/false | false |

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_n3" }
```

**DELETE /api/notification-channels/{channelId}**

鉴权：ADMIN/OPERATOR

响应示例：

```json
{ "code": 0, "message": "ok", "data": {}, "request_id": "req_n4" }
```

**POST /api/notification-channels/{channelId}/test**

鉴权：ADMIN/OPERATOR

说明：对 WEBHOOK 类型渠道发送测试消息。请求头包含 `X-Signature: HMAC-SHA256(body, secret)` 签名。

响应示例：

```json
{ "code": 0, "message": "ok", "data": { "sent": true }, "request_id": "req_n5" }
```

### 3.9 系统概览与配置

**GET /api/overview/dashboard**

鉴权：ADMIN/OPERATOR/VIEWER

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "devices_online": 12, "devices_offline": 1, "alerts_open": 2 },
  "request_id": "req_41"
}
```

**GET /api/telemetry/system-configs**

鉴权：ADMIN

说明：返回全部系统配置项，敏感值（jwt_secret, db_password, mqtt_password）显示为 `***`。

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": 1,
        "config_key": "retention_days",
        "config_value": "365",
        "description": "遥测数据保留天数",
        "updated_at": "2026-02-11T08:00:00.000Z"
      },
      {
        "id": 2,
        "config_key": "jwt_secret",
        "config_value": "***",
        "description": "JWT 签名密钥",
        "updated_at": "2026-02-11T08:00:00.000Z"
      }
    ]
  },
  "request_id": "req_42"
}
```

**PUT /api/telemetry/system-configs**

鉴权：ADMIN

说明：按 config_key UPSERT 一条配置。

请求体：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| config_key | string | 是 | 1-64 字符 | "retention_days" |
| config_value | string | 是 | 0-255 字符 | "365" |
| description | string | 否 | 0-255 字符 | "遥测数据保留天数" |

响应示例：

```json
{ "code": 0, "message": "ok", "data": { "id": 1 }, "request_id": "req_43" }
```

### 3.10 审计日志

**GET /api/audit-logs**

鉴权：ADMIN

查询参数：
| 字段 | 类型 | 必填 | 规则 | 示例 |
| --- | --- | --- | --- | --- |
| user_id | number | 否 | 正整数 | 1 |
| action | string | 否 | 1-64 字符 | "CONTROL_COMMAND" |
| start_time | string | 否 | ISO 8601 | "2026-02-11T08:00:00.000Z" |
| end_time | string | 否 | ISO 8601 | "2026-02-11T09:00:00.000Z" |
| page | number | 否 | >=1 | 1 |
| page_size | number | 否 | 1-200 | 20 |

响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": [{ "id": 3001, "action": "CONTROL_COMMAND" }]
  },
  "request_id": "req_43"
}
```
