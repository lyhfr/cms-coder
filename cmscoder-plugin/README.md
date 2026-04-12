# cmscoder-plugin

cmscoder 插件端，为 Claude Code 和 OpenCode 提供企业级 AI 编程能力。包含 IAM 登录闭环、本地会话管理、安全存储、配置同步、工作流增强等核心功能。

## 架构概览

```
cmscoder-plugin/
├── bin/                          # 入口脚本
│   └── cmscoder-init             #   安装/初始化脚本 (Node.js)
├── lib/                          # 公共核心层（纯 Node.js，零 npm 依赖）
│   ├── cmscoder.js               #   CLI 入口 + 模块导出
│   ├── auth.js                   #   登录/登出/刷新编排
│   ├── callback-server.js        #   HTTP 回环服务
│   ├── storage.js                #   安全存储 + 本地缓存
│   ├── http-client.js            #   API 调用封装
│   └── bootstrap.js              #   服务端配置同步
└── adapters/                     # Agent 适配层
    ├── claude-code/              # Claude Code 适配
    │   ├── CLAUDE.md             #   系统提示 + 企业规范
    │   ├── settings.json         #   hooks 配置
    │   ├── hooks/                #   事件钩子（薄 shell 包装）
    │   │   ├── session-start/    #     会话启动检查
    │   │   ├── pre-command/      #     命令前置检查
    │   │   └── post-command/     #     命令后置审计
    │   └── skills/               #   用户技能
    │       ├── cmscoder-login/   #     /cmscoder-login
    │       └── cmscoder-status/  #     /cmscoder-status
    └── opencode/                 # OpenCode 适配
        ├── README.md
        └── skills/
            ├── cmscoder-login/
            └── cmscoder-status/
```

## 外部依赖

| 依赖 | 用途 | 运行时要求 |
|------|------|-----------|
| **cmscoder-web-server** | 认证 API、配置同步、模型访问 | 必需，HTTP 可达 |
| **系统浏览器** | IAM 登录页面展示 | 必需，`open`/`xdg-open` |
| **Node.js** | 所有核心逻辑运行 | 必需，v14+ |
| **macOS Keychain** | 安全存储 token | macOS 系统自带，`security` 命令 |
| **Windows AES-256-GCM** | 安全存储 token | Node.js 内置 `crypto` 模块 |
| **Bash** | Claude Code hooks 薄包装 | 系统自带即可 |

**无 Python 依赖、无 curl 依赖、无 npm 依赖。**

### 与 cmscoder-web-server 的交互

所有对外 API 调用均通过 `POST/GET ${BACKEND_URL}/api/*`：

| 调用方 | 端点 | 方法 | 触发时机 |
|--------|------|------|---------|
| `auth.js` | `/api/auth/login` | POST | 用户触发登录 |
| `auth.js` | `/api/auth/exchange` | POST | 收到 login_ticket 后 |
| `auth.js` | `/api/auth/refresh` | POST | token 过期自动刷新 |
| `auth.js` | `/api/auth/logout` | POST | 用户触发登出 |
| `bootstrap.js` | `/api/plugin/bootstrap` | GET | 登录成功后 / 定期同步 |
| `check-auth.sh` | (间接 via ensure-session) | - | Claude Code 会话启动 |

### 与外部系统的交互（间接）

| 外部系统 | 交互方式 | 说明 |
|---------|---------|------|
| **公司 IAM** | 系统浏览器 | 插件不直接调用 IAM，由浏览器跳转完成 OAuth 授权码流程 |
| **cmscoder-web-server** | HTTPS | 所有认证、配置、模型请求的统一入口 |

## 核心流程

### 登录流程

```
用户触发 /cmscoder-login
    │
    ├── 1. startCallbackServer()
    │      └── Node.js http.createServer() 监听 127.0.0.1:随机端口
    │
    ├── 2. POST /api/auth/login
    │      Body: {localPort, agentType, pluginInstanceId}
    │      返回: {loginId, browserUrl, expiresAt}
    │
    ├── 3. 打开系统浏览器访问 browserUrl
    │      └── 浏览器 → web-server → IAM 授权 → 用户登录
    │          IAM 回调 → web-server → user-service 签发 login_ticket
    │          302 → http://127.0.0.1:<port>/callback?login_ticket=xxx
    │
    ├── 4. callback-server.js 接收回调
    │      └── Promise resolve(ticket)
    │
    ├── 5. POST /api/auth/exchange
    │      Body: {loginTicket, pluginInstanceId}
    │      返回: {accessToken, refreshToken, expiresIn, user}
    │
    ├── 6. 安全存储
    │      ├── secureStore: access_token, refresh_token, user_info
    │      └── localCache: session_meta, server_config
    │
    ├── 7. bootstrapSync()
    │      └── GET /api/plugin/bootstrap → 缓存配置和默认模型
    │
    └── 8. server.close()
```

