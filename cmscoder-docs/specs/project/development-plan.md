# 整体开发计划

## 1. 文档目标

在总纲和里程碑文档的基础上，形成面向研发执行的整体开发计划，明确功能拆分、先后顺序、依赖关系和验收边界。

## 2. 关联文档

- 总纲：[../shared/cmscoder-overview.md](../shared/cmscoder-overview.md)
- 架构设计：[../plugin/plugin-architecture.md](../plugin/plugin-architecture.md)、[../web-server/server-architecture.md](../web-server/server-architecture.md)
- 认证设计：[../user-service/iam-auth-session.md](../user-service/iam-auth-session.md)
- 里程碑规划：[./roadmap-and-delivery-plan.md](./roadmap-and-delivery-plan.md)
- 已完成功能：[./completed-features.md](./completed-features.md)

## 3. 开发原则

- 先完成基础能力，再进入首个端到端主功能
- 插件端和服务端按功能闭环联合规划，不按团队孤立切分
- 每个功能至少包含插件端工作项、服务端工作项和联调验收项
- “已完成”以代码、联调和验收结论为准，不以文档完成替代功能完成

## 4. 功能序列总览

### Feature 0：插件基础适配与初始化

定位：所有业务功能的基础前置能力。

目标：

- 建立 Claude Code / OpenCode 的插件公共适配层
- 完成插件启动、初始化、配置加载、状态展示和诊断骨架
- 建立与服务端交互所需的基础配置、日志和错误处理框架

### Feature 1：IAM 登录与会话闭环

定位：基础能力后的首个主要功能。

目标：

- 完成插件端本地回环服务、浏览器唤起、`login_ticket` 回调接收和安全存储
- 完成服务端 login session、代理回调处理、用户信息查询和会话签发
- 让用户可通过企业账号完成首次登录、恢复登录和登出

### Feature 2：统一模型接入

定位：登录闭环之后的首个核心业务能力。

目标：

- 服务端提供统一模型访问入口
- 插件端将模型访问重定向到 cmscoder 服务端
- 完成默认模型、模型列表和基础调用链路
- **【即将实现】Model Token 认证机制（apiKeyHelper + HMAC 签名 + 短期 JWT）**
- **【即将实现】跨平台本地代理（macOS/Linux UDS + Windows Named Pipe）**

### Feature 2 开发计划（即将开始）

#### 2.1 服务端（web-server）

**Phase 1：Model Token 签发端点**
- [ ] 实现 `POST /api/auth/model-token`（HMAC 签名校验 + JWT 签发）
- [ ] 实现 HMAC-SHA256 签名验证中间件
- [ ] 实现 timestamp + nonce 防重放机制
- [ ] 实现 plugin_secret 存储与校验
- [ ] 添加配置项：`modelTokenTTL`（默认 5分钟）、`enableIPBinding`

**Phase 2：Model Auth 中间件改造**
- [ ] 将 Composite Token 校验改造为 JWT 校验
- [ ] 实现 JWT 签名验证（HS256）
- [ ] 实现 JWT 过期校验
- [ ] 可选：实现 IP 绑定校验

**Phase 3：user-service 改造**
- [ ] 登录 exchange 时生成并返回 `plugin_secret`
- [ ] session 中存储 `plugin_secret`
- [ ] 可选：存储 `clientIP` 用于 IP 绑定

#### 2.2 插件端

**Phase 1：model-token 命令（Claude Code）**
- [ ] 实现 `cmscoder.js model-token` 命令
- [ ] 实现 HMAC-SHA256 签名构造
- [ ] 实现 access_token 过期自动刷新
- [ ] 输出 Model Token 到 stdout（供 apiKeyHelper 使用）

**Phase 2：本地代理（OpenCode）**
- [ ] 实现 `lib/model-proxy.js` 跨平台服务器
  - [ ] macOS/Linux：Unix Domain Socket
  - [ ] Windows：Named Pipe
- [ ] 实现动态密钥生成与校验
- [ ] 实现 Model Token 自动获取与缓存
- [ ] 实现 Token 过期前自动刷新
- [ ] 实现 `cmscoder.js model-proxy` 命令（start/stop/status）
- [ ] 实现 OpenCode 配置自动写入

**Phase 3：安全存储扩展**
- [ ] 安全存储增加 `plugin_secret` 键
- [ ] 移除 `model_api_key`、`composite_token`（如已存在）

#### 2.3 配置与文档

- [ ] 更新 `settings.json` 模板（Claude Code apiKeyHelper 配置）
- [ ] 更新 OpenCode 适配器文档
- [ ] 更新服务端配置文档

