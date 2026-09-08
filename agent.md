# M1 开发要求：Chromium、性能与网络

## 目标

在不破坏现有单机 Docker 使用方式的前提下，为 Neko 建立 M1 能力：低延迟 Chromium 远程桌面、少端口 WebRTC 连接、无公网 IP 的 FRP 支持，以及带认证的 Chromium 出站代理。

M1 的优先顺序：

1. 媒体性能、低延迟与资源控制；
2. WebRTC 端口与 NAT 连接复杂度；
3. Chromium 出站代理；
4. 可观测性、诊断和自动化测试。

## M1 范围

- 唯一支持的应用镜像：`ghcr.io/m1k1o/neko/chromium`。
- 支持 Linux x86_64 的 Docker Engine。
- 支持 Windows x86_64 的 Docker Desktop/WSL2 Linux 容器模式。
- 支持 HTTP CONNECT Basic 认证和 SOCKS5 用户名/密码认证的 Chromium 出站代理。
- 支持公网 UDP MUX、同端口 UDP+TCP MUX、Coturn 回退，以及 FRP/SakuraFrp 同端口 UDP+TCP 隧道。
- 支持按网络、编码耗时和资源压力进行媒体质量降级，避免视频队列无限积压。

## 明确不在 M1 范围内

- Firefox、Google Chrome、Edge、其他 Chromium 衍生浏览器和其他 Linux 桌面应用。
- Windows 原生媒体 Worker 或 Windows 原生 GStreamer/桌面输入栈。
- PAC、NTLM、Kerberos、双向 TLS 代理。
- 多节点调度、Kubernetes、完整 OIDC/LDAP、数据库持久化和前端框架全面迁移。
- 重写 Pion、GStreamer、WebRTC 或容器运行时。

## 架构规则

- 后端继续使用 Go；浏览器客户端及共享协议使用 TypeScript。
- 优先模块化单体。不要为 M1 拆分微服务或让控制面转发媒体流。
- 媒体 Worker 必须保持 WebRTC PeerConnection 的房间粘性；客户端直连 Worker，或经 TURN/FRP 中继。
- 所有新增配置必须有默认值、显式校验、脱敏日志和文档示例。
- 不得用全局 `HTTP_PROXY`/`HTTPS_PROXY` 驱动整个 Neko 容器。

## 网络规则

- 默认使用一个可配置的 UDP MUX 端口；TCP MUX 可使用相同端口号作为回退。
- Docker 映射、Neko MUX 监听和 FRP 的本地/远程媒体端口必须使用相同数字；禁止端口重映射。
- FRP/SakuraFrp 模式中，Neko `NAT1TO1` 配置为 FRP 节点可达的公网 IP。
- HTTP(S)/WebSocket 页面入口可使用独立端口或反向代理；它不能替代 WebRTC 媒体端口。
- 不要把“客户端 UDP 不可达”作为 Worker 启动失败条件；它应是运行时诊断和 TCP/TURN 回退条件。

## 代理规则

- 实现 `ProxyProvider`，并使用每房间本地代理 Agent 承载上游认证。
- Chromium 只连接本地 Agent；上游代理 URL、用户名和密码仅通过 Secret 注入 Agent。
- 禁止将代理密码放入 Chromium 命令行、持久化 profile、管理 API、指标标签或日志。
- Neko API、信令 WebSocket、容器回环地址、Docker 网关和 STUN/TURN/FRP 控制端点必须加入 bypass 规则。
- 代理健康检查必须验证认证成功，并返回脱敏状态；失败不得泄漏凭据。

## 性能规则

- 先建立基线，再调整编码器、分辨率、帧率或码率；不凭经验固定压缩比例。
- 默认评估基准：`1280x720@30`、2 位观看者、1 位控制者。
- 指标至少包括：连接成功率、首帧时间、端到端延迟、RTT、jitter、丢包、编码耗时、帧率、队列深度、CPU、内存、GPU 和出口带宽。
- 捕获、编码和发送链路必须有有界队列；过载时优先丢弃旧视频帧，保证交互延迟。
- 先测试 H.264 硬件编码、H.264 软件编码与 VP8 回退；VP9/AV1 在证明 CPU、延迟和 Chromium 解码兼容性后才可成为可选 profile。
- 直连、FRP 和 TURN 的性能预算及结果必须分别统计，禁止混合为单一 SLO。

## 验收

- Linux x86_64 与 Windows Docker Desktop/WSL2 都通过 Chromium 基础端到端测试。
- 无代理、HTTP CONNECT Basic、SOCKS5 用户名/密码三种出站模式均能访问测试站点，且不影响 Neko 信令和 WebRTC。
- 公网、FRP、TURN 三种网络模式均有可重复的部署模板和诊断结果。
- 基线网络下，首次视频帧 P95 不高于 3 秒，端到端延迟 P95 不高于 300 ms；跨洲、FRP、TURN、弱网使用单独预算。
- 常见故障必须可区分：端口映射错误、错误 NAT IP、UDP 被阻断、TURN/FRP 不可达、代理认证失败、错误 bypass、Worker 故障。

## 交付纪律

- 每次变更聚焦一个可回滚能力，并添加或更新对应测试。
- 不修改现有用户配置语义；如必须变更，提供兼容层、迁移说明和特性开关。
- 任何依赖 GPU、Docker Desktop、FRP 服务商或代理类型的能力，都须注明前置条件与降级路径。
- 提交前执行格式检查、单元测试、集成测试（可用时）和 `git diff --check`。
