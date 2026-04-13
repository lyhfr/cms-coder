# 插件端总体架构设计

## 1. 文档目标

定义 cmscoder 插件端的模块划分、职责边界、扩展点和关键运行流程。本文档为已实现方案的总结，不再是占位设计。

## 2. 关联总纲

- 来源章节：7、8.1、9.1、10、11、12、13、14
- 相关文档：[../shared/cmscoder-overview.md](../shared/cmscoder-overview.md)、[../user-service/iam-auth-session.md](../user-service/iam-auth-session.md)、[../project/roadmap-and-delivery-plan.md](../project/roadmap-and-delivery-plan.md)

## 3. 设计范围

### In Scope

- Agent 适配层（Claude Code / OpenCode）
- 登录与本地会话层（auth.js, cmscoder.js）
- 模型接入配置层（bootstrap-sync、endpoint 注入）
- 会话初始化增强层（session-start hook）
- 本地安全存储（storage.js — secureStore + localCache）
- 回环回调服务（callback-server.js）
- 安装与初始化（cmscoder-init）

### Out of Scope

- 服务端配额、审计、模型路由主逻辑
- 上游模型协议转换主逻辑
- 工作流增强和工具治理的具体实现（预留扩展点，MVP 阶段为占位）

## 4. 关键场景

| 场景 | 触发方式 | 涉及模块 |
|------|---------|---------|
| 首次安装 | 运行 `cmscoder-init` | `bin/cmscoder-init` |
| 首次登录 | `/cmscoder-login` | `lib/auth.js` → `lib/callback-server.js` → `lib/http-client.js` → web-server |
| 会话恢复 | Claude Code 启动 | `hooks/session-start/check-auth.sh` → `lib/cmscoder.js ensure-session` |
| 静默续期 | token 过期 / 401 响应 | `lib/auth.js` → `lib/http-client.js` → web-server |
| 状态查看 | `/cmscoder-status` | `lib/cmscoder.js status` → `lib/storage.js` |
| 用户注销 | `/cmscoder-logout` | `lib/auth.js` → web-server → `lib/storage.js` |
| 配置同步 | 登录成功后 / 定期 | `lib/bootstrap.js` → web-server |

## 5. 架构实现

### 5.1 目录结构

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

### 5.2 模块划分与职责

#### 5.2.1 CLI 入口 (lib/cmscoder.js)

| 模块 | 职责 | 主要函数/子命令 |
|------|------|---------|
| `cmscoder.js` | CLI 入口 + 模块导出 | `login`, `logout`, `refresh`, `status`, `token`, `ensure-session`, **model-token** |

直接运行时作为命令行工具，`require()` 时作为 API 模块。

#### 5.2.1.1 Model Token 动态获取机制（apiKeyHelper 模式）

**背景**：利用 Claude Code 原生 `apiKeyHelper` 机制，实现无静态 API Key 的模型认证。Model Token 为短期 JWT（默认 5分钟，可配置），通过 HMAC 签名请求获取。

**核心流程**：

```
Claude Code 调用 apiKeyHelper
        │
        ▼
┌─ cmscoder.js model-token ──────────────────────┐
│                                                  │
│  1. 从安全存储读取 access_token                   │
│  2. 从安全存储读取 plugin_secret                  │
│                                                  │
│  3. access_token 是否有效？                       │
│     ├─ 有效 → 继续                                │
│     └─ 过期 → 用 refresh_token 静默刷新           │
│               ├─ 成功 → 继续                      │
│               └─ 失败 → stderr 提示重新登录       │
│                                                  │
│  4. 构造签名请求：                                 │
│     timestamp = now()                            │
│     nonce = random()                             │
│     signature = HMAC_SHA256(                     │
│       access_token + timestamp + nonce,          │
│       plugin_secret                              │
│     )                                            │
│                                                  │
│  5. POST /api/auth/model-token                   │
│     Body: { accessToken, timestamp, nonce,       │
│             signature, pluginInstanceId }         │
│                                                  │
│  6. stdout → model_token（给 Claude Code 使用）   │
└──────────────────────────────────────────────────┘
```

