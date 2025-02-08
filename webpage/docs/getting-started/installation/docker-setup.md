---
sidebar_position: 1
---

# Docker Setup

<div align="center">
  <img src="/img/icons/firefox.svg" title="m1k1o/neko:firefox" width="60" height="auto"/>
  <img src="/img/icons/google-chrome.svg" title="m1k1o/neko:google-chrome" width="60" height="auto"/>
  <img src="/img/icons/chromium.svg" title="m1k1o/neko:chromium" width="60" height="auto"/>
  <img src="/img/icons/microsoft-edge.svg" title="m1k1o/neko:microsoft-edge" width="60" height="auto"/>
  <img src="/img/icons/brave.svg" title="m1k1o/neko:brave" width="60" height="auto"/>
  <img src="/img/icons/vivaldi.svg" title="m1k1o/neko:vivaldi" width="60" height="auto"/>
  <img src="/img/icons/opera.svg" title="m1k1o/neko:opera" width="60" height="auto"/>
  <img src="/img/icons/tor-browser.svg" title="m1k1o/neko:tor-browser" width="60" height="auto"/>
  <img src="/img/icons/remmina.png" title="m1k1o/neko:remmina" width="60" height="auto"/>
  <img src="/img/icons/vlc.svg" title="m1k1o/neko:vlc" width="60" height="auto"/>
  <img src="/img/icons/xfce.svg" title="m1k1o/neko:xfce" width="60" height="auto"/>
  <img src="/img/icons/kde.svg" title="m1k1o/neko:kde" width="60" height="auto"/>
</div>

