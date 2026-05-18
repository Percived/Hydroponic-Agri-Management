### 第4章 系统环境搭建与后端功能验证

在完成系统的总体设计与核心模块的详细代码编写后，本章将详细记录系统运行环境的搭建过程，并以后端 API 响应数据与后台运行日志为主要依据，全面展示后端核心功能的实际运转效果。需要说明的是，为便于直观地观察和验证后端各接口的功能，本课题同时开发了一个基于 Vue 3 的前端管理界面作为后端功能的测试客户端——所有演示中涉及的前端页面均为调用后端 RESTful API 的展示层，不影响本文对后端设计与实现的论述重心。本章旨在验证前文所述后端架构设计的可行性以及 API 接口的正确性。

#### 4.1 系统运行前配置

为确保系统各组件能够稳定、高效地协同工作，本系统采用容器化的部署思路。后端服务与基础设施中间件均需进行严格的初始化配置。

1. **基础设施容器化部署（Docker Compose）**
   本系统的核心基础设施包括 MySQL（关系型数据库）、InfluxDB（时序数据库）以及 EMQX（MQTT 消息中间件）。为了避免本地环境差异带来的兼容性问题，系统采用 `docker-compose.yml` 进行统一编排。
    - **MySQL 配置**：映射宿主机端口 `13307`，配置持久化数据卷，并在容器启动时自动挂载执行 `packages/backend/migrations/merged/all.up.sql` 初始化脚本，完成所有核心业务表（`users`、`devices`、`control_policies`、`crop_batches` 等 20 余张表）与 14 项指标数据字典的创建。
    - **InfluxDB 配置**：映射端口 `9087`，初始化特定的 Bucket 用于接收 `telemetry` 遥测数据，并配置相应的访问 Token 供后端 Go 语言 SDK 调用。Bucket 的数据保留策略设置为 30 天，以平衡历史查询需求与存储空间。
    - **EMQX 配置**：映射 MQTT 默认端口 `1883` 以及 Web Dashboard 端口 `18083`，测试环境下开启匿名认证以便设备快速接入，同时配置客户端的自动订阅规则。

2. **后端服务运行配置**
   后端服务采用 Go 语言（版本 1.24）开发。在启动 `cmd/api/main.go` 之前，需在项目根目录下配置 `configs/config.yaml` 文件，指定数据库连接 DSN（`root:root@tcp(127.0.0.1:13307)/hydroponic?charset=utf8mb4&parseTime=true`）、InfluxDB 连接参数（URL、Token、Org、Bucket）、EMQX Broker 地址（`tcp://127.0.0.1:18830`）以及 JWT 签名密钥（`jwt_secret`，要求长度 >= 32 字符）。通过执行 `go mod tidy` 拉取依赖后，执行 `go run cmd/api/main.go` 启动运行于 `3000` 端口的后端 API 服务。启动日志中可观察到 MySQL 连接池初始化、InfluxDB 健康检查、MQTT 客户端连接成功及各模块路由注册的完整信息。

3. **测试前端与模拟器配置**
   为验证后端接口功能，同时启动前端测试界面（`packages/frontend/`，Vue 3 + Element Plus，运行于 `8082` 端口）与 Go 语言编写的 MQTT 设备模拟器（`cmd/simulator/`）。模拟器可配置多个虚拟传感器与执行器节点，按指定频率向 EMQX 推送符合系统 MQTT 主题规范的模拟遥测数据，用于验证后端的消息接收、解析、落库及策略触发全链路。

#### 4.2 后端功能验证

在环境搭建完毕后，本节以后端 API 的实际响应数据与后台运行日志为核心，逐项验证系统核心功能的正确性。

**4.2.1 认证与权限拦截验证**

首先验证后端 Auth 模块的 JWT 认证与 RBAC 权限拦截机制。
1. **正常登录与 Token 签发**：使用 Postman 向 `POST /api/auth/login` 发送管理员凭证（`{“username”:”admin”,”password”:”admin123”}`）。后端查询 MySQL `users` 表校验 Bcrypt 密码后，返回成功响应：
   ```json
   {“code”:0, “message”:”ok”, “data”:{“token”:”eyJhbGciOiJIUzI1NiIs...”,”user”:{“id”:1,”username”:”admin”,”nickname”:”管理员”,”roles”:[“ADMIN”]}}, “request_id”:”req_a1b2c3”}
   ```
   通过 jwt.io 解析 Token，确认 Payload 中包含 `user_id`、`username`、`roles` 字段及 `exp` 过期时间。

2. **未认证拦截**：不带 `Authorization` 请求头直接访问 `GET /api/greenhouses`。后端 `AuthMiddleware` 拦截该请求，返回 `{“code”:10002, “message”:”未授权，请先登录”}`。验证了中间件对所有受保护路由的全局拦截有效。

