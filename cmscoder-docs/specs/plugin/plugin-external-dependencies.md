# 插件端外部依赖与交互清单

## 1. 文档目标

明确 cmscoder 插件端与所有外部系统的交互方式、协议、数据流和依赖关系。

## 2. 外部系统总览

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│   系统浏览器      │     │  cmscoder        │     │  操作系统安全存储  │
│   (IAM 登录)      │◄───►│  web-server      │◄───►│  (Keychain 等)    │
│                  │     │  (:39010 / 外部 9010) │     │                  │
└──────────────────┘     └────────┬─────────┘     └──────────────────┘
                                 │
                        ┌────────▼─────────┐
                        │  cmscoder        │
                        │  user-service    │
                        └────────┬─────────┘
                                 │
                        ┌────────▼─────────┐
                        │  公司 IAM         │
                        │  (OAuth 2.0)     │
                        └──────────────────┘
```

## 3. 详细依赖

### 3.1 cmscoder-web-server

**依赖性质**：强依赖，插件端所有服务端能力均通过 web-server 中转

**通信协议**：HTTPS + JSON

**认证方式**：
- 未登录阶段：无认证（`POST /api/auth/login` 无需 token）
- 登录后：`Authorization: Bearer <access_token>`

**交互 API**：

| API | 方法 | 认证 | 调用方 | 请求体 | 响应体 | 错误处理 |
|-----|------|------|--------|--------|--------|---------|
| `/api/auth/login` | POST | 无 | `lib/auth.js` | `{localPort, agentType, pluginInstanceId}` | `{loginId, browserUrl, expiresAt}` | 网络失败 → 提示重试；4xx → 提示配置错误 |
| `/api/auth/exchange` | POST | 无 | `lib/auth.js` | `{loginTicket, pluginInstanceId}` | `{accessToken, refreshToken, expiresIn, user}` | ticket 失效 → 提示重新登录 |
| `/api/auth/refresh` | POST | 无 | `lib/auth.js` | `{refreshToken}` | `{accessToken, refreshToken, expiresIn}` | refresh_token 失效 → 提示重新登录 |
| `/api/auth/logout` | POST | 无 | `lib/auth.js` | `{refreshToken}` | `{}` | 服务端不可达 → 仍清除本地状态 |
| `/api/plugin/bootstrap` | GET | Bearer | `lib/bootstrap.js` | 无 | `{user, defaultModel, featureFlags}` | 不可达 → 使用本地缓存，记录错误 |

**超时配置**：
- 登录相关 API：30 秒
- 刷新 API：10 秒
- Bootstrap API：5 秒

**重试策略**：
- 不自动重试（避免用户感知延迟和重复提交）
- 网络错误时明确提示用户

### 3.2 系统浏览器

**依赖性质**：强依赖，IAM 登录的唯一入口

**交互方式**：通过系统命令打开默认浏览器

| 平台 | 命令 | 用途 |
|------|------|------|
| macOS | `open` | 打开 IAM 授权页面 |
| Linux | `xdg-open` | 打开 IAM 授权页面 |
| Windows | `start` | 打开 IAM 授权页面 |

**浏览器交互流程**：
1. 插件构造 web-server 的授权 URL
2. 调用系统命令打开浏览器
3. 用户在浏览器中完成 IAM 登录
4. IAM 回调到 web-server
5. web-server 签发 login_ticket 后 302 到 `http://127.0.0.1:<port>/callback`
6. 浏览器访问本地回环服务，完成回调

**安全约束**：
- 浏览器 URL 中不包含 access_token 或 refresh_token
- 回调地址固定为 `http://127.0.0.1:<port>/callback`，不接受其他路径
- 不注入任何敏感信息到 URL 参数（仅有 state 用于 CSRF 防护）

### 3.3 操作系统安全存储

**依赖性质**：强依赖，本地 token 安全存储的基础

**后端自动检测**（`lib/storage.js` 的 `_detectBackend()`）：

| 平台 | 后端 | 说明 |
|------|------|------|
| macOS | Keychain | `security` CLI |
| Linux (有 libsecret) | Secret Service | `secret-tool` CLI |
| Windows | AES-256-GCM 加密文件 | Node.js `crypto` 模块（PBKDF2 派生密钥） |
| 其他 (开发环境) | 文件存储 | `~/.cmscoder/.secure-store/*`，权限 600 |

**macOS Keychain**：

| 操作 | 命令 | 参数 |
|------|------|------|
| 存储 | `security add-generic-password` | `-s "cmscoder" -a "<key>" -w "<value>" -U` |
| 读取 | `security find-generic-password` | `-s "cmscoder" -a "<key>" -w` |
| 删除 | `security delete-generic-password` | `-s "cmscoder" -a "<key>"` |

**Linux libsecret**：

| 操作 | 命令 | 参数 |
|------|------|------|
| 存储 | `secret-tool store` | `--label="cmscoder: <key>" cmscoder-key "<key>"` |
| 读取 | `secret-tool lookup` | `cmscoder-key "<key>"` |
| 清除 | `secret-tool clear` | `cmscoder-key "<key>"` |

**Windows AES-256-GCM**：

