# Chat2API Go+Wails 重构开发设计文档

> **文档版本**: v1.0
> **创建日期**: 2026-06-14
> **项目**: Chat2API Electron → Wails 重构
> **目标**: 将现有 Electron + React + TypeScript 应用迁移至 Wails + Go + React

---

## 一、项目概述

### 1.1 项目背景

Chat2API 是一个多平台 AI 服务统一管理工具，当前基于 Electron 框架开发。本文档旨在分析现有架构，设计将核心技术栈从 Electron 迁移至 Wails (Go) 的详细方案。

### 1.2 当前技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 桌面框架 | Electron 33+ | 主进程 + 渲染进程架构 |
| 前端 | React 18 + TypeScript | 渲染进程 UI |
| 构建工具 | Vite + electron-vite | 开发调试与打包 |
| 后端服务 | Koa | 内置代理服务器 |
| 状态存储 | electron-store | JSON 文件持久化 |
| UI 组件 | Radix UI + Tailwind CSS | 现代化组件库 |
| 状态管理 | Zustand | React 状态管理 |

### 1.3 重构目标

1. **性能提升**: Go 的并发处理和内存效率优于 Node.js
2. **二进制体积**: Wails 打包的应用程序积显著小于 Electron
3. **启动速度**: Go 程序启动速度更快
4. **开发体验**: 保持 React 前端，热重载支持
5. **跨平台**: 支持 macOS、Windows、Linux

---

## 二、现有架构分析

### 2.1 Electron 架构

```
┌─────────────────────────────────────────────────────────┐
│                    Electron Main Process                 │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────┐  │
│  │   Window    │ │   Tray      │ │    Updater      │  │
│  │   Manager   │ │   Manager   │ │    Manager      │  │
│  └─────────────┘ └─────────────┘ └─────────────────┘  │
│  ┌─────────────────────────────────────────────────────┐│
│  │              IPC Handler (handlers.ts)             ││
│  │  - providers.*    - accounts.*                    ││
│  │  - proxy.*         - oauth.*                      ││
│  │  - logs.*          - config.*                      ││
│  │  - session.*       - statistics.*                  ││
│  └─────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────┐│
│  │              Proxy Server (Koa)                     ││
│  │  - /v1/chat/completions   - /v1/models             ││
│  │  - /v0/management/*     - /health, /stats        ││
│  └─────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────┐│
│  │              Store Manager (electron-store)         ││
│  │  - config.json    - providers.json                ││
│  │  - accounts.json   - logs/                         ││
│  └─────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────┘
                           │ IPC
┌─────────────────────────────────────────────────────────┐
│                    Preload (contextBridge)              │
│  - 暴露 electronAPI 到渲染进程                          │
│  - 封装 IPC 通信                                        │
└─────────────────────────────────────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────┐
│                  Electron Renderer Process              │
│  ┌─────────────────────────────────────────────────────┐│
│  │                   React Application                 ││
│  │  - Pages: Dashboard, Providers, Proxy, Logs...     ││
│  │  - Components: UI 组件库                            ││
│  │  - Stores: Zustand 状态管理                         ││
│  │  - Hooks: 自定义 React Hooks                       ││
│  └─────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────┘
```

### 2.2 核心模块分析

| 模块 | 文件位置 | 功能 | 迁移复杂度 |
|------|----------|------|----------|
| Window Manager | src/main/window/ | 窗口创建管理 | ⭐ 低 |
| Tray Manager | src/main/tray/ | 系统托盘 | ⭐ 低 |
| Proxy Server | src/main/proxy/server.ts | HTTP 代理 + API | ⭐⭐⭐ 高 |
| IPC Handlers | src/main/ipc/handlers.ts | 进程通信 | ⭐⭐ 中 |
| Store Manager | src/main/store/ | 数据持久化 | ⭐⭐ 中 |
| OAuth Manager | src/main/oauth/ | 第三方认证 | ⭐⭐⭐ 高 |
| Providers | src/main/providers/ | AI 提供商适配 | ⭐⭐ 中 |
| Forwarder | src/main/proxy/forwarder.ts | 请求转发 | ⭐⭐⭐ 高 |

---

## 三、Wails 架构设计

### 3.1 目标架构

