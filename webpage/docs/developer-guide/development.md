---
description: 在本地构建和运行 Neko
slug: /developer-guide/development
---

# 本地开发

当前仓库使用直接的 Go、npm 和 Docker Compose 构建流程，不再维护旧的
`server/dev/`、`client/dev/` Docker 包装脚本。

## 环境要求

- Docker Engine 和 Docker Compose；
- Go 1.25 或更高版本；
- Node.js 18 或更高版本；
- npm。

## 构建服务端

```bash
cd server
go build -o bin/neko ./cmd/neko
CGO_ENABLED=0 go build -o bin/neko-proxy ./cmd/neko-proxy
```

仅检查服务端所有包是否能编译：

```bash
go build ./...
```

## 构建前端

```bash
cd client
npm ci
npm run build
```

生产资源会生成到 `client/dist/`。如果只需要开发服务器：

```bash
npm run serve -- --port 3001
```

开发服务器的 API 地址可通过 `VUE_APP_SERVER_PORT` 指定，例如：

```bash
VUE_APP_SERVER_PORT=8080 npm run serve -- --port 3001
```

## 启动本地 Demo

Demo 会使用当前分支编译出的服务端、代理和前端资源构建 Chromium 镜像：

```bash
export NEKO_DEMO_USER_PASSWORD='普通用户密码'
export NEKO_DEMO_ADMIN_PASSWORD='管理员密码'
export NEKO_DEMO_NAT_IP='127.0.0.1'
export NEKO_DEMO_MEDIA_BIND='127.0.0.1'

docker compose -f demo/compose.local.yaml up -d --build
```

访问 `http://127.0.0.1:8080`。修改 Go 或前端代码后，重新执行对应构建命令，
再运行相同的 Compose 命令即可更新 Demo。

停止 Demo：

```bash
docker compose -f demo/compose.local.yaml down
```

更多 FRP、Coturn 和连通性检查说明见 [demo/README.md](../../../demo/README.md)。

## 生成完整镜像

构建客户端和服务端后，可以使用根目录构建入口生成 Chromium 镜像：

```bash
./build --application chromium --yes
```

如需手动生成 Dockerfile：

```bash
go run utils/docker/main.go \
  -i Dockerfile.tmpl \
  -o Dockerfile \
  -client client/dist
docker build -t local/neko -f Dockerfile .
```

## 代码检查

提交前至少执行：

```bash
git diff --check

cd server
go build ./...

cd ../client
npm run build
```

前端 SDK 合约检查使用 `npm run test:sdk`，完整文档站点检查在 `webpage/` 目录
执行 `npm ci` 后使用 `npm run typecheck` 和 `npm run build`。