实现方案（无 PowerShell 依赖，纯 Node.js `crypto` 模块）：
1. 首次运行时生成 32 字节随机主密钥，存储于 `~/.cmscoder/.secure-store/.key`（权限 600）
2. 写入时：PBKDF2 派生密钥 → AES-256-GCM 加密 → `{salt(16) + iv(12) + ciphertext + authTag(16)}` → base64 存储
3. 读取时：base64 解码 → 解析 salt/iv/ciphertext/tag → PBKDF2 重新派生密钥 → AES-256-GCM 解密
4. 密钥派生参数：PBKDF2-SHA256，100,000 次迭代
5. 仅当前机器 + 当前用户可解密（主密钥不跨设备共享）

**Fallback（开发环境）**：

当 Keychain 和 libsecret 均不可用且非 Windows 时，退回到文件存储：
- 存储位置：`~/.cmscoder/.secure-store/`
- 文件权限：600（仅所有者可读写）
- **注意**：此模式仅用于开发环境，生产环境应使用 OS 级安全存储

**存储项清单**：

| Key | 内容 | 敏感度 | TTL |
|-----|------|--------|-----|
| `access_token` | 访问令牌 | 高 | 15 分钟 |
| `refresh_token` | 刷新令牌 | 高 | 7 天（轮换） |
| `user_info` | 用户摘要 (JSON) | 中 | 同 session |
| `session_meta` | 会话元数据 (JSON) | 低 | 同 session |

### 3.4 Node.js 运行时

**依赖性质**：强依赖，所有核心逻辑由 Node.js 实现

**要求**：Node.js ≥ 14（仅使用内置模块：`fs`, `path`, `os`, `crypto`, `http`, `https`, `child_process`, `url`）

**用途**：
- 认证编排（`lib/auth.js`）
- 回环 HTTP 服务器（`lib/callback-server.js`）
- 安全存储与缓存（`lib/storage.js`）
- HTTP API 客户端（`lib/http-client.js`）
- 配置同步（`lib/bootstrap.js`）
- CLI 入口（`lib/cmscoder.js`）

**进程管理**（回调服务器）：
- 由 `auth.js` 通过 `http.createServer()` 启动，监听 `127.0.0.1` 随机端口
- 收到回调后自动 resolve Promise 并关闭服务
- 5 分钟无请求自动退出
- 父进程退出时通过 `process.on('exit')` 清理

### 3.5 Bash

**依赖性质**：弱依赖，仅用于 Claude Code hooks 薄包装

**要求**：系统自带即可（Bash ≥ 3.2，macOS 默认版本即可）

**用途**：
- `hooks/session-start/check-auth.sh` — 调用 `node cmscoder.js ensure-session`
- `hooks/pre-command/` 和 `hooks/post-command/` — 占位脚本，预留扩展

**注意**：所有业务逻辑在 Node.js 中，Bash 仅一行 `node ...` 包装。

## 4. 间接依赖（通过 web-server 中转）

### 4.1 公司 IAM

**依赖性质**：间接依赖，插件端不直接与 IAM 交互

**交互链路**：
```
插件端 → 系统浏览器 → web-server → IAM
                                  ↓
插件端 ← 系统浏览器 ← web-server ← IAM
```

**插件端不持有**：
- IAM `client_id`
- IAM `client_secret`
- IAM `tokenURL`
- IAM `userInfoURL`

**插件端仅持有**：
- web-server 的授权跳转 URL 中的 `authorizeURL` 片段（由服务端构造完整 URL 返回）

### 4.2 cmscoder-user-service

**依赖性质**：间接依赖，通过 web-server 内部调用

**交互链路**：
```
插件端 → web-server → user-service
插件端 ← web-server ← user-service
```

插件端不需要知道 user-service 的地址或 API，所有内部路由由 web-server 处理。

## 5. 安装时依赖

| 依赖 | 用途 | 检测方式 |
|------|------|---------|
| `claude` (CLI) | 检测 Claude Code 是否安装 | `command -v claude` |
| `opencode` (CLI) | 检测 OpenCode 是否安装 | `command -v opencode` |
| Shell profile | 写入环境变量 | 检测 `.zshrc` / `.bashrc` / `.bash_profile` 是否存在 |

## 6. 网络要求

| 方向 | 协议 | 端口 | 说明 |
|------|------|------|------|
| 插件端 → web-server | HTTPS | 443 或自定义 | 所有 API 调用 |
| 插件端 → 浏览器 | 本地命令 | - | `open`/`xdg-open` |
| 浏览器 → web-server | HTTPS | 443 或自定义 | IAM 授权跳转 |
| 浏览器 → 插件端 | HTTP | 127.0.0.1:随机 | 回调 login_ticket |
| 浏览器 → IAM | HTTPS | 443 | 用户认证 |

## 7. 错误诊断

| 错误场景 | 可能原因 | 诊断方式 |
|---------|---------|---------|
| 后端不可达 | 网络不通 / 服务端未部署 / URL 配置错误 | `node -e "require('$CMSCODER_PLUGIN_DIR/lib/http-client.js').get('/api/plugin/bootstrap').then(console.log).catch(console.error)"` |
| 浏览器未打开 | 无图形界面 / `open` 命令不可用 | 检查 `$CMSCODER_BROWSER_CMD` 环境变量 |
| 回调超时 | 浏览器未访问回调地址 / 网络问题 | 检查 `node` 进程是否在运行 |
| Keychain 不可用 | 无图形界面 / 权限问题 | `security find-generic-password -s "cmscoder"` 测试 |
| Token 刷新失败 | refresh_token 过期 / 服务端异常 | 检查 `~/.cmscoder/cache/last_error` |
| Node.js 不可用 | 环境未安装 Node.js | `node --version` 测试 |
