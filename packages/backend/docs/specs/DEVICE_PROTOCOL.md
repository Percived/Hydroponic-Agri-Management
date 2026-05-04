# 设备数据协议（MQTT/HTTP）

本协议用于规范设备数据上报，覆盖字段、单位、异常值处理与上报频率。适用于 MVP v1.0。

## 1. MQTT 上报
Topic：`telemetry/{device_code}`

QoS：1

Payload（JSON）：
```json
{
  "device_code": "DEV-001",
  "collected_at": "2026-02-11T08:30:00.000Z",
  "metrics": [
    {"code": "TEMP", "value": 24.6, "unit": "C"},
    {"code": "HUMIDITY", "value": 55.2, "unit": "%"}
  ]
}
```

## 2. HTTP 上报
接口：`POST /api/telemetry`

Payload 与 MQTT 一致。

## 3. 字段说明
| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| device_code | string | 是 | 设备唯一编码 |
| collected_at | string | 否 | ISO 8601 时间，缺省由服务端填充 |
| metrics | array | 是 | 指标列表 |

metrics 子项：
| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| code | string | 是 | 指标编码 |
| value | number | 是 | 指标数值 |
| unit | string | 否 | 单位（可选，服务端以指标字典为准） |

## 4. 单位与范围（默认指标字典）
| 指标编码 | 说明 | 单位 | 物理范围 |
| --- | --- | --- | --- |
| TEMP | 温度 | C | -10 到 60 |
| HUMIDITY | 湿度 | % | 0 到 100 |
| LIGHT | 光照 | lux | 0 到 200000 |
| PH | 酸碱度 | pH | 0 到 14 |
| EC | 电导率 | mS/cm | 0 到 10 |
| CO2 | 二氧化碳 | ppm | 0 到 5000 |

## 5. 异常值处理
异常定义：缺失值、非数字、或超出指标物理范围。

处理规则：
1. 记录原始数值 `raw_value`。
2. 标记 `quality=1`（正常为 0）。
3. 统计类查询默认过滤 `quality=1` 数据，避免影响统计结果。

## 6. 上报频率
默认采样周期：60 秒。

可配置范围：5 秒到 3600 秒，对应设备字段 `sampling_interval_sec`。

心跳与在线判定：当设备连续 `3 * sampling_interval_sec` 未上报，标记离线并触发离线告警。