### 会话恢复流程

```
Claude Code 启动
    │
    └── session-start hook (check-auth.sh)
         │
         └── node cmscoder.js ensure-session
                ├── 检查 secureStore 是否有 access_token
                ├── 检查 localCache session_meta 是否过期
                ├── 如过期 → POST /api/auth/refresh
                │
                ├── 有效 → exit 0
                └── 无效 → exit 1，提示 "Run /cmscoder-login"
```

### Token 刷新流程

```
请求服务端前或收到 401
    │
    ├── getAccessToken()
    │      ├── 从 secureStore 读取 access_token
    │      ├── 检查 session_meta.expiresAt
    │      ├── 如未过期 → 直接返回
    │      └── 如已过期 → refreshSilent()
    │             ├── POST /api/auth/refresh {refreshToken}
    │             ├── 返回新 {accessToken, refreshToken, expiresIn}
    │             └── 更新 secureStore 和 localCache
    │
    └── 刷新失败 → 返回 null，提示重新登录
```

### 注销流程

```
用户触发 /cmscoder-logout
    │
    ├── POST /api/auth/logout {refreshToken}
    │      └── 服务端撤销 refresh_token，标记 session 失效
    │
    ├── secureStore.clearAll()
    │      └── 清除 access_token, refresh_token, user_info
    │
    └── localCache.clear()
           └── 清除 session_meta, server_config, bootstrap_data
```

## 核心模块说明

### lib/cmscoder.js

CLI 入口 + 模块导出。直接运行时作为命令行工具，`require()` 时作为 API 模块。

**子命令**:
- `login` — 完整登录流程
- `logout` — 注销流程
- `refresh` — 静默刷新
- `status` — 展示当前会话状态
- `token` — 输出有效 access_token 到 stdout
- `ensure-session` — 检查并恢复会话（用于 hooks，成功 exit 0 / 失败 exit 1）

### lib/auth.js

认证编排核心。完整的登录/登出/刷新逻辑。

**导出**:
- `login()` — 完整登录：启动回环 → 创建 session → 打开浏览器 → 等待 ticket → 交换 token → 存储 → bootstrap sync
- `logout()` — 注销：服务端撤销 + 清理本地数据
- `refreshSilent()` — 静默刷新，成功返回 true
- `ensureSession()` — 检查 + 自动刷新，无效返回 false
- `getAccessToken()` — 返回有效 token，无效返回 null
- `status()` — 格式化输出会话状态

### lib/callback-server.js

轻量 Node.js HTTP 服务，仅用于接收浏览器回跳的 login_ticket。

- 监听 `127.0.0.1` 随机端口
- 仅处理 `GET /callback?login_ticket=xxx` 一个请求
- 返回 `{ port: Promise, waitForTicket: Promise, close: Function }`
- 5 分钟超时自动退出
- 不暴露于公网

### lib/storage.js

跨平台安全凭证存储 + 本地缓存。

**secureStore** — 安全存储（自动检测后端）:

| 平台 | 后端 | 命令 |
|------|------|------|
| macOS | Keychain | `security add/find/delete-generic-password` |
| Windows | AES-256-GCM 加密文件 | Node.js `crypto`（PBKDF2 派生密钥） |
| Linux (有 libsecret) | Secret Service | `secret-tool store/lookup/clear` |
| 其他 (开发环境) | 文件 | `~/.cmscoder/.secure-store/*` (权限 600) |

**localCache** — 文件系统 JSON 文件:

**存储位置**: `~/.cmscoder/cache/`

**缓存项**:
- `session_meta` — 用户摘要、过期时间（用于快速检查会话有效性）
- `server_config` — 后端 URL、默认模型、功能开关
- `bootstrap_data` — 服务端 bootstrap 响应原文
- `default_model` — 默认模型 ID
- `last_error` — 最近一次错误