**Claude Code 配置**（`~/.claude/settings.json`）：

```json
{
  "apiKeyHelper": "node ~/.cmscoder/plugin/lib/cmscoder.js model-token",
  "env": {
    "ANTHROPIC_BASE_URL": "https://cmscoder.company.com/api/model/v1",
    "CLAUDE_CODE_API_KEY_HELPER_TTL_MS": "240000"
  }
}
```

**配置说明**：
- `apiKeyHelper`：Claude Code 每次需要 API Key 时执行的命令
- `CLAUDE_CODE_API_KEY_HELPER_TTL_MS`：调用间隔（建议略小于 modelTokenTTL，如 4分钟）
- 命令输出（stdout）作为 API Key 使用
- 命令错误（stderr + exit 非 0）会提示用户重新登录

**会话状态处理**：

| 场景 | 处理流程 |
|------|---------|
| **首次请求（刚登录）** | access_token 有效 → 直接构造签名请求获取 model_token |
| **长时间未对话（access_token 过期）** | apiKeyHelper 检测到过期 → 用 refresh_token 静默刷新 → 成功后获取 model_token |
| **Session 已过期（refresh_token 过期）** | 静默刷新失败 → apiKeyHelper 输出错误到 stderr，exit 非 0 → Claude Code 提示重新登录 |

**安全特性**：

| 机制 | 说明 |
|------|------|
| **短期 Token** | 默认 5分钟有效期，泄露窗口极短 |
| **HMAC 签名** | 需要 plugin_secret 构造有效请求 |
| **Timestamp + Nonce** | 防重放攻击 |
| **IP 绑定（可选）** | 服务端配置启用，防止 Token 被带走使用 |

#### 5.2.2 认证层 (lib/auth.js)

| 模块 | 职责 | 主要函数 |
|------|------|---------|
| `auth.js` | 登录/登出/刷新完整流程编排 | `login()`, `logout()`, `refreshSilent()`, `ensureSession()`, `getAccessToken()`, `status()` |

#### 5.2.3 存储层 (lib/storage.js)

| 模块 | 职责 | 主要函数 |
|------|------|---------|
| `storage.js` | 跨平台安全凭证存储 + 本地缓存 | `secureStore.set/get/delete/clearAll()`, `localCache.set/get/getJson/has/delete/clear/setSessionMeta/getSessionMeta/sessionValid/setServerConfig/getServerConfig` |

**secureStore** — 安全存储（自动检测后端）:

| 平台 | 后端 | 说明 |
|------|------|------|
| macOS | Keychain | `security` CLI |
| Linux (有 libsecret) | Secret Service | `secret-tool` CLI |
| Windows | AES-256-GCM 加密文件 | Node.js `crypto` 模块 |
| 其他 (开发环境) | 文件 | `~/.cmscoder/.secure-store/*` (权限 600) |

**localCache** — 文件系统 JSON 缓存（`~/.cmscoder/cache/`）:
- `session_meta` — 用户摘要、过期时间
- `server_config` — 后端 URL、默认模型、功能开关
- `bootstrap_data` — 服务端 bootstrap 响应原文
- `default_model` — 默认模型 ID
- `model_endpoint` — 模型端点 URL（bootstrap 时从 backendUrl 推导）
- `last_error` — 最近一次错误

**secureStore** — 安全存储的键:
- `access_token` — 会话访问凭证（= sessionId）
- `refresh_token` — 刷新凭证
- `user_info` — 用户信息 JSON
- `plugin_secret` — HMAC 签名密钥，用于获取 model_token

#### 5.2.4 回环服务层 (lib/callback-server.js)

| 模块 | 职责 | 说明 |
|------|------|------|
| `callback-server.js` | 轻量 Node.js HTTP 服务，接收 login_ticket | 监听 `127.0.0.1` 随机端口，5 分钟超时自动退出 |

