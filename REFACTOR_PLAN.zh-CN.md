# Neko 重构计划

## 1. 背景与目标

Neko 是一个在容器内运行浏览器或 Linux 桌面、通过 WebRTC 向多个用户实时传输音视频并接收控制输入的项目。

本计划的目标不是重写 WebRTC、GStreamer 或容器运行时，而是在保持单机 Docker 部署兼容的前提下，按如下优先级改进：

- 首先降低端到端延迟、首帧时间、CPU/内存占用与多人观看时的媒体资源消耗；
- 首先将 WebRTC/NAT/TURN 连接收敛为少端口、可自检、可回退的部署方式；
- M1 仅支持 Chromium 引擎及其官方 `ghcr.io/m1k1o/neko/chromium` 镜像，不承诺 Firefox、Google Chrome、Edge 或其他桌面应用兼容性；
- M1 支持 Chromium 的受控出站代理，包括 HTTP CONNECT Basic 认证与 SOCKS5 用户名/密码认证；
- M1 支持 Linux x86_64 Docker Engine，以及 Windows x86_64 上 Docker Desktop/WSL2 的 Linux 容器部署；不提供 Windows 原生媒体 Worker；
- 将房间、成员、会话、权限和控制权变成明确、可测试的领域模型；
- 形成单一、明确且可校验的 REST 与实时信令契约；
- 为 OIDC/LDAP、持久化、多房间调度和可观测性建立扩展点；
- 渐进升级前端，使媒体连接逻辑不再耦合到页面组件。

不在本计划范围内：将核心后端全量改写为另一种语言、把实时媒体强制拆成微服务、实现完整企业 IAM 或重写现有应用镜像。

M1 中“代理”特指容器内 Chromium 访问外部网站时的**出站代理**，不等同于 Neko 的反向代理、SakuraFrp 端口映射或 TURN 媒体中继。这三类网络路径必须分别配置与诊断。M1 支持 HTTP CONNECT Basic 与 SOCKS5 用户名/密码认证；PAC、NTLM、Kerberos 与双向 TLS 代理不在 M1 范围内。

### 1.1 第一优先级的可量化目标

在阶段 0 的基线确定后，按当前生产规格设定并冻结目标值。默认建议以 `1280x720@30`、2 个观看者和 1 个控制者为基准：

- WebRTC 连接成功率不低于 99%（可直连网络），首次视频帧 P95 不高于 3 秒；
- 有效媒体路径上的端到端延迟 P95 不高于 300 ms；
- 默认单端口 UDP MUX 部署；只有 UDP 受阻时使用同端口 TCP MUX 或 TURN；
- Chromium 经过带账号密码的 HTTP CONNECT 或 SOCKS5 出站代理时，网页访问正常，Neko 信令、容器本地地址和 WebRTC 候选收集不被误代理；
- 黑屏/断连诊断在管理界面中可归因，而非仅依赖容器日志；
- 码率、帧率、分辨率根据带宽和 CPU 压力主动降级，避免无限排队造成秒级延迟。

这些是发布门槛而非承诺常数：跨洲、TURN 中继、弱网及高分辨率场景应使用独立的性能预算。

## 2. 技术选型

| 范围 | 选择 | 原因 |
| --- | --- | --- |
| 控制面与媒体 Worker | Go | 项目现有语言；与 Pion WebRTC、Linux/Docker、并发模型契合。 |
| 浏览器客户端、管理端、SDK | TypeScript | 浏览器 WebRTC API 原生支持良好，可共享协议类型与校验。 |
| 存储 | PostgreSQL + Redis（可选） | PostgreSQL 保存持久业务数据与审计；Redis 支撑 TTL 租约、短期会话、事件分发。 |
| 实时媒体 | 保留 Pion + GStreamer | 不改动稳定且性能敏感的媒体链路，只将其封装为适配器。 |
| TURN | Coturn | 在受限 NAT/UDP 网络中作为中继回退。 |
| Chromium 出站代理 | `ProxyProvider` + 每房间本地代理 Agent | 支持 HTTP CONNECT Basic、SOCKS5 用户名/密码认证；凭据不进入 Chromium 命令行、profile 或日志。 |
| 运行平台 | Linux x86_64；Windows x86_64 + Docker Desktop/WSL2 | 统一使用 Linux 容器及 GStreamer/Pion 运行时，避免维护 Windows 原生媒体栈。 |

Rust 只在将来出现明确的原生性能或安全瓶颈时，用于局部采集/输入适配模块；不作为主服务端替代方案。

## 3. 目标架构

