# 模型接入与协议转换设计

## 1. 文档目标

定义插件端到服务端再到天启大模型平台的模型访问路径、协议转换与错误处理规则。

## 2. 关联总纲

- 来源章节：3、6.1、8.1.3、8.2.3、8.2.4、8.2.5、9、12、14.2、18
- 相关文档：[../web-server/server-architecture.md](../web-server/server-architecture.md)

## 3. 设计范围

- 插件端模型入口统一重定向
- 服务端统一模型 API（OpenAI 兼容格式）
- 临时 Model API Key 生成、校验与防滥用
- 模型 ID 映射、默认模型、模型白名单
- usage、trace、错误码统一

## 4. 关键场景或流程

- 插件端登录 exchange 时获取临时 Composite Token
- Composite Token 与 user session 绑定，session 过期/登出即失效
- 插件端配置 Code Agent 使用 cmscoder 模型端点
- Code Agent 发起标准 OpenAI 格式请求到 `/api/model/v1/chat/completions`，携带 Composite Token
- web-server 通过 ModelAuth 中间件解析 composite token，校验 key 有效性 + session 状态 + 绑定关系
- 校验通过后转发请求到上游天启平台
- 服务端回传统一响应、usage 与 trace 信息
- 上游失败时返回统一错误语义

## 5. 设计要点

- 插件端不直接访问天启大模型平台
- 服务端统一补充系统级鉴权信息
- Composite Token 绑定 user session + agentType + pluginInstanceId
- 协议转换层必须向插件屏蔽底层差异
- Composite Token 防止 key 泄露后被滥用

## 6. 接口、数据或配置

### 6.1 模型端点

| 端点 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/model/v1/chat/completions` | POST | Bearer Composite Token | OpenAI 兼容，支持流式 SSE |
| `/api/model/v1/models` | GET | Bearer Composite Token | 列出可用模型 |

### 6.2 Model API Key 与 Composite Token

#### 6.2.1 Model API Key

- 格式：`cmscoder_` + 32 字符 hex
- 生成时机：登录 exchange 成功时
- 绑定：userId + sessionId + agentType + pluginInstanceId
- 有效期：与 access_token 同步
- 吊销：登出或 session 过期时自动失效

#### 6.2.2 Composite Token（防滥用）

插件端实际使用的凭证不是原始的 Model API Key，而是 **Composite Token**：

```
cmscoderv1_<base64(modelApiKey:accessToken)>
```

**设计动机**：Model API Key 是静态字符串，泄露后可被任意客户端使用。Composite Token 将 Model API Key 和 Access Token 绑定为单个凭证，确保即使其中一方泄露，另一方仍然缺失，无法使用。

**安全校验流程**：
1. web-server 解析 composite token，提取 modelApiKey 和 accessToken
2. 校验 modelApiKey 有效
3. 校验 modelApiKey 绑定的 sessionId == accessToken（同一 session）
4. 校验 accessToken 对应的 session 仍有效
5. 校验通过后转发请求到上游

**Claude Code 配置**：将 composite token 作为 `apiKey` 配置到 `customProviders`，无需修改插件端协议。

**为什么选择绑定 accessToken 而非 pluginInstanceId**：
- Claude Code 的 `customProviders` 只支持一个 `apiKey` 字段，无法同时发送两个凭证
- accessToken 是短时凭证，随 session 刷新而轮换
- pluginInstanceId 是静态标识，无法提供同样的防泄露保障

### 6.3 配置

- 服务端配置 `[model]` 段：`upstreamBaseURL`、`upstreamApiKey`、`defaultModel`
- 插件端缓存：`model_endpoint`（bootstrap 时自动设置）
- 插件端安全存储：`model_api_key`（exchange 时保存，仅供参考）
- 插件端安全存储：`composite_token`（exchange 时保存，实际使用）

## 7. 非功能要求

- 接口语义稳定，便于多 Agent 复用
- 异常处理一致，便于问题诊断
- 支持后续新增模型平台或路由策略

## 8. 风险与待确认

- 天启平台 API 格式与 OpenAI 格式的映射关系需确认
- 生产环境是否需要 Redis 替代内存缓存以支持多实例

## 9. 验收标准

- 所有模型请求可统一走服务端入口
- 插件端不暴露上游系统级密钥
- Composite Token 仅在 session 有效期间可用
- 登出后 Composite Token 立即失效
- 模型映射、错误码、usage 与 trace 信息具备一致性
- Composite Token 防滥用：即使 modelApiKey 或 accessToken 单独泄露，无法用于模型端点
