---
sidebar_position: 1
---

# Nvidia GPU Acceleration

:::danger
There is a known issue with EGL and Chromium-based browsers, see [WebGL not working for Nvidia Google Chrome 112.x](https://github.com/m1k1o/neko/issues/279).
That means currently only Firefox is supported for Nvidia GPU acceleration.
:::

You need to have [NVIDIA Container Toolkit](https://github.com/NVIDIA/nvidia-container-toolkit) installed, start the container with `--gpus all` flag and use images built for nvidia (see [Docker Images](/docs/docker-images)).

```bash
docker run -d --gpus all \
  -p 8080:8080 \
  -p 56000-56100:56000-56100/udp \
  -e NEKO_SCREEN=1920x1080@30 \
  -e NEKO_PASSWORD=neko \
  -e NEKO_PASSWORD_ADMIN=admin \
  -e NEKO_EPR=56000-56100 \
  -e NEKO_NAT1TO1=192.168.1.10 \
  -e NEKO_ICELITE=1 \
  -e NEKO_VIDEO_CODEC=h264 \
  -e NEKO_HWENC=nvenc \
  --shm-size=2gb \
  --cap-add=SYS_ADMIN \
  --name neko \
  ghcr.io/m1k1o/neko/nvidia-google-chrome:latest
```

If you want to use docker-compose, you can use this example:

```yaml
services:
  neko:
    image: "ghcr.io/m1k1o/neko/nvidia-google-chrome:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
    - "8080:8080"
    - "56000-56100:56000-56100/udp"
    cap_add:
    - SYS_ADMIN
    environment:
      NEKO_SCREEN: '1920x1080@30'
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 56000-56100
      NEKO_NAT1TO1: 192.168.1.10
      NEKO_VIDEO_CODEC: h264
      NEKO_HWENC: nvenc
    deploy:
      resources:
        reservations:
          devices:
          - driver: nvidia
            count: 1
            capabilities: [gpu]
```

- You can verify that GPU is available inside the container by running `docker exec -it neko nvidia-smi` command.
- You can verify that GPU is used for encoding by searching for `nvh264enc` in `docker logs neko` output.
- If you don'ŧ specify `NEKO_HWENC: nvenc` environment variable, CPU encoding will be used but GPU will still be available for browser rendering.

Broadcast pipeline is not hardware accelerated by default. You can use this pipeline created by [@evilalmus](https://github.com/m1k1o/neko/issues/276#issuecomment-1498362533).

```yaml
NEKO_BROADCAST_PIPELINE: |
  flvmux name=mux 
    ! rtmpsink location={url} pulsesrc device={device} 
    ! audio/x-raw,channels=2 
    ! audioconvert 
    ! voaacenc 
    ! mux.
  ximagesrc display-name={display} show-pointer=false use-damage=false 
    ! video/x-raw,framerate=30/1 
    ! videoconvert 
    ! queue 
    ! video/x-raw,format=NV12 
    ! nvh264enc name=encoder preset=low-latency-hq gop-size=25 spatial-aq=true temporal-aq=true bitrate=2800 vbv-buffer-size=2800 rc-mode=6 
    ! h264parse config-interval=-1 
    ! video/x-h264,stream-format=byte-stream,profile=high 
    ! h264parse 
    ! mux.
```