```text
Browser Client / Admin Console
   |  HTTPS, WebSocket                 | WebRTC (media/data)
   v                                   v
+------------------ Control Plane -------------------+     +------------------+
| API, OIDC adapter, rooms, permissions, scheduling  |---->| TURN (optional)  |
| PostgreSQL, Redis, audit, metrics                   |     +------------------+
+-----------------------------+----------------------+              ^
                              | provision / health                  |
                              v                                     |
                  +----------- Media Worker / Room -----------------+
                  | Pion, GStreamer, X11/Wayland, input, container  |
                  | Chromium + optional egress-proxy policy          |
                  +--------------------------------------------------+
```

第一阶段仍是一个可执行的模块化单体：控制面和 Worker 可以在同一进程/容器中运行。后续仅在多房间或多节点确有需求时，将 Worker 独立为可调度单元。

### 3.1 服务端目录目标

```text
server/internal/
  domain/          # Room、Member、Session、ControlLease 等纯模型和规则
  application/     # JoinRoom、GrantControl、ProvisionRoom 等用例
  ports/           # Repository、Authenticator、Worker、EventPublisher 等接口
  adapters/
    http/          # REST/OpenAPI
    websocket/     # 实时信令协议适配器
    webrtc/        # Pion 适配器
    capture/       # GStreamer/桌面采集适配器
    runtime/       # Docker/Kubernetes Worker 实现
    persistence/   # memory/file/PostgreSQL/Redis 实现
    auth/          # local/OIDC/LDAP 实现
  observability/   # metrics、tracing、structured logs
```

### 3.2 核心状态机

房间：`Provisioning -> Ready -> Active -> Draining -> Terminated`，任意状态可进入 `Failed`。

控制权：使用具有 `holder`、`epoch`、`expiresAt` 的 `ControlLease`。所有抢占、续租、释放操作必须幂等，并通过单调递增的 `epoch` 拒绝过期客户端事件。

## 4. 分阶段实施

### 阶段 0：性能与连通性基线（1–2 周）

1. 固定 M1 唯一应用镜像为 `ghcr.io/m1k1o/neko/chromium`；建立 Linux x86_64 与 Windows x86_64 Docker Desktop/WSL2 的 Chromium、软件/硬件编码能力矩阵。
2. 建立可重复压测矩阵：720p/1080p、1/2/5 个观看者、直连/UDP 阻断/TURN/FRP、软件/硬件编码，以及无代理/HTTP CONNECT Basic/SOCKS5 用户名密码出站代理。
3. 采集端到端延迟、首帧时间、帧率、编码耗时、队列深度、码率、丢包、CPU、内存、GPU 与出口带宽。
4. 固定镜像、Go、Node、Pion、GStreamer 和前端依赖版本；为每次构建产出可追溯镜像标签。
5. 梳理 EPR、UDP MUX、TCP MUX、NAT 1:1、STUN/TURN/FRP 的实际行为和默认配置，记录可复现的黑屏案例。

验收：关键场景有可对比的性能/连通性数据；仪表盘能够按 `room_id`、`peer_id`、连接方式查看指标；确定阶段 1 的优化目标。

### 阶段 1：媒体性能优化与少端口连接（3–5 周）

1. 将 WebRTC 网络配置收敛为 `ConnectivityProvider`，默认采用 UDP MUX；为受限网络启用同端口 TCP MUX；TURN 仅作策略允许的回退。
2. 启动前执行预检：监听端口、Docker 端口协议、候选公网 IP、NAT 1:1 设置、STUN/TURN 可用性；预检失败阻止实例进入 Ready 状态。
3. 将捕获、编码与 WebRTC 发送链路拆出清晰的背压边界：队列超过时丢弃旧视频帧而非持续累积，确保交互延迟优先于逐帧完整性。
4. 统一媒体质量策略：以网络估算、编码耗时和 Worker CPU/GPU 压力为输入，按 `分辨率 -> 帧率 -> 码率` 顺序降级与恢复；每次切换设置冷却时间，避免抖动。
5. 抽象 Chromium 所需的编码器能力（软件、VAAPI、NVENC、Intel QSV 等）；运行时能力探测与健康回退，缺少硬件设备时自动使用软件编码。
6. 对每个 peer 记录 ICE、DTLS、首帧、RTCP、重连和关闭原因；客户端展示可操作诊断结果及推荐措施。
7. 实现 `ProxyProvider` 与每房间本地代理 Agent：Chromium 仅连接本地 Agent，Agent 再连接带账号密码的上游 HTTP CONNECT 或 SOCKS5 代理。地址、账号和密钥分别由配置与 Secret 注入，凭据不得出现在 Chromium 命令行、持久化 profile、管理 API 或日志中。
8. Neko 服务、信令 WebSocket、容器回环地址、Docker 网关和 STUN/TURN/FRP 控制端点默认进入 `NO_PROXY`/绕过列表；不得依赖全局 `HTTP_PROXY`、`HTTPS_PROXY` 环境变量驱动整个容器。
9. 提供 Linux Docker Engine 与 Windows Docker Desktop/WSL2 的部署模板；Windows 模板说明 Docker Linux 容器模式、Windows Defender Firewall、SakuraFrp 客户端到 Docker 服务的本地可达性及 GPU 加速限制。
10. 提供四类官方部署模板：公网 UDP MUX、UDP+TCP MUX、Coturn 回退、FRP/SakuraFrp 同端口 UDP+TCP 隧道；模板使用明确端口与防火墙清单。

