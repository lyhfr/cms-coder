# 项目里程碑与研发拆分计划

## 1. 文档目标

基于总纲中的 MVP 和阶段建议，形成 cmscoder 的标准交付拆分结构。

## 2. 关联总纲

- 来源章节：14、15、16、18
- 相关文档：本目录全部专题 spec，执行细化见 [development-plan.md](./development-plan.md)，当前进度见 [completed-features.md](./completed-features.md)

## 3. 交付范围

- MVP 范围定义
- 阶段性里程碑
- 工作流与专题 spec 的对应关系
- 研发、测试、验收的拆分基线
- 基础能力与首个主要功能的优先级划分

## 4. 阶段规划

### 阶段零：基础能力

- 插件端工程骨架、Claude Code / OpenCode 适配公共层
- 服务端工程骨架、配置体系、日志与基础观测
- 本地状态展示基础框架
- 通用错误处理、配置加载、环境区分与基础诊断能力

### 阶段一：首个主要功能 - IAM 登录

- 插件端本地回环服务
- 浏览器唤起与 login session 接入
- 回调接收、login ticket 交换、token 安全存储、登录状态恢复
- 服务端代理登录入口、login session/state 管理、IAM callback、code 换 token
- 用户信息查询、cmscoder 会话签发、登出与刷新

### 阶段二：接入打通

- Claude Code / OpenCode 双适配
- 服务端模型统一入口
- 基础状态展示与基础会话增强

### 阶段三：插件增强

- 工作流增强
- 上下文治理
- 工具前置控制

### 阶段四：企业治理增强

- 配额、限流、审计、追踪
- 模型策略与白名单

### 阶段五：高级 Harness 能力

- 企业工具、流程编排、审批、知识和治理体系扩展

## 5. 工作流拆分

- 架构设计：对应 [../plugin/plugin-architecture.md](../plugin/plugin-architecture.md) 与 [../web-server/server-architecture.md](../web-server/server-architecture.md)
- IAM 登录首个主功能：对应 [../user-service/iam-auth-session.md](../user-service/iam-auth-session.md)
- Agent 适配：对应 [../plugin/claude-code-adapter.md](../plugin/claude-code-adapter.md) 与 [../plugin/opencode-adapter.md](../plugin/opencode-adapter.md)
- 认证后续能力与模型网关：对应 [../shared/model-access-protocol.md](../shared/model-access-protocol.md)
- 增强与治理：对应 [../plugin/plugin-workflow-enhancement.md](../plugin/plugin-workflow-enhancement.md)、[../plugin/tool-governance-permission-control.md](../plugin/tool-governance-permission-control.md)、[../shared/quota-audit-observability.md](../shared/quota-audit-observability.md)

## 6. 首个主要功能开发规划

### 6.1 基础能力完成标准

- 插件端具备通用适配层、配置管理、日志与基础状态展示
- 服务端具备配置、日志、基础追踪和认证模块骨架
- 本地安全存储与服务端配置管理能力可用

### 6.2 IAM 登录功能拆分

#### 插件端

- 实现登录入口与状态展示
- 启动本地临时回环服务并管理端口生命周期
- 唤起系统默认浏览器
- 接收本地回调中的 `login_ticket` 并交换正式 token
- 将 token 写入安全存储
- 实现登录态恢复、刷新失败处理和登出

#### 服务端

- 实现 `<cmscoder-backend>/api/auth/login-sessions` 和浏览器授权入口
- 管理 `login session`、`state` 与 `localPort` 映射
- 实现 `<cmscoder-backend>/api/auth/callback` 回调处理
- 调 `<iam>` 完成 code 换 token 与用户信息查询
- 生成一次性 `login_ticket` 并回跳本地回环地址
- 实现 `<cmscoder-backend>/api/auth/exchange`
- 实现 `<cmscoder-backend>/api/auth/refresh`、`<cmscoder-backend>/api/auth/logout`

#### 联调验收

- 插件端与服务端完成完整登录闭环
- 认证失败、state 校验失败、回环回调失败有可观测错误
- 登录成功后能支撑后续模型访问

## 7. 验收门槛

- 每个阶段开始前，相关专题 spec 应完成并冻结核心边界
- 每个阶段结束时，至少完成功能验收、异常场景验收和运维诊断验收
- 进入下一阶段前，上一阶段的关键依赖和风险要有关闭动作

## 8. 风险与待确认

- 团队规模、迭代节奏和排期依赖尚未提供
- 服务端、插件端是否由不同团队负责待确认
- IAM 联调环境、回调白名单和应用配置开通时序待确认
- 是否需要单独补充测试计划、发布计划和迁移计划待确认

## 9. 验收标准

- 基础能力与首个主要功能的边界和顺序被清晰定义
- IAM 登录已被明确为插件端与服务端基础能力后的首个主功能
- MVP 范围、阶段目标和专题依赖关系可被清晰追踪
- 研发拆分可以直接映射到专题 spec
- 后续新增阶段或专题时可以沿用同一结构扩展
