# cmscoder-server

cmscoder 后端服务 monorepo，包含所有服务端组件。

## 服务列表

| 服务 | 目录 | 端口 | 说明 |
|------|------|------|------|
| **web-server** | `web-server/` | 39010 (内部) / 9010 (k8s NodePort) | 统一 API 网关、IAM 回调代理、插件端入口 |
| **user-service** | `user-service/` | 39011 (内部 ClusterIP) | IAM 认证代理、用户会话管理、token 签发与刷新 |

## 快速开始

```bash
# 构建所有服务
make build

# 运行单个服务
make run service=web-server
make run service=user-service

# 构建 Docker 镜像
make docker

# 清理依赖
make tidy
```

## Docker 本地验证

使用 Docker Compose 快速启动完整服务端环境：

```bash
# 在项目根目录执行
cp .env.example .env
# 编辑 .env，将 CMS_HOST_IP 改为你的 LAN IP（Linux）或 host.docker.internal（Docker Desktop）

docker compose up --build          # 前台启动
docker compose up -d --build       # 后台启动
docker compose logs -f             # 查看日志
docker compose down                # 停止
```

- web-server 通过 `http://<CMS_HOST_IP>:9010` 对外提供服务
- user-service 仅在 Docker 内部网络可达，不暴露到宿主机

## 服务间通信

| 调用方 | 被调方 | 协议 | 端点 |
|--------|--------|------|------|
| 插件端 | web-server | HTTP | `:39010` (本地) / `:9010` (k8s) |
| web-server | user-service | HTTP | `:39011` (集群内) |
| user-service | 公司 IAM | HTTPS | 配置文件中定义 |

## 技术栈

- Go 1.26
- GoFrame v2.10.0
- Redis（生产环境，当前使用内存缓存测试）
- MySQL/PostgreSQL (持久化存储)
- Kubernetes 部署

## 目录约定

- 每个服务独立 `go.mod`，服务间通过 HTTP 通信
- `go.work` 用于本地多服务同时开发
- 新增微服务直接在根目录下创建新目录，端口从 39012 递增
