# Web Server 设计摘要

`Web Server` 是 cmscoder 的统一后端入口，只负责公网路由、浏览器回调承接和插件端 API 暴露，不承载 IAM 细节和用户会话主逻辑。

## 职责

- 对外暴露 `/api/auth/*`、`/api/plugin/*` 等稳定接口
- 统一做 trace、限流、参数校验和 access token 验证
- 承接 IAM 回调地址 `/api/auth/iam/callback`
- IAM 授权跳转由 web-server 从配置直接构造
- 将 SSO 相关内部调用转发给 `User Service`

## Feature 1 建议模块

- `api/auth`：登录、callback、exchange、refresh、logout
- `api/plugin`：`/api/plugin/bootstrap` 等聚合接口
- `middleware/auth`：access token 验签或 introspection
- `clients/user-service-client`：调用 `User Service` 的内部 API

## 关键边界

- 不保存 IAM `client_secret`
- 不直接落用户 session 主数据
- 不向插件返回浏览器可见的 access token
- IAM 地址通过配置文件管理，不向 user-service 显式获取

详细设计见 [../cmscoder-docs/specs/user-service/iam-auth-session.md](../cmscoder-docs/specs/user-service/iam-auth-session.md)。

## 工程结构

```
web-server/
├── api/auth/v1/              # API 请求/响应定义 (GoFrame 路由元数据)
│   ├── auth_login.go                # POST /api/auth/login
│   ├── auth_login_authorize.go      # GET /api/auth/login/{loginId}/authorize
│   ├── auth_callback.go             # GET /api/auth/iam/callback
│   ├── auth_exchange.go             # POST /api/auth/exchange
│   ├── auth_refresh.go              # POST /api/auth/refresh
│   ├── auth_logout.go               # POST /api/auth/logout
│   ├── auth_me.go                   # GET /api/auth/me
│   └── plugin_bootstrap.go          # GET /api/plugin/bootstrap
├── internal/
│   ├── cmd/cmd.go                   # 主命令, 路由注册与中间件装配
│   ├── clients/userclient/          # User Service HTTP 客户端
│   ├── controller/auth/             # 控制器实现
│   └── middleware/                  # tracing / auth / ratelimit 中间件
├── manifest/
│   ├── config/config.toml           # 应用配置 (含 iam 段)
│   ├── docker/Dockerfile            # 容器镜像
│   └── deploy/kustomize/            # K8s 部署配置
├── hack/                            # GoFrame CLI 配置
├── main.go
└── Makefile
```
