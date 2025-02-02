---
sidebar_position: 2
---

# Raspberry Pi


To set up GPU acceleration on a Raspberry Pi for the Neko application, you can use the following Docker Compose configuration.

It is required to run in privileged mode to access the GPU. The `shm_size` should be increased on Raspberry Pi's with more than 1GB of RAM.

```yaml
services:
  neko:
    # see docs for more variants
    image: "ghcr.io/m1k1o/neko/arm-chromium:latest"
    restart: "unless-stopped"
    # increase on rpi's with more then 1gb ram.
    shm_size: "520mb"
    ports:
      - "8088:8080"
      - "52000-52100:52000-52100/udp"
    # note: this is important since we need a GPU for hardware acceleration alternatively
    #       mount the devices into the docker.
    privileged: true
    environment:
      NEKO_SCREEN: '1280x720@30'
      NEKO_PASSWORD: 'neko'
      NEKO_PASSWORD_ADMIN: 'admin'
      NEKO_EPR: 52000-52100
      # note: when setting NEKO_VIDEO, then variables NEKO_MAX_FPS and NEKO_VIDEO_BITRATE
      #       are not being used, you can adjust them in this variable.
      NEKO_VIDEO: |
        ximagesrc display-name=%s use-damage=0 show-pointer=true use-damage=false
          ! video/x-raw,framerate=30/1
          ! videoconvert
          ! queue
          ! video/x-raw,framerate=30/1,format=NV12
          ! v4l2h264enc extra-controls="controls,h264_profile=1,video_bitrate=1250000;"
          ! h264parse config-interval=3
          ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline
      NEKO_VIDEO_CODEC: h264
```