#### 5.2.5 OpenCode 适配（本地代理模式）

OpenCode 不支持 `apiKeyHelper`，采用**本地代理模式**：

**架构**：
```
OpenCode ──▶ localhost:8080 ──▶ cmscoder-web-server
            (本地代理)           (模型端点)
```

**本地代理实现**（`lib/model-proxy.js`）：

| 功能 | 说明 |
|------|------|
| 自动获取 Model Token | 启动时调用 `model-token` 命令获取 |
| Token 缓存与刷新 | 缓存 Token，过期前自动刷新 |
| 请求转发 | 添加 Authorization Header 后转发到服务端 |
| 会话恢复 | Token 获取失败时提示用户重新登录 |
| **来源验证** | **校验请求是否来自 OpenCode 进程** |

**安全防护机制（跨平台）**：

本地代理监听 `127.0.0.1`，但仅靠本地监听不足以防止攻击（用户可用 curl 直接请求）。增加以下防护：

| 防护层 | 机制 | 跨平台支持 |
|--------|------|-----------|
| **1. 动态密钥** | 启动时生成随机密钥，请求需携带 | ✅ 全平台 |
| **2. 来源 IP 限制** | 仅接受 127.0.0.1/::1 连接 | ✅ 全平台 |
| **3. Named Pipe（Windows）/UDS（Unix）** | 替代 TCP，文件权限控制 | ✅ Windows Named Pipe<br>✅ macOS/Linux UDS |

**推荐方案：动态密钥 + 平台最优传输**

| 平台 | 传输方式 | 安全机制 |
|------|---------|---------|
| **macOS** | Unix Domain Socket | 文件权限 600 + 动态密钥 |
| **Linux** | Unix Domain Socket | 文件权限 600 + 动态密钥 |
| **Windows** | Named Pipe | ACL 限制当前用户 + 动态密钥 |

**实现逻辑**：

```javascript
// 本地代理启动时
const proxySecret = generateRandomKey(); // 32字节随机

if (process.platform === 'win32') {
  // Windows: Named Pipe
  const pipePath = `\\\\.\\pipe\\cmscoder-model-proxy-${uid}`;
  server.listen(pipePath);
  // Named Pipe 自动继承进程 ACL，仅当前用户可访问
} else {
  // macOS/Linux: Unix Domain Socket
  const socketPath = `${os.homedir()}/.cmscoder/model-proxy.sock`;
  server.listen(socketPath);
  fs.chmodSync(socketPath, 0o600); // 仅当前用户可访问
}

// 将传输地址和密钥写入 OpenCode 配置
updateOpenCodeConfig({
  baseURL: getPlatformBaseURL(),  // 根据平台生成
  apiKey: proxySecret             // 每个请求需携带
});
```

**OpenCode 配置（自动写入）**：

| 平台 | 配置示例 |
|------|---------|
| macOS/Linux | `{"baseURL": "http+unix://%2FUsers%2Fuser%2F.cmscoder%2Fmodel-proxy.sock", "apiKey": "secret"}` |
| Windows | `{"baseURL": "http://localhost:8080", "apiKey": "secret"}` + 仅 127.0.0.1 |

**请求校验流程**：
```
1. 接收请求
2. 校验来源 IP 为 127.0.0.1/::1（Windows TCP 模式）
   或 Connection 来自 UDS/Named Pipe（Unix/Windows 命名管道模式）
3. 校验 X-Proxy-Secret Header 是否匹配
4. 校验 User-Agent 是否包含 "OpenCode"
5. 全部通过 → 添加 Model Token 后转发到服务端
```

**CLI 命令**：
```bash
# 启动本地代理（自动选择平台最优传输）
node ~/.cmscoder/plugin/lib/cmscoder.js model-proxy

# 停止本地代理
node ~/.cmscoder/plugin/lib/cmscoder.js model-proxy stop

# 查看状态
node ~/.cmscoder/plugin/lib/cmscoder.js model-proxy status
```

