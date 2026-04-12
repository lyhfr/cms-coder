# 插件端工作流增强设计

## 1. 文档目标

定义 cmscoder 插件端如何增强规划、设计、编码、测试、Review 等研发流程，以及当前的实现状态和扩展规划。

## 2. 关联总纲

- 来源章节：8.1.4、8.1.5、8.1.7、10.2、10.3、14.1、15、18
- 相关文档：[plugin-architecture.md](./plugin-architecture.md)、[claude-code-adapter.md](./claude-code-adapter.md)

## 3. 当前实现状态

### 3.1 已实现

| 能力 | 实现方式 | 状态 |
|------|---------|------|
| 系统提示注入 | `~/.claude/CLAUDE.md` | 已完成 |
| 企业规范注入 | CLAUDE.md 中的 Development Standards 和 Security Reminders | 已完成 |
| 会话启动增强 | `hooks/SessionStart` → `check-auth.sh` | 已完成 |
| 状态查看 Skill | `skills/cmscoder-status/SKILL.md` | 已完成 |
| 登录 Skill | `skills/cmscoder-login/SKILL.md` | 已完成 |
| 命令前置检查占位 | `hooks/pre-command/check-command.sh` | 占位脚本 |
| 命令后置审计占位 | `hooks/post-command/audit.sh` | 占位脚本 |

### 3.2 CLAUDE.md 注入内容

当前 `CLAUDE.md` 包含：
- cmscoder 简介和快速命令列表（`/cmscoder-login`、`/cmscoder-status`、`/cmscoder-logout`）
- 会话管理说明（自动检查、静默刷新、过期处理）
- 企业研发规范（YAGNI、DRY、KISS、Test-first、增量提交）
- 安全提醒（不提交密钥、使用环境变量、OWASP Top 10）

## 4. 工作流增强架构

### 4.1 会话启动增强

```
Claude Code 启动
    │
    └── SessionStart hook
         │
         └── node "$CMSCODER_PLUGIN_DIR/lib/cmscoder.js" ensure-session
              ├── 检查 secureStore access_token
              ├── 检查 localCache session_meta.expiresAt
              └── 如过期 → 静默刷新
              │
              ├── 有效 → 静默继续
              │
              └── 无效 → 输出提示信息
                     "cmscoder: Not authenticated. Run /cmscoder-login to sign in."
```

### 4.2 系统提示增强

CLAUDE.md 作为持久化系统提示，在每次会话中自动加载。内容包括：

1. **身份标识** — 说明当前运行在企业 cmscoder 环境下
2. **可用命令** — 列出用户可执行的 cmscoder 操作
3. **研发规范** — YAGNI、DRY、KISS 等工程原则
4. **安全约束** — 不提交密钥、审查生成代码

### 4.3 Skills 机制

Skills 是用户可主动触发的增强能力。当前实现的 Skills：

| Skill | 触发命令 | 用途 |
|-------|---------|------|
| `cmscoder-login` | `/cmscoder-login` | 企业 SSO 登录 |
| `cmscoder-status` | `/cmscoder-status` | 查看会话状态 |

后续可扩展的 Skills：

| Skill | 触发命令 | 用途 |
|-------|---------|------|
| `cmscoder-logout` | `/cmscoder-logout` | 注销并清理会话 |
| `cmscoder-config` | `/cmscoder-config` | 查看/修改配置 |
| `cmscoder-models` | `/cmscoder-models` | 查看可用模型列表 |
| `cmscoder-quota` | `/cmscoder-quota` | 查看额度使用情况 |

## 5. 上下文治理

### 5.1 当前策略

- CLAUDE.md 作为固定系统提示，不随会话增长而膨胀
- 会话状态通过外部文件（cache）查询，不注入到对话上下文
- Skills 按需触发，不自动注入到每个对话

### 5.2 后续治理策略（待实现）

| 问题 | 治理方案 |
|------|---------|
| 长会话上下文漂移 | 会话启动时从外部文件读取最新状态，不依赖历史对话 |
| 项目规范注入 | 从服务端 bootstrap 同步，按需按项目加载不同规范 |
| 上下文压缩 | 定期将对话中的关键决策点写入外部文件，下次会话时加载摘要 |
| 提示污染控制 | 区分系统级提示（CLAUDE.md）和项目级提示（项目根目录 CLAUDE.md），避免重复 |

## 6. 研发流程增强（规划）

### 6.1 阶段引导

通过 PreCommand/PostCommand hooks 实现研发阶段识别和引导：

```
用户输入/执行命令
    │
    └── PreCommand hook
         │
         ├── 分析当前 git 分支和最近提交
         ├── 识别研发阶段（规划/编码/测试/review）
         │      └── 基于分支命名、最近文件变更、测试执行情况
         │
         └── 如检测到阶段转换，给出引导提示
                "检测到进入测试阶段，建议：
                 1. 先编写测试用例
                 2. 运行已有测试确保未破坏
                 3. 检查测试覆盖率"
```

### 6.2 企业技能库

后续可按企业需求扩展技能集：

| 技能集 | 内容 | 装配方式 |
|--------|------|---------|
| 代码规范 | 命名规范、注释规范、提交规范 | 按项目配置 |
| 架构约束 | 分层规范、依赖注入规范 | 按团队配置 |
| 安全规范 | 安全编码规范、输入校验规范 | 全局配置 |
| 测试规范 | 测试命名、测试组织、Mock 规范 | 按项目配置 |

### 6.3 项目规范注入

从服务端 bootstrap 同步项目级规范：

```
GET /api/plugin/bootstrap
    │
    └── 响应包含:
         ├── defaultModel
         ├── featureFlags
         └── projectSpecs (待扩展)
              ├── codingStandards
              ├── architectureRules
              └── securityGuidelines
```

## 7. 配置与装配

### 7.1 配置来源

| 配置项 | 来源 | 优先级 |
|--------|------|--------|
| 后端 URL | `~/.cmscoder/plugin/config/backend_url` | 1 |
| 默认模型 | 服务端 bootstrap 同步 | 2 |
| 功能开关 | 服务端 bootstrap 同步 | 2 |
| 项目规范 | 服务端 bootstrap 同步（待扩展） | 3 |
| 本地覆盖 | `~/.cmscoder/config.local.json`（待实现） | 1 |

### 7.2 功能开关

通过服务端返回的 `featureFlags` 控制插件功能：

```json
{
  "featureFlags": {
    "loginEnabled": true,
    "statusEnabled": true,
    "workflowEnhancement": true,
    "toolGovernance": false,
    "contextGovernance": false
  }
}
```

插件端根据开关决定是否展示对应功能和菜单。

## 8. 非功能要求

- 增强能力对用户可感知但不过度打扰（session-start hook 仅在未登录时提示）
- 具备版本化与灰度发布能力（通过 featureFlags 控制）
- 长会话表现稳定，通过外部文件查询状态而非依赖对话历史
- 企业规范、项目规范和技能可按配置装配

## 9. 风险与待确认

- 研发流程增强的强约束与弱引导边界待明确（是提醒还是阻断？）
- 项目知识和团队规范的来源、同步方式待明确（从服务端同步还是本地配置？）
- 不同 Agent 是否支持统一的技能与命令扩展机制待验证（Claude Code skills vs OpenCode commands）

## 10. 验收标准

- 插件端能够在会话启动时提供增强能力（当前：认证检查）
- 企业规范、项目规范可通过 CLAUDE.md 注入（当前：基础规范）
- Skills 可按用户指令触发（当前：login、status）
- 上下文污染、提示漂移有明确治理手段（当前：外部查询，不注入对话历史）