验收：默认公网部署仅需 HTTP(S) 端口和一个 WebRTC MUX 端口；FRP 模式可在无公网 IP 的内网主机运行；Chromium 在无代理、HTTP CONNECT Basic、SOCKS5 用户名/密码三种配置下均能访问测试站点，代理凭据不会泄漏到日志或管理 API，且信令和 WebRTC 不被代理配置破坏；Linux x86_64 与 Windows x86_64 Docker Desktop/WSL2 均通过基础端到端场景；弱网和 CPU 压力下不发生无界延迟累积；满足第 1.1 节的基准性能门槛。

### 阶段 2：协议契约与领域抽取（2–4 周）

1. 保留现有 OpenAPI，并以 OpenAPI generator 生成 TypeScript API client。
2. 为 WebSocket 信令固定唯一 envelope：`{ event, payload }`；废弃的扁平消息和兼容桥不进入运行时。
3. 使用 JSON Schema 或 Protobuf 定义实时事件；在 Go 和 TypeScript 中生成类型及运行时校验。
4. 抽取 `Room`、`Member`、`Session`、`Permission`、`ControlLease`，禁止 handler 直接修改共享 map。
5. 所有命令返回规范错误码，例如 `ROOM_NOT_READY`、`CONTROL_CONFLICT`、`STALE_EPOCH`、`ICE_FAILED`。

验收：协议不兼容时返回明确错误；领域层不依赖 HTTP、WebSocket、Pion、Docker；发布前删除已废弃的协议分支、类型和测试夹具。

### 阶段 3：模块化服务端与持久化（3–6 周）

1. 按目标目录逐步把现有实现迁入 application/ports/adapters。
2. 实现内存、文件、PostgreSQL repository；先双写/影子读，再切换为主存储。
3. 引入 Redis 作为可选 ControlLease 和事件总线实现；单机模式维持内存实现。
4. 引入认证接口，先兼容 multiuser/file/object，再添加 OIDC；LDAP 后续单独交付。
5. 以特性开关控制新存储、新认证和新房间状态机。

验收：原有 Docker Compose 无需数据库也能运行；启用 PostgreSQL 后，用户、房间元数据、审计与会话策略可持久化；重启不会产生悬挂控制权。

### 阶段 4：前端 SDK 与渐进升级（3–6 周）

1. 提取 `@neko/protocol`、`@neko/sdk`、`@neko/ui` 三个包。
2. `@neko/sdk` 统一管理认证、REST、WebSocket、PeerConnection、重连和设备权限；组件只订阅状态和发起命令。
3. 把 Vuex 状态拆分为房间、连接、媒体、UI 四个状态机；避免媒体对象放入响应式全局状态。
4. 从 Vue 2/Vue CLI 迁移至 Vue 3、Vite、Pinia；每次只替换一个页面或功能域，完成切换后立即删除旧实现。
5. 为 SDK 提供嵌入式 API，支持第三方产品以受控方式创建/加入房间。

验收：UI 框架升级不改变 WebRTC 信令；SDK 能被独立测试；已弃用的页面和适配层不再保留。

### 阶段 5：Worker 调度与多房间（按需求，4–8 周）

1. 引入 `WorkerRuntime` 接口：Docker 为默认实现，Kubernetes/Nomad 为可选实现。
2. 控制面按 roomId 调度 Worker，保存 Worker endpoint、健康状态、容量与版本。
3. 所有 WebSocket/API 请求按 roomId 粘性路由；WebRTC PeerConnection 始终固定在同一 Worker。
4. 支持优雅排空：停止新用户进入、通知客户端、迁移或终止房间。
5. 不将视频流经过控制面；客户端直连 Worker 或经 TURN。

验收：Worker 单点故障仅影响其承载房间；控制面故障恢复后可重新发现 Worker；多节点不产生跨房间串流或控制权泄漏。

## 5. 测试与发布策略

