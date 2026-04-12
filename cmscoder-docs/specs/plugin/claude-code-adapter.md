# Claude Code 适配方案

## 1. 文档目标

明确 cmscoder 如何接入 Claude Code，包括配置注入点、模型代理方式、上下文增强与权限治理。本文档为已实现方案的总结。

## 2. 关联总纲

- 来源章节：4.1、6.1、8.1.1、8.1.3、8.1.4、8.1.6、11、14.1、18
- 相关文档：[plugin-architecture.md](./plugin-architecture.md)、[../user-service/iam-auth-session.md](../user-service/iam-auth-session.md)

## 3. 适配实现

### 3.1 配置注入点

Claude Code 使用 `~/.claude/` 目录管理本地配置。cmscoder 通过以下文件注入能力：

| 文件 | 位置 | 用途 |
|------|------|------|
| `CLAUDE.md` | `~/.claude/CLAUDE.md` | 系统提示：企业上下文、研发规范、cmscoder 命令说明 |
| `settings.json` | `~/.claude/settings.json` | hooks 配置：SessionStart/PreCommand/PostCommand 事件绑定 |
| `hooks/session-start/check-auth.sh` | `~/.claude/hooks/session-start/check-auth.sh` | 会话启动时检查认证状态 |
| `hooks/pre-command/check-command.sh` | `~/.claude/hooks/pre-command/check-command.sh` | 命令执行前检查（预留） |
| `hooks/post-command/audit.sh` | `~/.claude/hooks/post-command/audit.sh` | 命令执行后审计（预留） |
| `skills/cmscoder-login/SKILL.md` | `~/.claude/skills/cmscoder-login/SKILL.md` | `/cmscoder-login` 技能定义 |
| `skills/cmscoder-status/SKILL.md` | `~/.claude/skills/cmscoder-status/SKILL.md` | `/cmscoder-status` 技能定义 |

### 3.2 环境变量注入

安装脚本在 shell profile（`.zshrc`/`.bashrc`/`.bash_profile`）中写入：

```bash
export CMSCODER_PLUGIN_DIR="${HOME}/.cmscoder/plugin"
```

所有核心脚本通过 `CMSCODER_PLUGIN_DIR` 定位公共层文件，不依赖固定路径。

### 3.3 模型 endpoint 接管

MVP 阶段通过以下方式实现模型代理：

1. **安装阶段**：`cmscoder-init` 将 cmscoder 后端 URL 写入 `~/.cmscoder/plugin/config/backend_url`
2. **登录后**：`bootstrap_sync()` 从服务端获取默认模型配置，写入 `~/.cmscoder/cache/default_model`
3. **模型调用**：后续通过 Claude Code 的 token helper 或 provider 配置将模型请求重定向到 cmscoder 服务端（具体机制待 Claude Code token helper API 稳定后补充）

### 3.4 会话初始化增强

通过 `hooks/SessionStart` 实现：

```json
{
  "hooks": {
    "SessionStart": [{
      "matcher": ".*",
      "hooks": [
        "node \"$CMSCODER_PLUGIN_DIR/lib/cmscoder.js\" ensure-session || echo '[cmscoder] Session check completed'"
      ]
    }]
  }
}
```

会话启动时执行：
1. 检查 `~/.cmscoder/plugin/` 是否存在
2. 检查 secure-store 中是否有有效的 access_token
3. 如果 token 过期，尝试静默刷新
4. 如果刷新失败或未登录，在终端输出提示信息
5. 如果会话有效，静默继续

### 3.5 Skills 定义

#### `/cmscoder-login`

触发条件：用户输入 `/cmscoder-login`

执行流程：
1. 读取 `~/.cmscoder/plugin/config/backend_url` 获取后端地址
2. 启动回环回调服务器（Node.js `http.createServer()`，127.0.0.1:随机端口）
3. POST `/api/auth/login` 创建登录 session
4. 打开系统浏览器访问返回的 `browserUrl`
5. 等待浏览器回调中的 `login_ticket`
6. POST `/api/auth/exchange` 交换正式 token
7. 将 access_token、refresh_token、user_info 写入 secure-store
8. 调用 `bootstrapSync()` 同步配置
9. 关闭回环服务器

#### `/cmscoder-status`

触发条件：用户输入 `/cmscoder-status`

输出内容：
- 认证状态（已认证 / 未登录 / 会话过期）
- 用户信息（display name、email）
- 租户 ID
- 会话剩余时间
- 默认模型（如已同步）
- 最近错误（如有）

### 3.6 工具治理（预留）

`hooks/pre-command/` 和 `hooks/post-command/` 已创建占位脚本。后续可实现：

- 高风险命令（`rm -rf`、`DROP TABLE` 等）识别和提醒
- 文件写入前的冲突检查
- 命令执行后的结果审计记录

## 4. 外部交互

| 外部系统 | 交互方式 | 协议 | 用途 |
|---------|---------|------|------|
| **系统浏览器** | `open` 命令 | HTTPS | 打开 IAM 登录页面 |
| **macOS Keychain** | `security` CLI | 本地 API | 安全存储 token |
| **Windows AES-256-GCM** | Node.js `crypto` 模块 | 本地 | 安全存储 token（加密文件） |
| **Linux libsecret** | `secret-tool` CLI | 本地 API | 安全存储 token |
| **cmscoder-web-server** | Node.js `http/https` | HTTP/JSON | 认证、配置同步 |

## 5. 风险与待确认

- Claude Code 的 token helper 机制是否支持动态 endpoint 配置
- hooks 的 PreCommand/PostCommand 是否能在所有命令（包括 Bash 工具）上触发
- Claude Code 版本升级是否会影响 hooks 和 skills 的兼容性
- 是否需要支持 Claude Code 的权限模式（`--dangerously-skip-permissions` vs 细粒度权限）

## 6. 验收标准

- 用户在 Claude Code 中可通过 `/cmscoder-login` 完成企业 SSO 登录
- 用户无须手动配置任何模型厂商 API Key
- 会话启动时自动检查登录态并给出明确提示
- `/cmscoder-status` 可查看完整的会话状态
- 系统提示中包含企业研发规范和安全提醒
- hooks 和 skills 可独立升级，不影响 Claude Code 核心功能
