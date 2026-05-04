# MVP 需求清单（第一版范围）

本清单基于现有需求分析文档，给出第一版必须交付的范围，并用 MoSCoW（Must/Should/Could/Won't）标记优先级，便于推进实现与验收。

## 范围冻结声明
冻结版本：v1.0  
冻结日期：2026-02-11  
冻结原则：本版本只保证 Must 范围按期交付，Should/Could 仅在不影响 Must 进度与质量的前提下推进。

## 变更控制（防止范围漂移）
变更入口：任何新增需求必须提交变更申请（CR），明确需求描述与业务价值。  
影响评估：评估对工期、测试、数据结构、接口兼容性的影响。  
审批规则：新增 Must 需要评审确认并同步更新本清单与 API 规范。  
版本处理：无法在当前周期内落地的需求直接进入下一版本 Backlog。

## 优先级边界
Must：缺失即无法形成完整闭环或无法交付可用系统。  
Should：能显著提升可用性或管理效率，但可延后。  
Could：增强体验或扩展能力，不影响核心闭环。  
Won't：本期明确不做，避免争议。

## Must（必须实现）
- 设备注册/管理基础能力：新增、编辑、禁用设备，按类型/位置筛选
- 设备分组与温室/作物区绑定
- 设备在线状态监测与离线告警
- 数据接入：支持 MQTT 或 HTTP 上报（至少一种可用）
- 数据清洗与异常过滤（基础：缺失值/超范围标记或过滤）
- 数据存储与查询：实时/历史查询按设备、指标、时间范围
- 阈值规则：支持阈值触发与自动控制
- 控制指令下发与状态回显（含控制日志）
- 告警生成与处理闭环：告警列表/详情/状态更新
- 用户登录与鉴权（JWT），基础角色区分（管理员/操作员/只读）
- 审计日志记录关键操作（登录、控制、配置）

## Should（建议实现）
- 数据统计接口（均值/最大/最小）
- 数据保留/归档策略（基础配置）
- 告警订阅入口（SSE/WebSocket，至少预留接口）
- 设备指标字典与单位标准化可配置

## Could（可选实现）
- 控制策略模板（按作物/阶段复用）
- 多温室/多场地管理增强
- 断线重连与数据补传
- 权限细分到具体功能点

## Won't（本期不做）
- 预测模型与高级数据分析
- 多租户支持
- 多协议深度适配（Modbus 等）

## 建议验收维度（最小集）
- 模拟设备稳定上报，系统能持续接收并查询历史曲线
- 阈值触发后自动下发控制指令且有日志与回显
- 设备离线触发告警，告警可确认与关闭
- 不同角色权限隔离生效（只读用户无法控制）

## 关联接口（覆盖 Must 范围）
- 设备：`POST /api/devices`、`PUT /api/devices/{deviceId}`、`PATCH /api/devices/{deviceId}/status`、`GET /api/devices`
- 分组：`POST /api/device-groups`、`PUT /api/device-groups/{groupId}`、`GET /api/device-groups`
- 设备状态：`GET /api/devices/{deviceId}/health`
- 数据：`POST /api/telemetry`、`GET /api/telemetry/latest`、`GET /api/telemetry/history`
- 控制：`POST /api/controls/commands`、`GET /api/controls/commands/{commandId}`、`GET /api/controls/commands`
- 规则：`POST /api/controls/rules`、`PUT /api/controls/rules/{ruleId}`、`DELETE /api/controls/rules/{ruleId}`、`GET /api/controls/rules`
- 告警：`GET /api/alerts`、`GET /api/alerts/{alertId}`、`PATCH /api/alerts/{alertId}/status`
- 鉴权：`POST /api/auth/login`、`GET /api/users`、`POST /api/users`
- 审计：`GET /api/audit-logs`
