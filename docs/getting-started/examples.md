# Examples

## Firefox

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    environment:
      NEKO_SCREEN: '1920x1080@30'
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 52000-52100
      NEKO_NAT1TO1: <your-IP>
```

## Chromium

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:chromium"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    cap_add:
      - SYS_ADMIN
    environment:
      NEKO_SCREEN: '1920x1080@30'
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 52000-52100
      NEKO_NAT1TO1: <your-IP>
```

## VLC

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:vlc"
    restart: "unless-stopped"
    shm_size: "2gb"
    volumes:
      - "<your-video-folder>:/video" 
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    cap_add:
      - SYS_ADMIN
    environment:
      NEKO_SCREEN: '1920x1080@30'
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_EPR: 52000-52100
      NEKO_NAT1TO1: <your-IP>
```

## Raspberry Pi

Note! Since HW accelerated pipeline is using H264, you are only able to connect from browsers supporting H264 for WebRTC. At the time of implementing, [Firefox does not support this](https://developer.mozilla.org/en-US/docs/Web/Media/Formats/WebRTC_codecs#supported-foot-1). When omitting `NEKO_VIDEO` and `NEKO_H264` parameters, you get default CPU encoding with VP8.

```yaml
version: "3.4"
services:
  neko:
    image: "m1k1o/neko:arm-chromium"
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
          ! v4l2h264enc extra-controls="controls,h264_profile=0,video_bitrate=1250000;"
          ! h264parse config-interval=3
          ! video/x-h264,profile=baseline,stream-format=byte-stream
      NEKO_H264: 1
```

## Not using docker?

You can execute `neko --help` to see available arguments. In [Dockerfile](https://github.com/m1k1o/neko/blob/master/.docker/base/Dockerfile) you can find required dependencies and install them manually.
