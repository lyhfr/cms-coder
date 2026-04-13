# 模型接入与协议转换设计

## 1. 文档目标

定义插件端到服务端再到天启大模型平台的模型访问路径、协议转换与错误处理规则。

## 2. 关联总纲

- 来源章节：3、6.1、8.1.3、8.2.3、8.2.4、8.2.5、9、12、14.2、18
- 相关文档：[../web-server/server-architecture.md](../web-server/server-architecture.md)、[../user-service/iam-auth-session.md](../user-service/iam-auth-session.md)

## 3. 设计范围

- 插件端模型入口统一重定向
- 服务端统一模型 API（OpenAI 兼容格式）
- **动态 Model Token 签发、校验与防滥用（HMAC 签名 + 短期 JWT）**
- 模型 ID 映射、默认模型、模型白名单
- usage、trace、错误码统一

## 4. 关键场景或流程

### 4.1 模型请求认证流程（apiKeyHelper 模式）

```
Claude Code 调用 apiKeyHelper
        │
        ▼
┌─ cmscoder.js model-token ──────────────────────┐
│                                                  │
│  1. 从安全存储读取 access_token                   │
│  2. 从安全存储读取 plugin_secret                  │
│                                                  │
│  3. access_token 是否有效？                       │
│     ├─ 有效 → 继续                                │
│     └─ 过期 → 用 refresh_token 静默刷新           │
│               ├─ 成功 → 继续                      │
│               └─ 失败 → 提示重新登录              │
│                                                  │
│  4. 构造签名请求：                                 │
│     timestamp = now()                            │
│     nonce = random()                             │
│     signature = HMAC_SHA256(                     │
│       access_token + timestamp + nonce,          │
│       plugin_secret                              │
│     )                                            │
│                                                  │
│  5. POST /api/auth/model-token                   │
│     Body: { accessToken, timestamp, nonce,       │
│             signature, pluginInstanceId }         │
│                                                  │
│  6. stdout → model_token（给 Claude Code 使用）   │
└──────────────────────────────────────────────────┘
        │
        ▼
Claude Code 使用 model_token 调用模型端点
        │
        ▼
┌─ web-server Model Auth 中间件 ──────────────────┐
│                                                  │
│  1. 解析 JWT                                     │
│  2. 校验签名（服务端 JWT 密钥）                   │
│  3. 校验未过期（默认 5分钟 TTL，可配置）          │
│  4. 提取 userId, sessionId                       │
│  5. 转发到上游模型                                │
└──────────────────────────────────────────────────┘
```

### 4.2 设计要点

- 插件端不直接访问天启大模型平台
- 服务端统一补充系统级鉴权信息
- **Model Token 通过 HMAC 签名请求获取，确保请求来自合法插件端**
- **Model Token 为短期 JWT（默认 5分钟），大幅降低泄露后可使用窗口**
- 协议转换层必须向插件屏蔽底层差异
- **IP 绑定校验为可选项，通过服务端配置开启**

## 5. 接口、数据或配置

### 5.1 模型端点

