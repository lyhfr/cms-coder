# User Service 设计摘要

`User Service` 是 SSO、用户身份和会话管理的核心服务，负责 IAM 对接、session 签发、refresh token 轮换和注销。

## 职责

- 持有 IAM `client_id` / `client_secret`
- 创建和管理 `login session`、`state`、`login_ticket`
- 通过 IAM 完成授权码换 token 和用户信息查询
- Upsert 用户信息并签发 cmscoder access token / refresh token
- 处理 refresh、logout、session introspection

## 部署

- 内部端口: `39011`
- k8s 服务类型: `ClusterIP`（不暴露外部端口）
- 仅对 `cmscoder-web-server` 开放访问

## Feature 1 建议模块

- `application/login-session`
- `application/iam`
- `application/session`
- `application/ticket`
- `application/user-profile`
- `infrastructure/cache`（当前内存，生产待迁 Redis）
- `infrastructure/repository`
- `infrastructure/iam-client`

## 关键边界

- 只对内暴露接口，不直接对插件暴露公网地址
- SSO 票据和会话主状态统一收敛到本服务
- 所有认证事件都需要审计和 trace

详细设计见 [../../cmscoder-docs/specs/user-service/iam-auth-session.md](../../cmscoder-docs/specs/user-service/iam-auth-session.md)。
