# P0 功能设计文档

## 概述

本文档定义 P0 优先级功能的详细设计，包括首页仪表盘、设备控制命令、控制规则和告警中心。

## 1. 首页仪表盘

### 1.1 页面路由
- 路径: `/` (重定向到 `/dashboard`) 或 `/dashboard`
- 权限: ADMIN/OPERATOR/VIEWER

### 1.2 页面布局

```
┌─────────────────────────────────────────────────────────────┐
│ 系统概览                                   当前日期          │
├─────────────────────────────────────────────────────────────┤
│ 关键指标卡片 (4列网格):                                       │
│ ┌────────────┬────────────┬────────────┬────────────┐        │
│ │ 🟢 在线设备 │ 🔴 离线设备 │ ⚠️ 活跃告警 │ 📊 今日数据 │        │
│ └────────────┴────────────┴────────────┴────────────┘        │
│                                                              │
│ 全局指标概览 (4列网格):                                       │
│ ┌────────────┬────────────┬────────────┬────────────┐        │
│ │ 平均温度    │ 平均湿度    │ 平均pH值   │ 平均电导率  │        │
│ └────────────┴────────────┴────────────┴────────────┘        │
│                                                              │
│ 两列布局:                                                     │
│ ┌─────────────────────────┬─────────────────────────────┐    │
│ │ 设备分布饼图            │ 最近告警列表 (5条)           │    │
│ │ - 类型分布              │ - 级别/类型/消息/时间        │    │
│ │ - 分组分布              │ - 查看全部链接               │    │
│ └─────────────────────────┴─────────────────────────────┘    │
│                                                              │
│ 分组指标 (可折叠):                                           │
│ ┌─────────────────────────────────────────────────────────┐  │
│ │ A区: 温度 24.5°C | 湿度 65% | pH 6.8                    │  │
│ │ B区: 温度 25.1°C | 湿度 62% | pH 7.0                    │  │
│ └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 1.3 API 接口

| 接口 | 用途 |
|------|------|
| `GET /api/overview/dashboard` | 获取概览数据 |
| `GET /api/alerts?status=OPEN&page_size=5` | 获取最近告警 |
| `GET /api/telemetry/latest?device_id=...` | 获取分组指标 |

### 1.4 响应格式

```typescript
// GET /api/overview/dashboard
interface DashboardOverview {
  devices_online: number
  devices_offline: number
  alerts_open: number
  today_data_points: number
  device_type_distribution: { type: string; count: number }[]
  device_group_distribution: { group_name: string; count: number }[]
  avg_metrics?: {
    temp?: number
    humidity?: number
    ph?: number
    ec?: number
  }
}
```

---

## 2. 控制命令

### 2.1 页面路由
- 路径: `/controls/commands`
- 权限: ADMIN/OPERATOR

### 2.2 页面布局

```
┌─────────────────────────────────────────────────────────────┐
│ 控制命令                                    [下发命令] 按钮   │
├─────────────────────────────────────────────────────────────┤
│ 筛选: [设备 ▼] [状态 ▼] [查询] [重置]                         │
├─────────────────────────────────────────────────────────────┤
│ 命令历史表格:                                                 │
│ ┌────┬──────────┬────────┬────────┬────────┬─────────────────┐│
│ │ ID │ 目标设备  │ 类型    │ 负载   │ 状态   │ 创建时间        ││
│ └────┴──────────┴────────┴────────┴────────┴─────────────────┘│
│ 分页: [<] 1 2 3 ... [>]                                      │
└─────────────────────────────────────────────────────────────┘
```

### 2.3 下发命令弹窗

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| device_id | Select | 是 | 选择执行器设备 |
| command_type | Select | 是 | SWITCH/SET 等 |
| payload | JSON编辑器 | 是 | 如 {"state": "ON"} |

### 2.4 命令状态

| 状态 | 颜色 | 说明 |
|------|------|------|
| PENDING | 蓝色 | 待发送 |
| SENT | 黄色 | 已发送 |
| EXECUTED | 绿色 | 已执行 |
| FAILED | 红色 | 执行失败 |

### 2.5 API 接口

| 接口 | 用途 |
|------|------|
| `GET /api/controls/commands` | 命令列表 |
| `POST /api/controls/commands` | 下发命令 |
| `GET /api/controls/commands/{id}` | 查询命令状态 |
| `GET /api/devices?type=ACTUATOR` | 获取执行器列表 |

---

## 3. 控制规则

### 3.1 页面路由
- 路径: `/controls/rules`
- 权限: ADMIN/OPERATOR

### 3.2 页面布局

```
┌─────────────────────────────────────────────────────────────┐
│ 控制规则                                    [新增规则] 按钮   │
├─────────────────────────────────────────────────────────────┤
│ 规则列表表格:                                                 │
│ ┌────┬──────────┬──────┬──────┬──────┬───────┬────────────┐  │
│ │ ID │ 规则名称  │ 指标  │ 条件  │ 阈值  │ 状态  │ 操作       │  │
│ └────┴──────────┴──────┴──────┴──────┴───────┴────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 3.3 新增/编辑规则弹窗

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | Input | 是 | 规则名称 (1-64字符) |
| metric_code | Select | 是 | 监控指标 |
| operator | Select | 是 | 比较运算符 >, >=, <, <=, == |
| threshold | InputNumber | 是 | 阈值 |
| target_device_id | Select | 是 | 目标执行器 |
| action | JSON编辑器 | 是 | 执行动作 |
| enabled | Switch | 是 | 启用状态 |