Use the following docker images from [Docker Hub](https://hub.docker.com/r/m1k1o/neko) for x86_64:
- `m1k1o/neko:latest` or `m1k1o/neko:firefox` - for Firefox.
- `m1k1o/neko:chromium` - for Chromium (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:google-chrome` - for Google Chrome (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:ungoogled-chromium` - for [Ungoogled Chromium](https://github.com/Eloston/ungoogled-chromium) (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)) (by @whalehub).
- `m1k1o/neko:microsoft-edge` - for Microsoft Edge (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:brave` - for [Brave Browser](https://brave.com) (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)).
- `m1k1o/neko:vivaldi` - for [Vivaldi Browser](https://vivaldi.com) (needs `--cap-add=SYS_ADMIN`, see the [security implications](https://www.redhat.com/en/blog/container-tidbits-adding-capabilities-container)) (by @Xeddius).
- `m1k1o/neko:opera` for [Opera Browser](https://opera.com) (requires extra steps to enable DRM, see instructions [here](https://www.reddit.com/r/operabrowser/wiki/opera/linux_widevine_config/). libffmpeg is already configured.) (by @prophetofxenu)
- `m1k1o/neko:tor-browser` - for Tor Browser.
- `m1k1o/neko:remmina` - for remote desktop connection (by @lowne).
  - Pass env var `REMMINA_URL=<proto>://[<username>[:<password>]@]server[:port]` (proto being `vnc`, `rdp` or `spice`).
  - Or create your custom configuration with remmina locally (it's saved in `~/.local/share/remmina/path_to_profile.remmina`) and bind-mount it, then pass env var `REMMINA_PROFILE=<path_to_profile.remmina>`.
- `m1k1o/neko:vlc` - for VLC Video player (needs volume mounted to `/media` with local video files, or setting `VLC_MEDIA=/media` path).
- `m1k1o/neko:xfce` or `m1k1o/neko:kde` - for a shared desktop / installing shared software.
- `m1k1o/neko:base` - for custom base.

Dockerhub images are built using GitHub actions on every push and on weekly basis to keep all browsers up-to-date.

All images are also available on [GitHub Container Registry](https://github.com/m1k1o?tab=packages&repo_name=neko) for faster pulls:

- `ghcr.io/m1k1o/neko/firefox:latest`
- `ghcr.io/m1k1o/neko/chromium:latest`
- `ghcr.io/m1k1o/neko/google-chrome:latest`
- `ghcr.io/m1k1o/neko/ungoogled-chromium:latest`
- `ghcr.io/m1k1o/neko/microsoft-edge:latest`
- `ghcr.io/m1k1o/neko/brave:latest`
- `ghcr.io/m1k1o/neko/vivaldi:latest`
- `ghcr.io/m1k1o/neko/opera:latest`
- `ghcr.io/m1k1o/neko/tor-browser:latest`
- `ghcr.io/m1k1o/neko/remmina:latest`
- `ghcr.io/m1k1o/neko/vlc:latest`
- `ghcr.io/m1k1o/neko/xfce:latest`
- `ghcr.io/m1k1o/neko/kde:latest`

For ARM-based images (like Raspberry Pi - with GPU hardware acceleration, Oracle Cloud ARM tier). Currently, not all images are available for ARM, because not all applications are available for ARM. Please note, that `m1k1o/neko:arm-*` images from dockerhub are currently not maintained and they can contain outdated software. Please use images below:

- `ghcr.io/m1k1o/neko/arm-firefox:latest`
- `ghcr.io/m1k1o/neko/arm-chromium:latest`
- `ghcr.io/m1k1o/neko/arm-ungoogled-chromium:latest`
- `ghcr.io/m1k1o/neko/arm-vlc:latest`
- `ghcr.io/m1k1o/neko/arm-xfce:latest`

For images with VAAPI GPU hardware acceleration using intel drivers use:

- `ghcr.io/m1k1o/neko/intel-firefox:latest`
- `ghcr.io/m1k1o/neko/intel-chromium:latest`
- `ghcr.io/m1k1o/neko/intel-google-chrome:latest`
- `ghcr.io/m1k1o/neko/intel-ungoogled-chromium:latest`
- `ghcr.io/m1k1o/neko/intel-microsoft-edge:latest`
- `ghcr.io/m1k1o/neko/intel-brave:latest`
- `ghcr.io/m1k1o/neko/intel-vivaldi:latest`
- `ghcr.io/m1k1o/neko/intel-opera:latest`
- `ghcr.io/m1k1o/neko/intel-tor-browser:latest`
- `ghcr.io/m1k1o/neko/intel-remmina:latest`
- `ghcr.io/m1k1o/neko/intel-vlc:latest`
- `ghcr.io/m1k1o/neko/intel-xfce:latest`
- `ghcr.io/m1k1o/neko/intel-kde:latest`

For images with Nvidia GPU hardware acceleration using EGL (see example below) use (please note, there is a known issue with EGL and Chromium-based browsers, see [here](https://github.com/m1k1o/neko/issues/279)):

- `ghcr.io/m1k1o/neko/nvidia-firefox:latest`
- `ghcr.io/m1k1o/neko/nvidia-chromium:latest`
- `ghcr.io/m1k1o/neko/nvidia-google-chrome:latest`
- `ghcr.io/m1k1o/neko/nvidia-microsoft-edge:latest`
- `ghcr.io/m1k1o/neko/nvidia-brave:latest`

GHCR images are built using GitHub actions for every tag.

## Running Neko with Docker

To start a basic Neko container, use the following command:

```sh
docker run -d --rm \
  -p 8080:8080 \
  -p 56000-56100:56000-56100/udp \
  -e NEKO_EPR=56000-56100 \
  -e NEKO_PASSWORD=neko \
  -e NEKO_PASSWORD_ADMIN=admin \
  -e NEKO_NAT1TO1=<your-ip> \
  --shm-size=2g \
  m1k1o/neko:latest
```

### Explanation

- `-d` - Run the container in the background.
- `--rm` - Automatically remove the container when it exits.
- `-p 8080:8080` - Map the host's port `8080` to the container's port `8080`.
- `-p 56000-56100:56000-56100/udp` - Map the host's ports `56000-56100` to the container's ports `56000-56100` using UDP.
- `-e NEKO_EPR=56000-56100` - Set the range of ports for the WebRTC connection, it must match the port range mapped above.
- `-e NEKO_PASSWORD=neko` and `-e NEKO_PASSWORD_ADMIN=admin` - Set passwords for the user and admin user.
- `-e NEKO_NAT1TO1=<your-ip>` - Set the public or local IP address for the NAT1:1 connection.
- `--shm-size=2g` - Set the shared memory size to 2GB, otherwise, the browser may crash.
- `m1k1o/neko:latest` - The name of the image to run, change it to the desired image.

Now, open your browser and go to: `http://localhost:8080`. You should see the Neko interface.

## Using Docker Compose

You can also use Docker Compose to run Neko. Create a `docker-compose.yml` file with the following content:

```yaml title="docker-compose.yml"
services:
  neko:
    image: m1k1o/neko:latest
    shm_size: 2g
    ports:
      - "8080:8080"
      - "56000-56100:56000-56100/udp"
    environment:
      NEKO_EPR: 56000-56100
      NEKO_PASSWORD: neko
      NEKO_PASSWORD_ADMIN: admin
      NEKO_NAT1TO1: <your-ip>
```

Then, run the following command:

```sh
docker compose up -d
```