| 层级 | 内容 |
| --- | --- |
| 单元测试 | 房间状态机、权限矩阵、ControlLease、配置解析、鉴权策略。 |
| 契约测试 | OpenAPI、WebSocket schema、Go/TypeScript 生成类型的兼容性。 |
| 集成测试 | Docker + Chromium Neko + Coturn + FRP + 需认证的 HTTP CONNECT/SOCKS5 proxy；PostgreSQL/Redis 为后续可选组合。 |
| 端到端测试 | Linux x86_64 与 Windows Docker Desktop/WSL2：Chromium 房间登录、媒体首帧、轮流控制、刷新、断线恢复和出站代理访问。 |
| 故障测试 | UDP 阻断、错误 NAT IP、端口耗尽、TURN/FRP 不可达、代理认证失败、错误 bypass 规则、Worker 崩溃、Windows 主机防火墙拦截。 |
| 性能测试 | Chromium 720p/1080p、多参与者、不同编码器、直连/FRP/TURN、无代理/有代理；记录采集到发送的队列深度、CPU/GPU、出口带宽、P95 首帧与端到端延迟。 |

采用语义化版本和迁移文档。每次仅发布一个可回滚的架构变化；数据库迁移遵循 expand → migrate → contract，协议变更通过明确的迁移步骤完成，不长期保留旧分支。

## 6. 可观测性与安全基线

- Prometheus：room 数、peer 数、ICE 成功率、候选类型、首帧时间、RTT、jitter、重传、帧率、编码耗时、队列深度、重连、CPU、内存、GPU、容器生命周期。
- OpenTelemetry：REST 请求、房间 provisioning、认证、信令协商、Worker 调度链路。
- 结构化日志：强制带 `room_id`、`member_id`（脱敏）、`peer_id`、`worker_id`、`request_id`。
- 默认安全配置：HTTPS/WSS、强密码、非 root 容器、最小 Linux capabilities、禁用调试端点、受限 CORS、密钥外置。
- 代理凭据仅通过 Secret 注入；管理 API 只能返回已脱敏的代理 URL、类型、绕过规则和健康状态；本地代理 Agent 使用最小权限并只接受来自 Chromium 容器网络的连接。
- 审计事件：登录、创建/终止房间、控制权变动、管理员配置变更、文件/剪贴板策略变更。

## 7. 风险与决策原则

| 风险 | 应对 |
| --- | --- |
| 重构影响媒体稳定性 | 不重写 Pion/GStreamer；先通过接口包裹，保留旧路径与回滚开关。 |
| 微服务化增加时延和运维成本 | 前三阶段保持模块化单体；只在多节点需求明确时拆出 Worker。 |
| 新协议破坏旧客户端 | 发布前完成客户端升级和迁移检查；服务端对不支持的协议返回明确错误，不保留旧协议运行时分支。 |
| 数据库引入破坏轻量部署 | 存储接口与内存默认实现并存；PostgreSQL/Redis 均为可选。 |
| TURN 成本不可控 | 直连优先；将 TURN 作为策略可控回退，并按房间统计 relay 流量。 |
| FRP 节点 IP/端口变化 | 将 FRP 公网端点设为显式配置；启动时校验 Neko 的 NAT 候选端口与 FRP 同端口 TCP/UDP 隧道一致。 |
| 全局代理劫持控制面 | Chromium 采用显式代理参数；服务端和 WebRTC 端点采用明确 bypass，禁止将全局代理变量透传给所有子进程。 |
| 前端大迁移长期分叉 | 先提取 SDK 与共享协议，UI 逐页迁移，禁止并行维护两套媒体实现。 |

## 8. 当前实施状态（2026-09-09）

### 已完成

