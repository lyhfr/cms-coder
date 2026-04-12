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

### 6.1 对插件端端点

| 端点 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `POST /api/auth/login` | POST | 无 | 创建登录 session，返回 loginId 和 browserUrl |
| `GET /api/auth/login/{loginId}/authorize` | GET | 无 | 重定向浏览器到 IAM 授权页 |
| `GET /api/auth/iam/callback` | GET | 无 | IAM OAuth 回调，接收 code + state，转发到 user-service |
| `POST /api/auth/exchange` | POST | 无 | 用 login_ticket 交换 accessToken、refreshToken、compositeToken |
| `POST /api/auth/refresh` | POST | 无 | 刷新 access_token（token 轮换） |
| `POST /api/auth/logout` | POST | 无 | 注销会话（需 refreshToken 或 sessionId） |
| `GET /api/auth/me` | GET | Access Token | 获取当前用户信息 |
| `GET /api/plugin/bootstrap` | GET | Access Token | 获取插件引导配置（用户、功能开关、默认模型） |
| `POST /api/model/v1/chat/completions` | POST | Composite Token | OpenAI 兼容聊天补全，支持流式 SSE |
| `GET /api/model/v1/models` | GET | Composite Token | 列出可用模型 |

### 6.2 user-service 内部端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `POST /user-service/auth/login` | POST | 创建登录 session |
| `POST /user-service/auth/iam/callback/complete` | POST | 完成 IAM OAuth 回调 |
| `POST /user-service/auth/login-tickets/exchange` | POST | 交换 login_ticket 为会话凭证 |
| `POST /user-service/auth/sessions/refresh` | POST | 刷新会话（token 轮换） |
| `POST /user-service/auth/sessions/revoke` | POST | 注销会话 |
| `GET /user-service/auth/sessions/introspect` | GET | 校验 access_token |
| `POST /user-service/auth/model-keys/validate` | POST | 校验 model API key |

### 6.3 核心数据

- 用户会话：sessionId（= accessToken）、refreshToken、expiresAt、agentType、pluginInstance
- IAM 登录 state 映射：state → loginId
- login session：loginId、state、localPort、agentType、pluginInstanceId、status、expiresAt
- 一次性 login ticket：ticketId、loginId、pluginInstanceId、consumedAt、expiresAt
- 模型映射：modelApiKey → userId + sessionId + agentType + pluginInstance + expiresAt
- 策略配置：`[model]` 段（upstreamBaseURL、upstreamApiKey、defaultModel）
- 额度记录、审计日志（待实现）

## 7. 非功能要求

- 支持请求追踪、错误追踪和审计留痕
- 支持灰度、熔断、限流和模型白名单
- 服务端接口语义稳定，便于多 Agent 共用

## 8. 风险与待确认

- 模型路由规则与团队/项目策略来源待明确
- 审计日志保留策略与合规边界待确认
- 天启平台 API 格式与 OpenAI 格式的映射关系待确认

## 9. Model API 设计

### 9.1 临时 Model API Key 与 Composite Token

登录 exchange 时服务端自动生成，特性：

| 特性 | 说明 |
|------|------|
| Model Key 格式 | `cmscoder_` + 32 字符 hex（16 字节随机） |
| Composite Token 格式 | `cmscoderv1_<base64(modelApiKey:accessToken)>` |
| 绑定 | 与 user session + agentType + pluginInstanceId 绑定 |
| 有效期 | 与 access_token 同步，session 过期即失效 |
| 防滥用 | 解析 composite token 后同时校验 modelApiKey 和 accessToken，验证二者属于同一 session；登出时联动吊销 |
| 插件使用 | 插件端实际使用 composite token 作为模型端点的 Bearer token |

### 9.2 端点列表

| 端点 | 方法 | 认证方式 | 说明 |
|------|------|---------|------|
| `/api/model/v1/chat/completions` | POST | Composite Token (Bearer) | OpenAI 兼容聊天补全，支持流式 SSE |
| `/api/model/v1/models` | GET | Composite Token (Bearer) | 列出可用模型 |

### 9.3 安全链路

```
插件端登录 exchange
  → user-service 生成 modelApiKey（绑定 session）
  → user-service 生成 compositeToken = base64(modelApiKey:accessToken)
  → 返回 compositeToken 给插件端
  → 插件端存储到 secureStore
  → Code Agent 使用 compositeToken 调用 /api/model/v1/chat/completions
  → web-server 通过 ModelAuth 中间件解析 composite token
  → 提取 modelApiKey 和 accessToken
  → 校验 modelApiKey 有效
  → 校验 modelApiKey 的 sessionId == accessToken（同一 session 绑定）
  → 校验 session 仍有效
  → 校验通过后转发请求到上游天启平台
  → 登出时联动吊销 modelApiKey
```

### 9.4 Composite Token 防滥用机制

| 威胁场景 | 防护机制 |
|---------|---------|
| Model API Key 泄露 | 缺少 accessToken，无法通过 session 绑定校验 |
| Access Token 泄露 | 缺少 modelApiKey，无法解析 composite token |
| 两个 token 分别泄露 | 攻击者需要同时获取并正确组合，难度指数级增加 |
| 重放攻击 | accessToken 随 session 刷新轮换，旧 composite token 失效 |

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
