# cmscoder Spec 结构

按工程项目结构组织，分为以下目录：

| 目录 | 说明 |
|------|------|
| [`shared/`](./shared/) | 跨服务/跨端的公共规范（总纲、认证、模型协议、审计等） |
| [`plugin/`](./plugin/) | 插件端规范（架构、Claude Code/OpenCode 适配、工作流、工具治理等） |
| [`web-server/`](./web-server/) | web-server 服务规范（架构设计） |
| [`user-service/`](./user-service/) | user-service 服务规范 |
| [`project/`](./project/) | 项目管理文档（里程碑、开发计划、完成功能追踪） |

## 文件索引

### shared/

| 文件 | 说明 |
|------|------|
| `cmscoder-overview.md` | 系统总览、边界、目标、架构总图 |
| `model-access-protocol.md` | 模型接入与协议转换设计 |
| `quota-audit-observability.md` | 配额、审计与可观测设计 |

### plugin/

| 文件 | 说明 |
|------|------|
| `plugin-architecture.md` | 插件端总体架构 |
| `claude-code-adapter.md` | Claude Code 适配方案 |
| `opencode-adapter.md` | OpenCode 适配方案 |
| `plugin-workflow-enhancement.md` | 插件端工作流增强设计 |
| `tool-governance-permission-control.md` | 工具治理与权限控制设计 |
| `plugin-external-dependencies.md` | 插件端外部依赖与交互清单 |

### web-server/

| 文件 | 说明 |
|------|------|
| `server-architecture.md` | 服务端总体架构（web-server + user-service 分工） |

### user-service/

| 文件 | 说明 |
|------|------|
| `iam-auth-session.md` | IAM 认证与会话管理设计（OAuth 2.0 授权码 + 服务端 Proxy Callback） |

### project/

| 文件 | 说明 |
|------|------|
| `roadmap-and-delivery-plan.md` | 项目里程碑与研发拆分计划 |
| `development-plan.md` | 整体开发计划 |
| `completed-features.md` | 已完成功能与当前状态 |
