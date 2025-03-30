---
description: Example Docker Compose configurations for Neko.
---

# Examples

Here are some examples to get you started with Neko. You can use these examples as a reference to create your own configurations.

## Firefox {#firefox}

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/firefox:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    volumes:
      - <your-host-path>:/home/neko/.mozilla/firefox # persist firexfox settings
    environment:
      NEKO_DESKTOP_SCREEN: '1920x1080@30'
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
      NEKO_WEBRTC_NAT1TO1: <your-IP>
```

## Chromium {#chromium}

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/chromium:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    cap_add:
      - SYS_ADMIN
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    volumes:
      - <your-host-path>:/home/neko/.config/chromium # persist chromium settings
    environment:
      NEKO_DESKTOP_SCREEN: '1920x1080@30'
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
      NEKO_WEBRTC_NAT1TO1: <your-IP>
```

## VLC {#vlc}

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/vlc:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    cap_add:
      - SYS_ADMIN
    volumes:
      - "<your-video-folder>:/video" # mount your video folder
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    environment:
      NEKO_DESKTOP_SCREEN: '1920x1080@30'
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
      NEKO_WEBRTC_NAT1TO1: <your-IP>
```

## Raspberry Pi {#raspberry-pi}

```yaml title="config.yaml"
capture:
  video:
    codec: h264
    ids: [ main ]
    pipelines:
      main:
        gst_pipeline: |
          ximagesrc display-name=%s use-damage=0 show-pointer=true use-damage=false
            ! video/x-raw,framerate=30/1
            ! videoconvert
            ! queue
            ! video/x-raw,framerate=30/1,format=NV12
            ! v4l2h264enc extra-controls="controls,h264_profile=1,video_bitrate=1250000;"
            ! h264parse config-interval=3
            ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline
```

```yaml title="docker-compose.yaml"
services:
  neko:
    # see docs for more variants
    image: "ghcr.io/m1k1o/neko/chromium:latest"
    restart: "unless-stopped"
    # increase on rpi's with more then 1gb ram.
    shm_size: "520mb"
    ports:
      - "8088:8080"
      - "52000-52100:52000-52100/udp"
    # note: this is important since we need a GPU for hardware acceleration alternatively
    #       mount the devices into the docker.
    privileged: true
    volumes:
      - "./config.yaml:/etc/neko/neko.yaml"
    environment:
      NEKO_DESKTOP_SCREEN: '1280x720@30'
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: 'neko'
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: 'admin'
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
```