3. **越权拦截**：使用 VIEWER 角色的 Token 请求 `POST /api/policies`。后端中间件提取 Token 中的角色信息后，与路由注册时声明的权限要求（`AuthRequired(“ADMIN”, “OPERATOR”)`）进行交集比对，匹配失败后返回 `{“code”:10003, “message”:”权限不足，拒绝访问”}`。验证了 RBAC 细粒度鉴权有效。

（此处可插入图片：Postman 登录成功返回 Token 的 JSON 响应截图）
（此处可插入图片：Postman 越权访问被拒返回 10003 错误码截图）

**4.2.2 设备接入与遥测数据处理验证**

验证后端 MQTT Ingress 模块的物联网数据接入能力。
1. **设备注册**：通过 `POST /api/devices/register` 接口批量注册一台传感器设备（含 TEMP、HUMIDITY 两个通道）和一台执行器设备（含 FAN 通道）。后端在 MySQL `sensor_devices`、`sensor_channels`、`actuator_devices`、`actuator_channels` 四张表中分别创建对应记录，并返回完整的设备与通道信息（含自增 ID）。

2. **遥测数据接收与双库写入**：启动 Go 模拟器向 EMQX 的 `hydroponic/DEV001/telemetry/TEMP_CH1` 主题持续发布温度数据。观察后端控制台日志输出：
   ```
   [MQTT Ingress] telemetry received device=DEV001 channel=TEMP_CH1 metric=TEMP value=25.3
   ```
   随后通过 `GET /api/telemetry/channels/:id/latest` 接口查询 MySQL 中该通道的最新遥测记录，返回数据与模拟器推送值一致。同时通过 InfluxDB 的 Web UI 查询 `telemetry` Bucket 中的时序数据，确认数据点与上报时间戳完全对应，验证了双库同步写入机制的正确性。

3. **设备离线检测**：停止模拟器的心跳消息发送，等待约 600 秒后，通过 `GET /api/sensor-devices` 查询设备列表，该设备状态已自动更新为 `OFFLINE`。同时查询 `GET /api/alerts` 接口，确认系统自动生成了一条 `type=DEVICE_OFFLINE`、`level=CRITICAL` 的告警记录。

（此处可插入图片：后端控制台 MQTT Ingress 接收遥测数据的日志截图）
（此处可插入图片：InfluxDB Web UI 查询遥测数据点的截图）

**4.2.3 策略引擎与命令下发验证**

验证后端的自动化控制全链路——从策略配置到命令下发的完整闭环。
1. **策略创建**：通过 `POST /api/policies/full` 接口一次性提交一条完整的阈值策略。请求体包含策略基本信息（`policy_type=THRESHOLD`）、触发条件（`metric_code=TEMP, operator=GT, threshold_value=28, required_duration_sec=30`）和目标动作（`actuator_channel_id` 指向风机通道、`command_type=SWITCH, command_payload={“state”:”ON”}`）。后端在一个数据库事务中完成策略、条件与目标的原子性创建，返回策略 ID。

2. **策略触发验证**：将模拟温度值上调至 30°C 并稳定发送 30 秒以上。通过 `GET /api/policies/:id/executions` 查询策略执行记录，确认生成了一条 `decision=EXECUTED` 的执行记录。通过 `GET /api/commands` 查询命令列表，确认生成了一条 `command_type=SWITCH`、`status=ACKED` 的控制命令，`payload` 字段为 `{“state”:”ON”}`。同时，通过 `GET /api/alerts` 确认策略引擎同步创建了对应的告警记录。

3. **MQTT ACK 回执追踪**：模拟器在收到 MQTT 命令后，向 `hydroponic/DEV001/ack` 主题回复 ACK 消息。后端 Ingress 捕获后更新命令状态为 `ACKED` 并记录 `acked_at` 时间戳。通过 `GET /api/commands/:id/receipts` 查看回执记录，确认 `receipt_status` 与 `ack_code` 字段正确。

（此处可插入图片：策略执行与命令生成的数据库记录截图）
（此处可插入图片：命令状态从 PENDING 到 ACKED 流转的后端日志截图）

#### 4.3 本章小结

本章详细记录了水培农植信息管理系统的运行环境部署全过程，并以 Postman 请求响应、后端控制台日志及数据库查询结果为客观依据，对后端的认证鉴权、MQTT 设备通信、双库数据同步、自动策略引擎及命令闭环等核心功能进行了系统化验证。验证结果表明，后端 API 响应符合统一的 Envelope 规范，MQTT 消息上下行链路稳定可靠，策略引擎能够准确按照预设逻辑触发并完成从决策到执行的完整闭环。这充分验证了本文所设计的双库架构及业务模型的工程可行性。