```
┌─────────────────────────────────────────────────────────┐
│                      Wails Backend (Go)                 │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────┐  │
│  │   Window    │ │   Tray      │ │    Updater      │  │
│  │   Manager   │ │   Manager   │ │    Manager      │  │
│  └─────────────┘ └─────────────┘ └─────────────────┘  │
│  ┌─────────────────────────────────────────────────────┐│
│  │              Wails Runtime (JavaScript Bridge)      ││
│  │  - window.*      - events.*                        ││
│  │  - clipboard.*   - storage.*                        ││
│  └─────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────┐│
│  │              Go Application Layer                    ││
│  │  ┌──────────────┐ ┌──────────────┐ ┌────────────┐ ││
│  │  │ Proxy Server │ │ Provider     │ │ OAuth      │ ││
│  │  │ (net/http)   │ │ Manager      │ │ Manager    │ ││
│  │  └──────────────┘ └──────────────┘ └────────────┘ ││
│  │  ┌──────────────┐ ┌──────────────┐ ┌────────────┐ ││
│  │  │ Store        │ │ Session      │ │ Statistics │ ││
│  │  │ Manager      │ │ Manager      │ │ Manager    │ ││
│  │  └──────────────┘ └──────────────┘ └────────────┘ ││
│  └─────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────┘
                           │ Wails Bindings
┌─────────────────────────────────────────────────────────┐
│                      Frontend (React)                   │
│  - 保持现有 React 代码结构                              │
│  - 通过 window.go实时通信                              │
│  - Zustand 状态管理                                     │
└─────────────────────────────────────────────────────────┘
```

### 3.2 目录结构设计

```
chat2api-wails/
├── wails.json                    # Wails 项目配置
├── main.go                       # Wails 入口
├── build/                        # 构建资源
│   ├── appicon.png
│   └── .../
├── frontend/                     # 前端目录 (原 src/renderer)
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── stores/
│   │   ├── hooks/
│   │   └── ...
│   ├── package.json
│   └── vite.config.ts
├── internal/                     # Go 内部模块
│   ├── proxy/                    # 代理服务器
│   │   ├── server.go
│   │   ├── routes/
│   │   ├── forwarder.go
│   │   └── toolcalling/
│   ├── providers/               # AI 提供商
│   │   ├── manager.go
│   │   └── builtin/
│   ├── oauth/                    # OAuth 认证
│   │   ├── manager.go
│   │   └── adapters/
│   ├── store/                    # 数据存储
│   │   ├── manager.go
│   │   └── crypto.go
│   ├── session/                  # 会话管理
│   │   └── manager.go
│   ├── tray/                     # 系统托盘
│   │   └── manager.go
│   ├── updater/                  # 自动更新
│   │   └── manager.go
│   └── types/                    # 共享类型
│       └── types.go
└── pkg/                          # 公共包
    └── utils/
```

---

## 四、模块迁移设计

### 4.1 Proxy Server (代理服务器)

**现状**: Koa + TypeScript

**目标**: Go net/http

**核心路由**:

| 路径 | 方法 | 功能 |
|------|------|------|
| `/v1/chat/completions` | POST | 聊天完成 (流式/非流式) |
| `/v1/models` | GET | 模型列表 |
| `/v1/models/:model` | GET | 模型信息 |
| `/v0/management/*` | * | 管理 API |
| `/health` | GET | 健康检查 |
| `/stats` | GET | 统计信息 |

**Go 核心实现**:

```go
// internal/proxy/server.go
type ProxyServer struct {
    mux         *http.ServeMux
    server      *http.Server
    port        int
    host        string
    // 中间件
    apiKeyAuth  *ApiKeyAuthMiddleware
    logger      *Logger
    // 转发器
    forwarder   *Forwarder
}

func (s *ProxyServer) Start(port int, host string) error {
    s.setupRoutes()
    s.server = &http.Server{
        Addr:         fmt.Sprintf("%s:%d", host, port),
        Handler:      s.mux,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 60 * time.Second,
    }
    return s.server.ListenAndServe()
}
```

### 4.2 Store Manager (数据存储)

**现状**: electron-store (Node.js)

**目标**: Go JSON 文件存储 + AES-256-GCM 加密

**数据结构**:

```
~/.chat2api/
├── config.json      # 应用配置
├── providers.json   # 提供商配置
├── accounts.json    # 账户凭证 (加密)
└── logs/
    └── app.log      # 应用日志
```

**Go 核心实现**:

```go
// internal/store/manager.go
type StoreManager struct {
    basePath string
    mu       sync.RWMutex
    config   *AppConfig
    // 文件缓存
    files    map[string][]byte
}

func (s *StoreManager) Get(key string) ([]byte, error)
func (s *StoreManager) Set(key string, data []byte) error
func (s *StoreManager) SaveAccounts(accounts []Account) error
func (s *StoreManager) LoadAccounts() ([]Account, error)
```