- M1 范围清理：删除 Firefox、其他 Chromium 衍生浏览器及桌面应用的镜像资产、旧运行时开发配置、对应静态图标、多应用发布矩阵和失效的 v2 文档归档；默认 Compose、构建入口与当前版本文档统一为 Chromium 和单端口 UDP/TCP MUX。
- `747c26a1`：固化 M1 的 Chromium、性能、FRP、认证代理和平台范围。
- `38887ae6`：新增纯 Go 媒体端口计划，校验直连 MUX 与 FRP 的同端口 TCP/UDP 约束。
- `a096b0c3`：新增 `webrtc.connectivity.mode=frp` 启动前预检；FRP 模式要求显式唯一 NAT IP、同端口 MUX，且拒绝 EPR 混用。
- `6ce2169c`：新增认证出站代理配置模型，支持 HTTP CONNECT Basic 与 SOCKS5 用户名/密码，拒绝 URL 内嵌凭据并提供脱敏诊断地址。
- `13e2928d`：Chromium 镜像改用本地代理 Agent 启动包装器；仅允许 loopback Agent 地址，保留 NVIDIA 初始化入口。
- `86e9be01`：实现独立 `neko-proxy` Agent，完成 HTTP CONNECT Basic 与 SOCKS5 用户名/密码认证转发；密码从 Secret 文件读取，并接入 Chromium Supervisor 生命周期。
- `bfe2928f`：新增显式目标的代理认证预检、周期健康检查和回环 `/healthz` 脱敏诊断；首次检查失败会阻止 Chromium 启动。
- `47e7b8ac`：新增真实 Squid Basic 与 microsocks 用户名/密码代理的 Docker Compose 集成测试，并接入 PR 默认 CI。
- `f155a767`：捕获和 WebRTC Track 改用有界的“最新帧优先”队列；队列满载时淘汰旧帧而非阻塞编码或保留过时画面，并公开捕获侧及按会话音/视频 Track 划分的队列深度、淘汰计数指标。
- `b8987638`：新增显式 `low`、`balanced`、`high` Chromium 画质档位；生成 VP8/软件 H.264 标准管线，并在启动时拒绝与自定义或旧版视频管线参数混用。
- `fea72230`：新增 `auto`、`software`、`vaapi`、`nvenc` 编码器选择和进程内 GStreamer 元素探测；H.264 元素缺失时按确定顺序回退至 x264 或 VP8，并输出脱敏诊断。
- `e9143cf6`：捕获层支持同 codec 的运行时管线候选；硬件 H.264 管线创建或进入播放状态失败时重试 x264，不在 WebRTC 协商后切换 codec。
- `b25ef523`：GStreamer 播放启动等待状态转换完成；硬件设备初始化失败会在首个观看者创建管线时返回错误，从而触发同 codec 回退。
- `70a0706b`：配置阶段对 VAAPI/NVENC 执行无显示依赖的短 GStreamer 运行时探测；设备或驱动初始化失败时在选择编码器前记录原因并尝试同 codec 软件回退，同时保留硬件元素的备选探测顺序。
- `07c30420`：标准 Chromium M1 质量管线扩展为 VP8/H.264/H.265/AV1；新增 AV1/H.265 的软件、VAAPI 和 NVENC 编码候选，统一设备探测与同 codec 软件回退。
- `1816aecd`：增加 AV1/H.265 编码配置、硬件候选、软件回退和运行时探测的单元测试与竞态覆盖。
- `02b616c8`：增加 AV1/H.264/H.265 RTP codec 解析与 Pion 注册测试。
- `d402b3df`：新增可在目标运行时执行的编码器能力矩阵脚本、矩阵说明和配置文档。
- `ecc1dff6`：在运行时镜像中提供 GStreamer CLI 工具，并补齐 Intel VAAPI 开发配置。
- `5ff71fa4`：为编码矩阵增加 `software`、`vaapi`、`nvenc` 族筛选，支持硬件发布作业按 GPU 类型严格执行。
- `08dfc2ef`：补齐 VAAPI 低功耗 H.264 配置生成并增加回归覆盖。
- `231b6d3c`：将 H.265 SDP 能力提升到 Main Level 4.0（`level-id=120`），覆盖 `high` 档位 1080p30，并加入能力断言测试。
- 动态编码回退：Chromium 客户端在 `signal/request.payload.video_codecs` 上报接收能力；服务端按配置 codec、H.264、VP8 的顺序选择可用编码，并为 M1 profile 建立惰性 codec 采集变体，避免浏览器不支持 H.265/AV1 时在 SDP answer 阶段失败重连。

### 已验证