#### 2.4 联调验收

- [ ] Claude Code 完整链路：登录 → apiKeyHelper → 模型请求
- [ ] OpenCode 完整链路：登录 → 启动代理 → 模型请求
- [ ] Token 过期自动刷新验证
- [ ] IP 绑定功能验证（如启用）
- [ ] 跨平台测试（macOS、Linux、Windows）

### Feature 3：插件基础体验增强

目标：

- 启动时登录检查
- 基础状态展示与异常提示
- 基础会话增强

### Feature 4：工作流增强与上下文治理

目标：

- 任务分解、设计、编码、测试、Review 引导
- 项目规范注入
- 上下文压缩和长期上下文治理

### Feature 5：工具治理与权限控制

目标：

- 高风险命令识别
- allow / ask / deny 策略执行
- 文件写入、代码修改、测试执行等前置检查

### Feature 6：企业治理能力

目标：

- 配额、限流、审计、追踪
- 模型策略与白名单
- 后续对接企业治理体系

## 5. Feature 0 开发计划

### 5.1 插件端

- 建立 Claude Code / OpenCode 适配公共层
- 建立插件启动入口和初始化流程
- 实现本地配置加载、环境识别和基础日志
- 实现本地状态展示骨架
- 预留安全存储、登录和模型接入的扩展接口

### 5.2 服务端

- 建立服务端工程骨架和模块边界
- 建立配置管理、环境区分、日志和基础追踪
- 提供基础健康检查和版本信息接口
- 预留认证、模型网关和治理模块的接入点

### 5.3 验收标准

- 插件端可以完成启动、初始化、配置读取和基础状态展示
- 服务端可以完成启动、配置加载和基础健康检查
- 插件端与服务端具备稳定的基础联通方式

## 6. Feature 1 开发计划

### 6.1 插件端

- 提供登录入口与登录状态展示
- 启动本地回环服务并管理端口生命周期
- 调用 `<cmscoder-backend>/api/auth/login` 获取浏览器授权地址
- 唤起系统默认浏览器访问授权地址
- 接收 `http://127.0.0.1:<port>/callback` 回调中的 `login_ticket`
- 调用 `<cmscoder-backend>/api/auth/exchange` 交换正式 token 并写入安全存储
- 实现登录态恢复、登出和刷新失败处理

### 6.2 服务端

- 实现 `<cmscoder-backend>/api/auth/login`
- 管理 `login session`、`state` 与 `localPort` 映射
- 提供浏览器授权入口
- 实现 `<cmscoder-backend>/api/auth/callback`
- 调 `<iam>` 完成授权码换 token 和用户信息查询
- 签发 cmscoder 会话和一次性 `login_ticket`
- 实现 `<cmscoder-backend>/api/auth/exchange`
- 实现 `<cmscoder-backend>/api/auth/refresh` 和 `<cmscoder-backend>/api/auth/logout`

### 6.3 联调验收

- 插件端与服务端完成完整登录闭环
- 登录成功后插件端可稳定持有 cmscoder token
- 认证失败、state 失败、回环失败具备可观测错误
- 登录结果能直接支撑后续模型访问

## 7. 依赖关系

- Feature 1 依赖 Feature 0
- Feature 2 依赖 Feature 0 和 Feature 1
- Feature 3 依赖 Feature 0 和 Feature 1
- Feature 4、Feature 5、Feature 6 依赖前序接入能力稳定

## 8. 建议交付顺序

1. 完成 Feature 0 的插件基础适配与初始化
2. 完成 Feature 1 的 IAM 登录与会话闭环
3. 完成 Feature 2 的统一模型接入
4. 补齐 Feature 3 的基础体验增强
5. 再推进 Feature 4、Feature 5、Feature 6

## 8. Docker 验证环境

使用 Docker Compose 作为本地联调标准环境：

```bash
cp .env.example .env
# 编辑 .env，将 CMS_HOST_IP 改为宿主机 LAN IP（Linux）或 host.docker.internal（Docker Desktop）
docker compose up --build
```

验收标准：
- web-server 和 user-service 均可正常启动并通过健康检查
- web-server 可通过 `http://<HOST_IP>:9010` 从宿主机浏览器访问
- web-server 可通过 Docker 内部网络调用 user-service
- IAM 回调地址可通过 `CMS_HOST_IP` 环境变量灵活配置

## 9. 风险与待确认

- Feature 0 的插件初始化方式是否在 Claude Code 与 OpenCode 中足够统一
- IAM 联调环境、白名单和应用配置开通时序
- 安全存储能力在不同运行环境中的一致性
- 服务端与插件端是否由不同团队并行交付
