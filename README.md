# n.eko

Neko 是一个基于 Docker 和 WebRTC 的自托管协作浏览器。它在容器中运行
Chromium，并将桌面画面、音频和控制通道实时提供给浏览器客户端。

当前发布版本：`v3.1.6`

本版本重点改进了 Chromium 网络连接和登录启动流程、响应式房间界面、视频编解码
回退，以及本地开发和运行时结构。完整变更记录见
[发布说明](webpage/docs/release-notes.md)。

Neko 适合以下场景：

- 多人观看、聊天和发送表情；
- 远程演示、教学和协作浏览；
- 需要共享浏览器环境的内部工具；
- 将远程 Chromium 嵌入其他应用。

## 当前运行范围

当前版本只支持 Chromium 运行时：

- Docker Engine（Linux `amd64`、`arm64`）；
- Docker Desktop/WSL2（Windows，使用 Linux 容器）；
- WebRTC 直连、TCP/UDP MUX、TURN 和 FRP 网络模式；
- 软件编码回退，以及由宿主机和容器驱动决定的硬件编码能力；
- 按网络压力动态调整视频质量，并支持可用的 H.264、H.265 和 AV1 编码路径。

Firefox、VLC、其他桌面应用和 Windows 原生媒体 Worker 不在当前版本范围内。

## 主要功能

- 多用户和管理员认证；
- WebRTC 视频、音频和低延迟输入控制；
- 控制权租约，避免多人同时操作桌面；
- 聊天、表情、聊天记录持久化；
- 用户头像上传和持久化；
- 文件传输、剪贴板、键盘布局和屏幕分辨率管理；
- 直播状态管理；
- `low`、`balanced`、`high` 视频质量配置；
- 代理、TURN 和 FRP 部署支持；
- 单端口 UDP/TCP MUX，降低公网部署的端口配置复杂度。

## 快速启动

### 使用已发布镜像

本仓库发布的 Chromium 镜像地址为：

```text
ghcr.io/picronsin/neko/chromium:v3.1.6
```

如果使用上游官方镜像，可继续使用仓库中的 Compose 示例；如需固定到本版本，
将 `docker-compose.yaml` 中的 `image` 改为上面的版本地址。

复制项目根目录的 `docker-compose.yaml`，先修改以下配置：

```yaml
NEKO_MEMBER_MULTIUSER_USER_PASSWORD: "修改为普通用户密码"
NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: "修改为管理员密码"
NEKO_WEBRTC_NAT1TO1: "客户端可访问的服务器地址"
```

然后启动：

```bash
docker compose up -d
```

浏览器访问 `http://服务器地址:8080`。默认 Compose 会使用 Chromium 镜像，
并将 WebRTC UDP 和 TCP 回退复用到 `52000` 端口：

```text
52000/udp
52000/tcp
```

公网部署时，必须在防火墙和 Docker 中放行相同端口，并将
`NEKO_WEBRTC_NAT1TO1` 设置为客户端实际可访问的地址。无法使用直连时，
请配置 TURN 或 FRP。

停止服务：

```bash
docker compose down
```

### 本地开发 Demo

本地 Demo 会把当前分支编译出的服务端、代理、Chromium 启动配置和前端资源
覆盖到 Chromium 运行时中。需要 Go 1.25+、Node.js 18+、npm 和 Docker，然后执行：

```bash
cd server
go build -o bin/neko ./cmd/neko
CGO_ENABLED=0 go build -o bin/neko-proxy ./cmd/neko-proxy
cd ../client
npm ci
npm run build
cd ..
```

设置 Demo 凭据并启动：

```bash
export NEKO_DEMO_USER_PASSWORD='普通用户密码'
export NEKO_DEMO_ADMIN_PASSWORD='管理员密码'
export NEKO_DEMO_NAT_IP='127.0.0.1'
export NEKO_DEMO_MEDIA_BIND='127.0.0.1'

docker compose -f demo/compose.local.yaml up -d --build
```

访问 `http://127.0.0.1:8080`。Demo 使用命名卷保存头像和聊天记录，重建容器
不会删除这些数据。

更多本地 FRP、Coturn 和连通性检查说明见
[demo/README.md](demo/README.md)。

## 配置入口

配置主要通过环境变量或 Compose 注入。当前常用配置包括：

| 配置 | 作用 |
| --- | --- |
| `NEKO_DESKTOP_SCREEN` | 虚拟桌面分辨率和帧率，例如 `1280x720@30` |
| `NEKO_CAPTURE_VIDEO_PROFILE` | 视频质量档位：`low`、`balanced`、`high` |
| `NEKO_WEBRTC_UDPMUX` | WebRTC UDP MUX 端口 |
| `NEKO_WEBRTC_TCPMUX` | WebRTC TCP MUX 端口 |
| `NEKO_WEBRTC_NAT1TO1` | 客户端访问服务端时使用的地址 |
| `NEKO_WEBRTC_CONNECTIVITY_MODE` | `direct` 或 FRP 等连通模式 |
| `NEKO_MEMBER_MULTIUSER_USER_PASSWORD` | 普通用户密码 |
| `NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD` | 管理员密码 |

完整配置说明位于 [webpage/docs](webpage/docs)，其中包括认证、桌面、媒体、
WebRTC、插件和反向代理配置。

## 数据持久化

建议将容器内 `/home/neko/.local/share/neko` 挂载到 Docker volume 或宿主机目录，
用于保存：

- 聊天历史；
- 用户头像；
- 其他 Neko 应用数据。

Chromium 的书签、扩展和浏览器历史位于 Chromium profile 目录。如需保留浏览器
环境，还应额外挂载 `/home/neko/.config/chromium`。

## 仓库结构

```text
apps/chromium/       Chromium 启动配置和代理集成
client/              TypeScript/Vue 浏览器客户端
protocol/            输入协议和共享契约
runtime/             Xorg、PulseAudio 和容器运行时
server/              Go 服务端、WebRTC 和 REST API
demo/                本地构建及 FRP/TURN 示例
webpage/             项目文档站点
```

## 开发检查

服务端编译：

```bash
cd server
go build ./...
```

前端编译：

```bash
cd client
npm run build
```

文档站点编译：

```bash
cd webpage
npm ci
npm run build
```

协议输入说明见 [protocol/README.md](protocol/README.md)。项目当前的 M1、M2、
M3 实际完成情况和后续工作见 [REFACTOR_PLAN.zh-CN.md](REFACTOR_PLAN.zh-CN.md)。

## 许可证与安全

项目使用 [Apache License 2.0](LICENSE)。安全问题请参考
[SECURITY.md](SECURITY.md)，不要在公开 Issue 中发布敏感信息。