- `go test ./internal/connectivity ./internal/proxy` 通过。
- `apps/chromium/neko-chromium_test.sh` 通过。
- 内存测试服务组成的 HTTP CONNECT 与 SOCKS5 认证代理集成测试通过。
- `go test -race ./internal/proxy` 通过。
- `go build -o /tmp/neko-proxy ./cmd/neko-proxy` 通过。
- 安装本机构建依赖后，`go test ./...`、`go test -race ./pkg/mediaqueue ./internal/webrtc` 均通过。
- Chromium 启动脚本、Shell 语法、Compose/GitHub Actions YAML 解析和 `git diff --check` 通过。
- Windows x86_64 + WSL2 Docker 环境的 Chromium 演示通过：账号密码登录、`1280x720@30` 画面及同端口 `52000/TCP+UDP` WebRTC 链路可正常运行。
- 真实 Squid Basic 与 microsocks 用户名/密码 Docker Compose 套件通过，覆盖健康检查、HTTP 转发、CONNECT 隧道、错误凭据和诊断脱敏。
- 显式 `balanced` 档位通过 `go test ./...`、目标包竞态检测和实际 Chromium 容器验证；最终 VP8 管线稳定输出整数码率 `2500000`，容器健康且无重启。
- 编码器选择与回退通过单元/竞态测试；默认 VP8 演示通过真实 GStreamer registry 探测与管线语法检查，仍保持同端口 `52000/TCP+UDP`。registry 探测只代表元素可见，GPU 设备及驱动初始化仍需下一小步验证。
- 硬件编码运行时探测通过 `videotestsrc -> encoder -> h264parse -> fakesink` 短管线验证；无可用设备或驱动时会在配置阶段回退软件编码，并通过配置/质量单元测试覆盖失败诊断。
- 运行时管线候选已接入 `StreamSink`；同 codec 的硬件失败可重试软件编码，所有候选失败会返回明确错误并保留现有会话生命周期。
- 当前工作区 `go test ./...` 通过；新增 AV1/H.265 配置、编码器选择、运行时回退和矩阵覆盖测试通过。当前环境没有可用 GStreamer CLI 或 GPU 设备，因此真实硬件行需在映射 `/dev/dri` 或 `--gpus all` 的目标运行时执行 `server/integration/encoding/matrix.sh`。
- `go test -race ./internal/quality ./internal/config ./pkg/gst ./pkg/types/codec`、`go vet ./...`、编码矩阵脚本 Shell 语法和软件/VAAPI 族筛选模拟执行通过；真实 GPU 行仍需在目标运行时归档 TSV 结果。
- `StreamSink` 增加首帧耗时、实际样本率、实际码率、样本总数和管线回退计数指标；码率状态改为读写锁保护，避免多观看者并发访问竞态。
- 动态回退通过配置单元测试、Go 全量测试、TypeScript lint/build 验证；在 H.265 1080p30 演示容器中，模拟仅支持 H.264/VP8 的浏览器信令请求已选择 H.264 管线并生成 H.264 SDP offer。
- 带宽估计器接入视频 Track 队列占用、RTCP jitter 和累计丢包；压力持续超过滞回时降档，压力存在时禁止升档，并将切换原因写入日志。
- `83341b33`：将队列压力、RTCP jitter 和新增丢包接入 WebRTC 自适应升降档；网络压力持续超过不稳定时长才降档，且压力期间禁止升档。
- 新增 `demo/compose.frp.example.yaml`、`demo/frpc.toml.example` 和 `demo/compose.turn.example.yaml`，覆盖 SakuraFrp/FRP 同号 TCP+UDP 媒体隧道及 Coturn relay 端口模板，凭据均使用占位符。
- `45e5c91c`：新增本地 FRP/Coturn 连通性集成套件；FRP 验证无公网主机模型下同号 `52000/TCP+UDP` 转发，Coturn 通过真实认证 Allocate 验证 UDP/TCP `3478` 和 `49160-49170` relay 范围，并在失败时输出端口、认证或 relay 范围诊断。
- 浏览器 E2E 入口的 Shell/JavaScript 语法、Git 差异和客户端 lint 已通过；完整 Playwright 镜像首次下载及真实媒体回归需在具备镜像缓存和运行中 Chromium demo 的 Linux/WSL2 环境执行。
- `ad9b2032`：新增固定 Playwright Chromium 的真实浏览器 E2E 入口；覆盖账号密码登录、`{event,payload}` 信令 envelope、`system/init`、连接状态、远端视频首帧和废弃事件检测，并提供 1/2/5 观看者并发基线脚本及 JSON 结果归档。
- `0a5d2adf`：清理文件传输 UI 遗留的 `/file?pwd=...` 路径；上传、下载和删除统一使用带现有会话认证的 `/api/filetransfer`，不再把密码放入 URL。
- `9f36fde0`：E2E 基线入口增加可选 Prometheus `/metrics` 快照，并在并发基线中按观看者保存 JSON 与指标文件，便于将浏览器首帧数据和服务端队列/码率/ICE 指标对齐。
- `03dbcd7f`：客户端 `SignalingTransport` 与 Go `WebSocketMessage` 共享严格 envelope 规则；拒绝空/空白事件、`payload:null` 和额外顶层字段，同时保留合法数组 payload，避免重新引入已删除的扁平信令。
- `32fcf57e`：将前端 WebRTC `getStats`、RTT/丢包采样和网络质量分类提取为可复用的 `NetworkQualityMonitor` SDK；`NekoClient` 只负责把采样结果映射到状态模块，并导出给嵌入式客户端使用。
- `df63951e`：增加无额外测试依赖的 TypeScript SDK 合约测试，覆盖网络质量阈值、严格信令 envelope 和合法数组 payload，并纳入客户端 CI 构建工作流。
- `792d4f79`：将登录、登出和 Bearer token 注入提取为独立 `AuthClient`；`NekoClient` 不再直接管理认证 REST 细节，SDK 合约测试覆盖 token 注入与登出清理。
- `40f8dd1f`：删除客户端根状态中从未被读取的 `active/setActive` 字段，避免保留无效会话状态。
- `ad25e67c`：将显示名、认证密码、管理员锁和登录/登出动作迁入 namespaced `session` Vuex 模块；组件与 `NekoClient` 不再读取或写入根状态，完成一批会话/UI 状态边界拆分。

### 当前限制与下一步