### 4.3 OAuth Manager (OAuth 认证)

**现状**: Electron BrowserWindow 内打开 OAuth 页面

**目标**: Wails 内置 BrowserWindow 或系统浏览器

**支持提供商**:

| 提供商 | 认证类型 | 适配器 |
|--------|----------|--------|
| DeepSeek | User Token | ✅ |
| GLM | Refresh Token | ✅ |
| Kimi | JWT Token | ✅ |
| MiniMax | JWT Token | ✅ |
| Perplexity | Cookie | ✅ |
| Qwen | SSO Ticket | ✅ |
| Z.ai | JWT Token | ⚠️ 验证码风险 |

**Go 核心实现**:

```go
// internal/oauth/manager.go
type OAuthManager struct {
    providers map[string]OAuthAdapter
    loginWindow *wails.JSWindow
}

type OAuthAdapter interface {
    GetAuthURL() (string, error)
    HandleCallback(url string) (*TokenResult, error)
    RefreshToken(credentials map[string]string) (*TokenResult, error)
}
```

### 4.4 IPC 通信设计

**现状**: Electron IPC (ipcMain/ipcRenderer)

**目标**: Wails Call / Events 机制

**前端调用 Go 方法**:

```typescript
// 启动代理
const started = await window.go.proxy.Server.Start(8080, "127.0.0.1")

// 获取状态
const status = await window.go.proxy.Server.GetStatus()

// 监听事件
window.go.proxy.Events.OnStatusChanged((status: ProxyStatus) => {
    console.log('Status changed:', status)
})
```

**Wails 绑定示例**:

```go
// main.go
func (a *App) StartProxy(port int, host string) bool {
    return proxyServer.Start(port, host)
}

func (a *App) StopProxy() bool {
    return proxyServer.Stop()
}

func (a *App) GetProxyStatus() *ProxyStatus {
    return proxyServer.GetStatus()
}
```

---

## 五、数据迁移设计

### 5.1 配置格式映射

| Electron (JSON) | Wails (JSON) | 说明 |
|-----------------|--------------|------|
| config.json | config.json | 保持兼容 |
| providers.json | providers.json | 保持兼容 |
| accounts.json | accounts.json | 保持兼容，需加密存储 |

### 5.2 加密方案

凭证 (accounts.json) 使用 AES-256-GCM 加密:

```go
// internal/store/crypto.go
func Encrypt(data []byte, key []byte) ([]byte, error)
func Decrypt(data []byte, key []byte) ([]byte, error)
```

密钥从机器特定信息派生 (机器 ID + 用户名)。

---

## 六、UI 集成设计

### 6.1 前端保持策略

**保持不变**:

- React 18 + TypeScript
- Vite 构建
- Tailwind CSS
- Radix UI 组件
- Zustand 状态管理
- React Router

**需要修改**:

- `src/preload/index.ts` → 删除 (Wails 自动处理)
- `window.electronAPI` → `window.go`
- IPC 调用 → Wails 绑定调用

### 6.2 API 适配层

```typescript
// frontend/src/lib/wails-adapter.ts
interface WailsAPI {
  // 代理
  proxy: {
    start: (port?: number, host?: string) => Promise<boolean>
    stop: () => Promise<boolean>
    getStatus: () => Promise<ProxyStatus>
    onStatusChanged: (callback: (status: ProxyStatus) => void) => void
  }
  // 提供商
  providers: {
    getAll: () => Promise<Provider[]>
    add: (data: ProviderInput) => Promise<Provider>
    // ...
  }
  // 账户
  accounts: {
    getAll: (includeCredentials?: boolean) => Promise<Account[]>
    // ...
  }
  // OAuth
  oauth: {
    startLogin: (providerId: string, providerType: string) => Promise<OAuthResult>
    // ...
  }
  // 配置
  config: {
    get: () => Promise<AppConfig>
    update: (updates: Partial<AppConfig>) => Promise<boolean>
  }
}

declare global {
  interface Window {
    go: WailsAPI
  }
}
```

---

## 七、构建与部署

### 7.1 Wails 项目初始化

```bash
# 创建 Wails 项目
wails init -name chat2api -frontend "npm create vite@latest frontend -- --template react-ts"

# 目录结构调整
mv frontend/* .
rm -rf frontend
```

### 7.2 构建命令

```bash
# 开发模式
wails dev

# 生产构建
wails build

# 跨平台构建
wails build -platform windows/amd64
wails build -platform darwin/universal
wails build -platform linux/amd64
```

### 7.3 打包配置