**OpenCode 配置**（自动写入 `~/.opencode/config.json`）：
```json
{
  "provider": {
    "cmscoder-local": {
      "baseURL": "http+unix://%2FUsers%2Fuser%2F.cmscoder%2Fmodel-proxy.sock",
      "apiKey": "random_proxy_secret_32chars",
      "models": {
        "gpt-4": { "id": "gpt-4" }
      }
    }
  },
  "model": "cmscoder-local/gpt-4"
}
```

#### 5.2.6 HTTP 客户端 (lib/http-client.js)

| 模块 | 职责 | 主要函数 |
|------|------|---------|
| `http-client.js` | HTTP API 调用封装（Node.js 内置 http/https） | `post(path, body, token?)`, `get(path, token?)` |

#### 5.2.6 配置同步层 (lib/bootstrap.js)

| 模块 | 职责 | 主要函数 |
|------|------|---------|
| `bootstrap.js` | 服务端配置同步 | `sync(accessToken)`, `getCached()` |

#### 5.2.7 Agent 适配层 (adapters)

| 模块 | 职责 | 文件 |
|------|------|------|
| Claude Code | 系统提示、hooks、skills | `CLAUDE.md`, `settings.json`, `hooks/`, `skills/` |
| OpenCode | skills 定义（hooks 待 OpenCode 版本确认） | `skills/`, `README.md` |

### 5.3 安装与部署

安装脚本 `bin/cmscoder-init`（Node.js）负责：

1. 检测运行环境（Claude Code / OpenCode 是否在 PATH 中）
2. 将 `lib/` 文件复制到 `~/.cmscoder/plugin/lib/`
3. 将 adapter 文件复制到对应 Agent 配置目录
4. 在 shell profile（`.zshrc`/`.bashrc`/`.bash_profile`）中写入 `CMSCODER_PLUGIN_DIR` 环境变量
5. 保存后端 URL 配置到 `~/.cmscoder/plugin/config/backend_url`
6. 验证服务端连通性（`GET /api/plugin/bootstrap`）

**安装参数**：

| 参数 | 说明 |
|------|------|
| `--claude` | 仅安装 Claude Code 适配 |
| `--opencode` | 仅安装 OpenCode 适配 |
| `--all` | 安装所有支持的运行时 |
| `--dry-run` | 预演模式，不修改文件 |
| `--backend-url URL` | 指定 cmscoder 后端地址 |

## 6. 外部依赖

### 6.1 运行时依赖

| 依赖 | 用途 | 最低版本 | 安装方式 |
|------|------|---------|---------|
| **cmscoder-web-server** | 认证 API、配置同步、模型访问 | - | 服务端部署 |
| **系统浏览器** | IAM 登录页面展示 | - | macOS: `open`, Linux: `xdg-open`, Windows: `start` |
| **Node.js** | 所有核心逻辑运行 | ≥14 | 系统安装，零 npm 依赖 |
| **macOS Keychain** | 安全存储 token | - | macOS 系统自带 |
| **Windows AES-256-GCM** | 安全存储 token | Node.js 内置 `crypto` | Node.js 自带 |
| **Bash** | Claude Code hooks 薄包装 | 系统自带 | 系统自带 |

### 6.2 服务端 API 依赖

| API | 方法 | 调用方 | 用途 |
|-----|------|--------|------|
| `POST /api/auth/login` | POST | `lib/auth.js` | 创建登录 session，获取 browserUrl |
| `GET /api/auth/login/{loginId}/authorize` | GET | 浏览器 | 重定向到 IAM 授权页 |
| `GET /api/auth/iam/callback` | GET | 浏览器（IAM 回调） | 接收 IAM 授权码，转发给 user-service |
| `POST /api/auth/exchange` | POST | `lib/auth.js` | 用 login_ticket 交换正式会话凭证 + Composite Token |
| `POST /api/auth/refresh` | POST | `lib/auth.js` | 刷新 access_token |
| `POST /api/auth/logout` | POST | `lib/auth.js` | 注销会话 |
| `GET /api/auth/me` | GET | 需要 auth 中间件 | 获取当前用户信息 |
| `GET /api/plugin/bootstrap` | GET | `lib/bootstrap.js` | 同步服务端配置（模型列表、功能开关） |