- 已实现：为 H.264 VAAPI/NVENC 增加 GPU 设备及驱动初始化探测，并在实际管线无法进入可用状态时回退；仍需在真实 VAAPI/NVENC GPU 主机上完成硬件能力矩阵验证。
- 已实现：为 FRP/TURN 模板增加本地可重复的无公网 IP、relay 端口和故障诊断自动化测试；仍需在真实公网 FRP 节点、UDP 受阻网络和实际 Neko 媒体端点上补充跨网络验证。
- 已实现：真实浏览器 WebRTC 回归和性能基线采集入口；仍需在 Linux x86_64 与 Windows x86_64/WSL2 的运行中 demo 上执行 720p/1080p、1/2/5 观看者矩阵，并将结果与 `/metrics` 资源数据关联。
- 下一项实现：执行并固化 720p/1080p 多观看者性能基线，随后补充公网 FRP、UDP 受阻和 TURN 媒体端到端场景。

### UI 重构进行中

- 当前批次先行推进 UI 的视觉与交互基础层：深色蓝灰设计令牌、视频舞台容器、现代化登录卡片、语义化工具栏、响应式侧栏标签，以及键盘可访问的播放/音量/控制操作。
- 本轮优先项已完成：连接状态与 RTT/丢包质量指示、移动端侧栏抽屉与遮罩、核心 UI 文案国际化、登录表单行内校验与密码显隐、聊天/文件空状态及文件加载状态、设置页按播放/聊天/输入/管理/会话分组。
- 本批次只调整展示层、交互语义和布局，不改变 WebRTC 信令、媒体协议、连接状态机或端口契约；深层 SDK/状态拆分仍放在 M1 后段。
- `44ee598c`：完成 UI 基础层；`ee321149`、`5b123cfb`、`4193bbad`、`b172f693`、`aa23ebc3`：完成本轮状态反馈、移动端抽屉、设置分组、国际化、登录反馈和空/加载状态；`npm run lint -- --no-fix` 无错误，`npm run build` 构建成功，Docker 演示容器健康检查通过。构建仅提示既有未使用类型和 bundle 体积 warning，未阻断交付。

### 深层 SDK、信令与状态拆分（第一批已完成）

- `6ce487ba`：提取 `SignalingTransport`、`ConnectionStateMachine` 和媒体输入编码器；WebSocket 生命周期具备代际校验、JSON envelope 校验、发送结果反馈和异步错误归一化；PeerConnection 的 ICE、DataChannel、重协商和连接回调不再直接依赖页面组件。
- `015a2622`：媒体二进制协议统一为网络字节序并增加服务端长度校验；严格校验 EPR 端口范围，并拒绝 EPR 与 TCP/UDP MUX 的冲突组合。
- `26cd89e9`、`7c0fed2a`、`04725c07`：信令统一使用唯一的 `{ event, payload }` envelope；WebSocket 打开后主动发送 `signal/request`，使用标准 `iceservers` 字段和原生 ICE candidate。随后已删除扁平消息解析、legacy bridge 及其旧类型，避免运行时长期维护两套协议。连接状态模块将 `connected` 会话存活语义与 `reconnecting` 生命周期标签分离，避免短暂断网时 UI 销毁媒体会话。
- 当前增量：WebRTC 服务端在 ICE `disconnected` 后增加 5 秒代际保护窗口，恢复连接会取消销毁计时器，只有持续断开或进入 `failed` 才释放 Peer，避免网络切换期间被立即清理。
- 当前清理：删除旧 HTTP/WebSocket legacy 包、旧版 WebRTC data-channel handler、V2 配置迁移入口、旧媒体滚动编码和未使用的旧消息类型；服务端只创建 `data` channel，严格拒绝扁平/空 `payload` envelope，并对无 payload 事件统一省略 `payload` 字段。文件传输锁通过 `/api/room/settings` 更新插件设置，不再保留无效的 UI 操作。
- `02d2eaef`：客户端只保留 `/api` 认证、`/api/ws` 信令和现代事件/消息类型，删除客户端创建 data channel、扁平消息解析及废弃事件。
- `08536a32`：服务端删除 legacy HTTP、legacy WebRTC handler、V2 配置迁移入口、旧滚动编码和旧消息兼容字段。
- `2d871384`：同步迁移文档、Roadmap、配置生成脚本和开发要求，明确升级时不保留旧运行时分支。
- 验证：`vue-cli-service lint --no-fix` 与 `vue-cli-service build` 成功（仅有既有 Browserslist、bundle 体积提示）；WSL2 Go 工具链下 `go test ./...` 通过，覆盖完整 config/session/capture 与新增 AV1/H.265 选择逻辑。真实硬件管线仍需在映射 `/dev/dri` 或 `--gpus all` 的目标运行时执行矩阵脚本。