```json
// wails.json
{
  "name": "Chat2API",
  "outputfilename": "Chat2API",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "author": "Chat2API Team",
  "info": {
    "companyName": "Chat2API",
    "productName": "Chat2API",
    "copyright": "Copyright © 2026 Chat2API Team",
    "comments": "Multi-platform AI Service Unified Management Tool"
  }
}
```

---

## 八、迁移任务分解

### Phase 1: 基础设施 (1-2 周)

| 任务 | 描述 | 预估工时 |
|------|------|----------|
| T1.1 | 创建 Wails 项目骨架 | 2h |
| T1.2 | 配置 Go 模块和依赖 | 1h |
| T1.3 | 配置前端构建 (Vite) | 2h |
| T1.4 | 实现日志系统 | 4h |
| T1.5 | 实现配置存储 | 4h |

### Phase 2: 核心功能 (2-3 周)

| 任务 | 描述 | 预估工时 |
|------|------|----------|
| T2.1 | 实现代理服务器 | 2d |
| T2.2 | 实现提供商管理器 | 2d |
| T2.3 | 实现 OAuth 认证 | 2d |
| T2.4 | 实现账户管理 | 1d |
| T2.5 | 实现会话管理 | 1d |

### Phase 3: UI 集成 (1-2 周)

| 任务 | 描述 | 预估工时 |
|------|------|----------|
| T3.1 | 创建 Wails API 适配层 | 4h |
| T3.2 | 迁移 IPC 调用 | 4h |
| T3.3 | 实现事件系统 | 4h |
| T3.4 | UI 调试和修复 | 2d |

### Phase 4: 系统集成 (1 周)

| 任务 | 描述 | 预估工时 |
|------|------|----------|
| T4.1 | 实现系统托盘 | 4h |
| T4.2 | 实现窗口管理 | 4h |
| T4.3 | 实现自动更新 | 4h |
| T4.4 | 系统测试和修复 | 2d |

### Phase 5: 发布准备 (3-5 天)

| 任务 | 描述 | 预估工时 |
|------|------|----------|
| T5.1 | 跨平台构建测试 | 1d |
| T5.2 | 性能优化 | 1d |
| T5.3 | 文档更新 | 4h |
| T5.4 | 发布版本构建 | 4h |

---

## 九、风险与挑战

### 9.1 技术风险

| 风险 | 影响 | 缓解方案 |
|------|------|----------|
| OAuth 登录流程变化 | 高 | 使用应用内浏览器窗口，保持与现有流程一致 |
| 代理服务器兼容性 | 中 | 完整测试 OpenAI 客户端兼容性 |
| 数据迁移丢失 | 高 | 实现双向数据迁移工具 |
| WebAssembly 依赖 | 中 | 评估 sha3_wasm 迁移方案 |

### 9.2 已知的兼容性问题

1. **electron-store**: Node.js 特有的 `electron-store` 无法直接迁移，需要用 Go 原生实现
2. **electron-updater**: Wails 有内置更新机制，需重新实现
3. **Tray 行为差异**: Wails 的系统托盘 API 与 Electron 有所不同
4. **BrowserWindow OAuth**: OAuth 登录窗口需要使用 Wails 的窗口管理

---

## 十、附录

### 10.1 Go 依赖推荐

```go
// go.mod
module chat2api

go 1.21

require (
    github.com/wailsapp/wails/v2 v2.8.0
    golang.org/x/net v0.21.0
    github.com/google/uuid v1.6.0
    github.com/spf13/viper v1.18.2
    go.uber.org/zap v1.27.0
)
```

### 10.2 参考资料

- [Wails 官方文档](https://wails.io/docs/gettingstarted/installation)
- [Wails 2.x 中文文档](https://www.wails.top/)
- [Wails 示例项目](https://github.com/wailsapp/wails/tree/master/examples)

---

## 十一、总结

本设计文档详细分析了 Chat2API 从 Electron 迁移至 Wails 的完整方案，包括：

1. **架构分析**: 清晰映射现有 Electron 模块到 Wails/Go 实现
2. **模块设计**: 详细设计每个核心模块的 Go 实现方案
3. **数据迁移**: 设计配置格式兼容和加密存储方案
4. **UI 集成**: 保持 React 前端，创建 Wails API 适配层
5. **任务分解**: 将大迁移分解为可管理的阶段和任务

重构后的优势：
- **性能**: Go 的效率和并发处理能力
- **体积**: 更小的二进制文件
- **体验**: 更快的启动速度
- **维护**: 统一的 Go 技术栈
