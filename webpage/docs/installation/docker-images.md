---
description: M1 Chromium image and supported runtime targets.
---

# Docker Image

M1 publishes one application image: `ghcr.io/m1k1o/neko/chromium`.
It runs on Linux x86_64 Docker Engine and Windows x86_64 through Docker
Desktop/WSL2 Linux containers. Other browser and desktop images are no longer
part of this distribution.

Use a `2gb` shared-memory allocation for Chromium. The default deployment uses
one configurable WebRTC media port for both UDP and TCP fallback; see the
networking guide for direct, TURN and FRP deployments.

```bash
docker run --rm --shm-size=2g \
  -p 8080:8080 -p 52000:52000/udp -p 52000:52000/tcp \
  -e NEKO_MEMBER_MULTIUSER_USER_PASSWORD=change-me \
  -e NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD=change-me-too \
  -e NEKO_WEBRTC_UDPMUX=52000 -e NEKO_WEBRTC_TCPMUX=52000 \
  -e NEKO_WEBRTC_NAT1TO1=<public-or-frp-ip> \
  ghcr.io/m1k1o/neko/chromium:latest
```

Hardware encoder availability depends on the host GPU, driver and container
device mapping. When unavailable, the image uses the configured software
fallback.