### 6.3 数据流

```
用户 (Claude Code / OpenCode)
    │
    ▼
┌─────────────────────────────────┐
│  cmscoder-plugin                │
│                                 │
│  hooks/skills                   │
│    │                            │
│    ▼                            │
│  lib/cmscoder.js ───────┐      │
│    │                    │      │
│    ▼                    ▼      │
│  lib/auth.js       lib/storage.js
│  lib/callback-server  lib/bootstrap.js
│  lib/http-client              │
│                                 │
└──────────┬──────────────────────┘
           │ HTTPS
           ▼
┌─────────────────────────────────┐
│  cmscoder-web-server (:39010, k8s 外部 9010) │
│    → user-service (内部)        │
│      → 公司 IAM (OAuth 2.0)     │
└─────────────────────────────────┘
```

## 7. 安全约束

| 约束 | 实现方式 |
|------|---------|
| 不保存 IAM `client_secret` | 插件端代码中不包含任何密钥，`client_secret` 仅存在于 user-service 配置中 |
| 不直连 IAM | 所有 IAM 交互通过系统浏览器和 web-server 代理完成 |
| 浏览器回跳不携带 access_token | 回调地址仅携带一次性 `login_ticket`，由后端交换 |
| 本地 token 安全存储 | macOS 使用 Keychain，Windows 使用 AES-256-GCM 加密文件，Linux 使用 libsecret，开发环境 fallback 到权限 600 的文件 |
| 回环服务仅监听 127.0.0.1 | `callback-server.js` 绑定 `127.0.0.1`，不监听 `0.0.0.0` |
| login_ticket 一次性消费 | 服务端保证 login_ticket 消费后立即失效（内存标记，生产用 Redis） |
| Composite Token 防泄露 | 插件端使用 `cmscoderv1_` 复合凭证调用模型端点，即使 modelApiKey 或 accessToken 单独泄露，无法用于模型端点 |

## 8. 扩展点

| 扩展点 | 当前状态 | 说明 |
|--------|---------|------|
| `hooks/pre-command/` | 占位 | 工具治理：命令执行前策略检查 |
| `hooks/post-command/` | 占位 | 审计：命令执行后结果记录 |
| `adapters/opencode/hooks/` | 待实现 | OpenCode 版本确认后的 hooks 适配 |
| `adapters/claude-code/skills/` | 已有 login/status | 后续可扩展 cmscoder-logout, cmscoder-config 等 |
| `lib/bootstrap.js` | 已有 sync/getCached | 后续可扩展模型列表查询、额度查询 |

## 9. 非功能要求

- 本地敏感数据安全存储（Keychain / libsecret）
- 模块可替换，新增 Agent 不影响公共能力层（Shared Core + Adapters 架构）
- 启动链路可诊断（session-start hook 自动检查 + 明确错误提示）
- 回环服务 5 分钟超时自动退出，不残留后台进程
- 刷新失败时不自动无限重试，明确提示用户重新登录

## 10. 验收标准

- 用户可通过 `cmscoder-init` 一键完成安装
- Claude Code 中可通过 `/cmscoder-login` 触发企业 SSO 登录
- 登录成功后 token 安全存储于操作系统 Keychain
- Claude Code 启动时自动检查登录态并提示
- `/cmscoder-status` 可展示用户、租户、会话剩余时间
- 浏览器回跳链路不直接暴露 cmscoder access token
- 回环服务仅监听 `127.0.0.1`，不暴露于公网
- 注销后本地与服务端会话均可清理
- OpenCode adapter 骨架已就绪，skills 定义与 Claude Code 一致
