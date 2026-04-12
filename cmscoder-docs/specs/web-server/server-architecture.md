# 服务端总体架构设计

## 1. 文档目标

定义 cmscoder 服务端在认证、会话、模型访问、治理与审计方面的系统结构。

## 2. 关联总纲

- 来源章节：7、8.2、9.2、12、13、14、16
- 相关文档：[../shared/cmscoder-overview.md](../shared/cmscoder-overview.md)

## 3. 设计范围

### In Scope

- 身份校验
- 会话管理与 token 刷新
- 模型统一入口
- 协议转换
- 模型路由
- 系统级凭证管理
- 配额与限流
- 审计与可观测

### Out of Scope

- 替代插件端的本地交互职责
- 直接提供最终用户 IDE 体验

## 4. 关键场景

- 插件携带登录态访问服务端
- 服务端作为 IAM 代理回调方接收授权码，生成一次性登录票据并完成 cmscoder 会话签发
- 服务端验证用户身份并签发短时访问能力与可刷新的会话
- 请求被转换并转发到天启大模型平台
- 配额超限、策略拦截、上游异常时返回统一错误

## 5. 设计要点

### 逻辑分层

- 接入层：统一承接插件请求
- 认证代理层：承接 IAM 登录入口、代理回调、login session/state 管理、code 换 token、用户信息查询
- 身份层：校验用户、租户、项目、角色
- 会话层：管理访问 token、刷新 token、一次性 login ticket 和有效期
- 模型网关层：协议适配、模型映射、路由、重试与错误归一
- 治理层：配额、限流、白名单、审计、追踪

### 核心原则

- 插件端只访问服务端，不直连上游模型平台
- 基础能力完成后的首个主功能是 IAM 登录服务端闭环
- 系统级 Key 只存在于服务端
- 所有治理策略由服务端统一执行

## 6. 接口、数据或配置

- 对插件端：登录 session 初始化、浏览器授权入口、login ticket 交换、登录态校验、模型访问入口、模型列表、默认模型、额度摘要
- 对上游：天启平台标准模型 API、IAM 标准 OAuth2.0 / SSO、IAM 授权码换 token 与用户信息接口
- 核心数据：用户会话、IAM 登录 state 映射、login session、一次性 login ticket、模型映射、策略配置、额度记录、审计日志

## 7. 非功能要求

- 支持请求追踪、错误追踪和审计留痕
- 支持灰度、熔断、限流和模型白名单
- 服务端接口语义稳定，便于多 Agent 共用

## 8. 风险与待确认

- 服务端是否需要同时兼容 OpenAI 风格与 Claude 风格接口
- 模型路由规则与团队/项目策略来源待明确
- 审计日志保留策略与合规边界待确认

## 10. 本地验证部署

### Docker Compose

项目根目录提供 `docker-compose.yml`，用于本地开发和联调验证：

```bash
# 复制并编辑环境变量
cp .env.example .env
# 将 CMS_HOST_IP 改为宿主机 LAN IP（Linux）或 host.docker.internal（Docker Desktop）

# 启动
docker compose up --build
```

- **web-server**：端口映射 `9010:39010`，浏览器通过 `http://<HOST_IP>:9010` 访问
- **user-service**：仅 Docker 内部网络可达（`http://user-service:39011`），不暴露到宿主机
- IAM 回调 `redirectURI` 通过环境变量 `GF_IAM__REDIRECTURI` 动态覆盖，无需重建镜像

### Kubernetes

生产环境使用 kustomize 部署。参见 `manifest/deploy/kustomize/` 目录。

## 11. 验收标准

- 服务端可独立完成 IAM 登录代理回调、login ticket 交换和 cmscoder 会话签发
- 服务端可独立承接插件认证、模型访问和治理链路
- 上游异常与策略异常可以统一编码和追踪
- 系统级 Key 不会下发到插件端或最终用户环境