| 端点 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/model/v1/chat/completions` | POST | Bearer Model Token | OpenAI 兼容，支持流式 SSE |
| `/api/model/v1/models` | GET | Bearer Model Token | 列出可用模型 |
| `/api/auth/model-token` | POST | HMAC 签名 | 获取短期 Model Token |

### 5.2 Model Token 认证机制

#### 5.2.1 核心设计目标

1. **无静态 API Key**：用户无需手动配置，登录后自动获取
2. **防滥用**：即使 Token 被截获，可使用时间窗口极短（默认 5分钟）
3. **请求来源验证**：通过 HMAC 签名确保请求来自合法插件端

#### 5.2.2 Token 生命周期策略

| Token 类型 | 有效期 | 刷新方式 | 存储位置 | 说明 |
|-----------|--------|---------|---------|------|
| `access_token` | 15分钟 | `/api/auth/refresh` | 安全存储 | 会话访问凭证 |
| `refresh_token` | 7天 | 登录重新获取 | 安全存储 | 续期凭证 |
| `plugin_secret` | 随 session | 无需刷新 | 安全存储 | HMAC 签名密钥 |
| **model_token** | **可配置（默认 5分钟）** | **apiKeyHelper 自动获取** | **内存（不持久化）** | **短期模型调用凭证** |

**配置项**（服务端 `[model]` 段）：
```toml
[model]
modelTokenTTL = "5m"        # Model Token 有效期，默认 5分钟
enableIPBinding = false     # 是否启用 IP 绑定校验，默认关闭
apiKeyHelperInterval = "4m" # 建议的 apiKeyHelper 调用间隔
```

#### 5.2.3 获取 Model Token 流程

**请求**：`POST /api/auth/model-token`

```json
{
  "accessToken": "cmscoder_session_xxx",
  "timestamp": 1715500000,
  "nonce": "random_string_16_chars",
  "signature": "hmac_sha256_hex",
  "pluginInstanceId": "claude-code-v1.0"
}
```

**签名算法**：
```
signature = HMAC_SHA256(
  accessToken + timestamp + nonce,
  plugin_secret
)
```

**服务端校验**：
1. 校验 timestamp 在 ±30秒 内（防重放攻击）
2. 校验 nonce 未被使用过（短期缓存，如 5分钟）
3. introspect access_token 获取 session
4. 从 session 中取出 plugin_secret
5. 计算 expected_signature 并对比
6. 如启用 IP 绑定，校验请求 IP 与 session 记录的 IP 一致
7. 签发 model_token（短期 JWT）

**响应**：
```json
{
  "modelToken": "eyJhbGciOiJIUzI1NiIs...",
  "expiresIn": 300,
  "tokenType": "Bearer"
}
```

#### 5.2.4 Model Token 校验（模型端点）

模型端点 `/api/model/v1/*` 使用 Model Auth 中间件：

1. 从 Authorization Header 提取 Bearer Token
2. 解析 JWT（HS256 签名）
3. 校验 JWT 签名（服务端密钥）
4. 校验 JWT 未过期
5. 提取 claims：userId, sessionId, agentType
6. 转发请求到上游模型平台

**JWT Claims**：
```json
{
  "sub": "user_id",
  "sid": "session_id",
  "agent": "claude-code",
  "iat": 1715500000,
  "exp": 1715500300
}
```

### 5.3 客户端配置

#### 5.3.1 Claude Code（推荐：apiKeyHelper 模式）

通过 `settings.json` 配置 `apiKeyHelper`：

```json
{
  "apiKeyHelper": "node ~/.cmscoder/plugin/lib/cmscoder.js model-token",
  "env": {
    "ANTHROPIC_BASE_URL": "https://cmscoder.company.com/api/model/v1",
    "CLAUDE_CODE_API_KEY_HELPER_TTL_MS": "240000"
  }
}
```

**说明**：
- `apiKeyHelper`：Claude Code 每次需要 API Key 时执行的命令
- `CLAUDE_CODE_API_KEY_HELPER_TTL_MS`：调用间隔（建议略小于 modelTokenTTL，如 4分钟）
- 命令输出（stdout）作为 API Key 使用
- 命令错误（stderr + exit 非 0）会提示用户重新登录

#### 5.3.2 OpenCode（本地代理模式）

OpenCode 不支持 `apiKeyHelper`，也不支持配置文件热加载，采用**本地代理模式**：

```
OpenCode ──▶ 本地传输层 ──▶ 本地代理 ──▶ cmscoder-web-server
            (平台相关)       (动态密钥)     (模型端点)
```

**跨平台传输方案**：

| 平台 | 传输方式 | 安全机制 |
|------|---------|---------|
| **macOS** | Unix Domain Socket | 文件权限 600 + 动态密钥 |
| **Linux** | Unix Domain Socket | 文件权限 600 + 动态密钥 |
| **Windows** | Named Pipe | ACL 限制当前用户 + 动态密钥 |

**安全防护**：
- **动态密钥**：启动时生成 32 字节随机密钥，请求需携带
- **传输层隔离**：UDS/Named Pipe 替代 TCP，文件系统权限控制
- **来源校验**：校验 User-Agent 包含 "OpenCode"
- **IP 白名单**（Windows 回退）：仅接受 127.0.0.1/::1

**本地代理职责**：
1. 拦截 OpenCode 的模型请求
2. 自动获取/缓存 Model Token（5分钟有效）
3. Token 过期前自动刷新
4. 添加 Authorization Header 后转发到服务端
5. **校验请求来源（传输层 + 动态密钥）**

**CLI 命令**：
```bash
# 启动本地代理（自动选择平台最优传输）
node ~/.cmscoder/plugin/lib/cmscoder.js model-proxy

# 停止本地代理
node ~/.cmscoder/plugin/lib/cmscoder.js model-proxy stop

# 查看状态
node ~/.cmscoder/plugin/lib/cmscoder.js model-proxy status
```

**OpenCode 配置（自动写入，平台相关）**：

| 平台 | baseURL 示例 |
|------|-------------|
| macOS/Linux | `http+unix://%2FUsers%2Fuser%2F.cmscoder%2Fmodel-proxy.sock` |
| Windows | `http://localhost:8080`（仅 127.0.0.1） |

```json
{
  "provider": {
    "cmscoder-local": {
      "baseURL": "http+unix://%2FUsers%2Fuser%2F.cmscoder%2Fmodel-proxy.sock",
      "apiKey": "random_proxy_secret_32chars",
      "models": {
        "gpt-4": { "id": "gpt-4" }
      }
    }
  },
  "model": "cmscoder-local/gpt-4"
}
```

### 5.4 会话状态处理

**场景 1：首次请求（刚登录）**
- access_token 有效
- 直接构造签名请求获取 model_token

**场景 2：长时间未对话（access_token 过期）**
- apiKeyHelper 检测到 access_token 过期
- 先用 refresh_token 静默刷新 access_token
- 刷新成功后获取 model_token
- 刷新失败 → 提示重新登录

**场景 3：Session 已过期（refresh_token 过期）**
- 静默刷新失败
- apiKeyHelper 输出错误信息到 stderr，exit 非 0
- Claude Code 提示用户重新登录

## 6. 安全机制

### 6.1 防滥用设计

| 机制 | 说明 | 效果 |
|------|------|------|
| **短期 Token** | 默认 5分钟有效期 | 泄露后可使用窗口极短 |
| **HMAC 签名** | 需要 plugin_secret 构造有效请求 | 增加伪造请求难度 |
| **Timestamp + Nonce** | 防重放攻击 | 同一请求无法重复使用 |
| **IP 绑定（可选）** | 校验请求来源 IP | 防止 Token 被带走使用 |

### 6.2 攻击场景分析

| 攻击场景 | 防护效果 |
|---------|---------|
| 用户截获 model_token | 仅 5 分钟可用，过期后需重新获取 |
| 用户截获 access_token | 无法直接当 model_token 使用，需 plugin_secret 签名 |
| 用户获取 plugin_secret | 需要同时构造有效签名，且 access_token 可能过期 |
| 自动化工具滥用 | 需要复现整套签名逻辑，门槛高 |
| **综合评估** | **虽不能 100% 阻止，但手动滥用成本远高于直接使用插件** |

## 7. 非功能要求

- 接口语义稳定，便于多 Agent 复用
- 异常处理一致，便于问题诊断
- 支持后续新增模型平台或路由策略
- **配置灵活：Token 有效期、IP 绑定均可配置**

## 8. 风险与待确认

- 天启平台 API 格式与 OpenAI 格式的映射关系需确认
- 生产环境是否需要 Redis 替代内存缓存以支持多实例
- **IP 绑定在 VPN/移动网络场景下的兼容性**

## 9. 验收标准

- 所有模型请求可统一走服务端入口
- 插件端不暴露上游系统级密钥
- **Model Token 通过 apiKeyHelper 动态获取，无需手动配置**
- **Model Token 短期有效（可配置），泄露后可使用窗口可控**
- **HMAC 签名机制确保请求来源可信**
- 模型映射、错误码、usage 与 trace 信息具备一致性