本批次没有改变公网端口号或 MUX 配置语义；服务端重连宽限/去抖已完成。下一批补充生成式契约测试和迁移错误码，不再引入第二套信令 envelope。

### M1 后续开发执行计划

M1 的 UI 工作拆为两层：当前先交付不触及媒体链路的视觉与交互基础；媒体性能、网络基础设施和发布基线收敛后，再进行深层 SDK/状态拆分。后者采用可回滚的渐进式迁移，不改变信令、媒体协议和服务端端口契约。

| 顺序 | 增量 | 主要交付 | 验收与提交边界 |
| --- | --- | --- | --- |
| 1 | UI 视觉与交互基础（第一批已完成） | 深色蓝灰设计令牌、视频舞台、登录卡片、连接/网络质量反馈、移动端侧栏抽屉、国际化、空/加载状态、设置分组和可访问性基础；保持旧状态与协议接口 | 变更可通过提交回滚；不改变信令/媒体协议；lint/build 通过，Chromium 基础交互回归通过 |
| 2 | 显式质量 Profile（已完成） | `low`、`balanced`、`high`；仅显式启用；拒绝与自定义 GStreamer 管线混用 | 配置和管线生成单测通过；历史默认配置不变 |
| 3 | 编码器能力与回退（AV1/H.265 已接入，真实硬件矩阵待执行） | VP8、H.264、H.265、AV1 的软件/VAAPI/NVENC 候选；配置阶段探测硬件运行时能力，首个媒体管线保留同 codec 回退；矩阵脚本覆盖软件、VAAPI、NVENC 编码器元素 | 缺失 GPU/插件时可诊断并回退；需在真实 VAAPI/NVENC GPU 主机和 Windows/WSL2 透传环境执行矩阵并归档结果 |
| 4 | 自适应质量策略 | 显式 profile 生成质量梯度，复用带宽估计器；后续加入队列压力、RTT/jitter/丢包输入 | 压力下降档、恢复升档，切换原因可观测 |
| 5 | 性能指标闭环 | 编码耗时、首帧、实际帧率/码率、路径标签和资源指标 | `/metrics` 覆盖基线指标且不含高风险凭据标签 |
| 6 | FRP 与 TURN 集成（本地套件完成，公网矩阵待补） | SakuraFrp 同端口 TCP/UDP 模板、Coturn 回退模板、relay 范围校验和故障诊断套件 | 本地两条路径可重复部署并通过连通性测试；真实公网/UDP 受阻/Neko 媒体链路需补充 |
| 7 | 基线与发布验收（入口已完成） | 固定 Playwright Chromium 的登录/信令/首帧 E2E、1/2/5 观看者并发脚本、JSON 结果；执行 720p/1080p 与 Linux/WSL2 矩阵 | 首帧/连接/分辨率数据可比较；运行中 demo、GPU/公网和 UDP 受阻矩阵逐项关闭 |
| 8 | UI 深层 SDK/状态拆分（第一批已完成，持续迭代） | 已提取 TypeScript 信令传输、连接状态机、媒体输入编码器，并把连接状态迁入 namespaced 模块；下一批继续拆分房间/媒体/UI 状态，完成切换后删除旧页面和适配层 | 当前批次完成 lint/build、纯 Go 单测；后续通过 Chromium、认证代理、FRP/TURN、端口和性能回归，并检查无废弃运行时路径 |

## 9. 里程碑与成功标准

1. **M1：Chromium 性能、认证代理与单端口连通性**：仅支持 Chromium；支持 Linux x86_64 和 Windows x86_64 Docker Desktop/WSL2；默认 UDP MUX、TCP/TURN/FRP 回退、启动预检、带认证的 HTTP CONNECT/SOCKS5 出站代理、媒体背压和质量策略完成，并通过性能门槛；UI 视觉基础可先行，基础能力收敛后再于 M1 后段执行深层 UI SDK/状态拆分，不改媒体协议。
2. **M2：可测试的契约与领域核心**：状态机、协议 schema、双端类型生成和基础 CI 完成。
3. **M3：可持久化、可集成认证的模块化后端**：内存部署兼容；PostgreSQL/OIDC 为可选生产能力。
4. **M4：独立客户端 SDK**：在 M1 后段 UI 提取基础上完成跨框架、可独立发布的客户端 SDK；前端框架升级不触及媒体协议，嵌入式集成可复用 SDK。
5. **M5：按需扩展的房间 Worker**：在多节点环境中安全调度、粘性路由和优雅排空。

最终成功标准：保留单机 Docker 的易用性，优先让媒体链路在常见公网环境下以低延迟、少端口和可诊断的方式稳定工作；随后在认证集成、房间生命周期与多房间扩展上具备明确、可测试、可观测的工程能力。