### 3.4 API 接口

| 接口 | 用途 |
|------|------|
| `GET /api/controls/rules` | 规则列表 |
| `POST /api/controls/rules` | 新增规则 |
| `PUT /api/controls/rules/{id}` | 更新规则 |
| `DELETE /api/controls/rules/{id}` | 删除规则 |

---

## 4. 告警中心

### 4.1 页面路由
- 路径: `/alerts`
- 权限: ADMIN/OPERATOR/VIEWER (VIEWER 仅查看)

### 4.2 页面布局

```
┌─────────────────────────────────────────────────────────────┐
│ 告警中心                                                      │
├─────────────────────────────────────────────────────────────┤
│ 统计卡片: 🔴 开放: 5  🟡 已确认: 2  🟢 已关闭: 10              │
├─────────────────────────────────────────────────────────────┤
│ 筛选: [类型 ▼] [级别 ▼] [状态 ▼] [查询] [重置]                │
├─────────────────────────────────────────────────────────────┤
│ 告警列表表格:                                                 │
│ ┌────┬──────┬───────┬────────────┬────────┬────────┬──────┐  │
│ │ ID │ 类型  │ 级别  │ 消息        │ 设备   │ 状态  │ 操作 │  │
│ └────┴──────┴───────┴────────────┴────────┴────────┴──────┘  │
│ 分页: [<] 1 2 3 ... [>]                                      │
└─────────────────────────────────────────────────────────────┘
```

### 4.3 处理告警弹窗

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | Select | 是 | ACK/CLOSED |
| comment | Textarea | 否 | 处理备注 (0-255字符) |

### 4.4 告警状态流转

```
OPEN (开放) → ACK (已确认) → CLOSED (已关闭)
```

### 4.5 告警级别

| 级别 | 颜色 | 说明 |
|------|------|------|
| INFO | 蓝色 | 信息 |
| WARN | 黄色 | 警告 |
| CRITICAL | 红色 | 严重 |

### 4.6 告警类型

| 类型 | 说明 |
|------|------|
| THRESHOLD | 阈值告警 |
| OFFLINE | 离线告警 |
| DEVICE_ERROR | 设备故障 |

### 4.7 API 接口

| 接口 | 用途 |
|------|------|
| `GET /api/alerts` | 告警列表 |
| `GET /api/alerts/{id}` | 告警详情 |
| `GET /api/alerts/stats` | 告警统计 |
| `PATCH /api/alerts/{id}/status` | 更新状态 |

---

## 5. 路由配置

```typescript
// 新增路由
{ path: '/', redirect: '/dashboard' },
{ path: '/dashboard', name: 'Dashboard', component: () => import('@/views/dashboard/index.vue'), meta: { requiresAuth: true, title: '首页' } },
{ path: '/controls/commands', name: 'ControlCommands', component: () => import('@/views/controls/commands.vue'), meta: { requiresAuth: true, roles: [Role.ADMIN, Role.OPERATOR], title: '控制命令' } },
{ path: '/controls/rules', name: 'ControlRules', component: () => import('@/views/controls/rules.vue'), meta: { requiresAuth: true, roles: [Role.ADMIN, Role.OPERATOR], title: '控制规则' } },
{ path: '/alerts', name: 'Alerts', component: () => import('@/views/alerts/index.vue'), meta: { requiresAuth: true, title: '告警中心' } }
```

---

## 6. 文件结构

```
src/
├── api/
│   ├── control.ts          # 控制命令与规则 API
│   ├── alert.ts            # 告警 API
│   └── overview.ts         # 仪表盘概览 API
├── types/
│   ├── control.ts          # ControlCommand, ControlRule 类型
│   └── alert.ts            # Alert, AlertType, AlertLevel, AlertStatus 类型
├── stores/
│   └── alert.ts            # 告警状态管理 (可选)
├── views/
│   ├── dashboard/
│   │   └── index.vue       # 首页仪表盘
│   ├── controls/
│   │   ├── commands.vue    # 控制命令页
│   │   └── rules.vue       # 控制规则页
│   └── alerts/
│       └── index.vue       # 告警中心
└── components/
    └── layout/
        └── AppSidebar.vue  # 更新侧边栏菜单
```

---

## 7. 权限矩阵

| 功能 | ADMIN | OPERATOR | VIEWER |
|------|-------|----------|--------|
| 首页仪表盘 | ✅ | ✅ | ✅ |
| 控制命令 | ✅ | ✅ | ❌ |
| 控制规则 | ✅ | ✅ | ❌ |
| 告警中心-查看 | ✅ | ✅ | ✅ |
| 告警中心-处理 | ✅ | ✅ | ❌ |