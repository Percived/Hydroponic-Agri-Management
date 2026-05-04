# 水培农业管理系统 Web 前端

水培农业管理系统 Web 前端是为温室/大棚水培环境设计的管理平台，支持设备管理、实时监控、数据采集展示、控制指令下发等功能。

## 技术栈

- **框架**: Vue 3 + Composition API
- **UI 组件库**: Element Plus
- **构建工具**: Vite
- **状态管理**: Pinia
- **路由**: Vue Router
- **HTTP 客户端**: Axios
- **图表**: ECharts
- **语言**: TypeScript

## 功能特性

### 核心功能 (MVP)

- 用户登录认证
- 设备列表与详情管理
- 设备分组管理
- 遥测数据实时监控
- 遥测数据历史查询

### 扩展功能

- 首页仪表盘
- 控制命令下发
- 自动控制规则配置
- 告警中心
- 用户管理
- 审计日志

### 权限控制

| 角色 | 权限范围 |
|------|---------|
| 管理员 (ADMIN) | 全部功能 |
| 操作员 (OPERATOR) | 查询 + 控制 + 告警处理 |
| 只读用户 (VIEWER) | 仅查询 |

## 快速开始

### 环境要求

- Node.js >= 18.0.0
- npm >= 9.0.0

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

启动后访问 http://localhost:5173

### 生产构建

```bash
npm run build
```

### 预览构建

```bash
npm run preview
```

### 类型检查

```bash
npm run type-check
```

### 代码检查

```bash
npm run lint
```

## 项目结构

```
src/
├── api/           # API 请求模块 (auth.ts, device.ts, telemetry.ts 等)
├── assets/        # 静态资源和全局样式
├── components/    # 共享组件 (layout/, common/)
├── composables/   # Vue 组合式函数 (useAuth.ts, usePermission.ts)
├── router/        # 路由配置
├── stores/        # Pinia 状态管理 (auth.ts, device.ts)
├── types/         # TypeScript 类型定义
├── utils/         # 工具函数 (request.ts, format.ts)
└── views/         # 页面组件
```

## 架构说明

### API 层

每个业务域有独立的 API 模块，所有请求通过 `request.ts` 中的 Axios 实例统一处理：

- 自动添加 JWT Token 到 Authorization 请求头
- 统一错误处理 (401 跳转登录, 403 权限拒绝)
- 响应格式: `{ code, message, data, request_id }`

### 认证机制

- JWT Token 存储在 localStorage
- 用户信息存储在 Pinia store
- 路由守卫检查 Token 有效性

### 权限控制

- 通过路由 meta 和 `hasPermission()` 组合式函数控制
- 三种角色: ADMIN (完全访问), OPERATOR (查询+控制), VIEWER (仅查询)

### 状态管理

- Pinia store 管理全局状态 (认证、设备选择等)
- 页面级数据使用组件本地状态

## API 配置

- 开发环境: `http://localhost:8080/api`
- 通过 `.env.development` 和 `.env.production` 配置

## 文档

- [产品需求文档](docs/FRONTEND_PRD.md) - 完整产品需求说明
- [API 规范](docs/API_SPEC.md) - 完整 API 接口规范
- [MVP 设计方案](docs/plans/2026-04-20-mvp-frontend-design.md) - MVP 阶段设计决策

## 开发计划

### Phase 1 (MVP)

- [x] 登录页
- [ ] 设备列表 + 详情
- [ ] 设备分组
- [ ] 遥测实时数据
- [ ] 遥测历史数据

### Phase 2

- [ ] 首页仪表盘
- [ ] 控制命令
- [ ] 控制规则
- [ ] 告警中心

### Phase 3

- [ ] 用户管理
- [ ] 审计日志
- [ ] 系统配置

## License

MIT