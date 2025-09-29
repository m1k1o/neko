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

## Raspberry Pi GPU Acceleration {#raspberry-pi}

```yaml title="docker-compose.yaml"
services:
  neko:
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
    environment:
      NEKO_CAPTURE_VIDEO_PIPELINE: |
        ximagesrc display-name={display} show-pointer=true use-damage=false
          ! video/x-raw,framerate=25/1
          ! videoconvert ! queue
          ! video/x-raw,format=NV12
          ! v4l2h264enc
            name=encoder
            extra-controls="controls,h264_profile=1,video_bitrate=1250000;"
          ! h264parse config-interval=-1
          ! video/x-h264,stream-format=byte-stream
          ! appsink name=appsink
      NEKO_CAPTURE_VIDEO_CODEC: "h264"
      NEKO_DESKTOP_SCREEN: '1280x720@30'
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
```

## Nvidia GPU Acceleration {#nvidia}

Neko supports hardware acceleration using Nvidia GPUs. To use this feature, you need to have the Nvidia Container Toolkit installed on your system. You can find the installation instructions [here](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html). Check if your GPU supports hardware encoding with [this list](https://developer.nvidia.com/video-encode-decode-gpu-support-matrix).

This example shows how to accelerate video encoding and as well the browser rendering using the GPU. You can test if the GPU is used by running `nvtop` or `nvidia-smi`, which should show the GPU usage of both the browser and neko. In the browser, you can run the [WebGL Aquarium Demo](https://webglsamples.org/aquarium/aquarium.html) to test the GPU usage.

```yaml title="docker-compose.yaml"
services:
  neko:
    image: "ghcr.io/m1k1o/neko/nvidia-firefox:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    environment:
      NEKO_CAPTURE_VIDEO_PIPELINE: |
        ximagesrc display-name={display} show-pointer=true use-damage=false
          ! video/x-raw,framerate=25/1
          ! cudaupload ! cudaconvert ! queue
          ! video/x-raw(memory:CUDAMemory),format=NV12
          ! nvh264enc
            name=encoder
            preset=2
            gop-size=25
            spatial-aq=true
            temporal-aq=true
            bitrate=4096
            vbv-buffer-size=4096
            rc-mode=6
          ! h264parse config-interval=-1
          ! video/x-h264,stream-format=byte-stream
          ! appsink name=appsink
      NEKO_CAPTURE_VIDEO_CODEC: "h264"
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
    deploy:
      resources:
        reservations:
          devices:
          - driver: nvidia
            count: 1
            capabilities: [gpu]
```

See available [Nvidia Docker Images](/docs/v3/installation/docker-images#nvidia).

:::note
If your Nvidia GPU does not support CUDA, you can use the pipeline below without `cudaupload` and `cudaconvert`. This should work with older GPUs, but the performance might be lower.
:::

If you only want to accelerate the encoding, **not the browser rendering**, and you do not need [Cuda library](https://gstreamer.freedesktop.org/documentation/cuda/index.html?gi-language=c), you can use the default image with additional environment variables:

```yaml title="docker-compose.yaml"
services:
  neko:
    # highlight-next-line
    image: "ghcr.io/m1k1o/neko/firefox:latest"
    restart: "unless-stopped"
    shm_size: "2gb"
    ports:
      - "8080:8080"
      - "52000-52100:52000-52100/udp"
    environment:
      # highlight-start
      NVIDIA_VISIBLE_DEVICES: all
      NVIDIA_DRIVER_CAPABILITIES: all
      # highlight-end
      NEKO_CAPTURE_VIDEO_PIPELINE: |
        ximagesrc display-name={display} show-pointer=true use-damage=false
          ! video/x-raw,framerate=25/1
      # highlight-start
          ! videoconvert ! queue
          ! video/x-raw,format=NV12
      # highlight-end
          ! nvh264enc
            name=encoder
            preset=2
            gop-size=25
            spatial-aq=true
            temporal-aq=true
            bitrate=4096
            vbv-buffer-size=4096
            rc-mode=6
          ! h264parse config-interval=-1
          ! video/x-h264,stream-format=byte-stream
          ! appsink name=appsink
      NEKO_CAPTURE_VIDEO_CODEC: "h264"
      NEKO_DESKTOP_SCREEN: 1920x1080@30
      NEKO_MEMBER_MULTIUSER_USER_PASSWORD: neko
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: admin
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: 1
    deploy:
      resources:
        reservations:
          devices:
          - driver: nvidia
            count: 1
            capabilities: [gpu]
```
