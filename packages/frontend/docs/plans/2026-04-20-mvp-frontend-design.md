# 水培农业管理系统前端设计文档

## 概述

本文档定义了水培农业管理系统 Web 前端 MVP 阶段的设计方案。

## 1. 项目范围

### MVP 功能模块

| 模块 | 功能 | 优先级 |
|------|------|--------|
| 登录 | 用户名/密码登录 | P0 |
| 设备列表 | 分页、筛选、搜索、新增/编辑 | P0 |
| 设备详情 | 查看设备信息、遥测数据 | P0 |
| 设备分组 | 分组管理、设备分配 | P0 |
| 遥测实时数据 | 设备遥测监控、自动刷新 | P0 |
| 遥测历史数据 | 数据查询、图表/表格展示 | P0 |

### 后续阶段（不在 MVP 范围）

- Phase 2：首页仪表盘、控制中心、告警中心
- Phase 3：用户管理、审计日志

## 2. 技术栈

| 类型 | 选型 | 版本 |
|------|------|------|
| 框架 | Vue 3 + Composition API | ^3.4 |
| UI 组件库 | Element Plus | ^2.5 |
| 构建工具 | Vite | ^5.0 |
| 状态管理 | Pinia | ^2.1 |
| 路由 | Vue Router | ^4.2 |
| HTTP 客户端 | Axios | ^1.6 |
| 图表 | ECharts | ^5.5 |
| 语言 | TypeScript | ^5.0 |

## 3. 项目结构

```
hydroponic-frontend/
├── public/
│   └── favicon.ico
├── src/
│   ├── api/                    # API 请求封装
│   │   ├── auth.ts             # 登录认证
│   │   ├── device.ts           # 设备管理
│   │   ├── device-group.ts     # 设备分组
│   │   ├── telemetry.ts        # 遥测数据
│   │   └── request.ts          # Axios 实例配置
│   ├── assets/                 # 静态资源
│   │   └── styles/
│   │       ├── variables.scss  # 样式变量
│   │       └── global.scss     # 全局样式
│   ├── components/             # 公共组件
│   │   ├── layout/
│   │   │   ├── AppHeader.vue   # 顶部导航
│   │   │   ├── AppSidebar.vue  # 侧边栏
│   │   │   └── AppLayout.vue   # 整体布局
│   │   └── common/             # 通用业务组件
│   ├── composables/            # 组合式函数
│   │   ├── useAuth.ts          # 认证相关
│   │   └── usePermission.ts    # 权限判断
│   ├── router/                 # 路由配置
│   │   └── index.ts
│   ├── stores/                 # Pinia 状态管理
│   │   ├── auth.ts             # 用户认证状态
│   │   └── device.ts           # 设备相关状态
│   ├── types/                  # TypeScript 类型
│   │   ├── api.d.ts            # API 响应类型
│   │   └── device.d.ts         # 设备相关类型
│   ├── utils/                  # 工具函数
│   │   ├── format.ts           # 格式化函数
│   │   └── storage.ts          # 本地存储
│   ├── views/                  # 页面组件
│   │   ├── login/
│   │   │   └── index.vue
│   │   ├── devices/
│   │   │   ├── list.vue        # 设备列表
│   │   │   └── detail.vue      # 设备详情
│   │   ├── device-groups/
│   │   │   └── index.vue       # 设备分组
│   │   └── telemetry/
│   │       ├── realtime.vue    # 实时数据
│   │       └── history.vue     # 历史数据
│   ├── App.vue
│   └── main.ts
├── .env.development            # 开发环境变量
├── .env.production             # 生产环境变量
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## 4. 路由设计

| 路由 | 页面 | 权限 | 说明 |
|------|------|------|------|
| `/login` | 登录页 | 公开 | 用户登录 |
| `/` | 重定向 | - | 重定向到 `/devices` |
| `/devices` | 设备列表 | 全部 | 分页、筛选、搜索 |
| `/devices/:id` | 设备详情 | 全部 | 设备信息、遥测数据 |
| `/device-groups` | 设备分组 | 全部 | 分组管理 |
| `/telemetry/realtime` | 实时数据 | 全部 | 设备遥测监控 |
| `/telemetry/history` | 历史数据 | 全部 | 数据查询、图表展示 |

## 5. 认证与权限

### 5.1 认证流程

1. 用户输入用户名/密码登录
2. 成功后存储 JWT Token 到 localStorage
3. 存储用户信息到 Pinia store
4. 路由守卫检查 Token，未登录跳转登录页
5. 退出登录清除 Token 和用户信息

### 5.2 权限控制

**角色**：ADMIN、OPERATOR、VIEWER

**权限矩阵**：

| 功能 | ADMIN | OPERATOR | VIEWER |
|------|-------|----------|--------|
| 设备列表/详情 | ✅ | ✅ | ✅ |
| 新增/编辑设备 | ✅ | ❌ | ❌ |
| 设备分组管理 | ✅ | ❌ | ❌ |
| 遥测数据查看 | ✅ | ✅ | ✅ |

**实现方式**：
- 路由级：路由 meta 配置所需角色，守卫判断
- 组件级：`v-if="hasPermission('ADMIN')"` 控制按钮显隐

## 6. HTTP 请求封装

### 6.1 Axios 配置

- 基础地址：`http://localhost:8080/api`
- 请求拦截：自动添加 `Authorization: Bearer <token>`
- 响应拦截：统一错误处理

### 6.2 错误处理

| 场景 | 处理方式 |
|------|----------|
| 网络错误 | ElMessage 提示"网络异常" |
| 401 未登录 | 清除 Token，跳转登录页 |
| 403 无权限 | ElMessage 提示"无权限" |
| 404 资源不存在 | ElMessage 提示"资源不存在" |
| 业务错误 | 显示后端返回的 message |

## 7. 页面功能

### 7.1 登录页

- 用户名/密码输入表单
- 表单验证（用户名 3-32 字符，密码 6-64 字符）
- 登录按钮 loading 状态
- 登录失败显示错误提示

### 7.2 设备列表页

- 筛选区：类型、分类、分组、状态下拉 + 关键词搜索
- 表格：设备编码、名称、类型、分类、分组、状态、操作
- 分页：支持切换每页条数
- 操作：查看详情、编辑/禁用（仅 ADMIN）
- 新增设备按钮（仅 ADMIN）

### 7.3 设备详情页

- 基本信息卡片
- 在线状态显示
- 最新遥测数据卡片
- 操作按钮：编辑、禁用（仅 ADMIN）

### 7.4 设备分组页

- 分组卡片列表
- 新增/编辑分组弹窗
- 管理设备弹窗（穿梭框）

### 7.5 遥测实时数据页

- 设备选择下拉框
- 数据卡片：多指标展示
- 自动刷新开关（默认 30 秒）
- 趋势图：最近 1 小时数据

### 7.6 遥测历史数据页

- 查询条件：设备、指标、时间范围
- 统计摘要：平均值、最大值、最小值
- 展示切换：图表 / 表格
- 分页查询

## 8. API 接口

详见 `docs/API_SPEC.md`，MVP 阶段涉及以下接口：

- 认证：`POST /api/auth/login`
- 设备：`GET/POST/PUT/PATCH /api/devices/*`
- 分组：`GET/POST/PUT /api/device-groups/*`
- 遥测：`GET /api/telemetry/latest`、`GET /api/telemetry/history`、`GET /api/telemetry/stats`

---

*设计批准日期：2026-04-20*
