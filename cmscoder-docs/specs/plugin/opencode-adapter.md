# OpenCode 适配方案

## 1. 文档目标

明确 cmscoder 如何接入 OpenCode，包括配置注入点、模型代理方式、工作流增强与状态回传。本文档为已实现方案的骨架，具体 hooks 机制待 OpenCode 版本确认。

## 2. 关联总纲

- 来源章节：4.1、6.1、8.1.1、8.1.3、8.1.4、8.1.6、11、14.1、18
- 相关文档：[plugin-architecture.md](./plugin-architecture.md)、[../user-service/iam-auth-session.md](../user-service/iam-auth-session.md)

## 3. 适配实现

### 3.1 当前状态

OpenCode 适配器已完成骨架搭建，与 Claude Code 共享全部核心逻辑（auth、storage、loopback、bootstrap），仅在 adapter 层保留差异。

**已实现**：
- `adapters/opencode/skills/cmscoder-login/SKILL.md` — 登录技能定义
- `adapters/opencode/skills/cmscoder-status/SKILL.md` — 状态查看技能定义
- `adapters/opencode/README.md` — 适配说明

**待实现**（需 OpenCode 版本确认）：
- hooks 配置方式和目录结构
- 系统提示注入方式
- 命令注册机制（skills vs commands vs 其他）
- 配置写入点

### 3.2 共享核心

OpenCode 与 Claude Code 完全共享以下模块：

| 模块 | 路径 | 说明 |
|------|------|------|
| `cmscoder.js` | `lib/` | CLI 入口 + 模块导出（login/logout/refresh/status/ensure-session） |
| `auth.js` | `lib/` | 认证编排（login/logout/refresh/ensureSession/getAccessToken） |
| `storage.js` | `lib/` | 安全存储 + 本地缓存（secureStore + localCache） |
| `callback-server.js` | `lib/` | 回环 HTTP 服务（Node.js http.createServer） |
| `http-client.js` | `lib/` | API 调用封装（POST/GET） |
| `bootstrap.js` | `lib/` | 服务端配置同步 |

这意味着 OpenCode 的登录流程、token 管理、安全存储、配置同步与 Claude Code **完全一致**，差异仅在触发方式和配置写入位置。

### 3.3 Skills 定义

#### `/cmscoder-login`

与 Claude Code 一致。SKILL.md 中指引用户：
1. 执行 `node "$CMSCODER_PLUGIN_DIR/lib/cmscoder.js" login`

#### `/cmscoder-status`

与 Claude Code 一致。SKILL.md 中指引用户：
1. 执行 `node "$CMSCODER_PLUGIN_DIR/lib/cmscoder.js" status`

### 3.4 模型 endpoint 接管

与 Claude Code 一致：
1. 后端 URL 存储在 `~/.cmscoder/cache/server_config`（由 bootstrap sync 写入）
2. 登录后通过 `bootstrapSync()` 获取默认模型
3. 模型调用通过 OpenCode 的 provider 配置重定向到 cmscoder 服务端（具体机制待确认）

### 3.5 会话初始化增强

待 OpenCode hooks 机制确认后的实现方案：

```
OpenCode 启动
    │
    └── [hook: session-start]  (待确认触发机制)
         │
         └── node "$CMSCODER_PLUGIN_DIR/lib/cmscoder.js" ensure-session
              ├── 检查 secureStore → access_token
              ├── 检查 localCache → session_meta.expiresAt
              └── 如过期 → 静默刷新
```

## 4. 外部交互

| 外部系统 | 交互方式 | 协议 | 用途 |
|---------|---------|------|------|
| **系统浏览器** | `open`/`xdg-open` | HTTPS | 打开 IAM 登录页面 |
| **macOS Keychain / Windows AES-256-GCM / Linux libsecret** | `security` / Node.js `crypto` / `secret-tool` | 本地 API | 安全存储 token |
| **cmscoder-web-server** | Node.js `http/https` 模块 | HTTP/JSON | 认证、配置同步 |
| **OpenCode 运行时** | (待确认) | (待确认) | hooks 注册、skills 加载、配置写入 |

## 5. 待确认事项

| 事项 | 说明 | 影响范围 |
|------|------|---------|
| OpenCode 配置目录 | 是否在 `~/.config/opencode/` 或其他位置 | 安装路径 |
| hooks 机制 | 是否支持类似 Claude Code 的 SessionStart/PreCommand/PostCommand | 会话检查 |
| skills/commands | 使用何种机制注册 `/cmscoder-login` 等命令 | 用户交互 |
| 系统提示注入 | 是否支持类似 CLAUDE.md 的持久化系统提示 | 企业规范注入 |
| provider 配置 | 如何修改模型 endpoint 和认证方式 | 模型代理 |
| 版本兼容范围 | 哪些 OpenCode 版本支持所需的扩展能力 | 兼容性 |
| 权限模型 | OpenCode 是否有类似 Claude Code 的权限审批机制 | 工具治理 |

## 6. 实现计划

### Phase 1（已完成）

- [x] 共享核心层实现（auth.js, storage.js, callback-server.js, http-client.js, bootstrap.js, cmscoder.js）
- [x] OpenCode adapter 目录结构
- [x] Skills 定义（cmscoder-login、cmscoder-status）
- [x] 安装脚本中的 OpenCode 支持（`--opencode` 参数）

### Phase 2（待 OpenCode 版本确认）

- [ ] OpenCode hooks 注册和配置写入
- [ ] 系统提示注入方式
- [ ] 模型 provider 配置重定向
- [ ] 完整安装流程测试

## 7. 验收标准

- 用户在 OpenCode 中可直接使用企业登录能力
- OpenCode 的模型请求能够统一走 cmscoder 服务端
- 工作流增强、工具治理和状态展示具备最小可用闭环
- OpenCode 与 Claude Code 共享同一套认证核心，差异仅在 adapter 层
- 新增 OpenCode 版本适配不影响已有的 Claude Code 适配