### lib/http-client.js

HTTP API 客户端，使用 Node.js 内置 `http/https` 模块。

- `post(path, body, token?)` / `get(path, token?)`
- 自动设置 `Content-Type`, `X-Trace-Id`
- 从环境变量 → 缓存 → 配置文件读取 backend URL
- 解析 GoFrame 标准响应格式 `{code, data}`

### lib/bootstrap.js

服务端配置同步。

- `sync(accessToken)` — GET /api/plugin/bootstrap，缓存默认模型和功能开关
- `getCached()` — 返回缓存的 bootstrap 数据

## Claude Code 适配

### 安装位置

| 源文件 | 目标位置 | 说明 |
|--------|---------|------|
| `adapters/claude-code/CLAUDE.md` | `~/.claude/CLAUDE.md` | 系统提示 + 企业规范 |
| `adapters/claude-code/skills/*/SKILL.md` | `~/.claude/skills/*/SKILL.md` | 用户技能 |
| `adapters/claude-code/hooks/*/` | `~/.claude/hooks/*/` | 事件钩子 |

### Hooks 机制

Claude Code hooks 需要 shell 命令，因此保留极薄的 bash 封装（一行 `node ...`），所有业务逻辑在 JS 中：

| Hook | 触发时机 | 实现 |
|------|---------|------|
| `SessionStart` | 新会话启动 | `node cmscoder.js ensure-session`，失败时提示登录 |
| `PreCommand` | 命令执行前 | 占位（预留工具治理） |
| `PostCommand` | 命令执行后 | 占位（预留审计） |

### Skills

| Skill | 命令 | 说明 |
|-------|------|------|
| `cmscoder-login` | `/cmscoder-login` | 触发完整登录流程 |
| `cmscoder-status` | `/cmscoder-status` | 展示当前会话状态 |

## 安装

```bash
# 交互式安装
node bin/cmscoder-init

# 指定运行时
node bin/cmscoder-init --claude --backend-url https://cmscoder.example.com

# 预演（不修改文件）
node bin/cmscoder-init --dry-run --claude --backend-url https://cmscoder.example.com

# 安装所有支持的运行时
node bin/cmscoder-init --all --backend-url https://cmscoder.example.com
```

安装过程：
1. 检测运行环境（Claude Code / OpenCode）
2. 将 lib/ 文件复制到 `~/.cmscoder/plugin/lib/`
3. 将 adapter 文件复制到对应 Agent 配置目录
4. 在 shell profile 中写入 `CMSCODER_PLUGIN_DIR` 环境变量
5. 保存后端 URL 配置
6. 验证服务端连通性

## 数据流

```
┌─────────────────────────────────────────────────────┐
│                    用户交互                          │
│  /cmscoder-login  →  lib/cmscoder.js login          │
│  /cmscoder-status →  lib/cmscoder.js status         │
│  session-start    →  hooks/check-auth.sh            │
└──────────────────────┬──────────────────────────────┘
                       │
          ┌────────────┼────────────┐
          │            │            │
     ┌────▼────┐  ┌────▼─────┐  ┌───▼────┐
     │ auth.js │  │storage.js│  │bootstrap│
     │ callback│  │http-client│  │  .js   │
     └────┬────┘  └────┬─────┘  └───┬────┘
          │            │             │
          └────────────┼─────────────┘
                       │
              ┌────────▼────────┐
              │ cmscoder        │
              │ web-server      │
              │ (:39010, 外部 9010) │
              └────────┬────────┘
                       │
              ┌────────▼────────┐
              │ user-service    │
              │ (IAM, session)  │
              └────────┬────────┘
                       │
              ┌────────▼────────┐
              │ 公司 IAM        │
              │ (OAuth2.0)      │
              └─────────────────┘
```

## 安全约束

1. **不保存 `client_secret`** — 仅 user-service 持有
2. **不直连 IAM** — 所有 IAM 交互通过浏览器和 web-server 代理
3. **login_ticket 一次性消费** — 服务端保证不可重放
4. **本地 token 安全存储** — 优先使用 OS 级安全存储
5. **回环服务仅监听 127.0.0.1** — 不暴露于公网
6. **浏览器回跳不携带 access_token** — 只携带一次性 login_ticket
